// SPDX-License-Identifier: AGPL-3.0-only
// Native packaging: never label a cross-compiled desktop binary as native-tested.
import {spawnSync} from 'node:child_process';
import {mkdirSync, readFileSync, writeFileSync, copyFileSync, readdirSync, existsSync, rmSync} from 'node:fs';
import {join, resolve} from 'node:path';
import {createHash} from 'node:crypto';
import assert from 'node:assert/strict';
const go=process.env.HVAC_GO || 'go';
const platform=process.platform;
assert.ok(['linux','win32','darwin'].includes(platform));
const target=`${platform==='win32'?'windows':platform}-${process.arch}`;
const version=process.env.HVAC_VERSION || 'development';
assert.match(version,/^[a-zA-Z0-9._-]+$/);
const name=`open-residential-hvac-${version}-${target}`;
const output=resolve('dist/releases');
const stage=resolve(`dist/package/${name}`);
rmSync(stage,{recursive:true,force:true}); mkdirSync(stage,{recursive:true}); mkdirSync(output,{recursive:true});
function run(command,args,options={}) {
  const env=platform==='win32'?{...process.env,CGO_ENABLED:'0'}:process.env;
  const r=spawnSync(command,args,{encoding:'utf8',timeout:300000,env,...options});
  assert.equal(r.status,0,r.error?.message||r.stderr||r.stdout); return r.stdout;
}
const extension=platform==='win32'?'.exe':'';
for(const command of ['hvac','hvac-compare','hvac-import-hpxml']) {
  run(go,['build','-trimpath','-ldflags=-s -w','-o',join(stage,command+extension),`./cmd/${command}`]);
}
const tags=['desktop','production'];
if(platform==='linux')tags.push('webkit2_41');
let desktop=join(stage,'hvac-desktop'+extension);
if(platform==='darwin') {
  const app=join(stage,'Open Residential HVAC.app','Contents');
  mkdirSync(join(app,'MacOS'),{recursive:true});
  desktop=join(app,'MacOS','hvac-desktop');
  writeFileSync(join(app,'Info.plist'),`<?xml version="1.0" encoding="UTF-8"?><!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd"><plist version="1.0"><dict><key>CFBundleExecutable</key><string>hvac-desktop</string><key>CFBundleIdentifier</key><string>com.thomas-bray.open-residential-hvac</string><key>CFBundleName</key><string>Open Residential HVAC</string><key>CFBundlePackageType</key><string>APPL</string><key>CFBundleVersion</key><string>0.1.0</string><key>NSHighResolutionCapable</key><true/></dict></plist>`);
}
run(go,['build','-trimpath','-tags',tags.join(','),'-ldflags=-s -w','-o',desktop,'./cmd/desktop']);
for(const file of ['LICENSE','THIRD_PARTY_NOTICES.md','README.md'])copyFileSync(file,join(stage,file));
mkdirSync(join(stage,'docs')); copyFileSync('docs/methodology.md',join(stage,'docs/methodology.md'));
copyFileSync('docs/validation.md',join(stage,'docs/validation.md'));
copyFileSync('testdata/designload/buildings/ranch.yaml',join(stage,'example.yaml'));
// Include actual dependency license texts, not merely SPDX labels in a manifest.
const licenses=join(stage,'dependency-licenses'); mkdirSync(licenses);
run(go,['mod','download','all']);
const modules=run(go,['list','-m','-f','{{.Path}}\t{{.Version}}\t{{.Dir}}','all']).trim().split('\n').slice(1);
const inventory=[];
for(const line of modules) {
  const [module,version,directory]=line.split('\t');
  assert.ok(directory&&existsSync(directory),`missing module cache: ${module}`);
  const files=readdirSync(directory).filter(n=>/^(license|copying|notice)(\.|$)/i.test(n));
  assert.ok(files.length,`missing license for ${module}`);
  const dest=join(licenses,module.replaceAll('/','__'));mkdirSync(dest);
  for(const file of files)copyFileSync(join(directory,file),join(dest,file));
  inventory.push({module,version,license_files:files});
}
// The frontend's runtime Svelte license is distributed with its bundled code.
const svelte='app/frontend/node_modules/svelte';
const svelteLicense=readdirSync(svelte).find(n=>/^license/i.test(n));
assert.ok(svelteLicense,'Svelte license missing');
copyFileSync(join(svelte,svelteLicense),join(licenses,'Svelte-LICENSE'));
inventory.push({module:'npm:svelte',version:JSON.parse(readFileSync(join(svelte,'package.json'))).version});
writeFileSync(join(stage,'DEPENDENCIES.json'),JSON.stringify(inventory,null,2)+'\n');
const sha=run('git',['rev-parse','HEAD']).trim();
const dirty=run('git',['status','--porcelain']).trim()!=='';
assert.ok(!version.startsWith('v')||!dirty,'tagged release requires a clean source checkout');
writeFileSync(join(stage,'RELEASE.json'),JSON.stringify({version,target,source_commit:sha,source_dirty:dirty,go:run(go,['version']).trim(),methodology:'designload/0.1',source_url:`https://git.thomas-bray.com/thomas/open-residential-hvac/src/commit/${sha}`,native_build:true,signed:false},null,2)+'\n');
writeFileSync(join(stage,'INSTALL.txt'),'CLI: run hvac load example.yaml. Desktop: launch hvac-desktop (macOS: the .app bundle).\nLinux requires GTK3 and WebKit2GTK 4.1 runtime libraries; Windows requires WebView2.\nPackages are unsigned development builds. Consult README and validation status before engineering use.\nComplete corresponding source and build instructions are at the RELEASE.json source URL.\n');
run('node',['tools/cli-smoke.mjs',join(stage,'hvac'+extension)]);
const archive=join(output,name+'.tar.gz');
run('tar',['-czf',archive,'-C',resolve('dist/package'),name]);
const digest=createHash('sha256').update(readFileSync(archive)).digest('hex');
writeFileSync(archive+'.sha256',`${digest}  ${name}.tar.gz\n`);
console.log(`Packaged native ${target}: ${archive}`);
