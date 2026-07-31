<template>
  <!-- Grid View Asset Item -->
  <div v-if="commonStore.useGrid" 
    ref="assetItem" 
    :data-file-drop-target="!isUntracked ? '' : null"
    :id="!isUntracked ? 'drop-asset-' + asset.id : undefined"
    class="asset-item-main asset-item-grid" 
    v-return="launchSelectedAsset" 
    v-esc="handleEscKey" 
    v-stop-propagation
    :style="gridStyles" 
    :class="{
      'asset-item-grid-selected': stage.markedItems.includes(asset.id) && !isGhost,
      'asset-item-grid-cut': stage.cutItems.map((item) => item.id).includes(asset.id) && !isGhost,
      'asset-item-grid-only-selected': stage.markedItems.length === 1 && stage.firstSelectedItemId === asset.id && !isGhost,
      'asset-item-grid-last-selected': stage.lastSelectedItemId === asset.id && !isGhost,
      'asset-item-child': asset.parent_id,
      'file-drop-target-active': isHovered,
      'asset-item-untracked': isUntracked
    }" 
    @dblclick="launchAssetCommand()">

    <div class="main-asset-item-grid">
      
      <!-- Grid Status Menu Overlay -->
      <GridStatusMenu 
        v-if="gridStatusMenuVisible && !isUntracked && asset.status" 
        @statusSelected="handleGridStatusSelected"
        @close="closeGridStatusMenu"
      />
      
      <div class="main-asset-item-grid-thumb-container">

        <div v-if="asset.preview || osThumbnail" class="asset-item-preview-container">

          <div class="asset-item-preview-image">
            <img class="screenshot-thumb" :src="displayThumbnail">
          </div>

          <!-- Icon container at bottom left when preview is present -->
          <div class="asset-item-icon-container asset-item-icon-overlay">
            <img v-if="asset.icon" class="small-icons no-filter overlay-icons" :src="asset.icon">
            <img v-else-if="isUntracked" class="small-icons overlay-icons" :src="getAppIcon(getFileTypeIcon(asset))" @error="$event.target.src = getAppIcon('file')">
            <span v-else class="app-ext">
            </span>
          </div>
        </div>

        <div v-else class="asset-item-icon-container">
          <img class="gigantic-icons no-filter " :src="displayThumbnail">
        </div>

        <!-- Asset assignee overlay in top right corner -->
        <div v-if="!isUntracked && (!asset.is_resource || isCurrentUser) && !isEditing" class="asset-item-grid-assignee-overlay-top-right">
          <!-- Show assignee profile picture if assigned -->
          <div v-if="asset.assignee_id" @click="canManageAssetAssignment && prepAssignAsset(index, asset, $event)" v-stop-propagation class="asset-item-assignee">
            <span v-tooltip="userFullName" class="single-action-button">
              <div class="profile-picture-grid" :style="{ backgroundColor: profileColor(asset.assignee_id) }">
                <img v-if="userPhoto" class="profile-img-grid" :src="userPhoto">
                <img v-else class="profile-img-grid" :src="generateAvatar(asset.assignee_id)">
              </div>
            </span>
          </div>
        </div>
        
      </div>
      
      <!-- Bottom bar with asset type icon, name, and file status -->
      <div class="main-asset-item-grid-bottom-bar">
        
        <!-- Outermost container with relative positioning -->
        <div class="asset-item-grid-bottom-bar-wrapper">
          
          <!-- Middle container with slide-up transition and padding for file state -->
          <div v-if="!isEditing" class="asset-item-grid-slide-container">
            
            <!-- Row 1: Name/Meta (always visible) -->
            <div class="asset-item-grid-meta-row">
              <div v-if="settingsStore.showTypeIcons" class="asset-item-grid-type-icon" >
                <img v-if="isUntracked" class="small-icons" :src="getAppIcon('generic')">
                <img v-else class="small-icons" :src="getAppIcon(asset.asset_type_icon)" v-tooltip="assetTypeName">
              </div>
              
              <div class="main-asset-item-grid-meta" :style="{ fontStyle: isUntracked ? 'italic' : 'normal' }">
                {{ assetName }}
              </div>
            </div>
            
            <!-- Row 2: Action Buttons (shows on hover) -->
            <div class="asset-item-grid-actions-row">
              
              <!-- Untracked label for untracked items -->
              <div v-if="isUntracked" class="asset-item-grid-untracked-label">
                <span>{{ $t('blocks.untracked') }}</span>
              </div>
              
              <!-- Asset Status -->
              <div v-if="!isUntracked && asset.status" @click="openGridStatusMenu" class="asset-item-grid-status-display">
                <div class="asset-item-status-grid" :style="{ backgroundColor: asset.status.color }">
                  {{ asset.status.short_name }}
                </div>
                
              </div>
              
              <!-- View Checkpoints button -->
              <div v-if="!asset.is_link && !isUntracked && userStore.canDo('view_checkpoint')" class="asset-item-grid-checkpoints-button">
                <ActionButton :icon="getAppIcon('checkpoint-stone')" v-tooltip="$t('blocks.viewCheckpoints')" @click="viewCheckpoints(index, asset, $event)" />
              </div>
              
              <!-- Assign Asset button -->
              <div v-if="!isUntracked && canAssignAsset" class="asset-item-grid-assign-asset-button">
                <ActionButton :icon="getAppIcon('person-plus')" v-tooltip="$t('blocks.assignAsset')" @click="prepAssignAsset(index, asset, $event)" />
              </div>
              
            </div>
            
          </div>

          <!-- Editing mode -->
          <div v-else class="asset-item-grid-slide-container">
            <div class="asset-item-grid-meta-row rename-input-grid">
              <RenameInput 
                v-model="editableAssetName"
                :originalValue="asset.name || ''"
                :placeholder="$t('placeholders.assetName')"
                @confirm="confirmRename"
                @cancel="cancelRename"
              />
            </div>
          </div>
          
          <!-- File state section (absolute positioned, always visible) -->
          <div v-if="!isEditing" class="asset-item-grid-file-state-absolute">

            <div v-if="loadingAssetState" class="file-state">
              <ActionButton :isLoading="true" :icon="getAppIcon('loading')"  
                v-tooltip="$t('common.loading')" />
            </div>

            <div v-else-if="!isUntracked && userStore.canDo('pull_chunk')" class="file-state">
              <ActionButton v-if="asset.is_link" :icon="getAppIcon('square-arrow-right-up')" 
                v-tooltip="$t('blocks.visitLink')" @click="openLink()" />
              <ActionButton v-else-if="platformStore.isWeb" :icon="getAppIcon(isDownloading ? 'loading' : 'arrow-down-ramp')" 
                v-tooltip="isDownloading ? $t('blocks.downloading') : $t('common.download')" 
                :isLoading="isDownloading"
                @click="downloadAsset(index, asset, $event)" />
              <ActionButton v-else-if="asset.file_status == 'normal'" :icon="getAppIcon('circle-check-go')" :noFilter="true" 
                v-tooltip="$t('blocks.noChanges')"  />
              <ActionButton :icon="getAppIcon('circle-check')" :useAlert="true" :noFilter="true" 
                v-tooltip="$t('blocks.outdatedClickUpdate')" v-else-if="asset.file_status == 'outdated'" 
                @click="revertAsset(index, asset, $event)" />
              <ActionButton :icon="getAppIcon('plus-stone')" :useAlert="true" :noFilter="true" 
                v-tooltip="modifiedAssetTooltip"
                v-else-if="asset.file_status == 'modified'"
                @click="handleModifiedAssetCheckpointClick(index, asset, $event)" />
              <ActionButton :icon="getAppIcon('fetch')" v-tooltip="$t('blocks.fileMissingClickFetch')"
                v-else-if="asset.file_status == 'fetchable'" @click="revertAsset(index, asset, $event)" />
              <ActionButton :icon="getAppIcon('alert')" :noFilter="true" 
                v-tooltip="$t('blocks.assetMissingResync')" v-else-if="asset.file_status == 'missing'" />
            </div>

            <div v-else-if="isUntracked">
              <ActionButton v-if="canCreateFromUntracked" 
                @click="prepCreateCheckpoint(index, asset, $event)" :icon="getAppIcon('plus-stone')" :useDanger="true" 
                :noFilter="true" v-tooltip="$t('blocks.fileUntrackedClickAdd')" />
              <ActionButton v-else :icon="getAppIcon('dot-big')" :useDanger="true" :noFilter="true" 
                v-tooltip="$t('blocks.fileUntracked')" />
            </div>
          </div>
          
        </div>
      </div>
    </div>
  </div>

  <!-- List View Asset Item -->
  <div v-else 
    ref="assetItem" 
    :data-file-drop-target="!isUntracked ? '' : null"
    :id="!isUntracked ? 'drop-asset-' + asset.id : undefined"
    class="asset-item-main" 
    v-return="launchSelectedAsset" 
    v-esc="handleEscKey" 
    v-stop-propagation
    :style="itemHeightStyles" 
    :class="{
      'asset-item-selected': stage.markedItems.includes(asset.id) && !isGhost,
      'asset-item-cut': stage.cutItems.map((item) => item.id).includes(asset.id) && !isGhost,
      'asset-item-only-selected': stage.markedItems.length === 1 && stage.firstSelectedItemId === asset.id && !isGhost,
      'asset-item-last-selected': stage.lastSelectedItemId === asset.id && !isGhost,
      'asset-item-child': asset.parent_id,
      'file-drop-target-active': isHovered
    }" 
    @dblclick="launchAssetCommand()">

    <div v-if="settingsStore.showTypeIcons" class="asset-spacer" v-tooltip="assetTypeName" @click="console.log(asset)">
      <span v-if="isUntracked" class="single-action-button single-action-button-disabled">
        <img :class="[commonStore.viewMode === 'compact' ? 'large-icons' : 'small-icons', 'collection-collapsed']" :src="getAppIcon('generic')">
      </span>
      <span v-else class="single-action-button single-action-button-disabled">
        <img :class="[commonStore.viewMode === 'compact' ? 'large-icons' : 'small-icons', 'collection-collapsed']" :src="getAppIcon(asset.asset_type_icon)">
      </span>
    </div>

    <div class="main-asset-item-root">

      <div class="asset-item-container drop-zone">

        <div class="asset-item-icon-container" @click="console.log(asset)" >
          <img v-if="asset.icon" class="large-icons no-filter" :src="asset.icon">
          <img v-else-if="isUntracked" class="small-icons " :src="getAppIcon(getFileTypeIcon(asset))" @error="$event.target.src = getAppIcon('file')">
          <span v-else class="app-ext">
          </span>
        </div>

        <div class="asset-item-content selection-area">
          <div v-if="!isEditing" class="asset-item-details" :style="{ fontStyle: isUntracked ? 'italic' : 'normal' }">
            {{ assetName }}
          </div>

          <RenameInput 
            v-else
            v-model="editableAssetName"
            :originalValue="asset.name || ''"
            :placeholder="$t('placeholders.assetName')"
            @confirm="confirmRename"
            @cancel="cancelRename"
          />

          

          <div v-if="!isEditing && asset.is_link" class="weblink-pointer-container">
              <div class="weblink-pointer">
                {{ asset.pointer }}  
              </div>
          </div>

        </div>

        <template v-if="!isEditing && !asset.is_link">
          
          <!-- asset assignation -->
          <div v-if="!isUntracked && (!asset.is_resource || isCurrentUser)" class="asset-item-assignee-container">
            <ActionButton class="asset-item-assignee-button" v-if="!asset.is_link && userStore.canDo('view_checkpoint') && !statusMenuDisplayed"
              :icon="getAppIcon('checkpoint-stone')" v-tooltip="$t('blocks.viewCheckpoints')" @click="viewCheckpoints(index, asset, $event)" />

            <ActionButton class="asset-item-assignee-button" v-if="canAssignAsset && !statusMenuDisplayed && !asset.assignee_id"
              :icon="getAppIcon('person-plus')" v-tooltip="$t('blocks.assignAsset')" @click="prepAssignAsset(index, asset, $event)" />

            <div v-else-if="asset.assignee_id" @click="canManageAssetAssignment && prepAssignAsset(index, asset, $event)" v-stop-propagation
              class="asset-item-assignee">
              <span v-tooltip="userFullName" class="single-action-button">
                <div class="profile-picture" :style="{ backgroundColor: profileColor(asset.assignee_id) }">
                  <img v-if="userPhoto" class="profile-img" :src="userPhoto">
                  <img v-else class="profile-img" :src="generateAvatar(asset.assignee_id)">
                </div>
              </span>
            </div>
            
          </div>

          <div v-else-if="!isEditing" class="asset-item-assignee-container">
            <ActionButton class="asset-item-assignee-button" v-if="!asset.is_link && !isUntracked && userStore.canDo('view_checkpoint') && !statusMenuDisplayed"
              :icon="getAppIcon('checkpoint-stone')" v-tooltip="$t('blocks.viewCheckpoints')" @click="viewCheckpoints(index, asset, $event)" />

            <ActionButton class="asset-item-assignee-button" v-if="canAssignAsset && !statusMenuDisplayed && !asset.assignee_id && !isUntracked"
              :icon="getAppIcon('person-plus')" v-tooltip="$t('blocks.assignAsset')" @click="prepAssignAsset(index, asset, $event)" />
          </div>

          <!-- asset status -->
          <div v-if="!isEditing && !isUntracked && (!asset.is_resource || isCurrentUser)" class="asset-item-status-root">
            <StatusMenu @statusSelected="closeStatusMenu" v-if="statusMenuDisplayed" />

            <div :class="{ 'is-disabled': stage.operationActive }" v-else class="asset-item-status-container"
              v-stop-propagation @click="toggleDisplayStatusMenu(index, asset, $event)">
              <div class="asset-item-status" :style="{ backgroundColor: asset.status.color }">
                {{ asset.status.short_name }}
              </div>
            </div>
          </div>

          <div v-else-if="!isEditing && !isUntracked" class="asset-item-status-root">

            <div class="asset-item-status-container" v-stop-propagation>
              <div class="asset-item-status" :style="{ backgroundColor: asset.status.color, padding: 3 + 'px' }">
              </div>
            </div>
          </div>

          <!-- asset actions -->
          <div v-if="!isEditing && !isUntracked && !statusMenuDisplayed" class="asset-item-actions">
            <div v-if="loadingAssetState" class="file-state">
                <ActionButton :isLoading="true" :icon="getAppIcon('loading')" 
                  v-tooltip="$t('common.loading')" />
            </div>

            <div v-else-if="userStore.canDo('pull_chunk')" class="file-state">
              
              <ActionButton v-if="platformStore.isWeb" :icon="getAppIcon(isDownloading ? 'loading' : 'arrow-down-ramp')" 
                v-tooltip="isDownloading ? $t('blocks.downloading') : $t('common.download')" 
                :isLoading="isDownloading"
                @click="downloadAsset(index, asset, $event)" />
              <ActionButton :icon="getAppIcon('circle-check-go')" :noFilter="true" @click="handleClick(index, asset, $event)"
                v-tooltip="$t('blocks.noChanges')" v-else-if="asset.file_status == 'normal'" />
              <ActionButton :icon="getAppIcon('circle-check')" :useAlert="true" :noFilter="true" v-tooltip="$t('blocks.outdatedClickUpdate')"
                v-else-if="asset.file_status == 'outdated'" @click="revertAsset(index, asset, $event)" />
              <ActionButton :icon="getAppIcon('plus-stone')" :useAlert="true" :noFilter="true" v-tooltip="modifiedAssetTooltip"
                v-else-if="asset.file_status == 'modified'"
                @click="handleModifiedAssetCheckpointClick(index, asset, $event)" />
              <ActionButton :icon="getAppIcon('fetch')" v-tooltip="$t('blocks.fileMissingClickFetch')"
                v-else-if="asset.file_status == 'fetchable'" @click="revertAsset(index, asset, $event)" />
              <ActionButton :icon="getAppIcon('alert')" :noFilter="true" v-tooltip="$t('blocks.assetMissingResync')"
                v-else-if="asset.file_status == 'missing'" />
            </div>
          </div>
        </template>

        <div v-if="asset.is_link" class="asset-item-actions link-item-actions" >
          <ActionButton :icon="getAppIcon('square-arrow-right-up')" v-tooltip="$t('blocks.visitLink')" v-stop-propagation @click="openLink()" />
        </div>

        <div v-else-if="isUntracked" class="asset-item-actions">
          <ActionButton v-if="canCreateFromUntracked" @click="prepCreateCheckpoint(index, asset, $event)"
            :icon="getAppIcon('plus-stone')" :useDanger="true" :noFilter="true" v-tooltip="$t('blocks.fileUntrackedClickAdd')" />
          <ActionButton v-else :icon="getAppIcon('dot-big')" :useDanger="true" :noFilter="true" v-tooltip="$t('blocks.fileUntracked')" />
        </div>


      </div>

    </div>
  </div>
