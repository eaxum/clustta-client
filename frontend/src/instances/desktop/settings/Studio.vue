<template>
  <div class="settings-component-root">
    <div class="settings-component-container">

      <div class="settings-list">

        <div class="settings-section">
          <div class="settings-item" @click="" v-stop-propagation>
            <div class="settings-icon"><img class="small-icons" :src="getAppIcon('storefront')"></div>
            <div class="settings-content">
              <div class="settings-header">Studio Name</div>
              <div class="settings-body">{{ studioInfo.name }}</div>
            </div>
            <div class="settings-action"><img class="small-icons" :src="getAppIcon('chevron-right')"></div>
          </div>


          <div class="settings-item" @click="">
            <div class="settings-icon"><img class="small-icons" :src="getAppIcon('website')"></div>
            <div class="settings-content">
              <div class="settings-header">IP Address/URL</div>
              <div class="settings-body">{{ studioInfo.url }}</div>
            </div>
            <div class="settings-action"><img class="small-icons" :src="getAppIcon('chevron-right')"></div>
          </div>

          <div v-if="studioInfo?.alt_url" class="settings-item" @click="">
            <div class="settings-icon"><img class="small-icons" :src="getAppIcon('website')"></div>
            <div class="settings-content">
              <div class="settings-header">Alternate URL</div>
              <div class="settings-body">{{ studioInfo?.alt_url }}</div>
            </div>
            <div class="settings-action"><img class="small-icons" :src="getAppIcon('chevron-right')"></div>
          </div>

        </div>

      </div>

    </div>
  </div>
</template>

<script setup>
// imports
import { ref, computed, onMounted } from "vue";
import { SettingsService } from "@/../bindings/clustta/services/index";

// services
import utils from '@/services/utils';

// stores/state imports
import { useProjectStore } from '@/stores/projects';
import { useTrayStates } from '@/stores/TrayStates';
import { useIconStore } from '@/stores/icons';

// refs
const trayStates = useTrayStates();
const projectStore = useProjectStore();
const iconStore = useIconStore();

// vars
const autoStart = ref(trayStates.autoStart);
const clusttaVersion = ref("");

const studioInfo = computed(() => {
  return projectStore.selectedStudio
})

const getAppIcon = (iconName) => {
  const icon = iconStore.getAppIcon(iconName);
  return icon
};

onMounted(async () => {
  autoStart.value = trayStates.autoStart;
  clusttaVersion.value = await utils.getRawClusttaVersion();
});

</script>


<style scoped>
@import "@/assets/desktop.css";

.input-short {
  flex: 1;
  width: 100%;
}

.settings-component-root {
  width: 100%;
  height: 100%;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  gap: 5px;
  box-sizing: border-box;
  align-items: center;
  justify-content: center;
}

.settings-component-container {
  display: flex;
  flex-direction: column;
  box-sizing: border-box;
  height: 100%;
  overflow: hidden;
  box-sizing: border-box;
  width: 96%;
  gap: .5rem;
  align-items: center;
  color: var(--white);
  justify-content: space-between;
  border-radius: var(--large-radius);
  padding: 1rem;
}

.list-item {

  border-radius: 8px;
  box-sizing: border-box;
  cursor: pointer;
  display: flex;
  gap: 10px;
  justify-content: flex-start;
  align-items: center;
  padding: 5px 5px;
  width: 100%;
  color: #fff;
  overflow: hidden;
  text-wrap: nowrap;
}

.list-item:hover {
  background-color: rgb(121, 121, 121);
  background-color: #ffffff15;
}

.list-item:active {
  background-color: rgb(70, 70, 70);
  background-color: #00000013;
}

.settings-item {
  color: var(--white);
  box-sizing: border-box;
  overflow: hidden;
  min-height: 40px;
  display: flex;
  padding: .5rem 1rem;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  height: max-content;
  border-bottom: 1px solid rgba(192,192,192,0.5); 
  border-bottom: 1px solid color-mix(in srgb, var(--silver) 30%, transparent); 
}

.settings-item:hover {
  background-color: rgb(121, 121, 121);
  background-color: #ffffff15;
}

.settings-item:active {
  background-color: rgb(70, 70, 70);
  background-color: #00000013;
}

.settings-icon {
  box-sizing: border-box;
  display: flex;
  align-items: center;
  overflow: hidden;
  height: 100%;
  padding: .3rem;
  width: max-content;
}

.settings-content {
  box-sizing: border-box;
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  justify-content: center;
  overflow: hidden;
  height: 100%;
  padding: .4rem .2rem;
  flex: 1;
}

.settings-header {
  padding: .1rem;
}

.settings-body {
  color: var(--silver);
  padding: .1rem;
  font-size: 12px;
  opacity: .8;
}

.settings-action {
  box-sizing: border-box;
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  height: 100%;
  width: max-content;
}

.fixed-width {
  min-width: 200px;
  padding: .1rem;
  box-sizing: border-box;
}

.tray-page-content {
  position: relative;
  flex-direction: column;
  box-sizing: border-box;
  align-items: center;
}

.settings-list {
  position: relative;
  width: 96%;
  height: 100%;
  overflow-y: scroll;
  box-sizing: border-box;
  flex-direction: column;
  gap: 2rem;
}

.settings-list::-webkit-scrollbar {
  width: 4px;
}

.settings-list::-webkit-scrollbar-thumb {
  border-radius: 10px;
  background-color: rgba(255, 255, 255, 0.295);
}

.settings-list::-webkit-scrollbar-track {
  border-radius: 10px;
}

.settings-section {
  box-sizing: border-box;
  position: relative;
  background-color: var(--light-steel);
  overflow: hidden;
  height: max-content;
  border-radius: 8px;
  flex-direction: column;
  margin-bottom: 2rem;
}

.list-item {
  height: 50px;
  border-bottom: var(--transparent-line);
}
</style>

