import test from 'node:test';
import assert from 'node:assert/strict';
import { canExportFiles, getExportDragSelection } from './exportDrag.js';

const asset = { id: 'first', type: 'asset', file_path: 'C:/project/first.blend', file_status: 'normal' };
const other = { ...asset, id: 'second' };

test('dragging a selected asset preserves the entire selection', () => {
  assert.deepEqual(getExportDragSelection(asset, [asset, other]), [asset, other]);
  assert.deepEqual(getExportDragSelection(asset, [other]), [asset]);
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