</template>

<script setup>
// imports
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { Browser, Events } from "@wailsio/runtime";
import emitter from '@/lib/mitt';
import { getParentPath } from '@/lib/pathlib';
import { canActOnAsset, canCreateCheckpointForItem, isCheckpointBlockedByAssignment } from '@/lib/permissions';
import { isValidWeblink } from '@/lib/pointer';
import utils from '@/services/utils';
import { generateAvatar } from '@/lib/avatar';

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import GridStatusMenu from '@/instances/desktop/menus/GridStatusMenu.vue';
import RenameInput from '@/instances/desktop/components/RenameInput.vue';
import StatusMenu from '@/instances/desktop/menus/StatusMenu.vue';

// services
import { AssetService, CheckpointService, FSService, SyncService } from "@/services";

// composables
import { useAssetThumbnail, getFileTypeIcon } from '@/composables/useAssetThumbnail';

// stores
import { useAssetStore } from '@/stores/assets';
import { useBrowserTreeStore } from '@/stores/browserTree';
import { useCollectionStore } from '@/stores/collections';
import { useCommonStore } from '@/stores/common';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useDndStore } from '@/stores/dnd';
import { useIconStore } from '@/stores/icons';
import { useMenu } from '@/stores/menu';
import { useNotificationStore } from '@/stores/notifications';
import { usePaneStore } from '@/stores/panes';
import { usePlatformStore } from '@/stores/platform';
import { useProjectStore } from '@/stores/projects';
import { useSettingsStore } from '@/stores/settings';
import { useStageStore } from '@/stores/stages';
import { useTrayStates } from '@/stores/TrayStates';
import { useUserStore } from '@/stores/users';

