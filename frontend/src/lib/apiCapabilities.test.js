import test from 'node:test';
import assert from 'node:assert/strict';
import {
  PROJECT_PERMISSIONS,
  VERSIONED_DEPENDENCIES,
  currentAPIInfo,
  negotiateStudioAPI,
} from './apiCapabilities.js';

test('missing discovery data falls back to API 1', () => {
  const negotiated = negotiateStudioAPI(undefined);
  assert.equal(negotiated.version, '1');
  assert.deepEqual(negotiated.capabilities, []);
});

test('client selects the highest mutually supported API', () => {
  const negotiated = negotiateStudioAPI(currentAPIInfo());
  assert.equal(negotiated.version, '2');
  assert.deepEqual(
    negotiated.capabilities,
    [VERSIONED_DEPENDENCIES, PROJECT_PERMISSIONS],
  );
});

test('unknown capabilities remain additive', () => {
  const negotiated = negotiateStudioAPI({
    default_version: '1',
    supported_versions: ['1', '2'],
    capabilities_by_version: {
      1: [],
      2: ['audit_logs'],
    },
  });
  assert.equal(negotiated.version, '2');
  assert.deepEqual(negotiated.capabilities, ['audit_logs']);
});

