<template>
  <div class="data-table" :style="{ '--table-max-height': maxHeight, '--table-min-width': minWidth }">
    <div class="table-scroll">
      <table>
        <colgroup>
          <col v-for="column in columns" :key="column.key" :style="column.width ? { width: column.width } : {}" />
        </colgroup>
        <thead>
          <tr>
            <th v-for="column in columns" :key="column.key">{{ column.label }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-if="loading">
            <td :colspan="columns.length">{{ loadingText }}</td>
          </tr>
          <tr v-else-if="!rows.length">
            <td :colspan="columns.length">{{ emptyText }}</td>
          </tr>
          <tr v-for="(row, rowIndex) in rows" v-else :key="getRowKey(row, rowIndex)">
            <td v-for="column in columns" :key="column.key">
              <slot :name="`cell-${column.key}`" :row="row" :value="row[column.key]" :rowIndex="rowIndex">
                {{ formatValue(row[column.key]) }}
              </slot>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup>
const props = defineProps({
  columns: { type: Array, default: () => [] },
  rows: { type: Array, default: () => [] },
  rowKey: { type: [String, Function], default: 'id' },
  loading: { type: Boolean, default: false },
  loadingText: { type: String, default: 'Loading...' },
  emptyText: { type: String, default: 'No items' },
  maxHeight: { type: String, default: '300px' },
  minWidth: { type: String, default: '100%' },
});

const getRowKey = (row, index) => {
  if (typeof props.rowKey === 'function') return props.rowKey(row, index);
  return row?.[props.rowKey] ?? index;
};
const formatValue = (value) => Array.isArray(value) ? value.join(', ') : String(value ?? '');
</script>

<style scoped>
.data-table {
  width: 100%;
  overflow: hidden;
  outline: var(--transparent-line);
  outline-offset: -1px;
  border-radius: var(--large-radius);
}

.table-scroll {
  width: 100%;
  max-height: var(--table-max-height);
  overflow: auto;
  scrollbar-color: var(--surface-5) transparent;
  scrollbar-width: thin;
}

table {
  width: 100%;
  min-width: var(--table-min-width);
  border-collapse: collapse;
  table-layout: fixed;
  color: var(--text);
}

th,
td {
  padding: .75rem 1rem;
  text-align: left;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

th {
  position: sticky;
  top: 0;
  z-index: 1;
  background-color: var(--bg);
  color: var(--surface-5);
  font-size: .75rem;
  font-weight: 600;
  text-transform: uppercase;
}

td {
  border-top: 1px solid var(--surface-3);
  font-size: .875rem;
}

.table-scroll::-webkit-scrollbar {
  width: 4px;
  height: 4px;
}

.table-scroll::-webkit-scrollbar-thumb {
  border-radius: var(--small-radius);
  background-color: var(--surface-5);
}
</style>
