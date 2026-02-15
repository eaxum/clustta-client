<template>

  <div class="modal-container">
    
    <HeaderArea :title="$t('eula.title')" :icon="'clustta'" :showSearch="showSearch" />
    <div class="general-container general-container-wide">
      <pre ref="eulaContent" class="eula-content" @scroll="updateScrollPosition">
        {{ textContent }}
      </pre>
      <div class="pop-up-actions">
        <GeneralButton :label="$t('common.decline')" :fullWidth="true" @click="logUserOut()" :colored="false" />
        <GeneralButton :label="$t('common.accept')" :fullWidth="true" @click="acceptAgreement()" :colored="true" :isActive="isAtBottom"/>
      </div>

    </div>
  </div>
</template>

<script setup>
// imports
import { computed, onMounted, ref } from 'vue';
import utils from '@/services/utils';
import textContent from '@/data/EULA.txt?raw';

// components
import GeneralButton from '@/instances/common/components/GeneralButton.vue';
import HeaderArea from '@/instances/common/components/HeaderArea.vue';

// services
import { AuthService, SettingsService } from '@/services';

// stores
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useProjectStore } from '@/stores/projects';
import { useTrayStates } from '@/stores/TrayStates';
import { useUserStore } from '@/stores/users';

const modals = useDesktopModalStore();
const projectStore = useProjectStore();
const trayStates = useTrayStates();
const userStore = useUserStore();

// refs
const clientHeight = ref(0);
const clusttaVersion = ref('');
const eulaContent = ref(null);
const hasReachedBottom = ref(false);
const scrollHeight = ref(0);
const scrollTop = ref(0);

// constants
const showSearch = false;

// computed
// Returns whether the user has scrolled to the bottom of the EULA.
const isAtBottom = computed(() => {
  if (!eulaContent.value) return false;
  const isScrollable = scrollHeight.value > clientHeight.value;
  if (!isScrollable) return false;
  const tolerance = 1;
  const currentlyAtBottom = scrollTop.value + clientHeight.value >= scrollHeight.value - tolerance;
  if (currentlyAtBottom) {
    hasReachedBottom.value = true;
  }
  return hasReachedBottom.value;
});

// methods
// Accepts the EULA agreement and proceeds to setup.
const acceptAgreement = async () => {
  await SettingsService.SetEulaAccepted(true);
  await SettingsService.SetCurrentVersion(clusttaVersion.value);
  modals.disableAllModals();
  setDirectories();
};

// Logs the user out after declining the EULA.
const logUserOut = async () => {
  await AuthService.Logout()
    .then(() => {
      modals.disableAllModals();
      userStore.$reset();
      projectStore.$reset();
      trayStates.$reset();
    })
    .catch(() => {});
};

// Opens the directory onboarding modal if project directory is not set.
const setDirectories = async () => {
  const projectDirectory = await SettingsService.GetProjectDirectory();
  if (projectDirectory) return;
  modals.setModalVisibility('dirOnboardModal', true);
};

// Updates the scroll position for tracking EULA read progress.
const updateScrollPosition = () => {
  if (eulaContent.value) {
    scrollTop.value = eulaContent.value.scrollTop;
    clientHeight.value = eulaContent.value.clientHeight;
    scrollHeight.value = eulaContent.value.scrollHeight;
  }
};

// lifecycle hooks
onMounted(async () => {
  clusttaVersion.value = await utils.getRawClusttaVersion();
});
</script>

<style scoped>
@import "@/assets/desktop.css";

.pop-up-actions {
  justify-content: space-between;
  align-items: center;
  gap: .5rem;
}

.eula-content {
  font-size: 14px;
  display: flex;
  padding: 1rem;
  color: var(--white);
  max-height: 50vh;
  overflow: hidden;
  overflow-y: scroll;
  background-color: var(--steel);
  border-radius: var(--small-radius);
  margin: 0;
  white-space: pre-wrap;
  word-wrap: break-word;
  line-height: 1.5;
}

.eula-content::-webkit-scrollbar {
  width: 4px;
}

.eula-content::-webkit-scrollbar-thumb {
  border-radius: var(--small-radius);
  background-color: var(--light-steel);
}

.eula-content::-webkit-scrollbar-track {
  border-radius: var(--small-radius);
}

.general-container {
  gap: .3rem;
  display: flex;
  padding: .5rem;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  box-sizing: border-box;
}

.general-container-wide {
  min-width: 500px !important;
}

.modal-container {
  justify-content: flex-start;
  align-items: flex-start;
  max-height: 90vh;
}
</style>

