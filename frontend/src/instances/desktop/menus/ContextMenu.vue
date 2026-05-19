<template>
  <Teleport to="body">
    <transition name="menu-fade" @enter="startAnimation" @after-enter="endAnimation" @leave="startLeaveAnimation">
      <div v-if="menu.contextMenuVisible" ref="menuEl" :style="menuStyle" class="context-menu-container">
        <component v-for="menu in visibleMenus" :key="menu.name" :is="menu.component" />
      </div>

    </transition>
  </Teleport>
</template>

<script setup>
// imports
import { computed, onMounted, onUnmounted, reactive, ref } from 'vue';

// components
import AccountMenu from '@/instances/desktop/menus/AccountMenu.vue';
import AssetMenu from '@/instances/desktop/menus/AssetMenu.vue';
import AssetTypeFilterMenu from '@/instances/desktop/menus/AssetTypeFilterMenu.vue';
import AssigneeFilterMenu from '@/instances/desktop/menus/AssigneeFilterMenu.vue';
import AssignMenu from '@/instances/desktop/menus/AssignMenu.vue';
import CollectionMenu from '@/instances/desktop/menus/CollectionMenu.vue';
import CollectionTypeFilterMenu from '@/instances/desktop/menus/CollectionTypeFilterMenu.vue';
import CopyToProjectSubMenu from '@/instances/desktop/menus/CopyToProjectSubMenu.vue';
import DependencySearchFilterMenu from '@/instances/desktop/menus/DependencySearchFilterMenu.vue';
import ExtensionFilterMenu from '@/instances/desktop/menus/ExtensionFilterMenu.vue';
import ManageTagsMenu from '@/instances/desktop/menus/ManageTagsMenu.vue';
import MoveToCollectionSubMenu from '@/instances/desktop/menus/MoveToCollectionSubMenu.vue';
import ProjectItemMenu from '@/instances/desktop/menus/ProjectItemMenu.vue';
import ProjectMenu from '@/instances/desktop/menus/ProjectMenu.vue';
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

// menu components mapping
const menuComponents = {
  accountMenu: AccountMenu,
  assetMenu: AssetMenu,
  assetTypeFilterMenu: AssetTypeFilterMenu,
  assigneeFilterMenu: AssigneeFilterMenu,
  assignMenu: AssignMenu,
  collectionMenu: CollectionMenu,
  collectionTypeFilterMenu: CollectionTypeFilterMenu,
  copyToProjectSubMenu: CopyToProjectSubMenu,
  dependencySearchFilterMenu: DependencySearchFilterMenu,
  extensionFilterMenu: ExtensionFilterMenu,
  manageTagsMenu: ManageTagsMenu,
  moveToCollectionSubMenu: MoveToCollectionSubMenu,
  projectItemMenu: ProjectItemMenu,
  projectMenu: ProjectMenu,
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
  
  const viewport = { width: window.innerWidth, height: window.innerHeight };
  const menuRect = menuEl.value.getBoundingClientRect();
  const activeMenuWidth = menuRect.width;
  const activeMenuHeight = menuDimensions.height || menuRect.height;
  const margin = 15;

  let left = menu.position.x;
  let top = menu.position.y;

  if (left + activeMenuWidth > viewport.width - margin) {
    left = viewport.width - activeMenuWidth - margin;
  }
  if (left < margin) {
    left = margin;
  }

  if (top + activeMenuHeight > viewport.height - margin) {
    const spaceAbove = menu.position.y - margin;
    if (spaceAbove >= activeMenuHeight) {
      top = menu.position.y - activeMenuHeight;
    } else {
      top = viewport.height - activeMenuHeight - margin;
      if (top < margin) {
        top = margin;
        const maxHeight = viewport.height - (2 * margin);
        if (menuEl.value) {
          menuEl.value.style.maxHeight = maxHeight + 'px';
          menuEl.value.style.overflowY = 'auto';
        }
      }
    }
  }
  if (top < margin) {
    top = margin;
  }

  return { left: `${left}px`, top: `${top}px` };
});

// Returns list of currently visible menu components.
const visibleMenus = computed(() => {
  return Object.entries(menu.menuStates)
    .filter(([name, isVisible]) => isVisible)
    .map(([name]) => ({ name, component: menuComponents[name] }));
});

// methods
// Handles the end of the enter animation.
const endAnimation = (el) => {
  el.style.height = '';
  menu.isAnimating = false;
};

// Hides the context menu on document click.
const hideContextMenu = (event) => {
  if (menuEl.value) {
    menu.hideContextMenu();
  }
};

// Starts the enter animation for the menu.
const startAnimation = (el) => {
  menu.isAnimating = true;
  el.style.height = '0px';
  el.style.opacity = '0';
  void el.offsetHeight;
  el.style.height = el.scrollHeight + 'px';
  menuDimensions.height = el.scrollHeight;
  el.style.opacity = '1';
};

// Starts the leave animation for the menu.
const startLeaveAnimation = (el) => {
  menu.isAnimating = true;
  el.style.height = el.scrollHeight + 'px';
  void el.offsetHeight;
  el.style.height = '0px';
  el.style.opacity = '0';
};

// lifecycle hooks
onMounted(() => {
  menu.menuEl = menuEl.value;
  document.addEventListener('click', hideContextMenu);
});

onUnmounted(() => {
  document.removeEventListener('click', hideContextMenu);
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

.menu-fade-enter-active,
.menu-fade-leave-active {
  transition: opacity 0.2s, height 0.2s;
  overflow: hidden;
}

.menu-fade-enter-from,
.menu-fade-leave-to {
  opacity: 0;
  height: 0;
}


.horizontal-flex{
  padding: 0px;
}

</style>

