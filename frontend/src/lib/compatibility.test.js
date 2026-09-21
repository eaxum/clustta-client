import { test } from 'node:test';
import assert from 'node:assert/strict';
import { compatibilityProblem, compatibilityCacheKey, parseCompatibilityError, projectAccessProblem, projectCompatibilityProblem } from './compatibility.js';

test('only the complete current contract is compatible', () => {
  assert.equal(compatibilityProblem({ protocol: '1', schema: '2.2', project_schema: '2.2' }), null);
  assert.equal(compatibilityProblem(null).required_update, 'server');
  assert.equal(compatibilityProblem({ protocol: '1', schema: '2.1', project_schema: '2.1' }).required_update, 'server');
  assert.equal(compatibilityProblem({ protocol: '1', schema: '2.3', project_schema: '2.3' }).required_update, 'client');
});

test('cache separates accounts, Studios, hosts, and projects', () => {
  const project = { id: 'project', remote: 'https://studio.test/project' };
  const key = compatibilityCacheKey('alice', 'studio', project);
  assert.notEqual(key, compatibilityCacheKey('bob', 'studio', project));
  assert.notEqual(key, compatibilityCacheKey('alice', 'other', project));
  assert.notEqual(key, compatibilityCacheKey('alice', 'studio', { ...project, remote: 'https://other.test/project' }));
  assert.notEqual(key, compatibilityCacheKey('alice', 'studio', { ...project, id: 'other' }));
});

test('compatibility errors remain distinct from connectivity errors', () => {
  const problem = compatibilityProblem(null);
  assert.deepEqual(parseCompatibilityError(new Error(JSON.stringify(problem))), problem);
  assert.equal(parseCompatibilityError(new Error('connection refused')), null);
});

test('offline compatibility status requires a previously verified contract', () => {
  const contract = { protocol: '1', schema: '2.2', project_schema: '2.2' };
  const project = { is_offline: true };
  assert.equal(projectCompatibilityProblem(project, { verified: true, contract }), null);
  assert.equal(projectCompatibilityProblem(project, undefined).required_update, 'server');
  const problem = compatibilityProblem({ protocol: '1', schema: '2.3', project_schema: '2.3' });
  assert.equal(projectCompatibilityProblem(project, { verified: false, problem }), problem);
});

test('offline access rejects a replica that does not match the verified host schema', () => {
  const project = { is_offline: true, is_downloaded: true, local_schema: '2.1' };
  const cached = { verified: true, contract: { protocol: '1', schema: '2.2', project_schema: '2.2' } };
  assert.equal(projectCompatibilityProblem(project, cached).required_update, 'replica');
});

test('host project drift requires host maintenance', () => {
  assert.equal(compatibilityProblem({ protocol: '1', schema: '2.2', project_schema: '2.10' }).required_update, 'server');
});

test('a mismatched replica retains a sync blocker even when the client supports the host', () => {
  const compatibility = { protocol: '1', schema: '2.2', project_schema: '2.2' };
  assert.equal(projectCompatibilityProblem({ compatibility, is_downloaded: true, local_schema: '2.1' }).required_update, 'replica');
  assert.equal(projectCompatibilityProblem({ compatibility, is_downloaded: true, local_schema: '2.2' }), null);
});

test('a readable local replica can open when its host is outdated', () => {
  const compatibility = { protocol: '1', schema: '2.1', project_schema: '2.1' };
  const project = { compatibility, is_downloaded: true, local_schema: '2.2' };
  assert.equal(projectCompatibilityProblem(project).required_update, 'server');
  assert.equal(projectAccessProblem(project), null);
});

test('local access still blocks unreadable and unverified offline replicas', () => {
  const compatibility = { protocol: '1', schema: '2.1', project_schema: '2.1' };
  assert.equal(projectAccessProblem({ compatibility, is_downloaded: true, local_schema: '2.1' }).required_update, 'replica');
  assert.equal(projectAccessProblem({ is_offline: true, is_downloaded: true }, undefined).required_update, 'server');
});
