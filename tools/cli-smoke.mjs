// SPDX-License-Identifier: AGPL-3.0-only
// Exercise installed CLI boundaries, not internal calculation helpers.
import {spawnSync} from 'node:child_process';
import {mkdtempSync, writeFileSync, readFileSync, rmSync} from 'node:fs';
import {tmpdir} from 'node:os';
import {join, resolve} from 'node:path';
import assert from 'node:assert/strict';
const binary = resolve(process.argv[2] || 'bin/hvac');
const fixture = 'testdata/designload/buildings/ranch.yaml';
const temp = mkdtempSync(join(tmpdir(), 'hvac-cli-'));
function run(args, status=0) {
  const r = spawnSync(binary, args, {encoding: 'utf8', timeout: 15000});
  assert.equal(r.status, status, r.error?.message || r.stderr || r.stdout);
  return status === 0 ? r.stdout : r.stderr;
}
try {
  assert.match(run(['validate', fixture]), /Valid project/);
  const args = ['load', fixture, '--model', 'designload', '--format', 'json'];
  const result = JSON.parse(run(args));
  assert.equal(result.methodology.id, 'designload');
  assert.equal(run(args), run(args), 'serialized results must be deterministic');
  assert.ok(result.rooms.length > 1);
  const room = JSON.parse(run(['load', fixture, '--room', result.rooms[0].name, '--format', 'json']));
  assert.equal(room.rooms.length, 1);
  assert.equal(room.rooms[0].heating_load_w, result.rooms[0].heating_load_w);
  const node = JSON.parse(run(['explain', fixture, '--node', result.heating.id, '--format', 'json']));
  assert.equal(node.node.value_w, result.heating_load_w);
  assert.equal(node.methodology.id, 'designload');
  assert.match(run(['load', fixture, '--units', 'ip']), /Btu\/h/);
  const invalid = join(temp, 'invalid.json');
  writeFileSync(invalid, '{"version":999,"building":{}}');
  const failure = JSON.parse(run(['validate', invalid, '--format', 'json'], 1));
  assert.ok(failure.error || failure.errors?.length);
  writeFileSync(join(temp, 'result.json'), run(args));
  assert.deepEqual(JSON.parse(readFileSync(join(temp, 'result.json'))), result);
  console.log('CLI acceptance: validate, calculate, room selection, explanation, units, errors and offline deterministic output PASS');
} finally { rmSync(temp, {recursive: true, force: true}); }