const assetStore = useAssetStore();
const browserTreeStore = useBrowserTreeStore();
const collectionStore = useCollectionStore();
const commonStore = useCommonStore();
const dndStore = useDndStore();
const iconStore = useIconStore();
const menu = useMenu();
const modals = useDesktopModalStore();
const notificationStore = useNotificationStore();
const panes = usePaneStore();
const platformStore = usePlatformStore();
const projectStore = useProjectStore();
const settingsStore = useSettingsStore();
const stage = useStageStore();
const trayStates = useTrayStates();
const userStore = useUserStore();
const { t } = useI18n();

// props
const props = defineProps({
  collectionId: { type: String, default: '' },
  index: Number,
  isChild: { type: Boolean, default: false },
  isGhost: { type: Boolean, default: false },
  isUntracked: { type: Boolean, default: false },
  loadingAssetState: { type: Boolean, default: false },
  asset: Object,
});

// emits
const emit = defineEmits(['toggle-edit-mode', 'expand', 'refreshData']);

// refs
const editableAssetName = ref(props.asset.name || '');
const gridStatusMenuVisible = ref(false);
const isAwaitingResponse = ref(false);
const isDownloading = ref(false);
const isEditing = ref(false);
const isExpanded = ref(false);
const statusMenuDisplayed = ref(false);
const assetItem = ref(null);

// thumbnail
const { displayThumbnail, osThumbnail } = useAssetThumbnail(
  () => props.asset,
  { enabled: () => commonStore.useGrid, includeAssetIcon: true },
);

// computed
// Returns the capitalized asset type name.
const assetTypeName = computed(() => {
  return utils.capitalizeStr(props.asset?.asset_type_name);
});

