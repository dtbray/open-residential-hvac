// SPDX-License-Identifier: AGPL-3.0-only
import {spawnSync} from 'node:child_process';
import {mkdirSync, readdirSync, readFileSync} from 'node:fs';
import {resolve} from 'node:path';
import assert from 'node:assert/strict';
import {createHash} from 'node:crypto';
const binary = resolve(process.argv[2] || 'bin/hvac-compare');
mkdirSync('dist/validation', {recursive: true});
for (const [fixture, reference] of [
  ['wall-conduction-only', 'wall-conduction-hand'],
  ['leaky-house', 'leaky-house-psychrolib'],
]) {
  const r = spawnSync(binary, ['--report', `dist/validation/${reference}.json`,
    `testdata/designload/buildings/${fixture}.json`, `testdata/designload/reference/${reference}.json`], {encoding:'utf8'});
  assert.equal(r.status, 0, r.stdout + r.stderr);
}
// Real oracle comparisons are diagnostic until reviewed engineering tolerances exist.
// Assert their documented differences, input hashes and mapping, not false agreement.
for (const name of readdirSync('testdata/openstudio').filter(n => !n.includes('.'))) {
  const base = `testdata/openstudio/${name}`;
  const policy = JSON.parse(readFileSync(`${base}/comparison-policy.json`));
  const generation = JSON.parse(readFileSync(`${base}/generation.json`));
  for (const [file, key] of [['input.xml','input_sha256'],['resolved.xml','resolved_sha256'],['oracle.json','oracle_sha256']]) {
    assert.equal(createHash('sha256').update(readFileSync(`${base}/${file}`)).digest('hex'),generation[key], `${name} ${file} evidence changed`);
  }
  const raw = JSON.parse(readFileSync(`${base}/oracle.json`));
  const reference = JSON.parse(readFileSync(`${base}/reference.json`));
  const mapping = JSON.parse(readFileSync(`${base}/metric-mapping.json`));
  assert.deepEqual(Object.keys(mapping).sort(),Object.keys(reference.metrics_w).sort());
  for (const [metric, source] of Object.entries(mapping)) {
    const value = source.sum ? source.sum.reduce((sum,key)=>sum+reference.metrics_w[key],0)
      : raw[source.report][source.entry][source.field]*0.2930710701722222;
    assert.equal(reference.metrics_w[metric],value,`${name} ${metric} differs from raw oracle`);
  }
  const report = `dist/validation/${name}.json`;
  const r = spawnSync(binary, ['--relative', String(policy.relative_tolerance), '--absolute', String(policy.absolute_tolerance_w),
    '--report', report, `${base}/project.json`, `${base}/reference.json`], {encoding:'utf8'});
  assert.ok(r.status === 0 || r.status === 1, r.stderr);
  const actual = JSON.parse(readFileSync(report));
  const expectedReport = JSON.parse(readFileSync(`${base}/comparison.json`));
  assert.equal(actual.differences.length,expectedReport.differences.length);
  for (let index=0;index<actual.differences.length;index++) {
    const got=actual.differences[index], expected=expectedReport.differences[index];
    assert.equal(got.metric,expected.metric); assert.equal(got.pass,expected.pass);
    for(const key of ['reference_w','candidate_w','difference_w']) assert.ok(Math.abs(got[key]-expected[key])<=1e-8,`${name} ${got.metric} changed`);
  }
  assert.equal(actual.pass,expectedReport.pass);
  assert.deepEqual(actual.methodology,expectedReport.methodology);
  const candidate = spawnSync(resolve(process.argv[3] || 'bin/hvac'), ['load', `${base}/project.json`, '--format', 'json'], {encoding:'utf8'});
  assert.equal(candidate.status, 0, candidate.stderr);
  const result = JSON.parse(candidate.stdout);
  for (const [key, expected] of Object.entries(policy.candidate_totals_w)) {
    assert.ok(Math.abs(result[key]-expected) <= 1e-8, `${name} ${key} regression`);
  }
  console.log(`${name}: ${actual.differences.filter(d=>!d.pass).length} documented diagnostic differences`);
}
console.log('Stored reference comparisons PASS; reports in dist/validation');
