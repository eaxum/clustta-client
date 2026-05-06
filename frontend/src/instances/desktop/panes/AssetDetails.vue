<template>
  
  <div class="general-pane-root">
    <div class="general-pane-container">

      <div v-if="assetStore.selectedAsset?.preview" class="collection-thumb-container">
        <div class="collection-thumb">
          <img v-if="assetStore.selectedAsset.preview" class="screenshot-thumb" :src="assetStore.selectedAsset.preview">
          <img v-else class="screenshot-thumb" src="/page-states/no_image.png">
        </div>
      </div>

      <div class="pane-parameter-section">
        <div class="action-bar" v-if="userStore.canDo('update_asset')">

          <div class="action-bar-section">
            <ActionButton :isInactive="true" :icon="CiFilePlus" :label="$t('panes.type')" />
            <DropDownBox :items="assetStore.getAssetTypesNames" :selectedItem="assetStore.selectedAsset?.asset_type_name"
              :onSelect="changeAssetType" :fixedWidth="true" />
          </div>

          <div class="action-bar-section">
            <ActionButton :isInactive="true" :icon="CiClock" :label="$t('panes.status')" />
            <DropDownBox :items="projectStatuses" :selectedItem="assetStore.selectedAsset.status_short_name"
              :onSelect="setStatus" :fixedWidth="true" />
          </div>

          <div class="action-bar-section">
            <ActionButton :isInactive="true" :icon="CiShapes" :label="$t('panes.asset')" />

            <ToggleSwitch v-tooltip="!assetStore.selectedAsset.is_resource ? $t('panes.unsetAsAsset') : $t('panes.setAsAsset')"
              @click="toggleIsAsset" :switchValueProp="!assetStore.selectedAsset.is_resource" />
          </div>

        </div>

        <span v-if="userStore.canDo('update_asset')" class="menu-divider"></span>

        <div class="asset-details">
          <div class="pane-parameter-detail">
            <div class="simple-text-key">
              {{ $t('panes.parent') }}
            </div>
            <div class="simple-text-value">
              {{ assetStore.selectedAsset.collection_name }}
            </div>
          </div>

          <div v-if="!assetStore.selectedAsset.is_link" class="pane-parameter-detail">
            <div class="simple-text-key">
              {{ $t('panes.extension') }}
            </div>
            <div class="simple-text-value">
              {{ assetStore.selectedAsset.extension }}
            </div>
          </div>

          <div class="pane-parameter-detail">
            <div class="simple-text-key">
              {{ $t('panes.assignedTo') }}
            </div>
            <ActionButton v-if="assetStore.selectedAsset.assignee_id" :iconAfter="true" :label="userFullName" v-tooltip="$t('panes.seeAllAssets')" :buttonFunction="showAllAssets"/>
            <div v-else class="simple-text-value">
              {{ userFullName }}
            </div>
          </div>

          <div v-if="!assetStore.selectedAsset.is_link" class="pane-parameter-detail">
            <div class="simple-text-key">
              {{ $t('panes.checkpointComment') }}
            </div>
            <div class="simple-text-value">
              {{ lastCheckpoint.comment }}
            </div>
          </div>

          <div v-if="lastCheckpoint?.comment !== $t('panes.noCheckpoints') && !assetStore.selectedAsset.is_link"
            class="pane-parameter-detail">
            <div class="simple-text-key">
              {{ $t('panes.checkpointDate') }}
            </div>
            <div class="simple-text-value">
              {{ formatMtime(lastCheckpoint.created_at) }}
            </div>
          </div>
          

          <div v-if="!assetStore.selectedAsset.is_link" class="pane-parameter-detail">
            <div class="simple-text-key">
              {{ $t('panes.location') }}
            </div>
              <div class="simple-text-value truncate-path" v-tooltip="assetStore.selectedAsset.file_path">
                {{ assetStore.selectedAsset.file_path }}
              </div>
              <div v-if="!platformStore.isWeb" class="pane-parameter-actions">
                <ActionButton :icon="CiCopy" v-tooltip="$t('common.copyPath')" @click="copyAssetPath('asset')"/>
                <ActionButton :icon="CiFolderArrowUpRight" v-tooltip="$t('common.revealInExplorer')" :buttonFunction="revealInExplorer"/>
              </div>
          </div>

          <div v-if="!assetStore.selectedAsset.is_link" class="pane-parameter-detail">
            <div class="simple-text-key">
              {{ $t('panes.fileState') }}
            </div>
            <div class="simple-text-value">
              {{ assetStore.selectedAsset.file_status }}
            </div>
          </div>

          <div v-if="!assetStore.selectedAsset.is_link" class="pane-parameter-detail">
          <div class="simple-text-key">
          {{ $t('panes.size') }}
          </div>
          <div class="simple-text-value">
            {{  assetSize }}
          </div>
        </div>

          <div class="pane-parameter-detail tag-parameter">
            <div class="simple-text-key tag-key">
              {{ $t('panes.tags') }}
            </div>
            

            <div class="tag-section">
              <div class="asset-tag-list">
                <Chip v-for="tag in assetTags" :key="tag.id" :label="tag.name" :readonly="!userStore.canDo('update_asset')" :onRemove="() => removeTag(tag)" />
                <Chip v-if="userStore.canDo('update_asset') && !showTagInput" :icon="CiPlusCircle" :label="$t('panes.addTag')" :isStatic="false" :readonly="true" @click="openTagInput" />
                <span v-if="userStore.canDo('update_asset') && showTagInput" class="tag-input-chip">
                  <input ref="tagInput" v-model="tagInputValue" class="tag-chip-input" type="text" :placeholder="$t('panes.addTag')" :size="Math.max(tagInputValue.length, 6)" @keydown.enter.prevent="addTag" @keydown.escape.prevent="closeTagInput" />
                  <ActionButton :icon="CiCheck" v-tooltip="$t('common.confirm')" @click="addTag" />
                  <ActionButton :icon="CiClose" v-tooltip="$t('common.close')" @click="closeTagInput" />
                </span>
              </div>
            </div>

          </div>

        </div>

      </div>

      

    </div>
  </div>



