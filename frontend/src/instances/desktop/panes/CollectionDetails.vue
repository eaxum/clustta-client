<template>
  
  <div class="general-pane-root">
    <div class="general-pane-container">

      <div v-if="collectionStore.selectedCollection?.preview" class="collection-thumb-container">
        <div class="collection-thumb">
          <img v-if="collectionStore.selectedCollection.preview" class="screenshot-thumb"
            :src="'data:image/png;base64,' + collectionStore.selectedCollection.preview">
          <img v-else class="screenshot-thumb" src="/page-states/no_image.png">
        </div>
      </div>

      <div v-if="collectionStore.selectedCollection" class="pane-parameter-section">
        <div class="action-bar" v-if="userStore.canDo('update_collection')">

          <div class="action-bar-section">
            <ActionButton :isInactive="true" :icon="getAppIcon('folder')" :label="$t('panes.collectionType')" />
            <DropDownBox :items="collectionStore.getCollectionTypesNames"
              :selectedItem="collectionStore.selectedCollection.collection_type_name" :onSelect="changeCollectionType"
              :fixedWidth="true" />
          </div>
          <div v-if="projectStore.activeProject?.has_remote" class="action-bar-section">
            <ActionButton :isInactive="true" :icon="getAppIcon('shared')" :label="$t('panes.shared')" />

            <ToggleSwitch v-tooltip="collectionStore.selectedCollection.is_shared ? $t('panes.unmarkAsShared') : $t('panes.markAsShared')"
              @click="changeIsShared" :switchValueProp="collectionStore.selectedCollection.is_shared" />
          </div>

          <div v-if="projectStore.activeProject?.has_remote" class="vertical-flex assignees-search">
            <ActionButton :isInactive="true" :icon="getAppIcon('person')" :label="$t('panes.assignees')" />
            <CollaboratorSuggestions :displayEmail="false" :placeholder="placeholder" :allItems="projectUsers"
              @tagAdded="addUser" @tagRemoved="removeUser" />
          </div>
          <div class="assignees" v-if="projectStore.activeProject?.has_remote && collaboratorsList.length">
            <AssigneeItem v-stop-propagation v-for="(collaborator, index) in collaboratorsList" :key="index"
              :assigneeId="collaborator.id" :name="collaborator.full_name" :userPhoto="collaborator.photo"
              :avatarColor="collaborator.avatarColor">
              <template #actions>
                <span v-stop-propagation class="single-action-button" @click="removeUser(collaborator)"
                  v-tooltip="$t('common.unassign')">
                  <img class="small-icons" src="/icons/remove_collaborator.svg">
                </span>
              </template>
            </AssigneeItem>
          </div>

        </div>

        <span v-if="userStore.canDo('update_collection')" class="menu-divider"></span>

        <div class="pane-parameter-detail">
          <div class="simple-text-key">
            {{ $t('panes.parent') }}
          </div>
          <div class="simple-text-value">
            {{ parentName }}
          </div>
        </div>

        <div class="pane-parameter-detail">
          <div class="simple-text-key">
            {{ $t('panes.location') }}
          </div>
            <div class="simple-text-value truncate-path" v-tooltip="collectionStore.selectedCollection.file_path">
              {{ collectionStore.selectedCollection.file_path }}
            </div>
            <div v-if="!platformStore.isWeb" class="pane-parameter-actions">
              <ActionButton :icon="getAppIcon('copy')" v-tooltip="$t('common.copyPath')" @click="copyCollectionPath('collection')"/>
              <ActionButton :icon="getAppIcon('folder-arrow-up-right')" v-tooltip="$t('common.revealInExplorer')" :buttonFunction="revealInExplorer"/>
            </div>
        </div>

        <div class="pane-parameter-detail">
          <div class="simple-text-key">
          {{ $t('panes.assets') }}
          </div>
          <div class="simple-text-value">
           {{  assetsOnDiskCount }}
          </div>
        </div>

        <div class="pane-parameter-detail">
          <div class="simple-text-key">
          {{ $t('panes.collections') }}
          </div>
          <div class="simple-text-value">
           {{  collectionsOnDiskCount }}
          </div>
        </div>

        <div class="pane-parameter-detail">
          <div class="simple-text-key">
          {{ $t('panes.size') }}
          </div>
          <div class="simple-text-value">
            {{  collectionSize }}
          </div>
        </div>


      </div>
    </div>
  </div>



</template>

<script setup>
import { useIconStore } from '@/stores/icons';
const iconStore = useIconStore();

const getAppIcon = (iconName) => {
  const icon = iconStore.getAppIcon(iconName);
  return icon
};

import { CollectionService } from "@/services";
import { FSService } from '@/services';
import { Clipboard } from '@wailsio/runtime';
import ToggleSwitch from '@/instances/common/components/ToggleSwitch.vue';
import AssigneeItem from '@/instances/common/components/AssigneeItem.vue'

