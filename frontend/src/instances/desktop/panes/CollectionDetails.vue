<template>
  <div class="general-pane-header">
    <HeaderArea v-if="collectionStore.selectedEntity" :title="utils.capitalizeStr(collectionStore.selectedEntity?.name)"
      :icon="entityTypeIcon" />

    <ActionButton v-if="userStore.canDo('update_entity')" :icon="getAppIcon('parameters')" v-tooltip="'Rename Entity'"
      :buttonFunction="editEntity" />
  </div>

  <div class="general-pane-root">
    <div class="general-pane-container">

      <div v-if="collectionStore.selectedEntity?.preview" class="entity-thumb-container">
        <div class="entity-thumb">
          <img v-if="collectionStore.selectedEntity.preview" class="screenshot-thumb"
            :src="'data:image/png;base64,' + collectionStore.selectedEntity.preview">
          <img v-else class="screenshot-thumb" src="/page-states/no_image.png">
        </div>
      </div>

      <div v-if="collectionStore.selectedEntity" class="pane-parameter-section">
        <div class="action-bar" v-if="userStore.canDo('update_entity')">

          <div class="action-bar-section">
            <ActionButton :isInactive="true" :icon="getAppIcon('folder')" :label="'Collection type'" />
            <DropDownBox :items="collectionStore.getCollectionTypesNames"
              :selectedItem="collectionStore.selectedEntity.entity_type_name" :onSelect="changeEntityType"
              :fixedWidth="true" />
          </div>
          <div class="action-bar-section">
            <ActionButton :isInactive="true" :icon="getAppIcon('bookmark')" :label="'Library'" />

            <ToggleSwitch v-tooltip="collectionStore.selectedEntity.is_library ? 'Unmark as library' : 'Mark as a library'"
              @click="changeIsLibrary" :switchValueProp="collectionStore.selectedEntity.is_library" />
          </div>

          <div class="vertical-flex assignees-search">
            <ActionButton :isInactive="true" :icon="getAppIcon('two-persons')" :label="'Assignees'" />
            <CollaboratorSuggestions :displayEmail="false" :placeholder="placeholder" :allItems="projectUsers"
              @tagAdded="addUser" @tagRemoved="removeUser" />
          </div>
          <div class="assignees" v-if="collaboratorsList.length">
            <AssigneeItem v-stop-propagation v-for="(collaborator, index) in collaboratorsList" :key="index"
              :assigneeId="collaborator.id" :name="collaborator.full_name" :userPhoto="collaborator.photo"
              :avatarColor="collaborator.avatarColor">
              <template #actions>
                <span v-stop-propagation class="single-action-button" @click="removeUser(collaborator)"
                  v-tooltip="'Unassign'">
                  <img class="small-icons" src="/icons/remove_collaborator.svg">
                </span>
              </template>
            </AssigneeItem>
          </div>

        </div>

        <span v-if="userStore.canDo('update_entity')" class="menu-divider"></span>

        <div class="pane-parameter-detail">
          <div class="simple-text-key">
            Parent
          </div>
          <div class="simple-text-value">
            {{ parentName }}
          </div>
        </div>

        <div class="pane-parameter-detail">
          <div class="simple-text-key">
            Location
          </div>
            <div class="simple-text-value" >
              {{ collectionStore.selectedEntity.file_path }}
            </div>
            <div class="pane-parameter-actions">
              <ActionButton :icon="getAppIcon('copy')" v-tooltip="'Copy Path'" @click="copyEntityPath('entity')"/>
              <ActionButton :icon="getAppIcon('folder-arrow-up-right')" v-tooltip="'Reveal in Explorer'" :buttonFunction="revealInExplorer"/>
            </div>
        </div>

        <div class="pane-parameter-detail">
          <div class="simple-text-key">
          Assets
          </div>
          <div class="simple-text-value">
           {{  assetsOnDiskCount }}
          </div>
        </div>

        <div class="pane-parameter-detail">
          <div class="simple-text-key">
          Collections
          </div>
          <div class="simple-text-value">
           {{  collectionsOnDiskCount }}
          </div>
        </div>

        <div class="pane-parameter-detail">
          <div class="simple-text-key">
          Size 
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

