<template>
  <div ref="modalContainer" class="modal-container">

      <HeaderArea :title="title" :icon="getAppIcon('clustta')" :showSearch="false" />

    <div v-if="!isStudioRegistered" class="general-container">

      <div class="studio-info-text">
        <p>{{ $t('modals.clusttaCloudDesc') }}</p>
      </div>

      <FormInput
        v-model="studioName"
        :placeholder="$t('placeholders.studioName')"
        :error="studioNameError"
        :loading="checkingStudioNameAvailability"
        :valid="!!studioName && !studioNameError && !checkingStudioNameAvailability"
        :showValidation="!!studioName"
        @input="checkStudioName"
      />

      <FormInput
        v-model="deploymentCode"
        :placeholder="$t('placeholders.deploymentCode')"
        :error="deploymentCodeError"
        @input="deploymentCodeError = ''"
      />

      <!-- Notification section for deployment code info -->
      <div class="notification-area">
        <div class="horizontal-flex">
          <NotificationBox 
            type="info"
            :icon="getAppIcon('info')"
            :iconAlt="$t('common.info')"
            :title="$t('modals.getYourCode')"
            :message="$t('modals.discordCodeMessage')"
            :clickable="true"
            @click="openDiscordLink"
          />
        </div>
      </div>


      <!-- Temporarily disable ability to change studio type -->
        <!-- <div class="input-section">
            <div class="input-label-row">
                <label class="input-label">Location</label>
            </div>
            <div class="horizontal-flex">
                <DropDownBox :items="serverLocationNames" :selectedItem="serverLocationName" :onSelect="changeServerLocation" />
            </div>
        </div>

        <div class="input-section">
            <div class="input-label-row">
                <label class="input-label">Plan</label>
            </div>
            <div class="horizontal-flex">
                <DropDownBox :items="vmSizeNames" :selectedItem="vmSizeName" :onSelect="changeVmSize" />
            </div>
        </div> -->


      <div class="pop-up-actions">
        <GeneralButton :label="$t('common.back')" :fullWidth="true" :buttonFunction="goBack" :colored="false" />
        <GeneralButton 
          :label="$t('common.create')" 
          :fullWidth="true" 
          @click="createStudio" 
          :isActive="isValueChanged"
          :loading="isAwaitingResponse" 
        />
      </div>
    </div>

    <div v-else class="general-container">

      <div class="success-message">
        <p v-if="deploymentStatus?.status === 'completed'">{{ $t('modals.studioCreatedSuccess') }}</p>
        <p v-else>{{ $t('modals.studioCreatingMessage') }}</p>
      </div>

      <!-- Deployment Progress -->
      <div v-if="deploymentStatus && deploymentStatus?.status !== 'completed'" class="deployment-progress">
        
        <div class="deployment-progress-header">

            <div class="deployment-progress-status">
                <span class="single-action-button">
                <img class="small-icons loading-children-icon" :src="getAppIcon('loading')">
                </span>
                
                <div class="deployment-progress-title">{{ deploymentStatus.current_step }}</div>
            </div>

          
          <div class="deployment-progress-percentage">{{ deploymentStatus.progress }}%</div>
        </div>

        <div class="deployment-progress-bar-container">
            <div class="progress-bar-loader tint">
                <ProgressBar :taskProgress="deploymentStatus.progress" />
            </div>
        </div>
        
        
        
        <div v-if="deploymentStatus.status === 'failed'" class="deployment-error">
          {{ $t('common.error') }}: {{ deploymentStatus.error }}
        </div>
      </div>      
      
      <!-- UI Testing Button -->
      <!-- <div class="ui-testing-section">
        <GeneralButton 
          :label="'Toggle'" 
          :fullWidth="true" 
          @click="toggleDeploymentStatus" 
          :colored="false"
        />
      </div> -->
      
      <div class="pop-up-actions single-action">
        <GeneralButton :label="$t('common.finish')" :fullWidth="true" @click="launchStudio" :isActive="isValueChanged"
          :loading="isAwaitingResponse" />
      </div>
    </div>
    
  </div>
</template>

<script setup>
// imports
import { computed, onMounted, ref, watchEffect } from 'vue';
import { Browser } from "@wailsio/runtime";
import { useI18n } from 'vue-i18n';

// components
import DropDownBox from '@/instances/common/components/DropDownBox.vue';
import FormInput from '@/instances/desktop/components/FormInput.vue';
import GeneralButton from '@/instances/common/components/GeneralButton.vue';
import HeaderArea from '@/instances/common/components/HeaderArea.vue';
import InputAlert from '@/instances/common/components/InputAlert.vue';
import NotificationBox from '@/instances/common/components/NotificationBox.vue';
import ProgressBar from '@/instances/common/components/ProgressBar.vue';

