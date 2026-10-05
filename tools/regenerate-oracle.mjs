// SPDX-License-Identifier: AGPL-3.0-only
// Scheduled/manual validation only. Never runs as part of the engine.
import {readFileSync, existsSync, createWriteStream, createReadStream, mkdirSync, writeFileSync} from 'node:fs';
import {createHash} from 'node:crypto';
import {Readable} from 'node:stream';
import {pipeline} from 'node:stream/promises';
import {spawnSync} from 'node:child_process';
import {resolve} from 'node:path';
import assert from 'node:assert/strict';
const lock=JSON.parse(readFileSync('tools/oracle-lock.json'));
const cache=resolve('.tools/oracle');
mkdirSync(cache,{recursive:true}); mkdirSync('dist/oracle',{recursive:true});
async function hash(path) {
  const h=createHash('sha256'); for await (const chunk of createReadStream(path)) h.update(chunk); return h.digest('hex');
}
function run(command,args) {
  const r=spawnSync(command,args,{encoding:'utf8',timeout:180000});
  if(r.status!==0)throw new Error(`${command}: ${r.error?.message||r.stderr||r.stdout}`);
  return r.stdout;
}
// Independent downloads are concurrent; extraction and oracle runs are sequential.
await Promise.all([['hpxml.zip',lock.hpxml],['openstudio.tar.gz',lock.openstudio]].map(async([name,asset])=>{
  const destination=`${cache}/${name}`;
  if(!existsSync(destination)) {
    const r=await fetch(asset.url); if(!r.ok)throw new Error(`download ${r.status}`);
    await pipeline(Readable.fromWeb(r.body),createWriteStream(destination));
  }
  assert.equal(await hash(destination),asset.sha256,`untrusted ${name}`);
}));
run('python3',['-c',`import zipfile; zipfile.ZipFile(${JSON.stringify(cache+'/hpxml.zip')}).extractall(${JSON.stringify(cache)})`]);
run('tar',['--warning=no-unknown-keyword','-xzf',cache+'/openstudio.tar.gz','-C',cache]);
const binary=`${cache}/${lock.openstudio.binary}`;
assert.equal(run(binary,['--version']).trim(),lock.openstudio.version);
for(const name of lock.fixtures) {
  const source=`testdata/openstudio/${name}`;
  const destination=resolve(`dist/oracle/${name}`);
  const manifest=JSON.parse(readFileSync(`${source}/generation.json`));
  assert.equal(await hash(`${source}/input.xml`),manifest.input_sha256);
  assert.equal(await hash(`${cache}/OpenStudio-HPXML/weather/${manifest.weather_file}`),manifest.weather_sha256);
  const args=[`${cache}/OpenStudio-HPXML/workflow/run_simulation.rb`,'-x',resolve(`${source}/input.xml`),'-o',destination,'--skip-simulation','--output-format','json'];
  const log=run(binary,args); writeFileSync(`${destination}/generation.log`,log);
  // Default workflow validates both schema and schematron. No skip-validation flag.
  const actual=JSON.parse(readFileSync(`${destination}/run/results_design_load_details.json`));
  assert.deepEqual(actual,JSON.parse(readFileSync(`${source}/oracle.json`)),`${name} oracle changed`);
  console.log(`${name}: schema-validated pinned oracle reproduced exactly`);
}
