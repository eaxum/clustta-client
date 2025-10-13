<template>
  <div ref="modalContainer" class="modal-container">

    <div class="general-pane-header">
      <HeaderArea :title="title" :icon="getAppIcon('clustta')" :showSearch="false" />
    </div>

    <div v-if="!isStudioRegistered" class="general-container">

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
        </div>


      <div class="pop-up-actions">
        <GeneralButton :label="'Back'" :fullWidth="true" :buttonFunction="goBack" :colored="false" />
        <GeneralButton 
          :label="'Create'" 
          :fullWidth="true" 
          @click="createStudio" 
          :isActive="isValueChanged"
          :loading="isAwaitingResponse" 
        />
      </div>
    </div>

    <div v-else class="general-container">

      <div class="success-message">
        <p v-if="deploymentStatus?.status === 'completed'">Your Studio is created successfully!</p>
        <p v-else>Hang tight, we're creating your studio... This wont take long.</p>
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
          Error: {{ deploymentStatus.error }}
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
import { StudioService, DeploymentService } from '@/../bindings/clustta/services/index';

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
import DropDownBox from '@/instances/common/components/DropDownBox.vue';
import ProgressBar from '@/instances/common/components/ProgressBar.vue';

//header vars
let title = 'New ClusttaCloud Studio';

// stores/states
const modals = useDesktopModalStore();
const notificationStore = useNotificationStore();
const menu = useMenu();
const iconStore = useIconStore();
const projectStore = useProjectStore();
const stage = useStageStore();

//refs
const studioName = ref('');
const vmSize = ref('Standard_B1s');
const diskSizeGB = ref(30);
const isStudioRegistered = ref(false);
const isAwaitingResponse = ref(false);
const modalContainer = ref(null);
const createdStudio = ref(null);
const deploymentStatus = ref(null);
const deploymentId = ref('');

// Dummy deployment status for UI testing
const dummyStatuses = [
  null, // No deployment started
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

const currentDummyIndex = ref(0);

const serverLocation = ref('eastus');
const serverLocationName = ref('East US');

const serverLocations = computed(() => {
  return  [
    {name: 'East US', location: 'eastus'},
    {name: 'West US 2', location: 'westus2'},
    {name: 'Central US', location: 'centralus'},
    {name: 'West Europe', location: 'westeurope'},
    {name: 'North Europe', location: 'northeurope'},
  ];
});

const serverLocationNames = computed(() => {
  return serverLocations.value.map((location) => location.name)
});

const changeServerLocation = (selectedServerLocation) => {
    const selectedServer = serverLocations.value.find((item) => item.name === selectedServerLocation );
    serverLocationName.value = selectedServer.name;
    serverLocation.value = selectedServer.location
};

const vmSizeName = ref('Indie');

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

const vmSizeNames = computed(() => {
  return vmSizes.value.map((size) => size.name)
});

const changeVmSize = (selectedVmSize) => {
    const selectedSize = vmSizes.value.find((item) => item.name === selectedVmSize );
    vmSizeName.value = selectedSize.name;
    vmSize.value = selectedSize.vmSize;
};

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
  const baseValid = !studioNameEmpty.value && !studioNameInUse.value;
    return baseValid;
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
  isAwaitingResponse.value = true;
  
  try {
      // Deployment - first register studio, then deploy to Azure
      const studioResult = await StudioService.RegisterStudio(studioName.value, 'pending');
      createdStudio.value = studioResult;
      
      // Prepare deployment request
      const deploymentRequest = {
        studio_name: studioName.value,
        studio_url: 'pending', // Will be updated after deployment
        studio_secret_key: studioResult.secret_key,
        azure_region: serverLocation.value,
        vm_size: vmSize.value,
        disk_size_gb: diskSizeGB.value
      };
      
      // Start deployment
      const deploymentResult = await DeploymentService.DeployStudio(deploymentRequest);
      deploymentId.value = deploymentResult.deployment_id;
      
      // Start monitoring deployment progress
      monitorDeployment();
      
      isStudioRegistered.value = true;

  } catch (error) {
    console.error(error);
    notificationStore.errorNotification('Error creating studio', error);
  } finally {
    isAwaitingResponse.value = false;
  }
};

const monitorDeployment = async () => {
  const checkDeployment = async () => {
    try {
      const status = await DeploymentService.GetDeploymentStatus(deploymentId.value);
      deploymentStatus.value = status;
      
      if (status.status === 'completed') {
        // Update studio URL with the deployed VM's public IP
        createdStudio.value.url = `http://${status.public_ip}`;
        notificationStore.addNotification('Deployment Complete', 'Your Azure VM is ready!');
      } else if (status.status === 'failed') {
        notificationStore.errorNotification('Deployment Failed', status.error || 'Unknown error');
      } else {
        // Continue monitoring
        setTimeout(checkDeployment, 5000); // Check every 5 seconds
      }
    } catch (error) {
      console.error('Error monitoring deployment:', error);
      notificationStore.errorNotification('Monitoring Error', 'Failed to check deployment status');
    }
  };
  
  checkDeployment();
};

// Toggle through dummy deployment statuses for UI testing
const toggleDeploymentStatus = () => {
  currentDummyIndex.value = (currentDummyIndex.value + 1) % dummyStatuses.length;
  deploymentStatus.value = dummyStatuses[currentDummyIndex.value];
  
  // If completed status, update the studio URL
  if (deploymentStatus.value?.status === 'completed') {
    if (createdStudio.value) {
      createdStudio.value.url = `http://${deploymentStatus.value.public_ip}`;
    }
  }
  
  console.log('Deployment Status Updated:', deploymentStatus.value);
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
  /* gap: 1rem; */
}

.single-action{
  justify-content: flex-end;
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
</style>
