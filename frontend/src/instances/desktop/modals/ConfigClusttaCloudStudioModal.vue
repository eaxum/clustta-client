<template>
  <div ref="modalContainer" class="modal-container">

    <HeaderArea :title="title" :icon="getAppIcon('clustta')" :showSearch="false" />

    <div class="general-container">

      <div class="studio-info-text">
        <p>{{ $t('modals.clusttaCloudDesc') }}</p>
      </div>

      <FormInput v-model="studioName" :placeholder="$t('placeholders.studioName')" :error="studioNameError" :loading="checkingStudioNameAvailability" :valid="!!studioName && !studioNameError && !checkingStudioNameAvailability" :showValidation="!!studioName" @input="checkStudioName" />

      <div v-if="isLoadingPlans" class="plan-loading">Loading plans...</div>

      <div v-else class="plan-select-container">
        <div class="plan-select-label">Select a plan</div>

        <OptionCard v-for="plan in studioPlans" :key="plan.id" :title="formatPlanName(plan.name)" :description="planDescription(plan)" :selectable="true" :selected="selectedPlanId === plan.id" @select="selectedPlanId = plan.id" />
      </div>

      <div class="pop-up-actions">
        <GeneralButton :label="$t('common.back')" :fullWidth="true" :buttonFunction="goBack" :colored="false" />
        <GeneralButton :label="createButtonLabel" :fullWidth="true" :buttonFunction="createStudioAndCheckout" :isActive="canProceed" :loading="isAwaitingResponse" />
      </div>
    </div>

  </div>
</template>

<script setup>
// imports
import { computed, onMounted, ref, watchEffect } from 'vue';
import { useI18n } from 'vue-i18n';
import { Browser } from '@wailsio/runtime';

// components
import FormInput from '@/instances/desktop/components/FormInput.vue';
import GeneralButton from '@/instances/common/components/GeneralButton.vue';
import HeaderArea from '@/instances/common/components/HeaderArea.vue';
import OptionCard from '@/instances/common/components/OptionCard.vue';

// services
import { StudioService } from '@/services';

// stores
const { t } = useI18n();
const entitlementStore = useEntitlementStore();
const iconStore = useIconStore();
const menu = useMenu();
const modals = useDesktopModalStore();
const notificationStore = useNotificationStore();

import { useDesktopModalStore } from '@/stores/desktopModals';
import { useEntitlementStore } from '@/stores/entitlements';
import { useIconStore } from '@/stores/icons';
import { useMenu } from '@/stores/menu';
import { useNotificationStore } from '@/stores/notifications';

// constants
const title = t('modals.newClusttaCloudStudio');

const restrictedNames = ['clustta', 'eaxum', 'pixar', 'disney', 'dreamworks'];

// refs
const checkingStudioNameAvailability = ref(false);
const isAwaitingResponse = ref(false);
const isLoadingPlans = ref(false);
const isStudioNameTaken = ref(false);
const modalContainer = ref(null);
const selectedPlanId = ref(null);
const studioName = ref('');
const studioNameError = ref('');

// computed

// Returns the label for the create button based on the selected plan.
const createButtonLabel = computed(() => {
  const plan = studioPlans.value.find(p => p.id === selectedPlanId.value);
  if (!plan) return t('common.create');
  if (plan.name === 'studio_enterprise') return 'Contact Sales';
  return 'Create & Subscribe';
});

// Returns whether the form is ready to proceed.
const canProceed = computed(() => {
  return isStudioNameValid.value && !!selectedPlanId.value;
});

// Returns whether the studio name is valid.
const isStudioNameValid = computed(() => {
  return studioName.value !== '' && !studioNameError.value && !isStudioNameTaken.value && !checkingStudioNameAvailability.value;
});

// Returns only paid studio plans (excludes free).
const studioPlans = computed(() => {
  return entitlementStore.plans.filter(p => p.type === 'studio' && p.price_cents !== 0);
});

// methods