// services
import { DeploymentService, StudioService } from '@/services';

// stores
const { t } = useI18n();
const iconStore = useIconStore();
const menu = useMenu();
const modals = useDesktopModalStore();
const notificationStore = useNotificationStore();
const projectStore = useProjectStore();
const stage = useStageStore();

import { useDesktopModalStore } from '@/stores/desktopModals';
import { useIconStore } from '@/stores/icons';
import { useMenu } from '@/stores/menu';
import { useNotificationStore } from '@/stores/notifications';
import { useProjectStore } from '@/stores/projects';
import { useStageStore } from '@/stores/stages';

// constants
const title = t('modals.newClusttaCloudStudio');

// Dummy deployment status for UI testing
const dummyStatuses = [
  null,
  {
    deployment_id: 'dep_test_123',
    studio_name: 'Test Studio',
    status: 'queued',
    progress: 0,
    current_step: 'Queued',
    error: null
  },
  {
    deployment_id: 'dep_test_123',
    studio_name: 'Test Studio',
    status: 'running',
    progress: 20,
    current_step: 'Initializing',
    error: null
  },
  {
    deployment_id: 'dep_test_123',
    studio_name: 'Test Studio',
    status: 'running',
    progress: 50,
    current_step: 'Creating resources',
    error: null
  },
  {
    deployment_id: 'dep_test_123',
    studio_name: 'Test Studio',
    status: 'running',
    progress: 80,
    current_step: 'Configuring machine',
    error: null
  },
  {
    deployment_id: 'dep_test_123',
    studio_name: 'Test Studio',
    status: 'completed',
    progress: 100,
    current_step: 'Deployment complete',
    public_ip: '20.121.45.123',
    vm_name: 'test-studio-vm',
    error: null
  },
  {
    deployment_id: 'dep_test_123',
    studio_name: 'Test Studio',
    status: 'failed',
    progress: 30,
    current_step: 'Creating Azure resources',
    error: 'Failed to create resource group: insufficient permissions'
  }
];

// refs
const checkingStudioNameAvailability = ref(false);
const createdStudio = ref(null);
const currentDummyIndex = ref(0);
const deploymentCode = ref('');
const deploymentCodeError = ref('');
const deploymentId = ref('');
const deploymentStatus = ref(null);
const diskSizeGB = ref(30);
const isAwaitingResponse = ref(false);
const isStudioNameTaken = ref(false);
const isStudioRegistered = ref(false);
const modalContainer = ref(null);
const serverLocation = ref('eastus');
const serverLocationName = ref('East US');
const studioName = ref('');
const studioNameError = ref('');
const vmSize = ref('Standard_B1s');
const vmSizeName = ref('Indie');

// computed
const deploymentCodeEmpty = computed(() => {
  return deploymentCode.value === '';
});

const isValueChanged = computed(() => {
  const baseValid = !studioNameEmpty.value && !studioNameInUse.value && !deploymentCodeEmpty.value && !studioNameError.value && !deploymentCodeError.value;
  return baseValid;
});

const restrictedNames = computed(() => {
  return ['clustta', 'eaxum', 'pixar', 'disney', 'dreamworks'];
});

const serverLocationNames = computed(() => {
  return serverLocations.value.map((location) => location.name);
});

const serverLocations = computed(() => {
  return [
    { name: 'East US', location: 'eastus' },
    { name: 'West US 2', location: 'westus2' },
    { name: 'Central US', location: 'centralus' },
    { name: 'West Europe', location: 'westeurope' },
    { name: 'North Europe', location: 'northeurope' },
  ];
});

const studioNameEmpty = computed(() => {
  return studioName.value === '';
});

const studioNameInUse = computed(() => {
  return restrictedNames.value.includes(studioName.value.toLowerCase()) || isStudioNameTaken.value;
});

const vmSizeNames = computed(() => {
  return vmSizes.value.map((size) => size.name);
});

const vmSizes = computed(() => {
  return [
    {
      name: "Indie",
      vmSize: "Standard_B1s",
      price: "$11/month",
      description: "1 vCPU, 1GB RAM",
      details: "Designed for individual creators and small teams of 1-3 people."
    },
    {
      name: "Team",
      vmSize: "Standard_B2s",
      price: "$45/month",
      description: "2 vCPU, 4GB RAM",
      details: "Built for growing creative teams of 3-20 people working collaboratively."
    },
    {
      name: "Studio",
      vmSize: "Standard_B4ms",
      price: "$180/month",
      description: "4 vCPU, 16GB RAM",
      details: "Optimized for large studios and organizations with 20-50 team members."
    }
  ];
});

