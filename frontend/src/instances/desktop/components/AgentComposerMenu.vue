<template>
  <Teleport to="body">
    <div v-if="visible" class="agent-composer-menu" :style="menuStyle" role="listbox">
      <div v-if="title" class="agent-composer-menu-title">{{ title }}</div>
      <button v-for="(item, index) in items" :key="item.key" type="button"
        class="agent-composer-menu-item" :class="{ active: index === activeIndex }"
        @mousedown.prevent="$emit('select', item)">
        <span class="agent-composer-menu-label">{{ item.label }}</span>
        <span v-if="item.meta" class="agent-composer-menu-meta">{{ item.meta }}</span>
        <span v-if="item.description" class="agent-composer-menu-description">{{ item.description }}</span>
      </button>
      <div v-if="!items.length" class="agent-composer-menu-empty">{{ emptyText }}</div>
    </div>
  </Teleport>
</template>

<script setup>
defineProps({
  activeIndex: { type: Number, default: 0 },
  emptyText: { type: String, default: 'No matches' },
  items: { type: Array, default: () => [] },
  menuStyle: { type: Object, default: () => ({}) },
  title: { type: String, default: '' },
  visible: { type: Boolean, default: false },
});

defineEmits(['select']);
</script>

<style scoped>
.agent-composer-menu {
  position: fixed;
  z-index: 10000;
  display: flex;
  flex-direction: column;
  width: calc(100vw - 16px);
  max-height: 320px;
  padding: 0.3rem;
  overflow: hidden;
  overflow-y: auto;
  box-sizing: border-box;
  border: 1px solid var(--surface-4);
  border-radius: var(--large-radius);
  background: var(--surface-3);
}


.agent-composer-menu::-webkit-scrollbar {
  width: 4px;
}

.agent-composer-menu::-webkit-scrollbar-thumb {
  border-radius: var(--small-radius);
  background-color: var(--surface-2);
}

.agent-composer-menu::-webkit-scrollbar-track {
  border-radius: var(--small-radius);
  margin: 1rem 0;
}

.agent-composer-menu-title {
  padding: 0.35rem 0.5rem;
  color: var(--text-muted);
  font-size: 11px;
  font-weight: 500;
}

.agent-composer-menu-item {
  display: grid;
  grid-template-columns: max-content 1fr;
  gap: 0.15rem 0.4rem;
  width: 100%;
  padding: 0.45rem 0.55rem;
  border: 0;
  border-radius: var(--small-radius);
  color: var(--text);
  background: transparent;
  text-align: left;
  cursor: pointer;
}

.agent-composer-menu-item:hover,
.agent-composer-menu-item.active {
  background: var(--hover);
}

.agent-composer-menu-label {
  font-family: 'Consolas', 'Monaco', monospace;
  font-size: 12px;
  font-weight: 600;
}

.agent-composer-menu-meta,
.agent-composer-menu-description,
.agent-composer-menu-empty {
  color: var(--text-muted);
  font-size: 11px;
}

.agent-composer-menu-description {
  grid-column: 1 / -1;
}

.agent-composer-menu-empty {
  padding: 0.75rem;
  text-align: center;
}
</style>