</template>

<script setup>
// imports
import { ref, computed, onMounted, watch, onBeforeUnmount, nextTick } from 'vue';
import { useI18n } from 'vue-i18n';
import { FSService, TagService } from '@/services';
import { Clipboard } from '@wailsio/runtime';
import utils from '@/services/utils';
import emitter from '@/lib/mitt';
import { CiCheck, CiClock, CiClose, CiCopy, CiFilePlus, CiFolderArrowUpRight, CiPlusCircle, CiShapes } from '@clustta/icons-vue';
import { resolveIcon } from '@/lib/icon-map';

// store imports
import { useProjectStore } from '@/stores/projects';
import { useUserStore } from '@/stores/users';
import { useStageStore } from '@/stores/stages';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useAssetStore } from '@/stores/assets';
import { useStatusStore } from '@/stores/status';
import { useIconStore } from '@/stores/icons';
import { useNotificationStore } from '@/stores/notifications';
import { useCommonStore } from '@/stores/common';
import { usePlatformStore } from '@/stores/platform';
import { useTagStore } from '@/stores/tags';

// services
import { AssetService, CheckpointService } from "@/services";

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import Chip from '@/instances/common/components/Chip.vue';
import DropDownBox from '@/instances/common/components/DropDownBox.vue';
import ToggleSwitch from '@/instances/common/components/ToggleSwitch.vue';
import AssigneeItem from '@/instances/common/components/AssigneeItem.vue';

// stores
const assetStore = useAssetStore();
const stage = useStageStore();
const userStore = useUserStore();
const modals = useDesktopModalStore();
const statusStore = useStatusStore();
const projectStore = useProjectStore();
const iconStore = useIconStore();
const notificationStore = useNotificationStore();
const commonStore = useCommonStore();
const platformStore = usePlatformStore();
const tagStore = useTagStore();
const { t } = useI18n();

// refs
const assetTags = ref([]);
const multiStatusChange = ref(false);
const latestCheckpoint = ref(null);
const numberOfSelectedAssets = ref(0);
const showTagInput = ref(false);
const tagInputValue = ref('');
const tagInput = ref(null);

// computed properties
const projectStatuses = computed(() => {
  const allStatuses = statusStore.statuses;
  if (!userStore.canDo('set_done_asset')) {
    const limitedStatus = ['done', 'retake']
    return allStatuses.filter((item) => !limitedStatus.includes(item.short_name))
  } else {
    return allStatuses.map((status) => status.short_name.toUpperCase())
  }
});