import { ClipboardService, EntityService } from "@/../bindings/clustta/services";
import { FSService } from '@/../bindings/clustta/services/index';
import ToggleSwitch from '@/instances/common/components/ToggleSwitch.vue';
import AssigneeItem from '@/instances/common/components/AssigneeItem.vue'

// imports
import { ref, computed, onMounted, onBeforeUnmount, watch } from 'vue';
import utils from '@/services/utils';
import emitter from '@/lib/mitt';

// store imports
import { useUserStore } from '@/stores/users';
import { useCollectionStore } from '@/stores/collections';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useStageStore } from '@/stores/stages';
import { useProjectStore } from '@/stores/projects';

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

// vars
let placeholder = 'Search collaborators';

const collaboratorsList = computed(() => {
  return userStore.getProjectCollaborators
    .map(user => ({
      ...user,
      full_name: `${user.first_name} ${user.last_name}`,
      avatarColor: userStore.userProfileColor(user.id)
    }))
    .filter((user) => collectionStore.selectedEntity.assignee_ids?.includes(user.id));
});

const projectUsers = computed(() => {
  const availableUsers = userStore.getProjectCollaborators
    .map(user => ({
      ...user,
      full_name: `${user.first_name} ${user.last_name}`,
      avatarColor: userStore.userProfileColor(user.id)
    }))
    .filter((user) => !collectionStore.selectedEntity.assignee_ids?.includes(user.id));
  return availableUsers;
});

// Helper function to emit entity data updates
const emitEntityUpdates = (entityId, updates) => {
  const updateData = { itemId: entityId, updates };
  
  // Emit to both Browser and VirtuaItem components
  emitter.emit('update-root-data', updateData);
  emitter.emit('update-children', updateData);
};

const revealInExplorer = async () => {
  await FSService.MakeDirs(collectionStore.selectedEntity.file_path)
  FSService.RevealInExplorer(collectionStore.selectedEntity.file_path)
};

const copyEntityPath = async () => {
  let entity = collectionStore.selectedEntity;
  let entityDir = entity.file_path;
  entityDir = entityDir.replace(/\\/g, '/');
  await ClipboardService.WriteText(entityDir);
  const message = 'Path copied to clipboard';
  notificationStore.addNotification(message, "", "success");
};

const removeUser = (user) => {
  const userId = user.id;
  EntityService.Unassign(projectStore.activeProject.uri, collectionStore.selectedEntity.id, userId)
    .then((data) => {
      // Update local entity data
      collectionStore.selectedEntity.assignee_ids = collectionStore.selectedEntity.assignee_ids.filter(t => t !== userId);
      
      // Emit updates using helper function
      emitEntityUpdates(collectionStore.selectedEntity.id, [
        { property: 'assignee_ids', value: collectionStore.selectedEntity.assignee_ids }
      ]);
      
      projectStore.refreshActiveProject();
    })
    .catch((error) => {
      notificationStore.errorNotification('Error removing user', error);
      console.error('Error removing user:', error);
    });
};

const addUser = (user) => {
  const userId = user.id;

  if (collectionStore.selectedEntity.assignee_ids.includes(userId)) {
    return
  }
  else {
    EntityService.Assign(projectStore.activeProject.uri, collectionStore.selectedEntity.id, userId)
      .then((data) => {
        // Update local entity data
        collectionStore.selectedEntity.assignee_ids.push(userId);
        
        // Emit updates using helper function
        emitEntityUpdates(collectionStore.selectedEntity.id, [
          { property: 'assignee_ids', value: collectionStore.selectedEntity.assignee_ids }
        ]);
        
        projectStore.refreshActiveProject();
      })
      .catch((error) => {
        notificationStore.errorNotification('Error adding user', error);
        console.error('Error adding user:', error);
      });
  }
};

