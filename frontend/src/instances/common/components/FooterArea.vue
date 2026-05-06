<template>
  <div class="page-footer-area-container">
    <span v-if="showLogout" class="footer-item" @click="logUserOut" v-stop-propagation><CiLogout class="small-icons" />{{ $t('components.footerArea.logout') }}</span>

    <span v-if="showProject" class="footer-item" @click="modalStore.setModalVisibility('projectMenu', true)"
      v-stop-propagation>
      <CiBriefcase v-if="projectStore.projects.length" class="small-icons" :size="20" />
      <CiStorefront v-else class="small-icons" :size="20" />

      <div class="list-item-text" @mouseenter="trayStates.handleHover($event)"
        @mouseleave="trayStates.resetScroll($event)">
        <div style="overflow: hidden; text-overflow: ellipsis;">{{ projectStore.projects.length ?
          projectStore.getActiveProjectName : projectStore.getSelectedStudioName }}</div>
      </div>

    </span>

    <div class="footer-actions">

      <!-- <ActionButton :icon="CiArrowDownRamp" v-tooltip="'Sync Project'" @click="syncProject()"  /> -->

      <!-- <span v-if="showSync && hasRemote" class="single-action-button" :class="{ 'project-unsynced': unSynced }"
        @click="syncProject()" v-tooltip="'Sync Project'">
        <CiCloudSync :size="20" />
      </span> -->

      <!-- :class="{ 'button-disabled': processRunning }" -->
      <span v-if="showReturn" @click="goBack()" class="action-button" v-tooltip="$t('components.footerArea.goBack')"><CiArrowLeft class="small-icons" :size="20" />
      </span>
      <span v-if="showPin" @click="trayStates.togglePin(true)" class="action-button"
        v-tooltip="trayStates.pin ? $t('components.footerArea.unpin') : $t('components.footerArea.pin')">
        <CiUnpin v-if="trayStates.pin" class="small-icons" :size="20" />
        <CiPin v-else class="small-icons" :size="20" />
      </span>
    </div>


  </div>
</template>

<script setup>
import { CiArrowDownRamp, CiArrowLeft, CiBriefcase, CiCloudSync, CiLogout, CiPin, CiStorefront, CiUnpin } from '@clustta/icons-vue';

// imports
import { ref, computed } from 'vue';
import { useI18n } from 'vue-i18n';

// state imports
import { useTrayStates } from '@/stores/TrayStates';

// services
import { AuthService } from "@/services";

// store imports
import { useUserStore } from '@/stores/users';
import { useProjectStore } from '@/stores/projects';
import { useNotificationStore } from '@/stores/notifications';
import { useModalStore } from '@/stores/modals';
import { useCollectionStore } from '@/stores/collections';

// states
const trayStates = useTrayStates();

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue'

// stores
const userStore = useUserStore();
const projectStore = useProjectStore();
const notificationStore = useNotificationStore();
const modalStore = useModalStore();
const collectionStore = useCollectionStore();
const { t } = useI18n();

// refs
const searchQuery = ref('');
const props = defineProps({

  showReturn: Boolean,
  showPin: { type: Boolean, default: false },
  showLogout: Boolean,
  showSync: { type: Boolean, default: false },
  showPurge: Boolean,
  showProject: Boolean,

});

// computed properties
const unSynced = computed(() => { return projectStore.getActiveProject.is_unsynced });
const hasRemote = computed(() => {
  let project = projectStore.getActiveProject
  return project ? project.has_remote : false;
});

// methods

// const syncProject = async () => {
//   console.log(projectStore.getActiveProject.is_unsynced);
//   await sync_service.sync(
//     (progress) => {
//       notificationStore.updateProgress(progress);
//     }
//   )
//     .then(() => {
//       projectStore.activeProject.is_unsynced = false;
//       trayStates.refreshData()
//     }).catch((error) => {
//       console.error(error.message)
//       notificationStore.addNotification(
//         "Error Syncing Data",
//         error.message,
//         "error",
//         false
//       )
//     })
// };
const goBack = () => {

  trayStates.showTraySearch = false;
  trayStates.navigateBack();

};

const editParams = (itemType) => {
  modalStore.setModalVisibility('createMenu', false);
  modalStore.setModalVisibility(itemType, true);
}


const logUserOut = async () => {
  await AuthService.Logout()
    .then((data) => {
      projectStore.$reset()
      trayStates.$reset()
      trayStates.navigateToPage('Login');
    }
    )
    .catch((error) => {
      notificationStore.errorNotification(t('components.footerArea.logoutFailed'), error)
    });
}

</script>

<style scoped>
@import "@/assets/tray.css";

.project-unsynced {
  background-color: #bd2d2d;
}

.button-disabled {
  cursor: not-allowed;
}

.page-footer-area-container {
  position: relative;
  display: flex;
  align-items: center;
  flex-direction: row;
  box-sizing: border-box;
  justify-content: flex-end;
  width: 96%;
  max-height: 35px;
  height: 35px;
  /* height: 100%; */
  min-height: 35px;
  border-radius: 8px;
  /* background-color: blue; */
  overflow: hidden;
  gap: 10px;
  /* flex: 1; */
  /* background-color: transparent; */
}

.footer-actions {
  display: flex;
  /* background-color: red; */
  border-radius: 8px;
  height: 100%;

}

.project-name {

  max-height: 10px;
  background-color: darkgreen;
  overflow: hidden;
}

.footer-item {
  border-radius: 8px;
  box-sizing: border-box;
  cursor: pointer;
  display: flex;
  gap: 10px;
  justify-content: flex-start;
  align-items: center;
  padding: 5px 5px;
  width: 90%;
  color: #fff;
  /* background-color: darkmagenta; */
  overflow: hidden;
  text-wrap: nowrap;
  height: 100%;
}


.footer-item:hover {
  background-color: rgb(121, 121, 121);
  background-color: #ffffff15;
  /* background-color: darkviolet; */
}

.footer-item:active {
  background-color: rgb(70, 70, 70);
  background-color: #00000013;
}

.footer-item-text {
  overflow: hidden;
  text-overflow: ellipsis;
  background-color: darkorange;
}
</style>