// imports
import { ref, computed, onMounted, onBeforeUnmount, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import utils from '@/services/utils';
import emitter from '@/lib/mitt';

// store imports
import { useUserStore } from '@/stores/users';
import { useCollectionStore } from '@/stores/collections';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useStageStore } from '@/stores/stages';
import { useProjectStore } from '@/stores/projects';
import { usePlatformStore } from '@/stores/platform';

// components
import HeaderArea from '@/instances/common/components/HeaderArea.vue';
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import DropDownBox from '@/instances/common/components/DropDownBox.vue';
import CollaboratorSuggestions from '@/instances/common/components/CollaboratorSuggestions.vue';
import { useNotificationStore } from '@/stores/notifications';

// stores
const userStore = useUserStore();
const collectionStore = useCollectionStore();
const modals = useDesktopModalStore();
const stage = useStageStore();
const projectStore = useProjectStore();
const notificationStore = useNotificationStore();
const platformStore = usePlatformStore();

// i18n
const { t } = useI18n();

// vars
const placeholder = computed(() => t('placeholders.searchCollaborators'));

const collaboratorsList = computed(() => {
  return userStore.getProjectCollaborators
    .map(user => ({
      ...user,
      full_name: `${user.first_name} ${user.last_name}`,
      avatarColor: userStore.userProfileColor(user.id)
    }))
    .filter((user) => collectionStore.selectedCollection.assignee_ids?.includes(user.id));
});

const projectUsers = computed(() => {
  const availableUsers = userStore.getProjectCollaborators
    .map(user => ({
      ...user,
      full_name: `${user.first_name} ${user.last_name}`,
      avatarColor: userStore.userProfileColor(user.id)
    }))
    .filter((user) => !collectionStore.selectedCollection.assignee_ids?.includes(user.id));
  return availableUsers;
});

// Helper function to emit collection data updates
const emitCollectionUpdates = (collectionId, updates) => {
  const updateData = { itemId: collectionId, updates };
  
  // Emit to both Browser and VirtuaItem components
  emitter.emit('update-root-data', updateData);
  emitter.emit('update-children', updateData);
};

const revealInExplorer = async () => {
  await FSService.MakeDirs(collectionStore.selectedCollection.file_path)
  FSService.RevealInExplorer(collectionStore.selectedCollection.file_path)
};

const copyCollectionPath = async () => {
  let collection = collectionStore.selectedCollection;
  let collectionDir = collection.file_path;
  collectionDir = collectionDir.replace(/\\/g, '/');
  FSService.MakeDirs(collectionDir);
  await Clipboard.SetText(collectionDir);
  const message = t('notifications.pathCopied');
  notificationStore.addNotification(message, "", "success");
};

const removeUser = (user) => {
  const userId = user.id;
  CollectionService.Unassign(projectStore.activeProject.uri, collectionStore.selectedCollection.id, userId)
    .then((data) => {
      // Update local collection data
      collectionStore.selectedCollection.assignee_ids = collectionStore.selectedCollection.assignee_ids.filter(t => t !== userId);
      
      // Emit updates using helper function
      emitCollectionUpdates(collectionStore.selectedCollection.id, [
        { property: 'assignee_ids', value: collectionStore.selectedCollection.assignee_ids }
      ]);
      
      projectStore.refreshActiveProject();
    })
    .catch((error) => {
      notificationStore.errorNotification(t('notifications.errorRemovingUser'), error);
      console.error('Error removing user:', error);
    });
};

const addUser = (user) => {
  const userId = user.id;

  if (collectionStore.selectedCollection.assignee_ids.includes(userId)) {
    return
  }
  else {
    CollectionService.Assign(projectStore.activeProject.uri, collectionStore.selectedCollection.id, userId)
      .then((data) => {
        // Update local collection data
        collectionStore.selectedCollection.assignee_ids.push(userId);
        
        // Emit updates using helper function
        emitCollectionUpdates(collectionStore.selectedCollection.id, [
          { property: 'assignee_ids', value: collectionStore.selectedCollection.assignee_ids }
        ]);
        
        projectStore.refreshActiveProject();
      })
      .catch((error) => {
        notificationStore.errorNotification(t('notifications.errorAddingUserToProject'), error);
        console.error('Error adding user:', error);
      });
  }
};