// Checks if the studio name is available.
const checkStudioName = async () => {
  if (!studioName.value) {
    studioNameError.value = '';
    isStudioNameTaken.value = false;
    return;
  }

  if (restrictedNames.includes(studioName.value.toLowerCase())) {
    studioNameError.value = t('notifications.studioNameReserved');
    isStudioNameTaken.value = true;
    return;
  }

  checkingStudioNameAvailability.value = true;

  try {
    const nameExists = await StudioService.CheckStudioNameExists(studioName.value.toLowerCase());
    if (nameExists) {
      studioNameError.value = t('notifications.studioNameTaken');
      isStudioNameTaken.value = true;
    } else {
      studioNameError.value = '';
      isStudioNameTaken.value = false;
    }
  } catch (error) {
    studioNameError.value = '';
    isStudioNameTaken.value = false;
    console.error('Error checking studio name:', error);
  } finally {
    checkingStudioNameAvailability.value = false;
  }
};

// Creates the studio on the free tier, then redirects to Stripe Checkout for the selected plan.
const createStudioAndCheckout = async () => {
  if (!canProceed.value || isAwaitingResponse.value) return;

  const plan = studioPlans.value.find(p => p.id === selectedPlanId.value);
  if (!plan) return;

  if (plan.name === 'studio_enterprise') {
    // TODO: open enterprise contact form
    return;
  }

  isAwaitingResponse.value = true;

  try {
    // Create the studio (inactive until checkout completes)
    const result = await StudioService.RegisterStudio(studioName.value, '', 'cloud');

    // Get the studio ID from the creation response
    const studioId = result?.id || '';

    // Redirect to Stripe Checkout for the selected plan
    const checkoutUrl = await entitlementStore.createCheckout(plan.id, studioId);
    if (checkoutUrl) {
      Browser.OpenURL(checkoutUrl);
      notificationStore.addNotification('Checkout', 'Complete your payment in the browser. Your studio will be activated once payment is confirmed.', 'info', false);
    } else {
      notificationStore.addNotification('Error', 'Failed to start checkout. Please try again.', 'error', false);
    }

    modals.disableAllModals();
  } catch (error) {
    console.error(error);
    notificationStore.errorNotification(t('notifications.errorCreatingStudio'), error);
  } finally {
    isAwaitingResponse.value = false;
  }
};

// Returns a human-readable plan name.
const formatPlanName = (name) => {
  return name.split('_').map(w => w.charAt(0).toUpperCase() + w.slice(1)).join(' ');
};

// Returns the app icon for the given icon name.
const getAppIcon = (iconName) => {
  return iconStore.getAppIcon(iconName);
};

// Goes back to the studio type selection modal.
const goBack = () => {
  modals.setModalVisibility('selectNewStudioTypeModal', true);
};

// Returns a short description with price for the plan card.
const planDescription = (plan) => {
  if (plan.name === 'studio_enterprise') return 'Custom infrastructure — contact us for pricing';
  const price = '$' + (plan.price_cents / 100) + '/mo';
  const storage = formatStorage(plan.storage_bytes) + ' storage';
  const seats = plan.max_collaborators === -1 ? 'Unlimited seats' : plan.max_collaborators + ' seats';
  return price + ' · ' + storage + ' · ' + seats;
};

// Formats bytes to human-readable storage string.
const formatStorage = (bytes) => {
  if (bytes <= 0) return '0 B';
  const units = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(1024));
  return Math.round(bytes / Math.pow(1024, i)) + ' ' + units[i];
};

// watchers
watchEffect(() => {
  if (modalContainer.value) {
    menu.clickOutsideMask = modalContainer.value;
  }
});

// lifecycle hooks
onMounted(async () => {
  if (!entitlementStore.plans.length) {
    isLoadingPlans.value = true;
    await entitlementStore.fetchPlans();
    isLoadingPlans.value = false;
  }
});
</script>

<style scoped>
@import "@/assets/desktop.css";
@import "@/assets/modals.css";

.general-container {
  padding-top: 1rem;
  overflow-y: auto;
  max-height: 70vh;
}

.general-container::-webkit-scrollbar {
  width: 4px;
}

.general-container::-webkit-scrollbar-thumb {
  border-radius: var(--small-radius);
  background-color: var(--light-steel);
}

.general-container::-webkit-scrollbar-track {
  border-radius: var(--small-radius);
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

.plan-loading {
  text-align: center;
  padding: 1rem;
  color: var(--white);
  opacity: 0.6;
}

.plan-select-container {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  padding: 0.5rem 0;
  width: 100%;
}

.plan-select-label {
  font-size: 0.85rem;
  color: var(--white);
  opacity: 0.6;
  font-weight: 300;
}


</style>
