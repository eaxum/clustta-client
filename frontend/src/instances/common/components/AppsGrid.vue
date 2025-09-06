<template>
  <div class="apps-container-full">
    <div class="apps-container">
      <div ref="scrollableElement" class="apps-grid">
        <div v-for="(template) in appsWithIcons" :key="template.id" class="apps-grid-item-container"
          @click="selectApp(template.name, template.icon)">
          <div class="apps-grid-item">
            <img v-if="template.icon" class="app-icons" :src="template.icon">
            <p class="app-ext-text">{{ utils.capitalizeStr(template.name) }}</p>
          </div>
        </div>
      </div>
    </div>
    <ActionButton v-if="userStore.canDo('create_template')" :icon="getAppIcon('squares-plus')" :label="'Manage templates'"
      :buttonFunction="manageTemplates" />
  </div>

</template>

<script setup>
import { ref, computed, onMounted } from 'vue';
import utils from '@/services/utils';

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue'

// states
import { useTrayStates } from '@/stores/TrayStates';
import { useIconStore } from '@/stores/icons';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useTemplateStore } from '@/stores/template';
import { useStageStore } from '@/stores/stages';
import { useUserStore } from '@/stores/users';
import { useSettingsStore } from '@/stores/settings';

// states
const trayStates = useTrayStates();
const iconStore = useIconStore();
const modals = useDesktopModalStore();
const templateStore = useTemplateStore();
const stage = useStageStore();
const userStore = useUserStore();
const settings = useSettingsStore();

const props = defineProps({
  isDesktop: {
    type: Boolean,
    default: false
  }
});

const manageTemplates = () => {
  modals.disableAllModals();
  settings.activeModalName = 'Templates';
  stage.setStageVisibility('projectSettings', true);
	settings.setModalVisibility('templates', true);
};

const getAppIcon = (iconName) => {
	const icon = iconStore.getAppIcon(iconName);
	return icon
};

const selectApp = (name, icon) => {
  templateStore.selectedTemplateName = name;
  trayStates.popUpModalIcon = icon;
  trayStates.popUpModalTitle = ('New ' + name.replace(/_/g, " ") + ' asset').toLowerCase().replace(/(^\w|\s\w)/g, match => match.toUpperCase());
  trayStates.selectedApp = document.querySelector(`.apps-grid-item-selected`);
  modals.setModalVisibility('createTaskModal', true);
};

const replaceSymbols = (name) => {
  return name.replace(/_/g, " ").toLowerCase().replace(/(^\w|\s\w)/g, match => match.toUpperCase());
}

function sortApps(apps) {
  return apps.sort((app1, app2) => {
    return app1.name.localeCompare(app2.name);
  });
}

const scrollableElement = ref(null);

const scrollElement = (direction) => {
  if (scrollableElement.value) {
    const elementWidth = scrollableElement.value.clientWidth;
    canScrollRef();
    scrollableElement.value.scrollBy({
      left: elementWidth * direction,
      behavior: 'smooth',
    });
  }
};

const canScrollRightRef = ref(true);
const canScrollLeftRef = ref(false);

const canScrollRef = () => {
  if (document.querySelector('.apps-grid')) {
    const appsContainer = document.querySelector('.apps-grid');
    const { scrollWidth, clientWidth, scrollLeft } = appsContainer;
    canScrollLeftRef.value = scrollLeft > 0;
    canScrollRightRef.value = scrollWidth > clientWidth + scrollLeft;

  }
};

const iconDirectory = "/file-icons/";
const appsWithIcons = ref(iconStore.appsWithIcons);
const apps = templateStore.getTemplates;

onMounted(async () => {

  iconStore.fetchAppIcons();
  appsWithIcons.value = sortApps(appsWithIcons.value);

});

</script>

<style scoped>
/* @import "@/assets/tray.css"; */

.apps-container-full {
  box-sizing: border-box;
  width: 100%;
  display: flex;
  position: relative;
  display: flex;
  /* justify-content: space-between; */
  align-items: center;
  flex-direction: column;
  gap: 10px;
}

.apps-container-full-desktop {
  background-color: red;
}

.apps-container {
  box-sizing: border-box;
  overflow: hidden;
  /* background-color: chartreuse; */
  width: 100%;
  border-radius: var(--normal-radius);
  padding: .2rem;
  flex: 1;
  display: flex;
  position: relative;
  display: flex;
  justify-content: space-between;
  align-items: center;
  align-items: baseline;
  overflow: hidden;
  overflow-y: scroll;
  max-height: 400px;
}

.apps-container::-webkit-scrollbar {
  width: 4px;
}

.apps-container::-webkit-scrollbar-thumb {
  border-radius: 10px;
  background-color: rgba(255, 255, 255, 0.295);
}

.apps-container::-webkit-scrollbar-track {
  border-radius: 10px;
  /* background-color: rgba(0, 0, 0, 0.295); */
}

.apps-grid {
  /* background-color: deeppink; */
  width: 100%;
  display: flex;
  gap: .5rem;
  flex-wrap: wrap;
  box-sizing: border-box;
  /* grid-template-columns: auto auto auto auto auto; */
  /* height: 100px; */

  /* display: flex; */
}


.apps-grid-item-container {
  box-sizing: border-box;
  cursor: pointer;
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  overflow: hidden;
  /* background-color: cadetblue; */
  width: 100%;
  /* gap: 10px; */
  /* padding: .1rem; */
}



.apps-grid-item {
  width: 100%;
  box-sizing: border-box;
  background-color: var(--dark-steel);
  display: flex;
  align-items: center;
  justify-content: flex-start;
  border-radius: 8px;
  gap: .5rem;
  padding: .3rem .8rem;
  outline: var(--transparent-line);
  outline-offset: -1px;

}

.apps-grid-item:hover {
  background-color: rgb(121, 121, 121);
  background-color: #ffffff15;
  outline: var(--transparent-line);
  outline-offset: -1px;
}

.apps-grid-item:active {
  background-color: rgb(70, 70, 70);
  background-color: #00000013;
}

.apps-grid-item-selected {
  box-sizing: border-box;
  background-color: rgba(0, 0, 0, 0.216);
  /* border: 1px solid var(--white); */
  outline: var(--transparent-line);
  outline-offset: -1px;

}

.apps-grid-item-selected:hover {
  background-color: rgba(0, 0, 0, 0.216);

}

.app-ext {
  box-sizing: border-box;
  background-color: rgb(6, 86, 117);
  border-radius: 8px;
  padding: 2px;
  height: 40px;
  width: 40px;
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: center;
}

.app-ext-text {
  color: var(--white);
  font-size: 16px;
  overflow: hidden;
  overflow-wrap: break-word;
}
</style>