const parentName = computed(() => {
  const parentId = collectionStore.selectedEntity.parent_id
  const parent = collectionStore.getCollections.find((item) => item.id === parentId)
  return parent ? parent.entity_path.replace(/\//g, ' / ') : 'None'
});

const changeEntityType = async (entityTypeName) => {
  stage.operationActive = true;

  let newEntityType;
  const entityTypes = collectionStore.getCollectionTypes;
  newEntityType = entityTypes.find((item) => item.name === entityTypeName);

  const projectPath = projectStore.activeProject.uri;
  let entity = collectionStore.selectedEntity;

  await EntityService.ChangeType(projectPath, entity.id, newEntityType.id)
    .then((data) => {
      // Update local entity data
      entity.entity_type_name = newEntityType.name;
      entity.entity_type_icon = newEntityType.icon;
      entity.entity_type_id = newEntityType.id;
      
      // Emit updates using helper function
      emitEntityUpdates(entity.id, [
        { property: 'entity_type_name', value: newEntityType.name },
        { property: 'entity_type_icon', value: newEntityType.icon },
        { property: 'entity_type_id', value: newEntityType.id }
      ]);
    })
    .catch((error) => {
      console.error('Error:', error);
    });

  stage.operationActive = false;

};
const changeIsLibrary = async () => {
  stage.operationActive = true;
  const projectPath = projectStore.activeProject.uri;
  let entity = collectionStore.selectedEntity;

  await EntityService.ChangeIsLibrary(projectPath, entity.id, !collectionStore.selectedEntity.is_library)
    .then((data) => {
      // Update local entity data
      collectionStore.selectedEntity.is_library = !collectionStore.selectedEntity.is_library;
      
      // Emit updates using helper function
      emitEntityUpdates(entity.id, [
        { property: 'is_library', value: collectionStore.selectedEntity.is_library }
      ]);
      
      projectStore.refreshActiveProject();
    })
    .catch((error) => {
      console.error('Error:', error);
    });

  stage.operationActive = false;

};
// refs
const numberOfSelectedEntities = ref(0);

const entityTypeIcon = computed(() => {
  const icon = '/types-icons/' + collectionStore.selectedEntity?.entity_type_icon + '.svg';
  if (icon) {
    return icon
  } else {
    return '/types-icons/other.svg';
  }
});


const editEntity = () => {
  modals.setModalVisibility('editCollectionModal', true);
};

const collectionSize = ref(0);
const assetsOnDiskCount = ref(0);
const collectionsOnDiskCount = ref(0);

const collectionPath = computed(() => {
  const path = collectionStore.selectedEntity?.file_path;
  return path?.replace(/\\/g, '/')
});

const getCollectionSize = async() => {
  const size = await FSService.FolderSize(collectionPath.value);
  collectionSize.value = size;
}

const getItemsCount = async() => {
  let entity = collectionStore.selectedEntity;
  assetsOnDiskCount.value = await FSService.FileCount(entity?.file_path);
  collectionsOnDiskCount.value = await FSService.FolderCount(entity?.file_path);
}

const getProjectData = async () => {
  let project = projectStore.getActiveProject;
  if (!await FSService.Exists(project.uri)) return
  getItemsCount();
  getCollectionSize();
}

watch(() => collectionStore.selectedEntity, () => {
  collectionSize.value = 0;
  assetsOnDiskCount.value = 0;
  collectionsOnDiskCount.value = 0;
  getProjectData();
});


// onMounted
onMounted(() => {
  if (!collectionStore.selectedEntity) {
    collectionStore.selectedEntity = collectionStore.getCollections[0];
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

.menu-divider {
  height: 5px;
  margin-top: 10px;
  /* margin-bottom: 10px; */
  width: 100%
}

.assignees {
  /* display: flex; */
  flex-direction: column;
  width: 100%;
  align-items: flex-start;
  justify-content: flex-start;
  max-height: 100%;
  max-height: 200px;
  flex: 1;
  overflow: hidden;
  overflow-y: scroll;
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

.pane-parameter-section {
  flex: 1;
  height: 200px;
}

.pane-parameter-detail {
  display: flex;
  font-size: 14px;
  height: max-content;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  height: 30px;
  border-bottom: var(--transparent-line);
  overflow: hidden;
}

.simple-text-key {
  white-space: nowrap;
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


.simple-text-value {

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

.is-library-prompt {
  padding: 1rem .5rem;
}
</style>