<template>
  <div ref="modalContainer" class="modal-container">

    <div class="general-pane-header">
      <HeaderArea :title="title" :icon="getAppIcon('two-drives')" :showSearch="false" />
    </div>

    <div v-if="!isStudioCreated" class="general-container">

      <div class="input-section">
        <div class="input-label-row">
          <label class="input-label">Studio Name</label>
        </div>
        <div class="horizontal-flex">
          <input v-model="studioName" class="input-short" type="text" placeholder="Studio Name"  v-focus />
        </div>
      </div>

      <div class="input-section">
        <div class="input-label-row">
          <label class="input-label">Studio URL</label>
        </div>
        <div class="horizontal-flex">
          <input v-model="studioUrl" class="input-short" type="text" placeholder="Studio URL"  />
        </div>
      </div>

      <div class="pop-up-actions">
        <GeneralButton :label="'Back'" :fullWidth="true" :buttonFunction="goBack" :colored="false" />
        <GeneralButton :label="'Create'" :fullWidth="true" @click="createStudio" :isActive="isValueChanged"
          :loading="isAwaitingResponse" />
      </div>
    </div>

    <div v-else class="general-container">

      <div class="success-message">
        <p>Congratulations, your studio <strong>{{ createdStudio?.name }}</strong> has been created.</p>
        <p>Copy the data below and paste it in your '.env' file to complete the setup on your server.</p>
      </div>

      <div class="secret-key-container">
        <div class="secret-key-header">
          <div class="studio-info-label">Studio Secret Key</div>
          <div class="secret-key-actions">
            <ActionButton 
              :icon="getAppIcon(showSecretKey ? 'eye-cancel' : 'eye')" 
              :buttonFunction="toggleSecretKey"
              v-tooltip="showSecretKey ? 'Hide' : 'Show'"
            />
            <ActionButton 
              :icon="getAppIcon('copy')" 
              :buttonFunction="copySecretKey"
              :label="secretKeyCopied ? 'Copied!' : 'Copy'"
              :showLabel="true"
            />
          </div>
        </div>
        <div class="secret-key-value-container">
          <input 
            v-if="showSecretKey"
            :value="createdStudio?.secret_key"
            readonly
            class="secret-key-input"
            @click="selectSecretKey"
          />
          <input 
            v-else
            :value="'•'.repeat(createdStudio?.secret_key?.length || 0)"
            readonly
            class="secret-key-input"
            disabled
          />
        </div>
      </div>

      <div class="env-file-container">
        <div class="env-file-header">
          <div class="studio-info-label">Environment File</div>
          <div class="env-file-actions">
            <ActionButton 
              :icon="getAppIcon(showEnvFile ? 'eye-cancel' : 'eye')" 
              :buttonFunction="toggleEnvFile"
              v-tooltip="showEnvFile ? 'Hide' : 'Show'"
            />
            <ActionButton 
              :icon="getAppIcon('copy')" 
              :buttonFunction="copyEnvFile"
              :label="envCopied ? 'Copied!' : 'Copy'"
              :showLabel="true"
            />
          </div>
        </div>

        <div class="env-file-textarea-container" :class="{'env-file-textarea-container-closed' : !showEnvFile }">
          <textarea 
            v-if="showEnvFile"
            ref="envTextarea"
            readonly 
            class="env-file-textarea" 
            :value="envFileContent"
          ></textarea>
        </div>

      </div>
      
      <div class="pop-up-actions single-action">
        <GeneralButton :label="'Finish'" :fullWidth="true" @click="launchStudio" :isActive="isValueChanged"
          :loading="isAwaitingResponse" />
      </div>
    </div>
    
  </div>
</template>

<script setup>
// imports
import { ref, onMounted, computed, watchEffect } from 'vue';

// services
import { StudioService } from '@/../bindings/clustta/services/index';

//stores
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useNotificationStore } from '@/stores/notifications';
import { useMenu } from '@/stores/menu';
import { useIconStore } from '@/stores/icons';
import { useProjectStore } from '@/stores/projects';
import { useStageStore } from '@/stores/stages';

//components
import HeaderArea from '@/instances/common/components/HeaderArea.vue';
import GeneralButton from '@/instances/common/components/GeneralButton.vue';
import ActionButton from '@/instances/desktop/components/ActionButton.vue';

//header vars
let title = 'New Self Managed Studio';

// stores/states
const modals = useDesktopModalStore();
const notificationStore = useNotificationStore();
const menu = useMenu();
const iconStore = useIconStore();
const projectStore = useProjectStore();
const stage = useStageStore();

//refs
const studioName = ref('');
const studioUrl = ref('');
const isStudioCreated = ref(false);
const isAwaitingResponse = ref(false);
const modalContainer = ref(null);
const createdStudio = ref(null);
const envTextarea = ref(null);
const envCopied = ref(false);
const showEnvFile = ref(false);
const showSecretKey = ref(false);
const secretKeyCopied = ref(false);

