<template>
  <div class="tools-manager">
    <!-- Display selected tools -->
    <div v-if="tools.length > 0" class="tools-container">
      <div
        v-for="tool in tools"
        :key="tool.id"
        class="tool-item"
      >
        <img 
          v-if="tool.logo" 
          :src="getAppIcon(tool.logo)" 
          :alt="tool.name" 
          class="small-icons"
          @error="handleImageError"
        />
        <img 
          v-else 
          :src="getAppIcon('file')" 
          alt="Tool" 
          class="small-icons"
        />
        <span class="tool-name">{{ tool.name }}</span>
        <button
          v-if="isEditing"
          @click="removeTool(tool)"
          class="tool-remove-button"
          title="Remove"
        >
          <img class="remove-icon" :src="getAppIcon('close')" alt="Remove">
        </button>
      </div>
    </div>
    
    <!-- ItemSelector for adding new tools -->
    <ItemSelector
      v-if="isEditing"
      :selectedItems="tools"
      :allItems="allTools"
      :placeholder="'Search and add tools...'"
      :itemType="'tool'"
      @itemAdded="addTool"
    />
  </div>
</template>

<script setup>
import { ref } from 'vue';
import { useIconStore } from '@/stores/icons';
import ItemSelector from './ItemSelector.vue';

const iconStore = useIconStore();

const props = defineProps({
  tools: {
    type: Array,
    default: () => []
  },
  isEditing: {
    type: Boolean,
    default: false
  },
  allTools: {
    type: Array,
    default: () => [
      { id: '1', name: 'Blender', category: '3D', logo: 'box' },
      { id: '2', name: 'Maya', category: '3D', logo: 'box' },
      { id: '3', name: 'Cinema 4D', category: '3D', logo: 'box' },
      { id: '4', name: '3ds Max', category: '3D', logo: 'box' },
      { id: '5', name: 'Houdini', category: '3D', logo: 'box' },
      { id: '6', name: 'ZBrush', category: '3D', logo: 'palette' },
      { id: '7', name: 'Substance Painter', category: '2D', logo: 'palette' },
      { id: '8', name: 'Photoshop', category: '2D', logo: 'palette' },
      { id: '9', name: 'After Effects', category: 'Compositing', logo: 'layers' },
      { id: '10', name: 'Nuke', category: 'Compositing', logo: 'layers' },
      { id: '11', name: 'Unreal Engine', category: 'Game Engine', logo: 'gamepad' },
      { id: '12', name: 'Unity', category: 'Game Engine', logo: 'gamepad' },
      { id: '13', name: 'Marvelous Designer', category: '3D', logo: 'box' },
      { id: '14', name: 'DaVinci Resolve', category: 'Video Editing', logo: 'film' },
      { id: '15', name: 'Premiere Pro', category: 'Video Editing', logo: 'film' }
    ]
  }
});

const emit = defineEmits(['toolAdded', 'toolRemoved']);

const addTool = (tool) => {
  emit('toolAdded', tool);
};

const removeTool = (tool) => {
  emit('toolRemoved', tool.name);
};

const handleImageError = (event) => {
  // If image fails to load, replace with default icon
  event.target.src = iconStore.getAppIcon('box');
  event.target.classList.add('tool-logo-default');
};

const getAppIcon = (iconName) => {
  return iconStore.getAppIcon(iconName);
};
</script>

<style scoped>
@import "@/assets/desktop.css";

.tools-manager {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  width: 100%;
}

.tools-container {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
}

.tool-item {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 0.75rem;
  background-color: var(--steel);
  border-radius: var(--normal-radius);
  transition: background-color 0.2s;
}

.tool-item:hover {
  background-color: rgba(255, 255, 255, 0.1);
}

.tool-remove-button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background: none;
  border: none;
  cursor: pointer;
  padding: 0;
  margin-left: 0.25rem;
  opacity: 0.7;
  transition: opacity 0.2s;
}

.tool-remove-button:hover {
  opacity: 1;
}

.remove-icon {
  width: 12px;
  height: 12px;
  filter: brightness(0) invert(1);
}

.tool-name {
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--white);
  user-select: none;
}

.tool-logo-default {
  filter: brightness(0) invert(1);
  opacity: 0.7;
}
</style>