const singleAsset = computed(() => {
  numberOfSelectedAssets.value = stage.markedAssets.length;
  const isSingleAsset = stage.markedAssets.length <= 1 && assetStore.selectedAsset;
  return isSingleAsset
});

const selectedAssetName = computed(() => {
  if (assetStore.selectedAsset) {
    return singleAsset.value ? assetStore.selectedAsset.name : t('panes.multipleAssetsSelected')
  }
});

const selectedAssetIcon = computed(() => {
  if (assetStore.selectedAsset) {
    return singleAsset.value ? assetStore.selectedAsset.icon : '/icons/categories.svg'
  }
});

// methods
const addTag = async () => {
  const name = tagInputValue.value.trim();
  if (!name || !assetStore.selectedAsset) return;
  try {
    const updatedNames = await tagStore.addTagToAsset(assetStore.selectedAsset.id, name);
    assetStore.selectedAsset.tags = updatedNames;
    await loadAssetTags();
    tagInputValue.value = '';
    showTagInput.value = false;
  } catch (error) {
    notificationStore.addNotification(t('notifications.failedToAddTag'), 'error');
  }
};

// Closes the tag input and resets the value.
const closeTagInput = () => {
  showTagInput.value = false;
  tagInputValue.value = '';
};

const getAppIcon = (iconName) => {
  const icon = iconStore.resolveIcon(iconName);
  return icon
};

const emitAssetUpdates = (assetId, updates) => {
  const updateData = { itemId: assetId, updates };
  
  emitter.emit('update-root-data', updateData);
  emitter.emit('update-children', updateData);
};

const copyAssetPath = async (pathType) => {
  let asset = assetStore.selectedAsset;
  let assetPath = asset.file_path;
  assetPath = assetPath.replace(/\\/g, '/');
  let assetDir = assetPath.split('/').slice(0, -1).join('/');
  let resourcesFolder = assetDir + '/resources';
  let outputPath = assetDir + '/output';
  if (pathType === 'resources') {
    assetPath = resourcesFolder;
  } else if (pathType === 'output') {
    assetPath = outputPath;
  }
  await Clipboard.SetText(assetPath);
  const message = t('notifications.pathCopied');
  notificationStore.addNotification(message, "", "success");
};

const showAllAssets = () => {
  commonStore.activeWorkspace = 'Default'
  commonStore.resetFilters();
  commonStore.navigatorMode = false;
  if (assetStore.selectedAsset?.assignee_id) {
    const assignee = userStore.getUserData(assetStore.selectedAsset.assignee_id);
    if (assignee) {
      const assigneeFilter = {
        name: `${assignee.first_name} ${assignee.last_name}`,
        id: assignee.id,
        type: 'assignation',
        avatarColor: userStore.userProfileColor(assignee.id)
      };
      commonStore.assetFilters.push(assigneeFilter);
    }
  }
  commonStore.onlyAssets = true;
  emitter.emit('refresh-browser');
};

// Opens the tag input field and focuses it.
const openTagInput = () => {
  showTagInput.value = true;
  nextTick(() => {
    tagInput.value?.focus();
  });
};

// Removes a tag from the selected asset.
const removeTag = async (tag) => {
  if (!assetStore.selectedAsset) return;
  try {
    const updatedNames = await tagStore.removeTagFromAsset(assetStore.selectedAsset.id, tag.id);
    assetStore.selectedAsset.tags = updatedNames;
    await loadAssetTags();
  } catch (error) {
    notificationStore.addNotification(t('notifications.failedToRemoveTag'), 'error');
  }
};

const revealInExplorer = async () => {
  const assetId = assetStore.selectedAsset.id;
  if(assetStore.selectedAsset.file_status == "rebuildable"){
    await CheckpointService.Revert(projectStore.activeProject.uri, projectStore.getActiveProjectUrl, [assetId])
    .then( async (response) => {
      assetStore.rebuildableAssetsPath = assetStore.rebuildableAssetsPath.filter(assetPath => assetPath !== asset.asset_path)
      assetStore.outdatedAssetsPath = assetStore.outdatedAssetsPath.filter(assetPath => assetPath !== asset.asset_path);
      emitter.emit('get-project-data')
    })
    .catch((error) => {
      notificationStore.errorNotification(t('notifications.errorDownloadingAsset'), error);
      console.error(error);
    });

  } 
  AssetService.RevealAsset(projectStore.activeProject.uri, assetStore.selectedAsset.id);
};

