<template>
  <div class="modal-container" v-esc="closeModal">
    <HeaderArea :title="title" :icon="'edit'" :showSearch="showSearch" />
    <div class="general-container">

      <div class="collaborator-details">
        <div class="colaborator-picture" :style="{ backgroundColor: avatarColor }">
          <img class="colaborator-img" :src="user.photo ? user.photo : '/icons/default_profile_picture.svg'">
        </div>

        <div class="collaborator-data">
          <div class="input-label">{{ collaboratorName }}</div>
          <div class="input-label">{{ user.email }}</div>
        </div>
      </div>

      <ListBox :items="userStore.getRolesNames" :isUnique="dummy" :onSelect="selectRole"
        :selectedItem="newCollaboratorRole" :placeHolder="'None'" />
      <div class="pop-up-actions">
        <GeneralButton :label="'Cancel'" :buttonFunction="closeModal" :colored="false" />
        <GeneralButton :label="'Update'" :buttonFunction="updateRole" :isActive="roleChanged"
          :loading="isAwaitingResponse" />
      </div>
    </div>
  </div>
</template>

<script setup>

// imports
import { ref, onMounted, computed } from 'vue';

// services
import { ProjectService } from "@/../bindings/clustta/services";

// state imports
import { useTrayStates } from '@/stores/TrayStates';

// store imports
import { useUserStore } from '@/stores/users';
import { useNotificationStore } from '@/stores/notifications';
import { useProjectStore } from '@/stores/projects';
import { useDesktopModalStore } from '@/stores/desktopModals';

// components
import HeaderArea from '@/instances/common/components/HeaderArea.vue';
import ListBox from '@/instances/common/components/ListBox.vue';
import GeneralButton from '@/instances/common/components/GeneralButton.vue';

// states
const trayStates = useTrayStates();

// stores
const userStore = useUserStore();

const notificationStore = useNotificationStore();
const projectStore = useProjectStore();
const modals = useDesktopModalStore();

// vars
let title = 'Edit Collaborator';
let showSearch = false;

// refs
const collaboratorName = ref('');
const collaboratorRole = ref(userStore.getRolesNames[0]);
const newCollaboratorRole = ref(userStore.getRolesNames[0]);
const isAwaitingResponse = ref(false);
const avatarColor = ref('');

// computed properties
const roleChanged = computed(() => { return newCollaboratorRole.value !== collaboratorRole.value });
const user = computed(() => { return userStore.selectedUser });

// methods
const closeModal = () => {
  modals.setModalVisibility("editCollaboratorModal", false);
};

const dummy = (item) => {
  return false;
};

const selectRole = (role) => {
  newCollaboratorRole.value = role;
};

const updateRole = async () => {
  isAwaitingResponse.value = true;
  await ProjectService.ChangeRole(projectStore.activeProject.uri, userStore.selectedUser.id, newCollaboratorRole.value)
    .then(async (data) => {
      isAwaitingResponse.value = false;
      closeModal();
      notificationStore.addNotification("User updated Successfully.", "", "success")
      await trayStates.refreshData();
    })
    .catch((error) => {
      isAwaitingResponse.value = false;
      notificationStore.errorNotification("Error updating User", error);
    })
}

const handleEnterKey = (event) => {

};

// onMounted
onMounted(() => {
  if (userStore.selectedUser) {
    const user = userStore.selectedUser;
    collaboratorRole.value = user.role.name;
    newCollaboratorRole.value = user.role.name;
    collaboratorName.value = user.first_name + ' ' + user.last_name;
    avatarColor.value = userStore.userProfileColor(user.id);
  };
})

</script>

<style scoped>
@import "@/assets/desktop.css";

.general-container {
  gap: .6rem;
}

.colaborator-picture {
  background-color: red;
  height: 48px;
  min-width: 48px;
  overflow: hidden;
  display: flex;
  align-items: center;
  border-radius: 48px;
  aspect-ratio: 1/1;
  /* padding: 5px; */
}

.colaborator-img {
  width: 100%;
  height: 100%;
}

.collaborator-details {
  width: 100%;
  display: flex;
  box-sizing: border-box;
  gap: .5rem;
  padding: 0px .5rem;
  /* background-color: tomato; */
}

.collaborator-data {
  width: 100%;
  display: flex;
  flex-direction: column;
  padding: .5rem;
  box-sizing: border-box;
  gap: .2rem;
  /* font-weight: 100; */
  /* background-color: green; */
}

.add-category {
  display: flex;
  gap: .5rem;
  flex-direction: row;
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
  color: var(--white);
  font-size: 14px;
  white-space: nowrap;
  flex: 1;
}

.category-area {
  box-sizing: border-box;
  display: flex;
  flex-direction: column;
  gap: 1rem;
  color: var(--white);
  width: 98%;
}

.category-list {
  box-sizing: border-box;
  display: flex;
  padding: .5rem;
  align-items: center;
  flex-direction: column;
  gap: .2rem;
  background-color: rgba(0, 0, 0, 0.144);
  height: 290px;
  overflow: hidden;
  overflow-y: scroll;
  width: 100%;
  border-radius: 10px;
}

.category-list::-webkit-scrollbar {
  width: 4px;
}

.category-list::-webkit-scrollbar-thumb {
  border-radius: 10px;
  background-color: rgba(255, 255, 255, 0.295);
}

.category-list::-webkit-scrollbar-track {
  border-radius: 10px;
}

.category-item {
  color: var(--white);
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: .5rem;
  width: 100%;
  border-bottom: 1px solid rgba(255, 255, 255, 0.096);
  height: max-content;
  padding: .2rem;
}

.category-item-actions {
  display: flex;
}
</style>