// Asset actions require both the role capability and parent collection scope.
const canCreateCheckpoint = computed(() => {
  // console.log(props.asset)
  return canCreateCheckpointForItem(props.asset)
});
const canUpdateAsset = computed(() => {
  return canActOnAsset('update_asset', props.asset);
});
const canDeleteAsset = computed(() => canActOnAsset('delete_asset', props.asset));
const canAssignAsset = computed(() => canActOnAsset('assign_asset', props.asset));
const canUnassignAsset = computed(() => canActOnAsset('unassign_asset', props.asset));
const canManageAssetAssignment = computed(() => canAssignAsset.value || canUnassignAsset.value);
const canChangeAssetStatus = computed(() => canActOnAsset('change_status', props.asset));
const canCreateFromUntracked = computed(() => canCreateCheckpoint.value);

// Checks if assignment locking allows the current user to checkpoint this asset.
const canModify = computed(() => {
  return !isCheckpointBlockedByAssignment(props.asset);
});

// Returns tooltip text for the modified asset checkpoint action.
const modifiedAssetTooltip = computed(() => {
  if (!canModify.value) {
    return t('blocks.modifiedAssignedOther');
  }
  return t('blocks.modifiedClickCheckpoint');
});

// Returns the grid styles for the asset item.
const gridStyles = computed(() => ({
  minWidth: commonStore.gridSize + 'px',
  height: commonStore.gridSize + 'px',
}));

// Checks if the current user is the assigned user.
const isCurrentUser = computed(() => {
  const user = userStore.user;
  if (!user) {
    return false;
  }
  let currentUserId = user.id;
  return props.asset.assignee_id === currentUserId;
});

// Checks if the item is hovered for drag and drop.
const isHovered = computed(() => { return dndStore.targetItemId === props.asset.id; });

// Checks if the asset is focused for selection.
const isAssetInFocus = computed(() => {
  return stage.markedItems.length === 1 && stage.firstSelectedItemId === props.asset.id && !dndStore.draggedItem;
});

// Returns the height styles for the item in list view.
const itemHeightStyles = computed(() => ({
  height: `calc(100% - ${commonStore.listItemGap}px)`,
}));

// Checks if any operations are currently active.
const operationsActive = computed(() => {
  return stage.operationActive || !!modals.activeModal || !!menu.activeMenu || isEditing.value || stage.activeStage !== 'browser';
});