const toggleIsAsset = async () => {
  stage.operationActive = true;
  const projectPath = projectStore.activeProject.uri;
  let isAsset = assetStore.selectedAsset.is_resource;
  let asset = assetStore.selectedAsset;
    
  await AssetService.ToggleIsAsset(projectPath, asset.id,  isAsset)
    .then((data) => {

      assetStore.selectedAsset.is_resource = !isAsset;
      emitAssetUpdates(asset.id, [
        { property: 'is_resource', value: !isAsset }
      ]);
      
    })
    .catch((error) => {
      console.error('Error:', error);
    });
    
    stage.operationActive = false;
};

const changeAssetType = async (assetTypeName) => {
  stage.operationActive = true;

  let newAssetType;
  const assetTypes = assetStore.getAssetTypes;
  newAssetType = assetTypes.find((item) => item.name === assetTypeName);

  const projectPath = projectStore.activeProject.uri;
  let asset = assetStore.selectedAsset;

  await AssetService.UpdateAsset(projectPath, asset.id, asset.name, newAssetType.id, asset.is_resource, '', asset.tags)
    .then((data) => {
      asset.asset_type_name = newAssetType.name;
      asset.asset_type_icon = newAssetType.icon;
      asset.asset_type_id = newAssetType.id;
      
      emitAssetUpdates(asset.id, [
        { property: 'asset_type_name', value: newAssetType.name },
        { property: 'asset_type_icon', value: newAssetType.icon },
        { property: 'asset_type_id', value: newAssetType.id }
      ]);
      
    })
    .catch((error) => {
      console.error('Error:', error);
    });

  stage.operationActive = false;

};

const setStatus = async (statusName) => {
  stage.operationActive = true;
  const projectPath = projectStore.activeProject.uri;
  const status = statusStore.statuses.find(item => item.short_name === statusName.toLowerCase());
  let asset = assetStore.selectedAsset;
  
  await AssetService.ChangeStatus(projectPath, [asset.id], status.id)
    .then((data) => {
      asset.status_short_name = status.short_name;
      asset.status = status;
      
      emitAssetUpdates(asset.id, [
        { property: 'status_short_name', value: status.short_name },
        { property: 'status', value: status }
      ]);
      
    })
    .catch((error) => {
      console.error('Error:', error);
    });
    
  multiStatusChange.value = false;
  stage.operationActive = false;
};

const userFullName = computed(() => {
  let assigneeId = assetStore.selectedAsset.assignee_id
  let user = userStore.getUserData(assigneeId);
  if (assigneeId && user) {
    let fullname = `${user.first_name} ${user.last_name}`;
    return fullname
  } else if(!assigneeId) {
    return t('panes.nobody')
  } else {
    return t('notifications.removedUser')
  }
});

const lastCheckpoint = computed(() => {
  let asset = assetStore.selectedAsset;
  if (asset.is_link) return { comment: '', created_at: asset.created_at };
  
  if (latestCheckpoint.value) {
    return { 
      comment: latestCheckpoint.value.comment, 
      created_at: latestCheckpoint.value.created_at 
    };
  }
  
  return { comment: t('panes.noCheckpoints'), created_at: asset.created_at };
});

const loadLatestCheckpoint = async () => {
  latestCheckpoint.value = null;
  
  if (!assetStore.selectedAsset || assetStore.selectedAsset.is_link) {
    return;
  }

  try {
    const checkpoint = await CheckpointService.GetLatestCheckpoint(
      projectStore.activeProject.uri,
      assetStore.selectedAsset.id
    );
    latestCheckpoint.value = checkpoint;
  } catch (error) {
    console.log('No checkpoint found or error:', error);
    latestCheckpoint.value = null;
  }
};

const formatMtime = (mtime) => {
  const date = new Date(mtime);
  const day = date.getDate();
  const monthNames = ["Jan", "Feb", "Mar", "Apr", "May", "Jun",
    "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"];
  const month = monthNames[date.getMonth()];
  const year = date.getFullYear();

  return `${day} ${month} ${year}`;
};

const editAsset = () => {
  modals.setModalVisibility('editAssetModal', true);
};

// Loads full tag objects for the selected asset.
const loadAssetTags = async () => {
  if (!assetStore.selectedAsset) {
    assetTags.value = [];
    return;
  }
  try {
    assetTags.value = await TagService.GetAssetTags(projectStore.activeProject.uri, assetStore.selectedAsset.id);
  } catch (error) {
    assetTags.value = [];
  }
};

