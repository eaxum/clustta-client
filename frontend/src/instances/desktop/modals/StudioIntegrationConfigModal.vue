<template>
  <div class="modal-container" v-stop-propagation v-esc="closeModal" v-return="handleEnterKey">
    <HeaderArea :title="modalTitle" :icon="getAppIcon(integrationIcon)" :showSearch="false" />

    <div class="general-container">

      <div v-if="view?.last_error" class="error-banner">{{ view.last_error }}</div>

      <p class="card-help">{{ $t('settings.kitsuIntegrationHelp') }}</p>

      <div class="input-section">
        <FormInput v-model="form.api_url" @input="validateApiUrl" placeholder="https://your-studio.cg-wire.com/api" :error="errors.api_url" :needsValidation="true" :showValidation="!!form.api_url" :valid="!errors.api_url && !!form.api_url" :labelTop="true" />
      </div>

      <div class="input-section">
        <FormInput v-model="form.email" @input="validateEmail" placeholder="bots@studio.com" type="email" :error="errors.email" :needsValidation="true" :showValidation="!!form.email" :valid="!errors.email && !!form.email" :labelTop="true" />
      </div>

      <div class="input-section">
        <FormInput v-model="form.password"  :placeholder="passwordPlaceholder" :isSecret="true" :labelTop="true" />
      </div>

      <div class="pop-up-actions">
        <GeneralButton :label="$t('common.close')" :fullWidth="true" :buttonFunction="closeModal" :colored="false" :isActive="!isBusy" />

        <GeneralButton :label="$t('common.save')" :fullWidth="true" :buttonFunction="onSave" :isActive="canSubmit && !isBusy" :loading="integrationStore.isSavingStudioConfig" />
      </div>

    </div>
  </div>
</template>

<script setup>
// imports
import { computed, onMounted, reactive } from 'vue';
import { useI18n } from 'vue-i18n';

// components
import FormInput from '@/instances/desktop/components/FormInput.vue';
import GeneralButton from '@/instances/common/components/GeneralButton.vue';
import HeaderArea from '@/instances/common/components/HeaderArea.vue';

// stores
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useIconStore } from '@/stores/icons';
import { useIntegrationStore } from '@/stores/integrations';
import { useProjectStore } from '@/stores/projects';

const iconStore = useIconStore();
const integrationStore = useIntegrationStore();
const modals = useDesktopModalStore();
const projectStore = useProjectStore();
const { t } = useI18n();

// refs
const form = reactive({ api_url: '', email: '', password: '', enabled: true });
const errors = reactive({ api_url: '', email: '' });

// computed
const integrationId = computed(() => integrationStore.activeStudioIntegrationId || 'kitsu');

const integration = computed(() => integrationStore.getIntegration(integrationId.value));

const integrationIcon = computed(() => integration.value?.icon || integrationId.value);

const integrationName = computed(() => integration.value?.name || integrationId.value);

const studioId = computed(() => projectStore.selectedStudio?.id || '');

const view = computed(() => integrationStore.studioConfig[integrationId.value] || null);

const modalTitle = computed(() => t('settings.kitsuIntegration'));

const isBusy = computed(() => integrationStore.isSavingStudioConfig);

const hasNoErrors = computed(() => !errors.api_url && !errors.email);

const canSubmit = computed(() => !!form.api_url && !!form.email && (!!form.password || !!view.value?.configured) && hasNoErrors.value);

const passwordPlaceholder = computed(() => view.value?.configured ? t('settings.serviceAccountPasswordUnchanged') : t('settings.serviceAccountPasswordPlaceholder'));

// methods
// Returns the app icon path for the given icon name.
const getAppIcon = (iconName) => iconStore.getAppIcon(iconName);

// Validates api_url with a URL parser; sets errors.api_url.
const validateApiUrl = () => {
  if (!form.api_url) {
    errors.api_url = '';
    return;
  }
  try {
    const u = new URL(form.api_url);
    errors.api_url = (u.protocol === 'http:' || u.protocol === 'https:') ? '' : t('settings.errorInvalidApiUrl');
  } catch {
    errors.api_url = t('settings.errorInvalidApiUrl');
  }
};

// Validates email format; sets errors.email.
const validateEmail = () => {
  if (!form.email) {
    errors.email = '';
    return;
  }
  const emailRe = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
  errors.email = emailRe.test(form.email) ? '' : t('settings.errorInvalidEmail');
};

// Seeds the form with existing values if the integration is configured.
const seedForm = () => {
  if (view.value?.configured) {
    form.api_url = view.value.api_url || '';
    form.email = view.value.email || '';
    form.enabled = view.value.enabled;
    form.password = '';
  } else {
    form.api_url = '';
    form.email = '';
    form.password = '';
    form.enabled = true;
  }
  errors.api_url = '';
  errors.email = '';
};

// Closes the modal and clears the active integration id.
const closeModal = () => {
  modals.setModalVisibility('studioIntegrationConfigModal', false);
  integrationStore.setActiveStudioIntegration(null);
};

// Saves the form to the server and closes the modal on success.
const onSave = async () => {
  if (!studioId.value || !canSubmit.value || isBusy.value) return;
  const enabledForSave = view.value?.configured ? view.value.enabled : true;
  const payload = { api_url: form.api_url, email: form.email, password: form.password, enabled: enabledForSave };
  const result = await integrationStore.saveStudioIntegrationConfig(studioId.value, integrationId.value, payload);
  if (result) {
    closeModal();
  }
};

// Handles enter key press to submit the form when valid.
const handleEnterKey = () => {
  if (canSubmit.value && !isBusy.value) onSave();
};

// lifecycle hooks
onMounted(async () => {
  if (studioId.value && integrationId.value) {
    await integrationStore.loadStudioIntegrationConfig(studioId.value, integrationId.value);
  }
  seedForm();
});
</script>

<style scoped>
@import "@/assets/desktop.css";

.card-help {
  margin: 0;
  font-size: .85rem;
  line-height: 1.4;
  color: var(--text);
  padding-bottom: .5rem;
}

.error-banner {
  background-color: rgba(255, 80, 80, 0.12);
  color: #ff6b6b;
  border-radius: var(--small-radius);
  padding: .5rem .75rem;
  font-size: .85rem;
}

.input-section {
  width: 100%;
}
</style>