// methods

// Changes the server location setting.
const changeServerLocation = (selectedServerLocation) => {
  const selectedServer = serverLocations.value.find((item) => item.name === selectedServerLocation);
  serverLocationName.value = selectedServer.name;
  serverLocation.value = selectedServer.location;
};

// Changes the VM size setting.
const changeVmSize = (selectedVmSize) => {
  const selectedSize = vmSizes.value.find((item) => item.name === selectedVmSize);
  vmSizeName.value = selectedSize.name;
  vmSize.value = selectedSize.vmSize;
};

// Checks if the studio name is available.
const checkStudioName = async () => {
  if (!studioName.value) {
    studioNameError.value = '';
    isStudioNameTaken.value = false;
    return;
  }
  
  if (restrictedNames.value.includes(studioName.value.toLowerCase())) {
    studioNameError.value = t('notifications.studioNameReserved');
    isStudioNameTaken.value = true;
    return;
  }
  
  checkingStudioNameAvailability.value = true;

  try {
    const nameExists = await StudioService.CheckStudioNameExists(studioName.value.toLowerCase());
    console.log(nameExists);
    if (nameExists) {
      studioNameError.value = t('notifications.studioNameTaken');
      isStudioNameTaken.value = true;
    } else {
      studioNameError.value = '';
      isStudioNameTaken.value = false;
    }
    checkingStudioNameAvailability.value = false;
  } catch (error) {
    studioNameError.value = '';
    isStudioNameTaken.value = false;
    console.error('Error checking studio name:', error);
    checkingStudioNameAvailability.value = false;
  }
};

// Closes the modal.
const closeModal = () => {
  modals.disableAllModals();
};

// Creates a new Clustta Cloud studio.
const createStudio = async () => {
  isAwaitingResponse.value = true;
  deploymentCodeError.value = '';
  
  try {
    const [isValid, message] = await StudioService.VerifyDeploymentCode(deploymentCode.value);
    deploymentCode.value = '';
    
    if (!isValid) {
      deploymentCodeError.value = message || t('notifications.invalidDeploymentCode');
      notificationStore.errorNotification(t('notifications.invalidDeploymentCode'), deploymentCodeError.value);
      isAwaitingResponse.value = false;
      return;
    }
    
    const studioResult = await StudioService.RegisterStudio(studioName.value, 'pending');
    createdStudio.value = studioResult;
    
    const deploymentRequest = {
      studio_name: studioName.value,
      studio_url: 'pending',
      studio_secret_key: studioResult.secret_key,
      azure_region: serverLocation.value,
      vm_size: vmSize.value,
      disk_size_gb: diskSizeGB.value
    };
    
    const deploymentResult = await DeploymentService.DeployStudio(deploymentRequest);
    deploymentId.value = deploymentResult.deployment_id;
    
    monitorDeployment();
    
    isStudioRegistered.value = true;

  } catch (error) {
    console.error(error);
    notificationStore.errorNotification(t('notifications.errorCreatingStudio'), error);
  } finally {
    isAwaitingResponse.value = false;
  }
};

// Returns the app icon for the given icon name.
const getAppIcon = (iconName) => {
  return iconStore.getAppIcon(iconName);
};

// Goes back to the studio type selection modal.
const goBack = () => {
  modals.setModalVisibility('selectNewStudioTypeModal', true);
};

// Handles enter key press.
const handleEnterKey = (event) => {
  if (event.key === 'Enter' && isValueChanged.value) {
    createStudio();
  }
};

// Launches the studio after creation.
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
    console.log(result);
  }).catch((error) => {
    console.error('Error:', error);
  });
  isAwaitingResponse.value = false;
  closeModal();
};

// Monitors deployment progress.
const monitorDeployment = async () => {
  const checkDeployment = async () => {
    try {
      const status = await DeploymentService.GetDeploymentStatus(deploymentId.value);
      deploymentStatus.value = status;
      
      if (status.status === 'completed') {
        createdStudio.value.url = `http://${status.public_ip}`;
        notificationStore.addNotification(t('notifications.deploymentComplete'), t('notifications.vmReady'));
      } else if (status.status === 'failed') {
        notificationStore.errorNotification(t('notifications.deploymentFailed'), status.error || t('notifications.unknownError'));
      } else {
        setTimeout(checkDeployment, 5000);
      }
    } catch (error) {
      console.error('Error monitoring deployment:', error);
      notificationStore.errorNotification(t('notifications.monitoringError'), t('notifications.failedToCheckDeploymentStatus'));
    }
  };
  
  checkDeployment();
};

