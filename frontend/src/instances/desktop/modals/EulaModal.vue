<template>

  <div class="modal-container">
    
    <HeaderArea :title="title" :icon="'clustta'" :showSearch="showSearch" />
    <div class="general-container general-container-wide">
      <pre ref="eulaContent" class="eula-content" @scroll="updateScrollPosition">
        {{ textContent }}
      </pre>
      <div class="pop-up-actions">
        <GeneralButton :label="'Decline'" :fullWidth="true" @click="logUserOut()" :colored="false" />
        <GeneralButton :label="'Accept'" :fullWidth="true" @click="acceptAgreement()" :colored="true" :isActive="isAtBottom"/>
      </div>

    </div>
  </div>
</template>

<script setup>
// imports
import { ref, computed, onMounted } from 'vue';
import utils from "@/services/utils";

import textContent from '@/../../EULA.txt?raw'
import { AuthService, SettingsService } from '@/../bindings/clustta/services/index';

// stores
import { useProjectStore } from '@/stores/projects';
import { useTrayStates } from '@/stores/TrayStates';
import { useUserStore } from '@/stores/users';

// store/state imports
import { useDesktopModalStore } from '@/stores/desktopModals';

// components
import HeaderArea from '@/instances/common/components/HeaderArea.vue';
import GeneralButton from '@/instances/common/components/GeneralButton.vue';

// vars
let title = 'Clustta End User License Agreement';
let showSearch = false;


const trayStates = useTrayStates();
const projectStore = useProjectStore();
const userStore = useUserStore();

// refs
const clusttaVersion = ref("");
const eulaContent = ref(null)
const scrollTop = ref(0)
const clientHeight = ref(0)
const scrollHeight = ref(0)
const hasReachedBottom = ref(false);

const isAtBottom = computed(() => {
    if (!eulaContent.value) return false
    
    const isScrollable = scrollHeight.value > clientHeight.value
      if (!isScrollable) return false

    const tolerance = 1
    const currentlyAtBottom = scrollTop.value + clientHeight.value >= scrollHeight.value - tolerance
    
    if (currentlyAtBottom) {
    hasReachedBottom.value = true
    }
    
    return hasReachedBottom.value
})

const updateScrollPosition = () => {
    if (eulaContent.value) {
    scrollTop.value = eulaContent.value.scrollTop
    clientHeight.value = eulaContent.value.clientHeight
    scrollHeight.value = eulaContent.value.scrollHeight
    }
}

// states/stores
const modals = useDesktopModalStore();

// methods
const acceptAgreement = async () => {
  await SettingsService.SetEulaAccepted(true);
  await SettingsService.SetCurrentVersion(clusttaVersion.value);
  modals.disableAllModals();
  setDirectories();
};

const setDirectories = async () => {
    const projectDirectory = await SettingsService.GetProjectDirectory();
    if(projectDirectory) return
	  modals.setModalVisibility('dirOnboardModal', true);
};

const logUserOut = async () => {
	await AuthService.Logout()
		.then((data) => {
            modals.disableAllModals();
			userStore.$reset()
			projectStore.$reset();
			trayStates.$reset();
		}
		)
		.catch((error) => {
			// notificationStore.errorNotification("Logout Failed", error)
		});
}

onMounted(async () => {
  clusttaVersion.value = await utils.getRawClusttaVersion();
});


</script>

<style scoped>
@import "@/assets/desktop.css";

.close {
  /* background-color: tomato; */
  opacity: .5;
  border-radius: 60px;
  position: absolute;
  top: 8px;
  right: 8px;
}

.close:hover {
  opacity: 1;
  background-color: crimson;
  transform: scale(1.02);

}

.login-button-icon {
  filter: invert(100%);
  width: 20px;
  height: 20px;
}

.login-button-text {
  font-size: 14px;
}

.pop-up-actions {
  /* background-color: red; */
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
  /* font-family: 'Courier New', monospace; */
  line-height: 1.5;

}

.eula-content::-webkit-scrollbar {
	width: 6px;
}

.eula-content::-webkit-scrollbar-thumb {
	border-radius: 10px;
	background-color: var(--silver);
}

.eula-content::-webkit-scrollbar-track {
	border-radius: 10px;
}

.consent-section{
    width: 100%;
    color: var(--white);
    display: flex;
    /* background-color: hotpink; */
    padding: .2rem;
    margin-top: 1rem;
    align-items: center;
    justify-content: flex-end;
    gap: 1rem;
}

.consent-text{
    color: var(--white);
    font-size: 14px;
}

.general-container {
  gap: .3rem;
  display: flex;
  padding: .5rem;
  /* padding-top: .5rem; */
  align-items: center;
  justify-content: center;
  overflow: hidden;
  box-sizing: border-box;
  /* background-color: red; */
}

.general-container-wide {
  /* overflow: hidden;
  width: 40vw;
  max-width: 50vw;
  max-height: 80vh; */
  min-width: 500px !important;
}

.modal-container {
  justify-content: flex-start;
  align-items: flex-start;
  max-height: 90vh;
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
  color: white;
  font-size: 16px;
  white-space: nowrap;
  flex: 1;

}
</style>