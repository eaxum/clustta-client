<template>
  <Teleport to="body">
    <transition
      name="menu-fade"
      @before-enter="startAnimation"
      @after-enter="endAnimation"
      @before-leave="startLeaveAnimation"
      @after-leave="endAnimation"
    >
      <div v-if="menu.contextMenuVisible" ref="menuEl" :style="menuStyle" class="context-menu-container">
        <component v-for="menu in visibleMenus" :key="menu.name" :is="menu.component" />
      </div>

    </transition>
  </Teleport>
</template>

<script setup>
// imports
import { computed, onMounted, onUnmounted, reactive, ref, watch } from 'vue';

// components
import AccountMenu from '@/instances/desktop/menus/AccountMenu.vue';
import AssetMenu from '@/instances/desktop/menus/AssetMenu.vue';
import AssetTypeFilterMenu from '@/instances/desktop/menus/AssetTypeFilterMenu.vue';
import AssigneeFilterMenu from '@/instances/desktop/menus/AssigneeFilterMenu.vue';
import AssignMenu from '@/instances/desktop/menus/AssignMenu.vue';
import CheckpointSelectionMenu from '@/instances/desktop/menus/CheckpointSelectionMenu.vue';
import CollectionMenu from '@/instances/desktop/menus/CollectionMenu.vue';
import CollectionTypeFilterMenu from '@/instances/desktop/menus/CollectionTypeFilterMenu.vue';
import CheckpointTagMenu from '@/instances/desktop/menus/CheckpointTagMenu.vue';
import CopyToProjectSubMenu from '@/instances/desktop/menus/CopyToProjectSubMenu.vue';
import DependencySearchFilterMenu from '@/instances/desktop/menus/DependencySearchFilterMenu.vue';
import ExtensionFilterMenu from '@/instances/desktop/menus/ExtensionFilterMenu.vue';
import ExportColumnsMenu from '@/instances/desktop/menus/ExportColumnsMenu.vue';
import ManageTagsMenu from '@/instances/desktop/menus/ManageTagsMenu.vue';
import MoveToCollectionSubMenu from '@/instances/desktop/menus/MoveToCollectionSubMenu.vue';
import ProjectItemMenu from '@/instances/desktop/menus/ProjectItemMenu.vue';
import ProjectMenu from '@/instances/desktop/menus/ProjectMenu.vue';
import RoleSelectionMenu from '@/instances/desktop/menus/RoleSelectionMenu.vue';
import SelectionMenu from '@/instances/desktop/menus/SelectionMenu.vue';
import StateFilterMenu from '@/instances/desktop/menus/StateFilterMenu.vue';
import StatusFilterMenu from '@/instances/desktop/menus/StatusFilterMenu.vue';
import TagsFilterMenu from '@/instances/desktop/menus/TagsFilterMenu.vue';
import SortMenu from '@/instances/desktop/menus/SortMenu.vue';
import TypeFilterMenu from '@/instances/desktop/menus/TypeFilterMenu.vue';
import UntrackedItemMenu from '@/instances/desktop/menus/UntrackedItemMenu.vue';
import ViewMenu from '@/instances/desktop/menus/ViewMenu.vue';

// stores
import { useMenu } from '@/stores/menu';

const menu = useMenu();

// refs
const menuDimensions = reactive({ height: 0, width: 0 });
const menuEl = ref(null);
const viewportSize = reactive({ height: window.innerHeight, width: window.innerWidth });
let menuResizeObserver = null;

// menu components mapping
const menuComponents = {
  accountMenu: AccountMenu,
  assetMenu: AssetMenu,
  assetTypeFilterMenu: AssetTypeFilterMenu,
  assigneeFilterMenu: AssigneeFilterMenu,
  assignMenu: AssignMenu,
  checkpointSelectionMenu: CheckpointSelectionMenu,
  collectionMenu: CollectionMenu,
  collectionTypeFilterMenu: CollectionTypeFilterMenu,
  checkpointTagMenu: CheckpointTagMenu,
  copyToProjectSubMenu: CopyToProjectSubMenu,
  dependencySearchFilterMenu: DependencySearchFilterMenu,
  extensionFilterMenu: ExtensionFilterMenu,
  exportColumnsMenu: ExportColumnsMenu,
  manageTagsMenu: ManageTagsMenu,
  moveToCollectionSubMenu: MoveToCollectionSubMenu,
  projectItemMenu: ProjectItemMenu,
  projectMenu: ProjectMenu,
  roleSelectionMenu: RoleSelectionMenu,
  selectionMenu: SelectionMenu,
  sortMenu: SortMenu,
  stateFilterMenu: StateFilterMenu,
  statusFilterMenu: StatusFilterMenu,
  tagsFilterMenu: TagsFilterMenu,
  typeFilterMenu: TypeFilterMenu,
  untrackedItemMenu: UntrackedItemMenu,
  viewMenu: ViewMenu,
};