const restrictedNames = computed(() => {
  let restrictedNames = ['clustta', 'eaxum', 'pixar', 'disney', 'dreamworks'];
  return restrictedNames;
});

const studioNameInUse = computed(() => {
  return restrictedNames.value.includes(studioName.value.toLowerCase());
});

const studioNameEmpty = computed(() => {
  return studioName.value === ''
});

const isValueChanged = computed(() => {
  return !studioNameEmpty.value && !studioNameInUse.value
});

const envFileContent = computed(() => {
  if (!createdStudio.value) return '';
  
  return `DATA_FOLDER=/home/server-user-name/data/
PROJECTS_FOLDER=/home/server-user-name/projects/
CLUSTTA_STUDIO_API_KEY=${createdStudio.value.secret_key || ''}
CLUSTTA_SERVER_NAME=${createdStudio.value.name || ''}
CLUSTTA_SERVER_URL=${createdStudio.value.url || ''}`;
});

const getAppIcon = (iconName) => {
  const icon = iconStore.getAppIcon(iconName);
  return icon
};

const goBack = () => {
  modals.setModalVisibility('selectNewStudioTypeModal', true);
};

const closeModal = () => {
  modals.disableAllModals();
};

const handleEnterKey = (event) => {
  if (event.key === 'Enter' && isValueChanged.value) {
    createStudio();
  }
};

const createStudio = async () => {

  await StudioService.RegisterStudio(studioName.value, studioUrl.value).then(async (result) => {

    console.log(result);
    createdStudio.value = result;
    isAwaitingResponse.value = false;
    isStudioCreated.value = true;

  }).catch((error) => {
    isAwaitingResponse.value = false
    console.log(error)
    notificationStore.errorNotification('Error creating project', error);
  });
};

const copyEnvFile = async () => {
  try {
    await navigator.clipboard.writeText(envFileContent.value);
    envCopied.value = true;
    setTimeout(() => {
      envCopied.value = false;
    }, 2000);
  } catch (error) {
    console.error('Failed to copy:', error);
    notificationStore.errorNotification('Failed to copy to clipboard', error);
  }
};

const toggleEnvFile = () => {
  showEnvFile.value = !showEnvFile.value;
};

const toggleSecretKey = () => {
  showSecretKey.value = !showSecretKey.value;
};

const copySecretKey = async () => {
  try {
    if (createdStudio.value?.secret_key) {
      await navigator.clipboard.writeText(createdStudio.value.secret_key);
      secretKeyCopied.value = true;
      setTimeout(() => {
        secretKeyCopied.value = false;
      }, 2000);
    }
  } catch (error) {
    console.error('Failed to copy:', error);
    notificationStore.errorNotification('Failed to copy to clipboard', error);
  }
};

const selectSecretKey = (event) => {
  event.target.select();
};

const launchStudio = async () => {
  isAwaitingResponse.value = true;
  await projectStore.loadStudios();
  let studio = projectStore.studios.find((item) => item.name === createdStudio.value.name);
  if (studio) {
    projectStore.selectedStudio = studio;
  } else {
    projectStore.selectedStudio = projectStore.studios[0];
  }

  await projectStore.loadProjects().then((result) => {
    console.log(result)
  }).catch((error) => {
    console.error('Error:', error);
  });
  isAwaitingResponse.value = false;
  closeModal();
};

watchEffect(() => {
  if (modalContainer.value) {
    menu.clickOutsideMask = modalContainer.value;
  }
});

onMounted(async () => {
  // await projectTemplateStore.loadProjectTemplates();
});

</script>

<style scoped>
@import "@/assets/desktop.css";
@import "@/assets/modals.css";

.general-container {
  gap: 1rem;
}

.single-action{
  justify-content: flex-end;
}

.success-message {
  font-size: 14px;
  font-weight: 300;
  width: 100%;
  padding: 1rem;
  border-radius: 8px;
  color: var(--white);
  /* line-height: 1.6; */
}

.success-message p {
  margin: 0;
  padding: 0;
}

.success-message p:first-child {
  margin-bottom: 0.5rem;
}

.success-message strong {
  font-weight: 600;
  color: rgba(34, 197, 94, 1);
}
.input-section {
  width: 100%;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  justify-content: flex-start;
  gap: 0.5rem;
  color: var(--white);
}

.input-label-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.modal-info {
  display: flex;
  flex-direction: column;
  max-width: 100%;
  justify-content: flex-start;
  align-self: stretch;
  width: 464px;
  align-items: flex-start;
  box-sizing: border-box;

}

.modal-text-container {
  display: flex;
  flex-direction: column;
  max-width: 100%;
  justify-content: flex-start;
  align-self: stretch;
  width: 464px;
  align-items: flex-start;
  /* margin-top: 20px; */
}

.modal-title {
  max-width: 100%;
  align-self: stretch;
  width: 464px;
  color: rgba(16, 24, 40, 1);
  color: white;
  font-size: 18px;
  line-height: 28px;
  letter-spacing: 0%;
  text-align: left;
}

.input-header {
  /* background-color: lightblue; */
  width: 100%;
  display: flex;
  align-items: center;
  margin: 10px 0px;
}

