<template>
  <div class="entity-item-wrapper" @click="console.log(task)">
    <!-- thumbnail -->
    <div v-if="commonStore.showThumbs" class="entity-item-preview-container">
      <div class="entity-item-preview-image">
        <img v-if="task.preview" class="screenshot-thumb" :src="task.preview">
      </div>
    </div>
    <!-- name and app icon -->
    <div class="entity-item-data">
      <div class="entity-item-content">
        <div v-if="task.task_type_icon && !commonStore.showThumbs" class="entity-item-type-container">
          <img  class="large-icons" :src="getAppIcon(task.task_type_icon)">
        </div>
        <div class="entity-item-info">
          <div class="entity-item-text">
            {{ utils.capitalizeStr(task.name) }}
          </div>
        </div>
        <div class="entity-item-icon-container">
          <img v-if="task.icon" class="large-icons no-filter" :src="task.icon">
        </div>
      </div>
      <!-- assignee -->
      <div v-if="task.assignee_id" class="entity-item-assignee">
        <div class="single-action-button" v-tooltip="userFullName">
          <div class="profile-picture" :style="{ backgroundColor: profileColor(task.assignee_id) }">
              <img v-if="userPhoto" class="profile-img" :src="userPhoto">
              <img v-else class="profile-img" :src="getAppIcon('person')">
          </div>
        </div>
      </div>
      <div v-else class="entity-item-assignee">
        <div class="single-action-button">
          <div class="entity-item-unassigned">
            <!-- Not Assigned -->
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
// imports
import { computed, ref, onMounted } from 'vue';;
import utils from '@/services/utils';

// states/store imports
import { useTrayStates } from '@/stores/TrayStates';
import { useMenu } from '@/stores/menu';
import { usePaneStore } from '@/stores/panes';
import { useStageStore } from '@/stores/stages';
import { useUserStore } from '@/stores/users';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useNotificationStore } from '@/stores/notifications';
import { useCommonStore } from '@/stores/common';
import { useIconStore } from '@/stores/icons';

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue'
import { useAssetStore } from '@/stores/assets';

// states/stores
const userStore = useUserStore();
const trayStates = useTrayStates();
const menu = useMenu();
const panes = usePaneStore();
const stage = useStageStore();
const assetStore = useAssetStore();
const commonStore = useCommonStore();
const iconStore = useIconStore();

const getAppIcon = (iconName) => {
  const icon = iconStore.getAppIcon(iconName);
  return icon
};

// props
const props = defineProps({
  task: Object,
  index: Number,
});

// refs

const userFullName = computed(() => {
  let user = userStore.getUserData(props.task.assignee_id);
  if (!user) {
    return 'No User'
  } else {
    return `${user.first_name} ${user.last_name}`;
  }
});

const userPhoto = computed(() => {
  const userPhoto = userStore.userProfilePhoto(props.task.assignee_id);
  return userPhoto
});

const profileColor = (uuid) => {
  const parts = uuid.split('-');
  return '#' + parts[0];
};



</script>

<style scoped>
@import "@/assets/desktop.css";

.entity-item-wrapper {
  display: flex;
  gap: .2rem;
  color: var(--white);
  align-items: center;
  box-sizing: border-box;
  height: 80px;
  width: 100%;
  width: 250px;
  overflow: hidden;
  border-radius: var(--small-radius);
  background-color: var(--dark-steel);
  /* background-color: darkorange; */
  padding: .2rem;
}


.entity-item-wrapper:hover {
  outline: var(--transparent-line);
  outline-offset: -1.5px;
}

.entity-item-root {
  display: flex;
  flex-direction: column;
  gap: .5rem;
  color: var(--white);
  align-items: center;
  box-sizing: border-box;
  width: 100%;
  /* flex: 1; */
  overflow: hidden;
  /* max-width: 360px; */
  height: min-content;
  justify-content: space-between;
  /* transition: all .3s ease-out; */
}

/* .entity-item-root:hover>*:last-child {
  opacity: 0;
  transition: opacity 0.2s ease-in-out;
} */

.entity-item-root-minimized {
  width: 100%;
}

.entity-item-container {
  display: flex;
  flex-direction: column;
  gap: .5rem;
  color: var(--white);
  align-items: center;
  padding: .4rem;
  box-sizing: border-box;
  width: 100%;
  height: min-content;
  justify-content: space-between;
  transition: all .3s ease-out;
  /* background-color: violet; */
  border-radius: var(--normal-radius);
}