// Returns the display name for the asset.
const assetName = computed(() => {
  const asset = props.asset;
  const extension = commonStore.hideExtensions ? '' : asset.name ? asset.extension : '';
  const assetName = asset.name ? asset.name : asset.extension;
  const isDirectParent = props.asset.id === asset.collection_id;
  const assetPath = asset.asset_path?.replace(/\//g, ' / ').replace(/^( \/ )?/, '');

  if (commonStore.showFullPath) {
    return assetPath + extension;
  }
  if (props.isChild) {
    if (commonStore.showChildCollections) {
      return assetName + extension;
    } else {
      return isDirectParent ? (assetName + extension) : assetPath;
    }
  } else {
    if (commonStore.viewSearchQuery) {
      return assetPath + extension;
    } else {
      return assetName + extension;
    }
  }
});

// Returns the full name of the assigned user.
const userFullName = computed(() => {
  let user = userStore.getUserData(props.asset.assignee_id);
  if (!user) {
    return t('notifications.removedUser');
  } else {
    return `${user.first_name} ${user.last_name}`;
  }
});

// Returns the profile photo URL of the assigned user.
const userPhoto = computed(() => {
  return userStore.userProfilePhoto(props.asset.assignee_id);
});

// events
Events.On('rename-item', async () => {
  if (operationsActive.value) return;
  if (isAssetInFocus.value && canUpdateAsset.value) {
    startRename();
  }
});

Events.On('edit-item', async () => {
  if (operationsActive.value) return;
  if (isAssetInFocus.value && canUpdateAsset.value) {
    modals.setModalVisibility('editAssetModal', true);
  }
});

Events.On('add-checkpoint', async () => {
  if (operationsActive.value) return;
  if (isAssetInFocus.value && canCreateCheckpoint.value) {
    prepCreateCheckpoint();
  }
});

Events.On('free-item-space', async () => {
  if (operationsActive.value) return;
  if (isAssetInFocus.value) {
    if (props.asset.type === 'asset') {
      prepFreeUpSpacePopUpModal();
    } else if (props.asset.type === 'untracked_asset') {
      prepDeleteUntrackedAssetPopUpModal();
    }
  }
});

Events.On('delete-item', async () => {
  if (operationsActive.value) return;
  if (isAssetInFocus.value && canDeleteAsset.value) {
    panes.setPaneVisibility('projectDetails', true);
    deleteAsset();
  }
});

// methods
// Cancels the current rename operation.
const cancelRename = () => {
  editableAssetName.value = props.asset.name || '';
  toggleEditMode();
};

// Shows a popup modal when user cannot modify the asset.
const canModifyPopUpModal = () => {
  trayStates.popUpModalTitle = t('common.warning');
  trayStates.popUpModalMessage = t('notifications.cannotModifyAsset');
  trayStates.popUpModalIcon = 'help';
  trayStates.popUpModalFunction = null;
  modals.setModalVisibility('popUpModal', true);
};

// Shows a popup modal when user cannot create a checkpoint.
const cannotCreateCheckpointPopUpModal = () => {
  trayStates.popUpModalTitle = t('common.warning');
  trayStates.popUpModalMessage = t('notifications.cannotCreateCheckpoint');
  trayStates.popUpModalIcon = 'help';
  trayStates.popUpModalFunction = null;
  modals.setModalVisibility('popUpModal', true);
};

// Handles the modified asset checkpoint action.
const handleModifiedAssetCheckpointClick = (index, asset, event) => {
  if (!canModify.value) {
    canModifyPopUpModal();
    return;
  }

  if (!canCreateCheckpoint.value) {
    cannotCreateCheckpointPopUpModal();
    return;
  }

  prepCreateCheckpoint(index, asset, event);
};

// Closes the grid status menu and updates asset status.
const closeGridStatusMenu = () => {
  gridStatusMenuVisible.value = false;
};

// Closes the status menu and updates asset status properties.
const closeStatusMenu = () => {
  props.asset.status = assetStore.selectedAsset.status;
  props.asset.status_id = assetStore.selectedAsset.status_id;
  props.asset.status_short_name = assetStore.selectedAsset.status_short_name;
  statusMenuDisplayed.value = false;
};

// Confirms and applies the rename operation.
const confirmRename = async () => {
  isAwaitingResponse.value = true;
  await updateAssetName();
  toggleEditMode();
};

// Deletes the asset or prepares to delete untracked asset.
const deleteAsset = async () => {
  if (props.asset.type === 'asset') {
    let assetId = assetStore.selectedAsset.id;
    AssetService.DeleteAsset(projectStore.activeProject.uri, assetId, true)
      .then(async (response) => {
        assetStore.selectedAsset = null;
        stage.markedItems = [];
        emitter.emit('refresh-browser');
      })
      .catch((error) => {
        notificationStore.errorNotification(t('notifications.assetFailedToDelete'), error);
      });
    let longMessage = t('notifications.movedToTrash', { item: assetStore.selectedAsset.name });
    notificationStore.addNotification(t('notifications.movedToTrash', { item: 'Asset' }), longMessage, "success", true);
  } else if (props.asset.type === 'untracked_asset') {
    prepDeleteUntrackedAssetPopUpModal();
  }
};

// Deletes an untracked item from the file system.
const deleteUntrackedItem = () => {
  FSService.DeleteFile(props.asset.file_path);
  projectStore.removeUntrackedAsset(props.asset.id);
  emitter.emit('refresh-browser');
  modals.disableAllModals();
};

// Downloads an asset in web mode.
const downloadAsset = async (index, asset, event) => {
  if (isDownloading.value) return;
  
  handleClick(index, asset, event);
  const assetId = asset.id;
  const fileName = `${asset.name}${asset.extension}`;
  
  isDownloading.value = true;
  
  try {
    const { CheckpointService: WebCheckpointService } = await import('@/services/adapters/checkpointservice.js');
    await WebCheckpointService.DownloadAsset(
      projectStore.activeProject.uri,
      assetId,
      null
    );
    
    notificationStore.addNotification(
      t('notifications.downloadComplete'),
      t('notifications.downloadedSuccessfully', { fileName }),
      "success",
      true
    );
  } catch (error) {
    console.error('Download error:', error);
    notificationStore.errorNotification(t('notifications.downloadFailed'), error.message || error);
  } finally {
    isDownloading.value = false;
  }
};

// Downloads a checkpoint from the server.
const downloadCheckpoint = (checkpointId) => {
  notificationStore.cancleFunction = SyncService.CancelSync;
  notificationStore.canCancel = true;
  SyncService.DownloadCheckpoint(projectStore.activeProject.uri, projectStore.getActiveProjectUrl, checkpointId)
    .then((response) => {
      emit('refreshCheckpoints');
    })
    .catch((error) => {
      console.log(error);
      notificationStore.errorNotification(t('notifications.errorDownloadingCheckpoint'), error);
    });
};

// Emits asset data updates to related components.
const emitAssetUpdates = (assetId, updates) => {
  const updateData = { itemId: assetId, updates };
  
  emitter.emit('update-root-data', updateData);
};

// Frees up disk space by deleting working files.
const freeUpSpace = async () => {
  let asset = assetStore.selectedAsset;
  let assetDir = asset.file_path.replace(/\\/g, '/');
  await FSService.DeleteFile(assetDir)
    .then((response) => {
      asset.file_status = 'fetchable';
      assetStore.fetchableAssetsPath.push(asset.asset_path + asset.extension);
      assetStore.outdatedAssetsPath = assetStore.outdatedAssetsPath.filter(assetPath => assetPath !== asset.asset_path + asset.extension);
      assetStore.modifiedAssetsPath = assetStore.modifiedAssetsPath.filter(assetPath => assetPath !== asset.asset_path + asset.extension);
      emitter.emit('refresh-browser');
    })
    .catch((error) => {
      console.error(error);
    });
  modals.disableAllModals();
};

// Returns the icon path for the given icon name.
const getAppIcon = (iconName) => {
  const icon = iconStore.getAppIcon(iconName);
  return icon;
};

// Opens the dependency graph modal for the asset.
const goToDependencies = (index, asset, event) => {
  handleClick(index, asset, event);
  assetStore.selectAsset(asset);
  modals.setModalVisibility('dependencyGraphModal', true);
};

// Handles asset click event and closes status menu.
const handleClick = (index, asset, event) => {
  closeStatusMenu();
  const id = asset.id;
};

// Handles clicks outside menus to close them.
const handleClickOutside = (event) => {
  if (statusMenuDisplayed.value) {
    statusMenuDisplayed.value = false;
  }
  if (gridStatusMenuVisible.value) {
    gridStatusMenuVisible.value = false;
  }
};

// Handles escape key to cancel operations.
const handleEscKey = () => {
  if (isEditing.value) {
    cancelRename();
  }
  if (statusMenuDisplayed.value) {
    statusMenuDisplayed.value = false;
  }
  if (gridStatusMenuVisible.value) {
    gridStatusMenuVisible.value = false;
  }
};

// Handles grid status selection and updates asset status.
const handleGridStatusSelected = () => {
  props.asset.status = assetStore.selectedAsset.status;
  props.asset.status_id = assetStore.selectedAsset.status_id;
  props.asset.status_short_name = assetStore.selectedAsset.status_short_name;
  closeGridStatusMenu();
};

// Launches the selected asset if conditions are met.
const launchSelectedAsset = () => {
  if (isEditing.value) return;
  if (isAssetInFocus.value && !modals.activeModal) {
    launchAssetCommand();
  }
};

// Launches the asset file or opens web link.
const launchAssetCommand = async () => {
  if (!userStore.canDo('pull_chunk')) {
    return;
  }
  const asset = props.asset;
  if (asset.is_link && isValidWeblink(asset.pointer)) {
    Browser.OpenURL(asset.pointer);
  } else {
    let file_path = asset.pointer ? asset.pointer : asset.file_path;
    if (await FSService.Exists(file_path)) {
      FSService.LaunchFile(file_path);
    } else {
      CheckpointService.Revert(projectStore.activeProject.uri, projectStore.getActiveProjectUrl, [asset.id])
        .then(async (response) => {
          if (!response?.restored_asset_ids?.includes(asset.id)) return;
          browserTreeStore.markAssetsAvailable(response.restored_asset_ids);
          await FSService.LaunchFile(file_path);
        })
        .catch((error) => {
          console.log(error);
          notificationStore.errorNotification(t('notifications.errorFetchingAsset'), error);
        });
    }
  }
};

// Triggers rename from the menu.
const menuRename = () => {
  if (isAssetInFocus.value && canUpdateAsset.value) {
    startRename();
  }
};

// Opens a web link in the browser.
const openLink = () => {
  const asset = props.asset;
  if (asset.is_link && isValidWeblink(asset.pointer)) {
    Browser.OpenURL(asset.pointer);
  }
};

// Opens the grid status menu.
const openGridStatusMenu = (event) => {
  if (!canChangeAssetStatus.value) return;
  const id = props.asset.id;
  const asset = props.asset;
  assetStore.selectAsset(asset);
  stage.markedAssets = [id];
  gridStatusMenuVisible.value = !gridStatusMenuVisible.value;
};

// Prepares to assign the asset to a user.
const prepAssignAsset = (index, asset, event) => {
  if (!canManageAssetAssignment.value) return;
  handleClick(index, asset, event);

  const id = asset.id;
  assetStore.selectAsset(asset);
  stage.markedAssets = [id];
  menu.showContextMenu(event, 'assignMenu', true);
};

// Prepares the create checkpoint modal.
const prepCreateCheckpoint = (index, mask, event) => {
  const asset = props.asset;
  assetStore.selectedAsset = asset;
  handleClick(index, asset, event);
  modals.setModalVisibility('createCheckpointModal', true);
};

// Prepares the delete untracked asset popup modal.
const prepDeleteUntrackedAssetPopUpModal = () => {
  trayStates.popUpModalTitle = t('common.delete');
  trayStates.popUpModalMessage = t('confirmations.deleteItemPermanently');
  trayStates.popUpModalIcon = 'trash';
  trayStates.popUpModalFunction = deleteUntrackedItem;
  modals.setModalVisibility('popUpModal', true);
};

// Prepares the free up space popup modal.
const prepFreeUpSpacePopUpModal = () => {
  trayStates.popUpModalTitle = t('notifications.freeUpAssetSpace');
  trayStates.popUpModalMessage = t('confirmations.deleteWorkingFiles', { item: 'asset' });
  trayStates.popUpModalIcon = 'broom';
  trayStates.popUpModalFunction = freeUpSpace;
  modals.setModalVisibility('popUpModal', true);
};

// Generates a color from a UUID.
const profileColor = (uuid) => {
  const parts = uuid.split('-');
  return '#' + parts[0];
};

// Reverts a asset to its last checkpoint.
const revertAsset = async (index, asset, event) => {
  handleClick(index, asset, event);
  const assetId = asset.id;

  notificationStore.cancleFunction = SyncService.CancelSync;
  notificationStore.canCancel = true;

  CheckpointService.Revert(projectStore.activeProject.uri, projectStore.getActiveProjectUrl, [assetId])
    .then(async (response) => {
      browserTreeStore.markAssetsAvailable(response?.restored_asset_ids);
    })
    .catch((error) => {
      console.log(error);
      notificationStore.errorNotification(t('notifications.errorRevertingAsset'), error);
    });
};

// Starts the rename operation.
const startRename = () => {
  toggleEditMode();
};

// Toggles the edit mode for renaming.
const toggleEditMode = (event) => {
  if (statusMenuDisplayed.value) {
    statusMenuDisplayed.value = false;
  }
  isEditing.value = !isEditing.value;
  emit('toggle-edit-mode', isEditing.value);
  
  if (isEditing.value) {
    nextTick(() => {
      const inputElement = document.querySelector('.input-short');
      if (inputElement) {
        inputElement.focus();
        inputElement.select();
      }
    });
  }
};

// Opens the status menu for changing asset status.
const toggleDisplayStatusMenu = (index, asset, event) => {
  handleClick(index, asset, event);
  if (!canChangeAssetStatus.value) return;
  assetStore.isAssetAssetStatus = true;
  assetStore.selectAsset(asset);
  statusMenuDisplayed.value = true;
};

// Triggers the rename operation if conditions are met.
const triggerRename = () => {
  if (operationsActive.value) return;
  if (isAssetInFocus.value && canUpdateAsset.value) {
    startRename();
  }
};

// Updates the asset name in the backend.
const updateAssetName = async () => {
  isAwaitingResponse.value = true;

  let assetId = props.asset.id;
  let asset = props.asset;

  if (props.asset.type === 'asset') {
    await AssetService.RenameAsset(projectStore.activeProject.uri, assetId, editableAssetName.value)
      .then((data) => {
        asset.name = editableAssetName.value;
        emitAssetUpdates(assetId, [
          { property: 'name', value: editableAssetName.value },
          { property: 'file_status', value: 'outdated' },
        ]);

        props.asset.file_status = 'outdated';
        
        isAwaitingResponse.value = false;
      })
      .catch((error) => {
        isAwaitingResponse.value = false;
        console.error('Error:', error);
      });
  } else if (props.asset.type === 'untracked_asset') {
    let oldPath = props.asset.file_path;
    let newPath = getParentPath(props.asset.file_path) + "/" + editableAssetName.value + props.asset.extension;
    let asset = projectStore.findUntrackedAsset(props.asset.id);
    await FSService.Rename(oldPath, newPath)
      .then((data) => {
        emitAssetUpdates(assetId, [
          { property: 'name', value: editableAssetName.value },
          { property: 'file_path', value: newPath }
        ]);
        
        isAwaitingResponse.value = false;
      })
      .catch((error) => {
        isAwaitingResponse.value = false;
        console.error('Error:', error);
      });
  }
};

// Opens the checkpoints view panel.
const viewCheckpoints = (index, asset, event) => {
  stage.markedItems = [asset.id];
  assetStore.selectAsset(asset);
  emitter.emit('view-checkpoints');
  panes.showDetailsPane = true;
};

// watchers
watch(() => isAssetInFocus.value, (newItems, oldItems) => {
  if (isEditing.value) {
    isEditing.value = false;
    editableAssetName.value = props.asset.name || '';
  }
  if (statusMenuDisplayed.value) {
    statusMenuDisplayed.value = false;
  }
  if (gridStatusMenuVisible.value) {
    gridStatusMenuVisible.value = false;
  }
}, { deep: true });

// lifecycle hooks
onMounted(async () => {
  emitter.on('renameAsset', menuRename);
  document.addEventListener('click', handleClickOutside);
});

onBeforeUnmount(() => {
  emitter.off('renameAsset', menuRename);
  document.removeEventListener('click', handleClickOutside);
});

</script>

<style scoped>
@import "@/assets/desktop.css";

.single-action-button-disabled {
  pointer-events: none;
}

.asset-collapsed {
  transform: rotate(-90deg);
}

.asset-expanded {
  transform: rotate(0deg);
}

.chevron-inactive {
  opacity: .2;
}

.asset-item-main {
  z-index: 100000;
  display: flex;
  gap: .2rem;
  color: var(--text);
  align-items: center;
  padding-left: .5rem;
  box-sizing: border-box;
  width: 100%;
  justify-content: flex-end;
  align-items: center;
  border-radius: 10px;
  overflow: hidden;
  padding-right: 0px;

  background-color: var(--surface-2);
  outline: var(--transparent-line);
  outline-offset: -1px;
  border-radius: var(--large-radius);
  transition: all .2s ease-out;
}

.asset-item-main:hover {
  background-color: var(--surface-3);
  border-radius: var(--small-radius);
  outline: 1px solid var(--surface-4);
}

.asset-item-main:hover  .main-asset-item-grid-thumb-container{
  border-radius: var(--tiny-radius);
}

.asset-item-selected {
  outline: var(--transparent-line);
  outline-offset: -1px;
  background-color: var(--selected-soft);
}

.asset-item-selected:hover {
  background-color: var(--selected);
}

.asset-item-cut{
  opacity: .5;
}

.asset-item-grid {
  align-items: flex-end;
  padding-left: 0px;
  padding: .5rem;
  background-color: var(--surface-2);
  outline: var(--transparent-line);
  outline-offset: -1.5px;
}

.asset-item-grid-selected {
  outline: var(--transparent-line);
  outline-offset: -1px;
  background-color: var(--selected-soft);
}

.asset-item-grid-selected:hover {
  background-color: var(--selected);
}

.asset-item-grid-cut {
  opacity: .5;
}

.asset-item-grid-last-selected {
  outline: var(--transparent-line);
  outline-offset: -1px;
  background-color: var(--selected);
}

.asset-item-grid-only-selected {
  outline: var(--transparent-line);
  outline-offset: -1px;
  background-color: var(--selected);
}

.asset-item-grid-only-selected:hover {
  outline: var(--transparent-line);
  outline-offset: -1px;
  background-color: var(--selected);
}

.main-asset-item-grid {
  position: relative;
  display: flex;
  flex-direction: column;
  height: 100%;
  width: 100%;
  overflow: hidden;
}

.main-asset-item-grid-bottom-bar {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: .2rem;
  padding-top: .5rem;
  box-sizing: border-box;
  overflow: visible;
  position: relative;
  flex-shrink: 0;
  transition: all 0.2s ease-out;
}

.asset-item-grid-bottom-bar-wrapper {
  position: relative;
  display: flex;
  flex-direction: column;
  width: 100%;
  box-sizing: border-box;
  gap: .3rem;
  height: 32px;
  transition: all 0.2s ease-out;
}

.asset-item-grid:hover .asset-item-grid-bottom-bar-wrapper:not(:has(.rename-input-grid)) {
  height: 70px;
}

.asset-item-grid-slide-container {
  display: flex;
  flex-direction: column;
  width: 80%;
  box-sizing: border-box;
  gap: .3rem;
  flex: 1;
  justify-content: space-between;
  transition: all 0.2s ease-in-out;
}

.asset-item-grid-slide-container:has(.rename-input-grid) {
  width: 100%;
}

.asset-item-grid:hover .asset-item-grid-slide-container {
  width: 100%;
  /* transition: all 0.2s ease-out; */
}

.asset-item-grid-meta-row {
  display: flex;
  align-items: center;
  gap: .3rem;
  width: 100%;
  min-height: 32px;
  height: 32px;
  flex-shrink: 0;
  box-sizing: border-box;
  overflow: hidden;
}

.asset-item-grid-actions-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: .5rem;
  width: 100%;
  min-height: 32px;
  height: 32px;
  flex-shrink: 0;
  box-sizing: border-box;
  overflow: hidden;
  width: 80%;
}

