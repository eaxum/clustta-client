import test from 'node:test';
import assert from 'node:assert/strict';
import { computed, reactive } from 'vue';
import { canExportFiles, getExportDragSelection } from './exportDrag.js';

const asset = { id: 'first', type: 'asset', file_path: 'C:/project/first.blend', file_status: 'normal' };
const other = { ...asset, id: 'second' };

test('dragging a selected asset preserves the entire selection', () => {
  assert.deepEqual(getExportDragSelection(asset, [asset, other], {}), [asset, other]);
  assert.deepEqual(getExportDragSelection(asset, [other], {}), [asset]);
});

test('downloaded selection becomes exportable from cache updates without another click', () => {
  const selectedItems = [asset, other].map((item) => ({ ...item, file_status: 'fetchable' }));
  const itemsByKey = reactive(Object.fromEntries(
    selectedItems.map((item) => [`asset:${item.id}`, { ...item }]),
  ));
  const selection = computed(() => getExportDragSelection(selectedItems[0], selectedItems, itemsByKey));
  const exportable = computed(() => canExportFiles(selection.value));

  assert.equal(exportable.value, false);
  itemsByKey['asset:first'].file_status = 'normal';
  assert.equal(exportable.value, false);
  itemsByKey['asset:second'] = { ...other, file_status: 'normal' };
  assert.equal(exportable.value, true);
  assert.equal(selectedItems[0].file_status, 'fetchable');
  assert.equal(selectedItems[1].file_status, 'fetchable');

  itemsByKey['asset:first'].file_status = 'missing';
  assert.equal(exportable.value, false);
});

test('cache lookup preserves uncached selection members and distinguishes item types', () => {
  const collection = { id: asset.id, type: 'collection' };
  const selection = getExportDragSelection(other, [collection, other], {
    [`asset:${asset.id}`]: asset,
  });
  assert.deepEqual(selection, [collection, other]);
  assert.equal(canExportFiles(selection), false);
});

test('mixed selections never silently export only eligible assets', () => {
  for (const invalid of [
    { ...other, type: 'collection' }, { ...other, file_status: 'fetchable' }, { ...other, file_status: 'missing' },
    { ...other, pointer: 'https://example.com' }, { ...other, is_link: true },
    { ...other, trashed: true }, { ...other, file_path: '' },
  ]) assert.equal(canExportFiles([asset, invalid]), false);
  assert.equal(canExportFiles([asset, other]), true);
  assert.equal(canExportFiles([]), false);
});