.entity-item-container-minimized {
  justify-content: center;
}

.entity-item-container-selected {
  outline: 1.5px solid rgb(255, 255, 255);
  outline-offset: -1.5px;
}

.entity-item-container-selected:hover {
  outline: 1.5px solid rgb(255, 255, 255);
  outline-offset: -1.5px;
}

.entity-item-preview-container {

  display: flex;
  box-sizing: border-box;
  align-items: center;
  justify-content: center;
  padding: .1rem;
  overflow: hidden;
  min-width: 60px;
  /* background-color: firebrick; */
  /* width: 100%; */
  height: 100%;
  aspect-ratio: 4 / 3;
}

.entity-item-preview-image {
  display: flex;
  position: relative;
  box-sizing: border-box;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  height: 100%;
  width: 100%;
  /* aspect-ratio: 4 / 3; */
  background-color: var(--black-steel);
  border-radius: 5px;
  pointer-events: none;
}

.entity-item-type-container {
  display: flex;
  box-sizing: border-box;
  align-items: center;
  justify-content: center;
  min-width: min-content;
  padding: .1rem;
  overflow: hidden;
  height: 100%;
  /* background-color: firebrick; */
}

.entity-item-icon-container {
  display: flex;
  box-sizing: border-box;
  align-items: center;
  justify-content: center;
  min-width: min-content;
  padding: .1rem;
  overflow: hidden;
  /* height: 100%; */
  /* background-color: firebrick; */
}

.entity-item-content {
  display: flex;
  box-sizing: border-box;
  align-items: center;
  height: 100%;
  /* background-color: goldenrod; */
  justify-content: space-between;
  width: 100%;
  gap: .3rem;
  padding: 0px .3rem;
  overflow: hidden;
}

.entity-item-data {
  display: flex;
  flex-direction: column;
  box-sizing: border-box;
  align-items: flex-start;
  height: 100%;
  /* background-color: goldenrod; */
  justify-content: space-between;
  justify-content: flex-start;
  /* align-items: flex-start; */
  width: 100%;
  gap: .3rem;
  padding: 0px .3rem;
  overflow: hidden;
}

.entity-item-info {
  display: flex;
  box-sizing: border-box;
  align-items: center;
  height: 100%;
  /* background-color: black; */
  justify-content: space-between;
  width: 100%;
  gap: .3rem;
  overflow: hidden;
}

.entity-item-text {
  padding: .2rem;
  flex-wrap: nowrap;
  overflow: hidden;
  box-sizing: border-box;
  align-items: center;
  height: 100%;
  height: min-content;
  white-space: nowrap;
  text-overflow: ellipsis;
  /* background-color: forestgreen; */
  font-size: 14px;
  /* font-weight: 200; */
}

.entity-item-meta {
  display: flex;
  padding: .2rem;
  box-sizing: border-box;
  align-items: center;
  gap: .4rem;
  width: 100%;
  height: 100%;
  overflow: hidden;
  /* background-color: rosybrown; */
}

.entity-item-type-indicator {
  color: rgb(219, 219, 219);
  color: var(--white);
  background-color: rgba(0, 0, 0, 0.216);
  padding: .3rem;
  border-radius: 5px;
  font-size: 12px;
}

.entity-item-tag {
  display: flex;
  box-sizing: border-box;
  overflow: hidden;
  padding: .1rem .4rem;
  font-size: 12px;
  background-color: black;
  border-radius: 20px;
}


.entity-item-assignee {
  display: flex;
  box-sizing: border-box;
  align-items: center;
  justify-content: flex-end;
  width: min-content;
  width: 100%;
  min-width: max-content;
  gap: .7rem;
  /* padding: 0px .3rem; */
  height: 100%;
  /* background-color: tomato; */
  /* flex: 1; */
}

.entity-item-assignee-name {
  font-size: 14px;
}

.entity-item-unassigned {
  font-size: 14px;
  opacity: .5;
  /* font-weight: 100; */
  font-style: italic;
}

.entity-item-actions {
  display: flex;
  box-sizing: border-box;
  align-items: center;
  justify-content: space-between;
  width: min-content;
  min-width: max-content;
  gap: .7rem;
  height: 100%;
  background-color: darkcyan;
  /* flex: 1; */
}
</style>