// computed properties
// Calculates menu position to keep it within viewport bounds.
const menuStyle = computed(() => {
  if (!menuEl.value) return {};
  
  const menuRect = menuEl.value.getBoundingClientRect();
  const activeMenuWidth = menuDimensions.width || menuRect.width;
  const activeMenuHeight = menuDimensions.height || menuRect.height;
  const margin = 15;
  const style = {};

  let left = menu.position.x;
  let top = menu.position.y;

  if (left + activeMenuWidth > viewportSize.width - margin) {
    left = viewportSize.width - activeMenuWidth - margin;
  }
  if (left < margin) {
    left = margin;
  }

  if (top + activeMenuHeight > viewportSize.height - margin) {
    const spaceAbove = menu.position.y - margin;
    if (spaceAbove >= activeMenuHeight) {
      top = menu.position.y - activeMenuHeight;
    } else {
      top = viewportSize.height - activeMenuHeight - margin;
      if (top < margin) {
        top = margin;
        style.maxHeight = `${viewportSize.height - (2 * margin)}px`;
        style.overflowY = 'auto';
      }
    }
  }
  if (top < margin) {
    top = margin;
  }

  return { ...style, left: `${left}px`, top: `${top}px` };
});

// Returns list of currently visible menu components.
const visibleMenus = computed(() => {
  return Object.entries(menu.menuStates)
    .filter(([name, isVisible]) => isVisible)
    .map(([name]) => ({ name, component: menuComponents[name] }));
});

// methods
const endAnimation = () => {
  menu.isAnimating = false;
};

// Hides the context menu on document click.
const hideContextMenu = (event) => {
  if (!menuEl.value || menuEl.value.contains(event.target)) return;
  menu.hideContextMenu();
};

const startAnimation = () => {
  menu.isAnimating = true;
};

const startLeaveAnimation = () => {
  menu.isAnimating = true;
};

const updateViewportSize = () => {
  viewportSize.height = window.innerHeight;
  viewportSize.width = window.innerWidth;
};

watch(menuEl, (element, previousElement) => {
  menu.menuEl = element;
  if (!menuResizeObserver) return;
  if (previousElement) menuResizeObserver.unobserve(previousElement);
  if (element) menuResizeObserver.observe(element);
}, { flush: 'post' });

// lifecycle hooks
onMounted(() => {
  menu.menuEl = menuEl.value;
  if (typeof ResizeObserver !== 'undefined') {
    menuResizeObserver = new ResizeObserver((entries) => {
      const entry = entries[0];
      if (!entry) return;
      menuDimensions.height = entry.borderBoxSize?.[0]?.blockSize || entry.contentRect.height;
      menuDimensions.width = entry.borderBoxSize?.[0]?.inlineSize || entry.contentRect.width;
    });
    if (menuEl.value) menuResizeObserver.observe(menuEl.value);
  }
  document.addEventListener('pointerdown', hideContextMenu, true);
  window.addEventListener('resize', updateViewportSize);
});

onUnmounted(() => {
  menuResizeObserver?.disconnect();
  document.removeEventListener('pointerdown', hideContextMenu, true);
  window.removeEventListener('resize', updateViewportSize);
});

</script>

<style>
.menu-item-text {
  font-weight: 400;
}

[data-theme="dark"] .menu-item-text {
  font-weight: 200;
}

.context-menu-container {
  z-index: 1000;
  display: flex;
  position: fixed;
  flex-direction: column;
  align-items: center;
  gap: .3rem;
  box-sizing: border-box;
  width: max-content;
  height: max-content;
  max-height: 70vh;
  overflow: hidden;
  overflow-y: scroll;
  border-radius: var(--large-radius);
  outline: var(--transparent-line);
  outline-offset: -1px;
  backdrop-filter: blur(55px);
}

.context-menu-container::-webkit-scrollbar {
  width: 8px;
  display: none;
}

.context-menu-container::-webkit-scrollbar-thumb {
  border-radius: 10px;
  background-color: var(--surface-1);
}

.context-menu-container::-webkit-scrollbar-track {
  margin: 10px;
  border-radius: 10px;
}

@keyframes context-menu-bubble-up {
  0% {
    opacity: 0;
    transform: translateY(-6px) scale(0.985);
  }

  50% {
    opacity: 1;
    transform: translateY(0) scale(1.01);
  }

  100% {
    opacity: 1;
    transform: translateY(0) scale(1);
  }
}

.menu-fade-enter-active {
  animation: context-menu-bubble-up 120ms cubic-bezier(0.2, 0.8, 0.2, 1) both;
  transform-origin: top left;
  will-change: transform, opacity;
}

.menu-fade-leave-active {
  transition: opacity 100ms ease, transform 100ms ease;
  transform-origin: top left;
}

.menu-fade-leave-to {
  opacity: 0;
  transform: translateY(-3px) scale(0.985);
}

@media (prefers-reduced-motion: reduce) {
  .menu-fade-enter-active {
    animation: none;
  }

  .menu-fade-leave-active {
    transition: none;
  }
}


.horizontal-flex{
  padding: 0px;
}

</style>

