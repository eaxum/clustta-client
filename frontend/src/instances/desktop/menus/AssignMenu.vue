<template>
  <div ref="collectionMenu" class="filter-menu-container">
    <div class="input-section">
      <div class="horizontal-flex">
        <input ref="searchUserInput" v-stop-propagation v-model="searchUserTerm" class="input-short" type="text"
          :placeholder="$t('placeholders.searchUser')" />
      </div>
    </div>

    <div class="assignee-scroll-container">
      <!-- Current Assignee -->
      <div v-if="assignee && !multipleAssets" class="current-assignee-section" :class="{ 'has-others': collaboratorsList.length > 0 }">
        <div class="section-label">{{ $t('menus.assigned') }}</div>
        <div class="assignee-list-container current-assignee">
          <AssigneeItem 
            :name="assignee.name" 
            :assigneeId="assignee.id"
            :photo="assignee.photo" 
            :avatarColor="assignee.avatarColor"
          >
            <template #actions>
              <span v-if="canUnassignAssets" v-stop-propagation class="single-action-button" @click="unassignAsset()" v-tooltip="$t('common.unassign')">
                <img class="small-icons" :src="getAppIcon('person-minus')">
              </span>
            </template>
          </AssigneeItem>
        </div>
      </div>

      <!-- Project Collaborators -->
      <div v-if="canAssignAssets && collaboratorsList && collaboratorsList.length" class="assignee-list-container">
        <AssigneeItem 
          v-stop-propagation
          v-for="(collaborator, index) in collaboratorsList" 
          :key="index" 
          :assigneeId="collaborator.id"
          :name="collaborator.name" 
          :userPhoto="collaborator.photo" 
          :avatarColor="collaborator.avatarColor"
          :isLoading="loadingUserIds.includes(collaborator.id)"
          @click="assignAsset(collaborator.id)"
        />
      </div>

      <!-- Studio Users Divider -->
      <div v-if="canAssignAssets && searchUserTerm && filteredStudioUsers.length && collaboratorsList.length" class="studio-users-divider">
        <span class="divider-text">{{ $t('menus.studioMembers') }}</span>
      </div>

      <!-- Studio Users (not in project) -->
      <div v-if="canAssignAssets && searchUserTerm && filteredStudioUsers.length" class="assignee-list-container">
        <AssigneeItem 
          v-stop-propagation
          v-for="(user, index) in filteredStudioUsers" 
          :key="'studio-' + index" 
          :assigneeId="user.id"
          :name="user.name" 
          :userPhoto="user.photo" 
          :avatarColor="user.avatarColor"
          :isLoading="loadingUserIds.includes(user.id)"
          @click="assignStudioUser(user)"
        />
      </div>

      <!-- No Results -->
      <div v-if="searchUserTerm && !collaboratorsList.length && !filteredStudioUsers.length" class="no-results">
        {{ $t('menus.noResults') }}
      </div>
    </div>
  </div>
</template>

