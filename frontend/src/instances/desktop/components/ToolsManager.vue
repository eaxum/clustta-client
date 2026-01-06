<template>
  <div class="tools-manager">
    <!-- Display selected tools -->
    <div v-if="tools.length > 0" class="tools-container">
      <div
        v-for="tool in tools"
        :key="tool.id"
        class="tool-item"
      >
        <div class="tool-icon-container">
          <img class="large-icons no-filter" :src="getToolLogoPath(tool)">
        </div>

        <span class="tool-name">{{ tool.tool_name }}</span>
        <ActionButton
          :icon="getAppIcon('close')"
          :buttonFunction="() => removeTool(tool)"
          :showIcon="true"
          :showLabel="false"
          :noFilter="false"
        />
      </div>
    </div>
    
    <!-- Empty state -->
    <div v-else-if="!isEditing" class="empty-state">
      No tools to display
    </div>
    
    <!-- ItemSelector for adding new tools -->
    <div v-if="isEditing">
      <ItemSelector
        v-if="tools.length < 5"
        :selectedItems="tools"
        :allItems="normalizedAllTools"
        :placeholder="'Search and add tools...'"
        :itemType="'tool'"
        @itemAdded="addTool"
      />
      <div v-else class="limit-message">
        <img :src="getAppIcon('info')" alt="Info" class="limit-icon" />
        <span>Maximum of 5 tools reached. Remove a tool to add another.</span>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue';
import { useIconStore } from '@/stores/icons';
import { useUserStore } from '@/stores/users';
import { useProfileStore } from '@/stores/profile';
import { useNotificationStore } from '@/stores/notifications';
import { ProfileService } from "@/services";
import ItemSelector from './ItemSelector.vue';
import ActionButton from './ActionButton.vue';
import { getToolLogo } from '@/utils/iconMappers';

const iconStore = useIconStore();
const userStore = useUserStore();
const profileStore = useProfileStore();
const notificationStore = useNotificationStore();

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
    default: () => []
  }
});

// Normalize allTools to have consistent 'name' field for ItemSelector
const normalizedAllTools = computed(() => {
  // Filter out tools that are already selected
  const selectedToolIds = props.tools.map(t => t.id);
  return props.allTools
    .filter(tool => !selectedToolIds.includes(tool.id))
    .map(tool => ({
      ...tool,
      name: tool.tool_name || tool.name,
      category: tool.tool_category || tool.category
    }));
});

const addTool = (tool) => {
  ProfileService.AddUserTool(userStore.user.id, {
    tool_id: tool.id,
    proficiency_level: tool.proficiency_level || 'intermediate'
  })
    .then(() => {
      // Transform the tool to match the expected structure
      const transformedTool = {
        ...tool,
        tool_id: tool.id,
        tool_name: tool.tool_name || tool.name,
        tool_category: tool.tool_category || tool.category
      };
      profileStore.addTool(transformedTool);
      notificationStore.addNotification("Tool added", "Tool added successfully.", "success", false);
    })
    .catch((err) => {
      notificationStore.errorNotification("Failed to add tool", err?.message || err);
    });
};

const removeTool = (tool) => {
  ProfileService.RemoveUserTool(userStore.user.id, tool.tool_id)
    .then(() => {
      profileStore.removeTool(tool.id);
      notificationStore.addNotification("Tool removed", "Tool removed successfully.", "success", false);
    })
    .catch((err) => {
      notificationStore.errorNotification("Failed to remove tool", err?.message || err);
    });
};

// Get tool logo path dynamically at render time
const getToolLogoPath = (tool) => {
  // Use the tool name to get the appropriate file icon
  const toolName = tool.tool_name || tool.ToolName || tool.name || '';
  return getToolLogo(toolName);
};

const handleImageError = (event) => {
  // If image fails to load, replace with default icon
  event.target.src = '/file-icons/default.svg';
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
  padding: 0.3rem;
  background-color: var(--steel);
  border-radius: var(--large-radius);
  transition: background-color 0.2s;
}

.tool-icon-container{
  padding: .3rem;
}

.tool-item:hover {
  background-color: rgba(255, 255, 255, 0.1);
}

.tool-name {
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--white);
  user-select: none;
}

.limit-message {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.75rem 1rem;
  background-color: rgba(255, 193, 7, 0.1);
  border: 1px solid rgba(255, 193, 7, 0.3);
  border-radius: var(--normal-radius);
  color: rgba(255, 193, 7, 0.9);
  font-size: 0.875rem;
}

.limit-icon {
  width: 16px;
  height: 16px;
  filter: invert(82%) sepia(89%) saturate(548%) hue-rotate(359deg) brightness(103%) contrast(98%);
  flex-shrink: 0;
}

.tool-logo-default {
  filter: brightness(0) invert(1);
  opacity: 0.7;
}

.empty-state {
  padding: .5rem;
  text-align: center;
  color: var(--white);
  opacity: .5;
  font-style: italic;
  font-size: 0.875rem;
}
</style>