.asset-item-grid:hover .asset-item-grid-actions-row {
  display: flex;
}

.asset-item-grid-type-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.asset-item-grid-file-state-absolute {
  position: absolute;
  bottom: 0;
  right: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  height: 32px;
  z-index: 1;
}

.main-asset-item-grid-thumb-container {
  position: relative;
  display: flex;
  overflow: hidden;
  height: 100%;
  width: 100%;
  background-color: rgba(0, 0, 0, 0.2);
  border-radius: var(--normal-radius);
  align-items: center;
  justify-content: center;
  transition: all 0.2s ease-out;
  flex: 1; 
}

.main-asset-item-grid-thumb-container::after {
  content: '';
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  height: 50%;
  background: linear-gradient(to top, rgba(0, 0, 0, 0.4) 0%, transparent 100%);
  pointer-events: none;
  z-index: 0;
}

.main-asset-item-grid-meta {
  display: block;
  flex: 1;
  text-align: left;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  font-size: 14px;
  font-weight: 300;
  min-width: 0;
  margin-left: .3rem;
  line-height: 1.2;
}

.asset-item-last-selected {
  outline: 1px solid rgb(255, 255, 255);
  outline: var(--transparent-line);
  outline-offset: -1px;
  background-color: var(--selected);
}

.asset-item-only-selected {
  outline: var(--transparent-line);
  outline-offset: -1px;
  background-color: var(--selected);
}

