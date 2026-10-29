import test from 'node:test';
import assert from 'node:assert/strict';
import { isActiveProjectChange } from './projectDataChanges.js';

const activeProject = {
  id: 'project-id',
  uri: 'C:/Projects/example.clst',
};

test('accepts changes for the active project', () => {
  assert.equal(isActiveProjectChange({
    project_id: activeProject.id,
    project_uri: activeProject.uri,
  }, activeProject), true);
});

test('rejects changes for another project', () => {
  assert.equal(isActiveProjectChange({
    project_id: 'other-project',
    project_uri: activeProject.uri,
  }, activeProject), false);
  assert.equal(isActiveProjectChange({
    project_id: activeProject.id,
    project_uri: 'C:/Projects/other.clst',
  }, activeProject), false);
});

test('rejects incomplete changes', () => {
  assert.equal(isActiveProjectChange(null, activeProject), false);
  assert.equal(isActiveProjectChange({}, activeProject), false);
  assert.equal(isActiveProjectChange({
    project_id: activeProject.id,
    project_uri: activeProject.uri,
  }, null), false);
});
