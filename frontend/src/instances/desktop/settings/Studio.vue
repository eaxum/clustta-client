<template>
  <div class="settings-component-root">
    <div class="settings-component-scroll">
    <div class="settings-component-container">

      <div class="settings-section-card">
        <div class="settings-section-card-header">
          <h2 class="settings-section-card-title">{{ $t('settings.studioInfo') }}</h2>
        </div>
        <div class="settings-section-card-content">

          <div class="settings-item" v-stop-propagation>
            <div class="settings-icon"><img class="small-icons" :src="getAppIcon('stall')"></div>
            <div class="settings-content">
              <div class="settings-header">{{ $t('settings.studioName') }}</div>
              <div class="settings-body">{{ studioInfo.name }}</div>
            </div>
          </div>

          <div class="settings-item" @click="launchUpdateStudioModal()" v-stop-propagation>
            <div class="settings-icon"><img class="small-icons" :src="getAppIcon('website')"></div>
            <div class="settings-content">
              <div class="settings-header">{{ $t('settings.ipAddressUrl') }}</div>
              <div class="settings-body">{{ studioInfo.url }}</div>
            </div>
            <div class="settings-action"><img class="small-icons" :src="getAppIcon('chevron-right')"></div>
          </div>

          <div v-if="studioInfo?.alt_url" class="settings-item" @click="launchUpdateStudioModal()" v-stop-propagation>
            <div class="settings-icon"><img class="small-icons" :src="getAppIcon('website')"></div>
            <div class="settings-content">
              <div class="settings-header">{{ $t('settings.alternateUrl') }}</div>
              <div class="settings-body">{{ studioInfo?.alt_url }}</div>
            </div>
            <div class="settings-action"><img class="small-icons" :src="getAppIcon('chevron-right')"></div>
          </div>

          <div class="settings-item" v-stop-propagation>
            <div class="settings-icon"><img class="small-icons" :src="getAppIcon('clustta')"></div>
            <div class="settings-content">
              <div class="settings-header">{{ $t('settings.clusttaServerVersion') }}</div>
              <div class="settings-body">{{ serverVersion || $t('common.loading') + '...' }}</div>
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
import { ref, computed, onMounted } from "vue";
import { useI18n } from 'vue-i18n';
import { SettingsService, StudioService } from "@/services";

// services
import utils from '@/services/utils';

// stores/state imports
import { useProjectStore } from '@/stores/projects';
import { useTrayStates } from '@/stores/TrayStates';
import { useIconStore } from '@/stores/icons';
import { useDesktopModalStore } from '@/stores/desktopModals';

// refs
const trayStates = useTrayStates();
const projectStore = useProjectStore();
const iconStore = useIconStore();
const modals = useDesktopModalStore();
const { t } = useI18n();

// vars
const autoStart = ref(trayStates.autoStart);
const clusttaVersion = ref("");
const serverVersion = ref("");

const studioInfo = computed(() => {
  return projectStore.selectedStudio
})

const fetchServerVersion = async () => {
  try {
    const studioUrl = studioInfo.value?.url;
    if (!studioUrl) {
      serverVersion.value = t('settings.noStudioConnected');
      return;
    }
    const version = await StudioService.GetServerVersion(studioUrl);
    serverVersion.value = version || t('settings.unknown');
  } catch (error) {
    serverVersion.value = t('settings.unavailable');
  }
};

const launchUpdateStudioModal = () => {
  modals.setModalVisibility('updateStudioModal', true);
};

const getAppIcon = (iconName) => {
  const icon = iconStore.getAppIcon(iconName);
  return icon
};

onMounted(async () => {
  autoStart.value = trayStates.autoStart;
  clusttaVersion.value = await utils.getRawClusttaVersion();
  await fetchServerVersion();
});

</script>


<style scoped>
@import "@/assets/desktop.css";

.settings-component-root {
  width: 100%;
  height: 100%;
  overflow: hidden;
  flex-direction: column;
  gap: 5px;
  box-sizing: border-box;
  align-items: center;
  justify-content: center;
  display: block;
  overflow-y: scroll;
  border-radius: var(--very-large-radius);
}

.settings-component-root::-webkit-scrollbar {
  width: 4px;
}

.settings-component-root::-webkit-scrollbar-thumb {
  border-radius: var(--small-radius);
  background-color: var(--light-steel);
}

.settings-component-root::-webkit-scrollbar-track {
  border-radius: var(--small-radius);
}

.settings-component-scroll {
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: center;
}

.settings-component-container {
  display: flex;
  flex-direction: column;
  box-sizing: border-box;
  height: 100%;
  gap: 1.5rem;
  width: 100%;
  padding-right: .2rem;
  border-radius: var(--large-radius);
}

.settings-item {
  color: var(--white);
  box-sizing: border-box;
  overflow: hidden;
  min-height: 50px;
  display: flex;
  padding: .5rem 1rem;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  height: max-content;
  background-color: var(--dark-steel);
  cursor: pointer;
  transition: background-color 0.2s ease;
  border-bottom: 1px solid var(--light-steel);
}

.settings-item:hover {
  background-color: #ffffff15;
}

.settings-item:active {
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
  font-size: 14px;
  font-weight: 400;
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
  justify-content: flex-end;
  overflow: hidden;
  height: 100%;
  width: max-content;
}
</style>