const assetSize = ref(0);

const assetPath = computed(() => {
  const path = assetStore.selectedAsset?.file_path;
  return path?.replace(/\\/g, '/')
});

const getAssetSize = async() => {
  const size = await FSService.FileStat(assetPath.value);
  assetSize.value = size.formattedSize;
}

const getProjectData = async () => {
  if (!await FSService.Exists(assetPath.value)){
    assetSize.value = t('panes.notOnDisk')
    return
  }
  getAssetSize();
  loadAssetTags();
  loadLatestCheckpoint();
}

watch(() => assetStore.selectedAsset, () => {
  assetSize.value = 0;
  getProjectData();
  loadAssetTags();
  loadLatestCheckpoint();
});


// onMounted
onMounted(() => {
  stage.markedAssets = [];
  
  getProjectData();
  loadLatestCheckpoint();
	emitter.on('get-project-data', getProjectData);
});

onBeforeUnmount(() => {
	emitter.off('get-project-data', getProjectData);
});


</script>


<style scoped>
@import "@/assets/desktop.css";

.pane-parameter-section {
  overflow: hidden;
  overflow-y: auto;
  padding-right: .5rem;
  height: 100%;
  width: 96%;
}

.pane-parameter-section::-webkit-scrollbar {
  width: 4px;
}

.pane-parameter-section::-webkit-scrollbar-thumb {
  border-radius: var(--small-radius);
  background-color: var(--light-steel);
}

.pane-parameter-section::-webkit-scrollbar-track {
  border-radius: var(--small-radius);
  margin: 1rem 0;
}

.asset-details{
  padding-right: 0;
}
.pane-parameter-detail {
  display: flex;
  font-size: 14px;
  height: max-content;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  min-height: 30px;
  height: min-content;
  border-bottom: var(--transparent-line);
}

.menu-divider {
  height: 5px;
  margin-top: 10px;
  width: 100%
}

.simple-text-key {
  white-space: nowrap;
}

.truncate-path {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  min-width: 0;
  flex: 1;
}

.status-box-container {
  width: 50%;
  height: 100%;
  height: max-content;
  /* height: 60px; */
}

.input-short {
  flex: 1;
  width: 100%;
}

.listbox-short {

  flex: 1;
  width: 130px;
}

.input-label {

  font-family: Inter, sans-serif;
  color: white;
  font-size: 14px;
  white-space: nowrap;
  flex: 1;

}

.pop-up-prompt {
  gap: 10px;
  /* background-color: bisque; */
  align-items: center;
  justify-content: center;
  /* height: 400px; */
  /* background-color: darkseagreen; */
}

.action-bar {
  position: relative;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: .6rem;
  width: max-content;
  width: 100%;
  /* justify-content: space-around; */
  height: max-content;
  padding: .2rem;
  /* background-color: black; */
  /* background-color: tomato; */
  align-items: flex-start;
  box-sizing: border-box;
}

.action-bar-section {
  display: flex;
  align-items: center;
  gap: .5rem;
  justify-content: space-between;
  width: 100%;
}

.multi-assign {
  /* background-color: tomato; */
  width: 100%;
  display: flex;
  align-items: flex-start;
  flex-direction: column;
}

.tag-section {
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: .5rem;
  overflow: hidden;
  box-sizing: border-box;
  align-items: center;
  justify-content: center;
  height: min-content;
}

.tag-input-chip {
  display: inline-flex;
  align-items: center;
  background-color: var(--steel);
  border-radius: var(--large-radius);
  font-size: 0.875rem;
  color: var(--white);
  overflow: hidden;
  height: min-content;
}

.tag-chip-input {
  background: transparent;
  border: none;
  outline: none;
  color: var(--white);
  font-size: 0.875rem;
  font-weight: 300;
  padding: 0.25rem 0.5rem;
  min-width: 3rem;
  max-width: 10rem;
  width: auto;
}

.tag-parameter {
  display: flex;
  justify-content: flex-start;
  align-items: flex-start;
  border-bottom: 0px;
}

.tag-key {
  display: flex;
  align-items: center;
  height: 30px;
}

.asset-tag-list {
  display: flex;
  flex-wrap: wrap;
  gap: .4rem;
  padding: .2rem;
  overflow: hidden;
  box-sizing: border-box;
  width: 100%;
}

</style>





