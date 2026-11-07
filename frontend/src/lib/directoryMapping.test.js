import assert from 'node:assert/strict';
import test from 'node:test';

import {
  filterTaskTypesForTemplate,
  fillDefaultTaskOutputs,
  getDefaultTaskOutput,
  getDirectoryTemplatePreviewValues,
  inferDirectoryTemplateEntity,
} from './directoryMapping.js';

const assetTemplate = {
  id: 'asset',
  template: 'Assets/<CollectionType>/<Asset>/<OutputName><TemplateExtension>',
};
const shotTemplate = {
  id: 'shot',
  template: 'Episodes/<Episode>/<Sequence>/<Shot>/<OutputName><TemplateExtension>',
};
const editTemplate = {
  id: 'edit',
  template: 'Edits/<Edit>/<OutputName><TemplateExtension>',
};

test('directory mapping identifies asset, shot, and edit templates', () => {
  assert.equal(inferDirectoryTemplateEntity(assetTemplate.id, assetTemplate.template), 'asset');
  assert.equal(inferDirectoryTemplateEntity(shotTemplate.id, shotTemplate.template), 'shot');
  assert.equal(inferDirectoryTemplateEntity(editTemplate.id, editTemplate.template), 'edit');
});

test('generic task types are available to asset and shot templates', () => {
  const taskTypes = [
    { name: 'Modeling', for_entity: 'task' },
    { name: 'Texture', for_entity: 'task' },
  ];

  assert.deepEqual(filterTaskTypesForTemplate(assetTemplate, taskTypes), taskTypes);
  assert.deepEqual(filterTaskTypesForTemplate(shotTemplate, taskTypes), taskTypes);
});

test('entity-scoped task types only appear on matching templates', () => {
  const assetTask = { name: 'Modeling', for_entity: 'asset' };
  const shotTask = { name: 'Animation', for_entity: 'shot' };
  const genericTask = { name: 'Review', for_entity: 'task' };
  const taskTypes = [assetTask, shotTask, genericTask];

  assert.deepEqual(filterTaskTypesForTemplate(assetTemplate, taskTypes), [assetTask, genericTask]);
  assert.deepEqual(filterTaskTypesForTemplate(shotTemplate, taskTypes), [shotTask, genericTask]);
});

test('edit-scoped task types only appear on edit templates', () => {
  const editTask = { name: 'Edit', for_entity: 'edit' };
  const shotTask = { name: 'Animation', for_entity: 'shot' };
  const taskTypes = [editTask, shotTask];

  assert.deepEqual(filterTaskTypesForTemplate(editTemplate, taskTypes), [editTask]);
});

test('task output defaults use the matching entity placeholder', () => {
  assert.equal(getDefaultTaskOutput(assetTemplate, 'Modeling'), '<Asset>-modeling');
  assert.equal(getDefaultTaskOutput(shotTemplate, 'Animation'), '<Shot>-animation');
  assert.equal(getDefaultTaskOutput(editTemplate, 'Edit'), '<Edit>-edit');
});

test('edit directory preview uses the edit entity name', () => {
  const values = getDirectoryTemplatePreviewValues(editTemplate);

  assert.equal(values.Edit, 'director-cut');
  assert.equal(values.OutputName, 'director-cut-edit');
});

test('directory mapping previews use matching task output examples', () => {
  assert.equal(getDirectoryTemplatePreviewValues(assetTemplate).OutputName, 'hero-modeling');
  assert.equal(getDirectoryTemplatePreviewValues(shotTemplate).OutputName, 'shot010-animation');
});

test('unclassified templates retain the asset default', () => {
  const customTemplate = { id: 'custom', template: '<CollectionType>/<OutputName>' };

  assert.equal(getDefaultTaskOutput(customTemplate, 'Modeling'), '<Asset>-modeling');
  assert.equal(getDirectoryTemplatePreviewValues(customTemplate).OutputName, 'hero-modeling');
});

test('auto task outputs fill blank enabled fields for each entity', () => {
  const assetOutputs = fillDefaultTaskOutputs(assetTemplate, [{ name: 'Modeling' }]);
  const shotOutputs = fillDefaultTaskOutputs(shotTemplate, [{ name: 'Animation' }]);

  assert.deepEqual(assetOutputs, { Modeling: '<Asset>-modeling' });
  assert.deepEqual(shotOutputs, { Animation: '<Shot>-animation' });
});

test('auto task outputs preserve custom values and skip disabled fields', () => {
  const template = {
    ...shotTemplate,
    task_outputs: { Animation: 'custom-animation' },
    task_output_enabled: { Lighting: false },
  };

  const outputs = fillDefaultTaskOutputs(template, [
    { name: 'Animation' },
    { name: 'Lighting' },
    { name: 'Compositing' },
  ]);

  assert.deepEqual(outputs, {
    Animation: 'custom-animation',
    Compositing: '<Shot>-compositing',
  });
});
