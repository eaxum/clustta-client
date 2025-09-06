<template>
  <div ref="taskItem" class="task-item-main" v-esc="handleEscKey" v-stop-propagation>
    <div class="task-spacer">
      <ProfilePhoto :assigneeId="collaborator.id" :userPhoto="collaborator.photo" />
    </div>

    <div class="main-task-item-root">

      <div class="task-item-container drop-zone">

        <div class="task-item-content selection-area">
          <div class="task-item-details">
            {{ userFullName }}
          </div>
        </div>

        <!-- task status -->
        <div v-if="!isCurrentUser" class="task-item-status-root ">
          <DropDownBox :selectedItem="collaborator.role_name" :items="collaboratorRoles" :onSelect="selectRole" />
        </div>

        <!-- task actions -->
        <div class="task-item-actions">
          <div class="file-state">
            <ActionButton :isDisabled="isCurrentUser" :icon="getAppIcon('trash')" @click="deleteCollaborator(collaborator.id)" v-tooltip="'Remove'" />
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
// imports
import { computed, ref, onMounted, onBeforeUnmount } from 'vue';

// states/store imports
import { useIconStore } from '@/stores/icons';
import { useNotificationStore } from '@/stores/notifications';
import { useProjectStore } from '@/stores/projects';
import { useStudioStore } from '@/stores/studio';
import { useUserStore } from '@/stores/users';

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import DropDownBox from '@/instances/common/components/DropDownBox.vue';
import ProfilePhoto from '@/instances/common/components/ProfilePhoto.vue'
import { StudioService } from "@/../bindings/clustta/services";

// states/stores
const iconStore = useIconStore();
const notificationStore = useNotificationStore();
const projectStore = useProjectStore();
const studioStore = useStudioStore();
const userStore = useUserStore();

// props
const props = defineProps({
  collaborator: Object,
  index: Number,
  entityId: { type: String, default: '' },
  isChild: { type: Boolean, default: false },
});

// refs
const taskItem = ref(null);
const collaboratorRoles = ref([
  'Admin',
  'User'
]);

const isCurrentUser = computed(() => {
  return userStore.user?.id === props.collaborator.id
})

const selectRole = (role) => {

  let userId = props.collaborator.id
  let selectedUser = studioStore.studioUsers.find((user) => user.id === userId);

  StudioService.ChangeCollaboratorRole( props.collaborator.id, projectStore.selectedStudio.id, role,)
  .then(() => {
      selectedUser.role_name = role;
    }).catch((error) => {
      notificationStore.errorNotification("Error Changing Role", error.response.data)
      console.log(error.response.data)
    })
}

const deleteCollaborator = (collaboratorId) => {

  let userId = props.collaborator.id

  StudioService.RemoveCollaborator( collaboratorId, projectStore.selectedStudio.id)
  .then(() => {
    studioStore.studioUsers = studioStore.studioUsers.filter((user) => user.id !== userId )
  }).catch((error) => {
    notificationStore.errorNotification("Error Deleting User", error.response.data)
    console.log(error)
  })
};

const userFullName = computed(() => {
  return `${props.collaborator.first_name} ${props.collaborator.last_name}`
})

// methods

const handleEnterKey = () => {
};

const handleEscKey = () => {
};

const getAppIcon = (iconName) => {
  const icon = iconStore.getAppIcon(iconName);
  return icon
};

const handleClickOutside = (event) => {
};

onMounted(() => {
  document.addEventListener('click', handleClickOutside);
});

onBeforeUnmount(() => {
  document.removeEventListener('click', handleClickOutside);
});

</script>

<style scoped>
@import "@/assets/desktop.css";

.single-action-button-disabled {
  pointer-events: none;
}

.task-item-main {
  z-index: 100000;
  display: flex;
  gap: .2rem;
  color: var(--white);
  align-items: center;
  padding-left: .5rem;
  box-sizing: border-box;
  width: 100%;
  height: min-content;
  justify-content: flex-end;
  align-items: flex-start;
  background-color: var(--dark-steel);
  border-radius: 10px;
  overflow: hidden;
  padding-right: 0px;

  outline: var(--transparent-line);
  outline-offset: -1px;
}

.task-item-main:hover {
  outline: var(--transparent-line);
  outline: 1px solid rgb(255, 255, 255);
  outline-offset: -1.5px;
}

.task-item-selected {
  outline: 1px solid rgb(255, 255, 255);
  outline: var(--transparent-line);
  outline-offset: -1px;
  background-color: var(--blue-steel);
}

.task-item-last-selected {
  outline: 1px solid rgb(255, 255, 255);
  outline: var(--transparent-line);
  outline-offset: -1px;
  background-color: var(--solid-blue-steel);
}

.task-item-only-selected {
  outline: 1px solid rgb(255, 255, 255);
  outline: var(--transparent-line);
  outline-offset: -1px;
  background-color: var(--solid-blue-steel);
}

.task-item-selected:hover {
  outline: 1px solid rgb(255, 255, 255);
  outline-offset: -1px;
}

.task-item-child {
  padding-left: 0px;
}

.main-task-item-root {
  display: flex;
  flex-direction: column;
  gap: .2rem;
  color: var(--white);
  align-items: center;
  padding: .3rem;
  box-sizing: border-box;
  width: 100%;
  height: min-content;
  justify-content: flex-end;
  border-radius: 10px;
  overflow: hidden;
  padding-right: 0px;
}

.task-item-container {
  display: flex;
  gap: .5rem;
  color: var(--white);
  align-items: center;
  padding: .2rem .4rem;
  box-sizing: border-box;
  width: 100%;
  height: 50px;
  justify-content: space-between;
  transition: all .3s ease-out;
}

.task-spacer {
  position: relative;
  width: min-content;
  width: 36px;
  height: 60px;
  display: flex;
  box-sizing: border-box;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  /* background-color: forestgreen; */
}

.task-item-preview-container {
  display: flex;
  box-sizing: border-box;
  align-items: center;
  justify-content: center;
  padding: .1rem;
  overflow: hidden;
  min-width: 60px;
  height: 100%;
  aspect-ratio: 16 / 9;
}

.task-item-preview-image {
  display: flex;
  box-sizing: border-box;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  height: 100%;
  aspect-ratio: 16 / 9;
  background-color: var(--light-steel);
  border-radius: 5px;
}

.task-item-icon-container {
  display: flex;
  box-sizing: border-box;
  align-items: center;
  justify-content: center;
  min-width: min-content;
  padding: .1rem;
  overflow: hidden;
  height: 100%;
}

.task-item-content {
  gap: .4rem;
  display: flex;
  box-sizing: border-box;
  align-items: center;
  justify-content: space-between;
  height: 100%;
  width: 100%;
  overflow: hidden;
}

.task-item-details {
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
  /* font-weight: 200; */
}

.task-item-status-container {
  display: flex;
  box-sizing: border-box;
  align-items: center;
  justify-content: center;
  width: min-content;
  padding: .4rem;
  height: 100%;
}

.task-item-status-root {
  /* Empty rule kept because it's used in the template */
}

.task-item-status {
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

.task-item-status:hover {
  border-radius: 10px;
  transform: scale(1.03);
}

.task-item-actions {
  display: flex;
  box-sizing: border-box;
  align-items: center;
  justify-content: space-between;
  justify-content: flex-end;
  width: min-content;
  min-width: max-content;
  gap: .7rem;
  height: 100%;
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
</style>