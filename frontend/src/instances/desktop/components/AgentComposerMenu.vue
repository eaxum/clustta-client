<template>
  <Teleport to="body">
    <div v-if="visible" ref="menuRef" class="agent-composer-menu" :style="menuStyle" role="listbox">
      <div v-if="title" class="agent-composer-menu-title">{{ title }}</div>
      <button v-for="(item, index) in items" :key="item.key" type="button"
        class="agent-composer-menu-item" :class="{ active: index === activeIndex, 'has-icon': item.icon }"
        @mouseenter="$emit('active-change', index)" @mousedown.prevent="$emit('select', item)">
        <img v-if="item.icon" class="agent-composer-menu-icon small-icons" :class="{ 'no-filter': item.noFilter }"
          :src="item.icon">
        <span class="agent-composer-menu-label">{{ item.label }}</span>
        <span v-if="item.meta" v-tooltip="item.meta" class="agent-composer-menu-meta">{{ item.meta }}</span>
        <span v-if="item.description" class="agent-composer-menu-description">{{ item.description }}</span>
      </button>
      <div v-if="!items.length" class="agent-composer-menu-empty">{{ emptyText }}</div>
    </div>
  </Teleport>
</template>

<script setup>
import { nextTick, ref, watch } from 'vue';

const props = defineProps({
  activeIndex: { type: Number, default: 0 },
  emptyText: { type: String, default: 'No matches' },
  items: { type: Array, default: () => [] },
  menuStyle: { type: Object, default: () => ({}) },
  title: { type: String, default: '' },
  visible: { type: Boolean, default: false },
});

defineEmits(['active-change', 'select']);

const menuRef = ref(null);

const scrollActiveItemIntoView = async () => {
  await nextTick();
  const menu = menuRef.value;
  const activeItem = menu?.querySelector('.agent-composer-menu-item.active');
  if (!menu || !activeItem) return;

  const itemTop = activeItem.offsetTop;
  const itemBottom = itemTop + activeItem.offsetHeight;
  const visibleTop = menu.scrollTop;
  const visibleBottom = visibleTop + menu.clientHeight;
  if (itemTop < visibleTop) {
    menu.scrollTop = itemTop;
  } else if (itemBottom > visibleBottom) {
    menu.scrollTop = itemBottom - menu.clientHeight;
  }
};

watch(() => props.activeIndex, scrollActiveItemIntoView);
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
  grid-template-columns: minmax(0, 1fr) minmax(0, 2fr);
  gap: 0.15rem 0.4rem;
  width: 100%;
  padding: 0.45rem 0.55rem;
  border: 0;
  border-radius: var(--small-radius);
  color: var(--text);
  background: transparent;
  font-family: inherit;
  font-size: 12px;
  text-align: left;
  cursor: pointer;
}

.agent-composer-menu-item.has-icon {
  grid-template-columns: 16px minmax(0, 1fr) minmax(0, 2fr);
  align-items: center;
}

.agent-composer-menu-icon {
  width: 14px;
  height: 14px;
}

.agent-composer-menu-item:hover,
.agent-composer-menu-item.active {
  background: var(--hover);
}

.agent-composer-menu-label {
  min-width: 0;
  overflow: hidden;
  font-family: inherit;
  font-size: 12px;
  font-weight: 600;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.agent-composer-menu-meta {
  min-width: 0;
  overflow: hidden;
  font-family: inherit;
  color: var(--text-muted);
  direction: rtl;
  font-size: 12px;
  text-align: left;
  text-overflow: ellipsis;
  white-space: nowrap;
}

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