const parentName = computed(() => {
  const parentId = collectionStore.selectedCollection.parent_id
  const parent = collectionStore.getCollections.find((item) => item.id === parentId)
  return parent ? parent.collection_path.replace(/\//g, ' / ') : t('common.none')
});

const changeCollectionType = async (collectionTypeName) => {
  stage.operationActive = true;

  let newCollectionType;
  const collectionTypes = collectionStore.getCollectionTypes;
  newCollectionType = collectionTypes.find((item) => item.name === collectionTypeName);

  const projectPath = projectStore.activeProject.uri;
  let collection = collectionStore.selectedCollection;

  await CollectionService.ChangeType(projectPath, collection.id, newCollectionType.id)
    .then((data) => {
      // Update local collection data
      collection.collection_type_name = newCollectionType.name;
      collection.collection_type_icon = newCollectionType.icon;
      collection.collection_type_id = newCollectionType.id;
      
      // Emit updates using helper function
      emitCollectionUpdates(collection.id, [
        { property: 'collection_type_name', value: newCollectionType.name },
        { property: 'collection_type_icon', value: newCollectionType.icon },
        { property: 'collection_type_id', value: newCollectionType.id }
      ]);
    })
    .catch((error) => {
      console.error('Error:', error);
    });

  stage.operationActive = false;

};
const changeIsShared = async () => {
  stage.operationActive = true;
  const projectPath = projectStore.activeProject.uri;
  let collection = collectionStore.selectedCollection;

  await CollectionService.ChangeIsShared(projectPath, collection.id, !collectionStore.selectedCollection.is_shared)
    .then((data) => {
      // Update local collection data
      collectionStore.selectedCollection.is_shared = !collectionStore.selectedCollection.is_shared;
      
      // Emit updates using helper function
      emitCollectionUpdates(collection.id, [
        { property: 'is_shared', value: collectionStore.selectedCollection.is_shared }
      ]);
      
      projectStore.refreshActiveProject();
    })
    .catch((error) => {
      console.error('Error:', error);
    });

  stage.operationActive = false;

};
// refs
const numberOfSelectedCollections = ref(0);

const collectionTypeIcon = computed(() => {
  const icon = '/types-icons/' + collectionStore.selectedCollection?.collection_type_icon + '.svg';
  if (icon) {
    return icon
  } else {
    return '/types-icons/other.svg';
  }
});


const editCollection = () => {
  modals.setModalVisibility('editCollectionModal', true);
};

const collectionSize = ref(0);
const assetsOnDiskCount = ref(0);
const collectionsOnDiskCount = ref(0);

const collectionPath = computed(() => {
  const path = collectionStore.selectedCollection?.file_path;
  return path?.replace(/\\/g, '/')
});

const getCollectionSize = async() => {
  const size = await FSService.FolderSize(collectionPath.value);
  collectionSize.value = size;
}

const getItemsCount = async() => {
  let collection = collectionStore.selectedCollection;
  assetsOnDiskCount.value = await FSService.FileCount(collection?.file_path);
  collectionsOnDiskCount.value = await FSService.FolderCount(collection?.file_path);
}

const getProjectData = async () => {
  let project = projectStore.getActiveProject;
  if (!await FSService.Exists(project.uri)) return
  getItemsCount();
  getCollectionSize();
}

watch(() => collectionStore.selectedCollection, () => {
  collectionSize.value = 0;
  assetsOnDiskCount.value = 0;
  collectionsOnDiskCount.value = 0;
  getProjectData();
});


// onMounted
onMounted(() => {
  if (!collectionStore.selectedCollection) {
    collectionStore.selectedCollection = collectionStore.getCollections[0];
  }
  getProjectData();
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
  align-items: center;
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


.menu-divider {
  height: 5px;
  margin-top: 10px;
  /* margin-bottom: 10px; */
  width: 100%
}

.assignees {
  flex-direction: column;
  width: 100%;
  align-items: flex-start;
  justify-content: flex-start;
  max-height: 100%;
  flex: 1;
  overflow: hidden;
}

.assignees::-webkit-scrollbar {
  width: 4px;
}

.assignees::-webkit-scrollbar-thumb {
  border-radius: 10px;
  background-color: var(--white);
}

.assignees::-webkit-scrollbar-track {
  border-radius: 10px;
}

.assignees-search {
  width: 100%;
  display: flex;
  /* flex-direction: column; */
  gap: .5rem;
  align-items: center;
  justify-content: flex-start;
}

.compound-input-section {
  /* background-color: royalblue; */
  /* flex: 1; */
  display: flex;
  flex-direction: row;
  align-items: center;
  gap: .4rem;
  width: 100%;
  justify-content: space-between;
  justify-content: space-around;
}


.pane-parameter-detail {
  display: flex;
  font-size: 14px;
  height: max-content;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  min-height: 30px;
  border-bottom: var(--transparent-line);
  overflow: hidden;
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

.simple-text-value-container{
  text-overflow: ellipsis;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  display: flex;
  overflow: hidden;
  text-overflow: ellipsis;
}

.action-bar {
  position: relative;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: .6rem;
  width: max-content;
  width: 100%;
  height: max-content;
  padding: .2rem;
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

.is-shared-prompt {
  padding: 1rem .5rem;
}
</style>
