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
import { ref, computed, onMounted, onUnmounted, nextTick, reactive } from 'vue';
import { useMenu } from '@/stores/menu';

// components
import TypeFilterMenu from '@/instances/desktop/menus/TypeFilterMenu.vue'
import AssetTypeFilterMenu from '@/instances/desktop/menus/AssetTypeFilterMenu.vue'
import CollectionTypeFilterMenu from '@/instances/desktop/menus/CollectionTypeFilterMenu.vue'
import StatusFilterMenu from '@/instances/desktop/menus/StatusFilterMenu.vue'
import ExtensionFilterMenu from '@/instances/desktop/menus/ExtensionFilterMenu.vue'
import TagsFilterMenu from '@/instances/desktop/menus/TagsFilterMenu.vue'
import AssigneeFilterMenu from '@/instances/desktop/menus/AssigneeFilterMenu.vue'
import StateFilterMenu from '@/instances/desktop/menus/StateFilterMenu.vue'

import AssignMenu from '@/instances/desktop/menus/AssignMenu.vue'
import ProjectMenu from '@/instances/desktop/menus/ProjectMenu.vue'
import ProjectItemMenu from '@/instances/desktop/menus/ProjectItemMenu.vue'
import AssetMenu from '@/instances/desktop/menus/AssetMenu.vue'
import UntrackedItemMenu from '@/instances/desktop/menus/UntrackedItemMenu.vue'
import CollectionMenu from '@/instances/desktop/menus/CollectionMenu.vue'
import AccountMenu from '@/instances/desktop/menus/AccountMenu.vue'

const menuComponents = {
  typeFilterMenu: TypeFilterMenu,
  assetTypeFilterMenu: AssetTypeFilterMenu,
  collectionTypeFilterMenu: CollectionTypeFilterMenu,
  statusFilterMenu: StatusFilterMenu,
  stateFilterMenu: StateFilterMenu,
  extensionFilterMenu: ExtensionFilterMenu,
  tagsFilterMenu: TagsFilterMenu,
  assigneeFilterMenu: AssigneeFilterMenu,

  projectMenu: ProjectMenu,
  projectItemMenu: ProjectItemMenu,
  collectionMenu: CollectionMenu,
  assetMenu: AssetMenu,
  untrackedItemMenu: UntrackedItemMenu,
  assignMenu: AssignMenu,
  accountMenu: AccountMenu,
};

const visibleMenus = computed(() => {
  return Object.entries(menu.menuStates)
    .filter(([name, isVisible]) => isVisible)
    .map(([name]) => ({
      name,
      component: menuComponents[name],
    }));
});

const menu = useMenu();
const menuDimensions = reactive({
  height: 0,
  width: 0
});

const hideMenu = () => {
  menu.disableAllMenus();
};

const menuEl = ref(null);

const menuStyle = computed(() => {
  if (!menuEl.value) return {};
  
  const viewport = {
    width: window.innerWidth,
    height: window.innerHeight
  };
  
  const menuRect = menuEl.value.getBoundingClientRect();
  const activeMenuWidth = menuRect.width;
  const activeMenuHeight = menuDimensions.height || menuRect.height;

  let left = menu.position.x;
  let top = menu.position.y;
  const margin = 15; // Safety margin from viewport edges

  // Handle horizontal clipping - ensure menu stays within viewport
  if (left + activeMenuWidth > viewport.width - margin) {
    left = viewport.width - activeMenuWidth - margin;
  }
  
  // Ensure menu doesn't go off the left edge
  if (left < margin) {
    left = margin;
  }

  // Handle vertical clipping - smarter positioning to avoid bottom clipping
  if (top + activeMenuHeight > viewport.height - margin) {
    // Try to position above the cursor/trigger if there's space
    const spaceAbove = menu.position.y - margin;
    if (spaceAbove >= activeMenuHeight) {
      top = menu.position.y - activeMenuHeight; // Small gap above trigger
    } else {
      // Position at the bottom of viewport with scrolling if needed
      top = viewport.height - activeMenuHeight - margin;
      
      // If menu is still too tall, set max height and enable scrolling
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
  
  // Ensure menu doesn't go off the top edge
  if (top < margin) {
    top = margin;
  }

  return {
    left: `${left}px`,
    top: `${top}px`,
  };
});

const startAnimation = (el) => {
  menu.isAnimating = true;
  el.style.height = '0px';
  el.style.opacity = '0';
  void el.offsetHeight;
  el.style.height = el.scrollHeight + 'px';
  menuDimensions.height = el.scrollHeight;
  el.style.opacity = '1';
};

const endAnimation = (el) => {
  el.style.height = '';
  menu.isAnimating = false;
};

const startLeaveAnimation = (el) => {
  menu.isAnimating = true;
  el.style.height = el.scrollHeight + 'px';
  void el.offsetHeight;
  el.style.height = '0px';
  el.style.opacity = '0';
};


const hideContextMenu = (event) => {
  if (menuEl.value) {
    menu.hideContextMenu();
  }
};

onMounted(() => {
  menu.menuEl = menuEl.value;
  document.addEventListener('click', hideContextMenu);
});

onUnmounted(() => {
  document.removeEventListener('click', hideContextMenu);
});

</script>

<style>

.menu-item-text{
  font-weight: 400;
}

[data-theme="dark"] .menu-item-text{
  font-weight: 200;
}

.menu-overlay-mask {
  top: 0;
  left: 0;
  position: absolute;
  z-index: 3;
  width: 100%;
  height: 100%;
  display: flex;
  transition: opacity 0.3s ease;
  background-color: rgba(0, 0, 0, 0.5);
  align-items: center;
  justify-content: center;
  backdrop-filter: blur(3px);
  box-sizing: border-box;
}

.context-menu {
  position: fixed;
  background-color: white;
  border: 1px solid #ccc;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);
  border-radius: 4px;
  overflow: hidden;
  z-index: 1000;
}

.context-menu-container {
  z-index: 1000;
  display: flex;
  position: fixed;
  flex-direction: column;
  color: white;
  align-items: center;
  gap: .3rem;
  box-sizing: border-box;
  width: max-content;
  height: max-content;
  max-height: 90vh;
  overflow: hidden;
  border-radius: var(--normal-radius);
  outline: var(--transparent-line);
  outline-offset: -1px;
  background-color: var(--light-steel);
  background-color: var(--dark-glass);
  backdrop-filter: blur(50px);
}

.context-menu-container::-webkit-scrollbar {
  width: 8px;
}

.context-menu-container::-webkit-scrollbar-thumb {
  border-radius: 10px;
  background-color: var(--black-steel);
}

.context-menu-container::-webkit-scrollbar-track {
  margin: 10px;
  border-radius: 10px;
}

.menu-item {
  padding: 8px 16px;
  cursor: pointer;
}

.menu-item:hover {
  background-color: #f0f0f0;
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
</style>