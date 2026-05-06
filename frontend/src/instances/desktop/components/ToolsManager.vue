<template>
  <div class="tools-manager">
    <!-- Display selected tools -->
    <div v-if="tools.length > 0" class="tools-container">
      <Chip
        v-for="tool in tools"
        :key="tool.id"
        :icon="getToolLogoPath(tool)"
        :label="tool.tool_name"
        :onRemove="() => removeTool(tool)"
        :readonly="readonly"
        :useImage="true"
      />
    </div>
    
    <!-- Empty state -->
    <div v-else-if="!isEditing" class="empty-state">
      {{ $t('components.toolsManager.noTools') }}
    </div>
    
    <!-- ItemSelector for adding new tools -->
    <div v-if="isEditing">
      <ItemSelector
        v-if="tools.length < 5"
        :selectedItems="tools"
        :allItems="normalizedAllTools"
        :placeholder="$t('components.toolsManager.searchPlaceholder')"
        :itemType="'tool'"
        :iconFilter="false"
        @itemAdded="addTool"
      />
      <div v-else class="limit-message">
        <CiInfo :size="20" class="limit-icon" />
        <span>{{ $t('components.toolsManager.maxReached') }}</span>
      </div>
    </div>
  </div>
</template>

<script setup>
import { CiInfo } from '@clustta/icons-vue';
import { resolveIcon } from '@/lib/icon-map';
import { ref, computed } from 'vue';
import { useI18n } from 'vue-i18n';
import { useIconStore } from '@/stores/icons';
import { useUserStore } from '@/stores/users';
import { useProfileStore } from '@/stores/profile';
import { useNotificationStore } from '@/stores/notifications';
import { ProfileService } from "@/services";
import ItemSelector from './ItemSelector.vue';
import Chip from '@/instances/common/components/Chip.vue';
import { getToolLogo } from '@/utils/iconMappers';

const iconStore = useIconStore();
const userStore = useUserStore();
const profileStore = useProfileStore();
const notificationStore = useNotificationStore();

const { t } = useI18n();

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
  },
  readonly: {
    type: Boolean,
    default: false
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
      notificationStore.addNotification(t('components.toolsManager.toolAdded'), t('components.toolsManager.toolAddedMessage'), "success", false);
    })
    .catch((err) => {
      notificationStore.errorNotification(t('components.toolsManager.failedToAddTool'), err?.message || err);
    });
};

const removeTool = (tool) => {
  ProfileService.RemoveUserTool(userStore.user.id, tool.tool_id)
    .then(() => {
      profileStore.removeTool(tool.id);
      notificationStore.addNotification(t('components.toolsManager.toolRemoved'), t('components.toolsManager.toolRemovedMessage'), "success", false);
    })
    .catch((err) => {
      notificationStore.errorNotification(t('components.toolsManager.failedToRemoveTool'), err?.message || err);
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
  return iconStore.resolveIcon(iconName);
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

.empty-state {
  padding: .5rem;
  text-align: center;
  color: var(--white);
  opacity: .5;
  font-style: italic;
  font-size: 0.875rem;
}
</style>