.asset-item-only-selected:hover {
  outline: var(--transparent-line);
  outline-offset: -1px;
  background-color: var(--selected);
}

.asset-item-child {
  padding-left: 0px;
}

.main-asset-item-root {
  display: flex;
  flex-direction: column;
  gap: .2rem;
  color: var(--text);
  align-items: center;
  padding: .1rem;
  box-sizing: border-box;
  width: 100%;
  height: min-content;
  justify-content: flex-end;
  border-radius: 10px;
  overflow: hidden;
  padding-right: 0px;

}

.asset-item-container {
  display: flex;
  gap: .5rem;
  color: var(--text);
  align-items: center;
  padding: .2rem .4rem;
  box-sizing: border-box;
  width: 100%;
  height: 50px;
  justify-content: space-between;
  transition: all .3s ease-out;
}

.asset-spacer {
  position: relative;
  width: min-content;
  width: 36px;
  height: 60px;
  height: 100%;
  display: flex;
  box-sizing: border-box;
  align-items: center;
  overflow: hidden;
}

.asset-spacer-empty {
  background-color: moccasin;
}

.checkboxes {
  width: 18px;
  height: 18px;
  border-radius: 4px;
  border: 2px solid yellow;
  background: #FFF;
  padding: 10px;
}

.action-column {
  text-align: center;
  padding: 2px;
  box-sizing: border-box;
  transition: all .3s ease-in;
}

.checkbox-container {
  display: flex;
  box-sizing: border-box;
  align-items: center;
  justify-content: center;
  padding: .1rem;
  overflow: hidden;
  min-width: min-content;
  height: 100%;
}

.asset-item-preview-container {
  position: relative;
  display: flex;
  box-sizing: border-box;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  min-width: 60px;
  height: 100%;
  width: 100%;
}

.asset-item-preview-image img,
.asset-item-icon-container img {
  transition: opacity 0.2s ease-in-out;
}