// Opens the Discord link for getting deployment codes.
const openDiscordLink = () => {
  Browser.OpenURL('https://discord.gg/NuR4uAuTZd');
};

// Toggle through dummy deployment statuses for UI testing.
const toggleDeploymentStatus = () => {
  currentDummyIndex.value = (currentDummyIndex.value + 1) % dummyStatuses.length;
  deploymentStatus.value = dummyStatuses[currentDummyIndex.value];
  
  if (deploymentStatus.value?.status === 'completed') {
    if (createdStudio.value) {
      createdStudio.value.url = `http://${deploymentStatus.value.public_ip}`;
    }
  }
  
  console.log('Deployment Status Updated:', deploymentStatus.value);
};

// watchers
watchEffect(() => {
  if (modalContainer.value) {
    menu.clickOutsideMask = modalContainer.value;
  }
});

// lifecycle
onMounted(async () => {
  // await projectTemplateStore.loadProjectTemplates();
});
</script>

<style scoped>
@import "@/assets/desktop.css";
@import "@/assets/modals.css";

.single-action{
  justify-content: flex-end;
}

.general-container{
  padding-top: 1rem;
}
.success-message {
  font-size: 16px;
  /* font-weight: 300; */
  width: 100%;
  padding: .5rem;
  border-radius: 8px;
  color: var(--white);
  box-sizing: border-box;
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



.input-short {
  flex: 1;
  width: 100%;
}



[data-theme="dark"] .input-short{
  font-weight: 200;
}



.input-label {
  font-family: Inter, sans-serif;
  color: var(--white);
  font-size: 14px;
  font-weight: 400;
  white-space: nowrap;
  opacity: 0.9;
}

.studio-name-input{
  height: min-content;
  align-items: center;
  justify-content: center;
}

.form-input-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  min-width: 24px;
}

.alert-icons {
  width: 20px;
  height: 20px;
}



/* Deployment Progress */
.deployment-progress {
    box-sizing: border-box;
    width: 100%;
    display: flex;
    flex-direction: column;
    gap: .5rem;
    padding: .5rem;
    /* background-color: var(--midnight-steel); */
    border-radius: 8px;
    /* border: 1px solid var(--steel); */
    /* margin: 1rem 0; */
}

.deployment-progress-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.5rem;
}

.deployment-progress-title {
  color: var(--white);
}

.deployment-progress-percentage {
  color: var(--white);
}

.deployment-progress-bar-container {
  position: relative;
  width: 100%;
  height: 20px;
  margin-bottom: 0.5rem;
}

.progress-bar-loader {
  position: relative;
  width: 100%;
  height: .2rem;
  border-radius: 999px;
  /* background-color: white; */
}

@keyframes loadingRotate {
  from {
      transform: rotate(0deg);
  }
  to {
      transform: rotate(360deg);
  }
}

.single-action-button{
  align-content: center;
  justify-content: center;
}

.loading-children-icon {
  width: 20px;
  height: 20px;
  overflow: hidden;
  padding: 0px;
  animation: loadingRotate .5s linear infinite;
}

.deployment-progress-status {
    display: flex;
    box-sizing: border-box;
    align-items: center;
    gap: .2rem;
    font-size: 0.9rem;
    color: var(--white);
}

.deployment-error {
  color: var(--error);
  font-size: 0.9rem;
  margin-top: 0.5rem;
  padding: 0.5rem;
  background-color: rgba(255, 0, 0, 0.1);
  border-radius: 4px;
}

/* UI Testing Section */
.ui-testing-section {
    box-sizing: border-box;
    width: 100%;
    padding: 1rem;
    background-color: rgba(255, 165, 0, 0.1);
    border-radius: 8px;
    border: 1px dashed #ffa500;
    margin: 1rem 0;
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    align-items: center;
}

.current-status-info {
  font-size: 0.9rem;
  color: var(--white);
  text-align: center;
}

/* Notification area */
.notification-area{
  width: 100%;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  gap: .5rem;
}

.studio-info-text {
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  color: var(--white);
  font-size: 14px;
  padding: .5rem 0;
  box-sizing: border-box;
}

.studio-info-text p {
  margin: 0;
}
</style>