<script setup>
// imports
import { computed, onBeforeUnmount, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import emitter from '@/lib/mitt';
import { canActOnAssets } from '@/lib/permissions';
import utils from '@/services/utils';

// components
import AssigneeItem from '@/instances/common/components/AssigneeItem.vue';

// services
import { AssetService, ProjectService } from "@/services";

// stores
import { useAssetStore } from '@/stores/assets';
import { useIconStore } from '@/stores/icons';
import { useMenu } from '@/stores/menu';
import { useNotificationStore } from '@/stores/notifications';
import { useProjectStore } from '@/stores/projects';
import { useStageStore } from '@/stores/stages';
import { useStudioStore } from '@/stores/studio';
import { useUserStore } from '@/stores/users';

const { t } = useI18n();
const assetStore = useAssetStore();
const iconStore = useIconStore();
const menu = useMenu();
const notificationStore = useNotificationStore();
const projectStore = useProjectStore();
const stage = useStageStore();
const studioStore = useStudioStore();
const userStore = useUserStore();

// refs
const collectionMenu = ref(null);
const loadingUserIds = ref([]);
const searchUserInput = ref(null);
const searchUserTerm = ref('');

// computed properties
// Returns the current assignee data formatted for display.
const assignee = computed(() => {
  if (!asset.value || !asset.value.assignee_id) return;

  const user = userStore.getUserData(asset.value.assignee_id);
  return {
    name: `${user.first_name} ${user.last_name}` || user,
    photo: user.photo || "",
    avatarColor: userStore.userProfileColor(user.id),
    id: user.id,
  };
});

// Returns formatted list of project collaborators, excluding current assignee.
const collaboratorsList = computed(() => {
  const allCollaborators = projectCollaborators.value;
  if (multipleAssets.value || !asset.value || !asset.value.assignee_id) {
    const availableCollaborators = allCollaborators.filter((item) => item.username.toLowerCase().includes(searchUserTerm.value));
    return utils.sortAlphabetically(formatCollaborators(availableCollaborators));
  }

  const filteredCollaborators = allCollaborators.filter((item) => item.id !== assignee.value.id && item.username.toLowerCase().includes(searchUserTerm.value));
  return utils.sortAlphabetically(formatCollaborators(filteredCollaborators));
});

// Returns studio users who are not in the current project, filtered by search term.
const filteredStudioUsers = computed(() => {
  if (!searchUserTerm.value) return [];
  
  const projectUserIds = projectCollaborators.value.map(user => user.id);
  const studioUsers = studioStore.studioUsers || [];
  const query = searchUserTerm.value.toLowerCase();
  
  const users = studioUsers
    .filter(user => !projectUserIds.includes(user.id))
    .filter(user => {
      const fullName = `${user.first_name} ${user.last_name}`.toLowerCase();
      return fullName.includes(query) || 
             user.email?.toLowerCase().includes(query) ||
             user.username?.toLowerCase().includes(query);
    })
    .map(user => ({
      name: `${user.first_name} ${user.last_name}` || user,
      photo: user.photo || "",
      email: user.email,
      avatarColor: userStore.userProfileColor(user.id),
      id: user.id,
    }));
  
  return utils.sortAlphabetically(users);
});

// Checks if multiple assets are selected.
const multipleAssets = computed(() => stage.markedItems.length > 1);

// Returns the list of project collaborators.
const projectCollaborators = computed(() => userStore.getProjectCollaborators);

// Returns the currently selected asset.
const asset = computed(() => assetStore.selectedAsset);
const targetAssets = computed(() => {
  if (multipleAssets.value) return stage.selectedItems.filter(item => item.type === 'asset');
  return asset.value ? [asset.value] : [];
});
const canAssignAssets = computed(() => canActOnAssets('assign_asset', targetAssets.value));
const canUnassignAssets = computed(() => canActOnAssets('unassign_asset', targetAssets.value));

// methods
// Assigns a asset to a user, handling single or multiple selection.
const assignAsset = (assigneeId) => {
  if (!canAssignAssets.value) return;
  if (!multipleAssets.value) {
    assignSingleAsset(assigneeId);
  } else {
    assignMultipleAssets(assigneeId);
  }
};

// Assigns multiple assets to a user.
const assignMultipleAssets = async (assigneeId) => {
  if (!canAssignAssets.value) return;
  let assetIds = stage.markedItems;

  return await AssetService.AssignAssets(projectStore.activeProject.uri, assetIds, assigneeId)
    .then(async (result) => {
      for (const assetId of assetIds) {
        emitAssetUpdates(assetId, [
          { property: 'assignee_id', value: assigneeId },
          { property: 'is_resource', value: false }
        ]);
      }
      menu.disableAllMenus();
      notificationStore.notifyMetadataUpdate(result, t('notifications.assetsAssigned'));
      return result;
    })
    .catch((error) => {
      console.log(error);
      notificationStore.errorNotification(t('notifications.errorAssigningAsset'), error);
      return undefined;
    });
};

// Assigns a single asset to a user.
const assignSingleAsset = async (assigneeId) => {
  if (!canAssignAssets.value) return;
  let selectedAsset = asset.value;
  let assetId = selectedAsset.id;
  let user = collaboratorsList.value.find((item) => item.id === assigneeId);
  let userId = user ? user.id : "";
  
  return await AssetService.AssignAsset(projectStore.activeProject.uri, assetId, userId)
    .then(async (result) => {
      selectedAsset.assignee_id = userId;
      selectedAsset.is_resource = false;
      emitAssetUpdates(assetId, [
        { property: 'assignee_id', value: userId },
        { property: 'is_resource', value: false }
      ]);
      menu.disableAllMenus();
      notificationStore.notifyMetadataUpdate(result, t('notifications.assetAssigned'));
      return result;
    })
    .catch((error) => {
      console.log(error);
      notificationStore.errorNotification(t('notifications.errorAssigningAsset'), error);
      return undefined;
    });
};

// Adds a studio user to the project and assigns the asset to them.
const assignStudioUser = async (user) => {
  if (!canAssignAssets.value) return;
  if (loadingUserIds.value.includes(user.id)) return;
  
  loadingUserIds.value.push(user.id);
  
  try {
    const roles = userStore.getRolesNames || [];
    const defaultRole = roles.find(role => role.toLowerCase() === 'artist') || roles[0];
    
    if (!defaultRole) {
      notificationStore.errorNotification(t('common.error'), t('notifications.noRolesAvailable'));
      return;
    }
    
    await ProjectService.AddUser(projectStore.activeProject.uri, user.email, defaultRole);
    await userStore.reloadUsers();
    
    const result = !multipleAssets.value
      ? await assignSingleAsset(user.id)
      : await assignMultipleAssets(user.id);

    if (result && !result.requires_sync) {
      notificationStore.addNotification(t('notifications.userAddedAndAssigned'), "", "success");
    }
  } catch (error) {
    console.error(error);
    notificationStore.errorNotification(t('notifications.errorAddingUserToProject'), error);
  } finally {
    loadingUserIds.value = loadingUserIds.value.filter(id => id !== user.id);
  }
};

// Emits asset update events to Browser and VirtualItem components.
const emitAssetUpdates = (assetId, updates) => {
  const updateData = { itemId: assetId, updates };
  emitter.emit('update-root-data', updateData);
  emitter.emit('update-children', updateData);
};

// Formats collaborator data for display.
const formatCollaborators = (arr) => {
  return arr.map((user, index) => ({
    name: `${user.first_name} ${user.last_name}` || user,
    photo: user.photo || "",
    avatarColor: userStore.userProfileColor(user.id),
    id: user.id,
    index: index.toString(),
  }));
};

// Returns the icon path for a given icon name.
const getAppIcon = (iconName) => {
  return iconStore.getAppIcon(iconName);
};

// Unassigns a asset, handling single or multiple selection.
const unassignAsset = () => {
  if (!canUnassignAssets.value) return;
  if (!multipleAssets.value) {
    unassignSingleAsset();
  } else {
    unassignMultipleAssets();
  }
};

// Unassigns multiple assets.
const unassignMultipleAssets = async () => {
  if (!canUnassignAssets.value) return;
  let assetIds = stage.markedItems;

  await AssetService.UnassignAssets(projectStore.activeProject.uri, assetIds)
    .then(async (result) => {
      for (const assetId of assetIds) {
        let asset = assetStore.findAsset(assetId);
        asset.assignee_id = null;
        emitAssetUpdates(assetId, [{ property: 'assignee_id', value: null }]);
      }
      menu.disableAllMenus();
      notificationStore.notifyMetadataUpdate(result, t('notifications.assetsUnassigned'));
    })
    .catch((error) => {
      console.log(error);
      notificationStore.errorNotification(t('notifications.errorUnassigningAsset'), error);
    });
};

// Unassigns a single asset.
const unassignSingleAsset = async () => {
  if (!canUnassignAssets.value) return;
  let selectedAsset = asset.value;
  let assetId = selectedAsset.id;
  
  await AssetService.UnassignAsset(projectStore.activeProject.uri, assetId)
    .then(async (result) => {
      selectedAsset.assignee_id = null;
      emitAssetUpdates(assetId, [{ property: 'assignee_id', value: null }]);
      notificationStore.notifyMetadataUpdate(result, t('notifications.assetUnassigned'));
      menu.disableAllMenus();
    })
    .catch((error) => {
      console.log(error);
      notificationStore.errorNotification(t('notifications.errorUnassigningAsset'), error);
    });
};

// lifecycle hooks
onMounted(() => {
  searchUserInput.value?.focus();
  menu.assetMenuWidth = collectionMenu.value.getBoundingClientRect().width;
  menu.collectionMenu = collectionMenu.value;
});

onBeforeUnmount(() => {
  menu.assetMenuWidth = collectionMenu.value.getBoundingClientRect().width;
  menu.assetMenuHeight = collectionMenu.value.getBoundingClientRect().height;
});
</script>

<style scoped>
@import "@/assets/desktop.css";
@import "@/assets/menu.css";

.input-section {
  min-height: min-content;
}

.input-short {
  flex: 1;
  width: 100%;
  font-size: 14px;
}

.assignee-scroll-container {
  flex-direction: column;
  gap: .3rem;
  max-height: 50vh;
  overflow: hidden;
  overflow-y: auto;
  width: 100%;
}

.assignee-scroll-container::-webkit-scrollbar {
  width: 4px;
}

.assignee-scroll-container::-webkit-scrollbar-thumb {
  border-radius: var(--small-radius);
  background-color: var(--surface-4);
}

.assignee-scroll-container::-webkit-scrollbar-track {
  border-radius: var(--small-radius);
}

.assignee-list-container {
  box-sizing: border-box;
  align-items: center;
  flex-direction: column;
  gap: .2rem;
  overflow: hidden;
  width: 100%;
  border-radius: 10px;
}

.current-assignee {
  overflow: hidden;
  min-height: min-content;
}

.current-assignee-section {
  display: flex;
  flex-direction: column;
  gap: .3rem;
}

.current-assignee-section.has-others {
  padding-bottom: .4rem;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
  margin-bottom: .2rem;
}

.section-label {
  font-size: 11px;
  color: var(--text);
  opacity: 0.5;
  padding-left: .2rem;
}

.no-results {
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 1rem;
  font-size: 12px;
  color: var(--text);
  opacity: 0.5;
}

.studio-users-divider {
  display: flex;
  align-items: center;
  width: 100%;
  padding: .3rem 0;
  gap: .5rem;
}

.studio-users-divider::before,
.studio-users-divider::after {
  content: '';
  flex: 1;
  height: 1px;
  background-color: var(--surface-inverse);
  opacity: 0.2;
}

.divider-text {
  font-size: 11px;
  color: var(--text);
  opacity: 0.5;
  white-space: nowrap;
}
</style>