.screenshot-thumb{
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.asset-item-preview-image img[src*="data:image"],
.asset-item-icon-container img[src*="data:image"] {
  animation: fadeIn 0.3s ease-in-out;
}

@keyframes fadeIn {
  from {
    opacity: 0;
  }
  to {
    opacity: 1;
  }
}

.asset-item-preview-image {
  display: flex;
  box-sizing: border-box;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  height: 100%;
  width: 100%;
}

.asset-item-icon-container {
  display: flex;
  box-sizing: border-box;
  align-items: center;
  justify-content: center;
  min-width: min-content;
  padding: .1rem;
  overflow: hidden;
  height: 100%;
}

.asset-item-icon-overlay {
  position: absolute;
  bottom: 0;
  left: 0;
  width: 32px;
  height: 32px;
  border-radius: 6px;
  min-width: unset;
  z-index: 1;
}

.overlay-icons{
  width: 100%;
  height: 100%;
}

.asset-item-content {
  gap: .4rem;
  /* flex-direction: column; */
  display: flex;
  box-sizing: border-box;
  align-items: center;
  justify-content: space-between;
  height: 100%;
  width: 100%;
  overflow: hidden;
}

.asset-item-meta-container {
  width: 100%;
  display: none;
  justify-content: flex-end;
}

.asset-item-main:hover .asset-item-meta-container {
  display: flex;
  align-items: center;
  gap: .5rem;
}

.asset-item-meta {
  display: flex;
  padding: .2rem;
  box-sizing: border-box;
  align-items: center;
  gap: .4rem;
  height: 100%;
  overflow: hidden;
  background-color: rosybrown;
  font-weight: 100;
}

.asset-item-details-old {
  padding: .2rem;
  flex-wrap: nowrap;
  overflow: hidden;
  box-sizing: border-box;
  align-items: center;
  justify-content: space-between;
  height: 100%;
  height: min-content;
  white-space: nowrap;
  text-overflow: ellipsis;
}

.asset-item-details {
  padding: .2rem;
  flex-wrap: nowrap;
  overflow: hidden;
  box-sizing: border-box;
  align-items: center;
  justify-content: space-between;
  height: 100%;
  height: min-content;
  white-space: nowrap;
  text-overflow: ellipsis;
  color: var(--text);
  justify-content: flex-end;
  /* direction: rtl; */
  text-align: left;
  font-size: 14px;
}

.weblink-pointer-container {
  display: none;
  flex-wrap: nowrap;
  text-wrap: nowrap;
  justify-content: flex-end;
  flex: 1;
  color: var(--text);
  padding: .2rem .2rem;
  border-radius: var(--tiny-radius);
  font-size: 12px;
  overflow: hidden;
  align-items: center;
  justify-content: flex-start;
  height: max-content;
  box-sizing: border-box;
  overflow: hidden;
  opacity: .5;
}

.weblink-pointer {
  width: 100%;
  overflow: hidden;
  box-sizing: border-box;
  align-items: flex-start;
  height: 100%;
  font-weight: 300;
  text-overflow: ellipsis;
}

.asset-item-main:hover .weblink-pointer-container {
  /* text-decoration: underline; */
  display: flex;
}

.asset-item-assignee-container{
  display: flex;
  gap: .5rem;
  min-width: min-content;
}

.asset-item-assignee-button {
  display: none;
}

.asset-item-main:hover .asset-item-assignee-button {
  /* text-decoration: underline; */
  display: flex;
}

.input-short {
  width: 100%;
  height: 100%;
}

.asset-item-tag {
  display: flex;
  box-sizing: border-box;
  overflow: hidden;
  padding: .1rem .4rem;
  font-size: 12px;
  background-color: black;
  border-radius: 20px;
}


.asset-item-status-container {
  display: flex;
  box-sizing: border-box;
  align-items: center;
  justify-content: center;
  width: min-content;
  padding: .4rem;
  height: 100%;
}

.asset-item-status {
  display: flex;
  border-radius: var(--normal-radius);
  box-sizing: border-box;
  align-items: center;
  justify-content: center;
  width: min-content;
  width: 60px;
  padding: .4rem .4rem;
  height: max-content;
  background-color: firebrick;
  font-size: 12px;
  text-transform: uppercase;
  font-weight: 700;
  color: black;
  transition: all 0.2s ease-out;
}

.asset-item-status:hover {
  border-radius: 6px;
}

.asset-item-actions {
  display: flex;
  box-sizing: border-box;
  align-items: center;
  justify-content: flex-end;
  width: min-content;
  min-width: max-content;
  gap: .7rem;
  height: 100%;
  min-width: var(--actions-width);
}

.link-item-actions{
  min-width: max-content;
}

.untracked-item-action {
  width: 100%;
  display: none;
  justify-content: flex-end;
}

.untracked-item-alert {
  width: 100%;
  display: flex;
  justify-content: flex-end;
}

.asset-item-main:hover .untracked-item-alert {
  display: none;
  align-items: center;
  gap: .5rem;
}

.asset-item-main:hover .untracked-item-action {
  display: flex;
  align-items: center;
  gap: .5rem;
}

.file-state {
  display: flex;
  box-sizing: border-box;
  align-items: center;
  justify-content: space-between;
  width: min-content;
  min-width: max-content;
  gap: .7rem;
  height: 100%;
}

.single-action-button{
  align-content: center;
  justify-content: center;
}

.asset-item-assignee {
  display: flex;
  box-sizing: border-box;
  align-items: center;
  justify-content: space-between;
  width: min-content;
  min-width: max-content;
  gap: .7rem;
  height: 100%;
}

.asset-item-grid-status-display {
  display: flex;
  align-items: center;
  justify-content: center;
}

.asset-item-status-grid {
  display: flex;
  border-radius: var(--normal-radius);
  box-sizing: border-box;
  align-items: center;
  justify-content: center;
  width: min-content;
  min-width: 60px;
  padding: .4rem .4rem;
  height: max-content;
  font-size: 12px;
  text-transform: uppercase;
  font-weight: 700;
  color: black;
  transition: all 0.2s ease-out;
}

.asset-item-status-grid:hover {
  border-radius: 6px;
}

.asset-item-grid-untracked-label {
  display: flex;
  align-items: center;
  justify-content: flex-start;
  flex: 1;
  padding: 0 .5rem;
}

.asset-item-grid-untracked-label span {
  font-style: italic;
  font-size: 14px;
  color: var(--text);
  opacity: 0.7;
}

.asset-item-grid-checkpoints-button,
.asset-item-grid-assign-asset-button {
  display: flex;
  align-items: center;
  justify-content: center;
}

.asset-item-grid-assignee {
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.asset-item-grid-assignee-overlay-top-right {
  position: absolute;
  bottom: 8px;
  right: 8px;
  z-index: 1;
}

.asset-item-grid-assign-button {
  opacity: 0;
  transition: opacity 0.2s ease-in-out;
}

.asset-item-grid:hover .asset-item-grid-assign-button {
  opacity: 1;
}

.profile-picture-grid {
  width: 20px;
  height: 20px;
  border-radius: 50%;
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  border: 2px solid rgba(255, 255, 255, 0.8);
}

.profile-img-grid {
  width: 100%;
  height: 100%;
  object-fit: cover;
  border-radius: 50%;
}

.profile-picture-grid-small {
  width: 16px;
  height: 16px;
  border-radius: 50%;
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  border: 1px solid rgba(255, 255, 255, 0.6);
}

.profile-img-grid-small {
  width: 100%;
  height: 100%;
  object-fit: cover;
  border-radius: 50%;
}

.rename-input-grid {
  display: flex !important;
  align-items: center;
  gap: 0.3rem;
  width: 100%;
  box-sizing: border-box;
}

.rename-input-grid .input-grid {
  box-sizing: border-box;
  flex: 1;
  min-width: 0;
  font-size: 14px;
  border-radius: 12px;
  height: 100%;
  color: var(--text);
}

@keyframes loadingRotate {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}

.loading-asset-state-icon {
  width: 20px;
  height: 20px;
  overflow: hidden;
  animation: loadingRotate .5s linear infinite;
}

</style>