.input-count {
  background-color: none;
  font-size: 14px;
  color: white;
}

.modal-subtitle {
  /* background-color: beige; */
  /* max-width: 100%; */
  align-self: stretch;
  width: 464px;
  color: rgba(16, 24, 40, 1);
  color: white;
  font-size: 14px;
  /* line-height: 28px; */
  letter-spacing: 0%;
  text-align: left;
}



.modal-body {
  box-sizing: border-box;
  max-width: 100%;
  align-self: stretch;
  width: 464px;
  margin: 8px 0px;
  font-size: 14px;
  color: rgba(16, 24, 40, 1);
  line-height: 20px;
  letter-spacing: 0%;
  text-align: left;
}

.modal-actions {
  box-sizing: border-box;
  padding: 1rem 2rem;
  gap: 2rem;
  display: flex;
  flex-direction: row;
  max-width: 100%;
  align-self: stretch;
  align-items: center;
  justify-content: space-evenly;
  width: 464px;
  margin-top: 32px;
}

.div-10 {
  display: flex;
}

.task-options-container {
  position: relative;
  box-sizing: border-box;
  width: 100%;
  height: max-content;
  height: 40px;
  transition: all .2s ease-in-out;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  margin: 0;
}

.task-options-container-closed {
  height: 0px;
  padding: 0;
  margin-bottom: -1rem;
}

.input-short {
  flex: 1;
  width: 100%;
}

.compound-input{
  font-family: 'Inter', sans-serif;
  box-sizing: border-box;
  font-size: 16px;
  border-radius: 12px;
  padding: 10px;
  border: 0px;
  border-style: solid;
  outline: none;
  background-color: var(--midnight-steel);
  color: var(--white);
  display: flex;
}

.compound-input-text{
  display: flex;
  font-family: 'Inter', sans-serif;
  box-sizing: border-box;
  font-size: 16px;
  border: 0px;
  border-style: solid;
  outline: none;
  background-color: transparent;
  width: min-content;
  background-color: crimson;
}

.compound-input-append{
  display: flex;
  /* width: 100%; */
  background-color: springgreen;
}

[data-theme="dark"] .input-short{
  font-weight: 200;
}

.listbox-short {

  flex: 1;
  width: 130px;
}

.input-label {
  font-family: Inter, sans-serif;
  color: var(--white);
  font-size: 14px;
  font-weight: 500;
  white-space: nowrap;
  opacity: 0.9;
}

.pop-up-prompt {
  gap: 10px;
  align-items: center;
  justify-content: center;
}

.studio-details {
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: 1rem;
  padding: 0.5rem 0;
}

.studio-info-section {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.studio-info-label {
  font-size: 12px;
  font-weight: 500;
  color: var(--white);
  opacity: 0.7;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.studio-info-value {
  font-size: 16px;
  color: var(--white);
  background-color: var(--midnight-steel);
  padding: 10px 12px;
  border-radius: 8px;
  word-break: break-word;
}

.studio-id,
.studio-key {
  font-family: 'Courier New', monospace;
  font-size: 14px;
}

.secret-key-container {
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  margin-bottom: 1rem;
}

.secret-key-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.secret-key-actions {
  display: flex;
  gap: 0.5rem;
  align-items: center;
}

.secret-key-value-container {
  width: 100%;
}

.secret-key-input {
  width: 100%;
  background-color: var(--midnight-steel);
  color: var(--white);
  border: none;
  border-radius: 8px;
  padding: 10px 12px;
  font-family: 'Courier New', monospace;
  font-size: 14px;
  font-weight: 500;
  outline: none;
  cursor: pointer;
  box-sizing: border-box;
}

.secret-key-input:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}

.secret-key-input:focus {
  outline: 1px solid rgba(255, 255, 255, 0.2);
}

.env-file-container {
  width: 100%;
  box-sizing: border-box;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  overflow: hidden;
}

.env-file-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.env-file-actions {
  display: flex;
  gap: 0.5rem;
  align-items: center;
}

.env-file-textarea-container {
  box-sizing: border-box;
  width: 100%;
  height: min-content;
  background-color: var(--midnight-steel);
  color: var(--white);
  border-radius: 8px;
  padding: 8px;
}

.env-file-textarea-container-closed{
  padding: 0px;
}

.env-file-textarea {
  width: 100%;
  min-height: 140px;
  color: var(--white);
  border: none;
  font-family: 'Courier New', monospace;
  background-color: transparent;
  font-weight: 500;
  font-size: 13px;
  line-height: 1.6;
  resize: none;
  outline: none;
  white-space: pre;
  overflow-wrap: normal;
  overflow: hidden;
  box-sizing: border-box;
  overflow-x: auto;
  border-radius: 8px;
  cursor: pointer;
}

.env-file-textarea::-webkit-scrollbar {
  height: 6px;
}

.env-file-textarea::-webkit-scrollbar-thumb {
  border-radius: 10px;
  background-color: var(--steel);
}
</style>
