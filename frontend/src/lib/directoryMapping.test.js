import assert from 'node:assert/strict';
import test from 'node:test';

import {
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

test('directory mapping identifies asset and shot templates', () => {
  assert.equal(inferDirectoryTemplateEntity(assetTemplate.id, assetTemplate.template), 'asset');
  assert.equal(inferDirectoryTemplateEntity(shotTemplate.id, shotTemplate.template), 'shot');
});

test('task output defaults use the matching entity placeholder', () => {
  assert.equal(getDefaultTaskOutput(assetTemplate, 'Modeling'), '<Asset>-modeling');
  assert.equal(getDefaultTaskOutput(shotTemplate, 'Animation'), '<Shot>-animation');
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
