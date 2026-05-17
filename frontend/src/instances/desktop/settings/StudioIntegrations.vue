<template>
  <div class="settings-component-root">

    <div class="settings-component-scroll">

      <div class="settings-component-container">

        <ProfileCard :title="$t('settings.kitsuIntegration')" :showEditButton="!!view?.configured && !isEditing" :isEditing="isEditing" @toggleEdit="toggleEdit">
          <div class="card-section">

            <div class="card-header-row">
              <p class="card-help">{{ $t('settings.kitsuIntegrationHelp') }}</p>

              <div v-if="view?.configured" class="status-badge" :class="statusClass">{{ statusLabel }}</div>
            </div>

            <div v-if="view?.last_error" class="error-banner">{{ view.last_error }}</div>

            <div v-if="!isEditing && view?.configured" class="display-fields">
              <div class="display-field">
                <span class="display-label">{{ $t('settings.kitsuApiUrl') }}</span>
                <span class="display-value">{{ view.api_url }}</span>
              </div>

              <div class="display-field">
                <span class="display-label">{{ $t('settings.serviceAccountEmail') }}</span>
                <span class="display-value">{{ view.email }}</span>
              </div>

              <div v-if="view.last_validated_at" class="display-field">
                <span class="display-label">{{ $t('settings.lastValidatedAt') }}</span>
                <span class="display-value">{{ lastValidatedFormatted }}</span>
              </div>
            </div>

            <div v-else class="edit-fields">
              <FormInput v-model="form.api_url" @input="validateApiUrl" :label="$t('settings.kitsuApiUrl')" :placeholder="'https://your-studio.cg-wire.com/api'" :error="errors.api_url" :needsValidation="true" :showValidation="!!form.api_url" :valid="!errors.api_url && !!form.api_url" :labelTop="true" />

              <FormInput v-model="form.email" @input="validateEmail" :label="$t('settings.serviceAccountEmail')" :placeholder="'bots@studio.com'" type="email" :error="errors.email" :needsValidation="true" :showValidation="!!form.email" :valid="!errors.email && !!form.email" :labelTop="true" />

              <FormInput v-model="form.password" :label="$t('settings.serviceAccountPassword')" :placeholder="passwordPlaceholder" :isSecret="true" :labelTop="true" />

              <div class="enabled-row">
                <label class="form-label">{{ $t('settings.enabled') }}</label>
                <ToggleSwitch :switchValueProp="form.enabled" @click="form.enabled = !form.enabled" />
              </div>

              <div class="actions-row">
                <ActionButton :icon="getAppIcon('refresh')" :label="$t('common.test')" :showLabel="true" :isDisabled="isBusy || !canTest" :buttonFunction="onTest" :useOutline="true" :isLoading="integrationStore.isTestingStudioConfig" />

                <ActionButton :icon="getAppIcon('save')" :label="$t('common.save')" :showLabel="true" :isDisabled="isBusy || !canSubmit" :buttonFunction="onSave" :useBackground="true" :isLoading="integrationStore.isSavingStudioConfig" />

                <ActionButton v-if="view?.configured" :icon="getAppIcon('close')" :label="$t('common.cancel')" :showLabel="true" :isDisabled="isBusy" :buttonFunction="cancelEdit" />

                <ActionButton v-if="view?.configured" :icon="getAppIcon('trash')" :label="$t('common.delete')" :showLabel="true" :isDisabled="isBusy" :buttonFunction="onDelete" useAlert />
              </div>
            </div>
          </div>
        </ProfileCard>

      </div>

    </div>

  </div>
</template>

<script setup>
// imports
import { computed, onMounted, reactive, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import FormInput from '@/instances/desktop/components/FormInput.vue';
import ProfileCard from '@/instances/desktop/components/ProfileCard.vue';
import ToggleSwitch from '@/instances/common/components/ToggleSwitch.vue';

// stores
import { useIconStore } from '@/stores/icons';
import { useIntegrationStore } from '@/stores/integrations';
import { useProjectStore } from '@/stores/projects';

const iconStore = useIconStore();
const integrationStore = useIntegrationStore();
const projectStore = useProjectStore();

const { t } = useI18n();

// refs
const integrationId = 'kitsu';
const form = reactive({ api_url: '', email: '', password: '', enabled: true });
const errors = reactive({ api_url: '', email: '' });
const isEditing = ref(false);

// computed properties
const studioId = computed(() => projectStore.selectedStudio?.id || '');

const view = computed(() => integrationStore.studioConfig[integrationId] || null);

const isBusy = computed(() => integrationStore.isSavingStudioConfig || integrationStore.isTestingStudioConfig);

const hasNoErrors = computed(() => !errors.api_url && !errors.email);

const canTest = computed(() => form.api_url && form.email && form.password && hasNoErrors.value);

const canSubmit = computed(() => form.api_url && form.email && (form.password || view.value?.configured) && hasNoErrors.value);

const passwordPlaceholder = computed(() => view.value?.configured ? t('settings.serviceAccountPasswordUnchanged') : t('settings.serviceAccountPasswordPlaceholder'));

const statusClass = computed(() => {
  if (!view.value?.configured) return 'status-stopped';
  if (view.value.last_error) return 'status-error';
  if (view.value.status === 'running') return 'status-running';
  return 'status-stopped';
});

const statusLabel = computed(() => {
  if (!view.value?.configured) return t('settings.statusNotConfigured');
  if (view.value.last_error) return t('settings.statusError');
  return view.value.status === 'running' ? t('settings.statusRunning') : t('settings.statusStopped');
});

const lastValidatedFormatted = computed(() => {
  if (!view.value?.last_validated_at) return '';
  return new Date(view.value.last_validated_at * 1000).toLocaleString();
});

// methods
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

// Pulls the current configuration and seeds the form with existing values.
const refresh = async () => {
  if (!studioId.value) return;
  await integrationStore.loadStudioIntegrationConfig(studioId.value, integrationId);
  if (view.value?.configured) {
    form.api_url = view.value.api_url || '';
    form.email = view.value.email || '';
    form.enabled = view.value.enabled;
    form.password = '';
    isEditing.value = false;
  } else {
    isEditing.value = true;
  }
};

// Switches to edit mode and re-seeds the form from the latest view.
const toggleEdit = () => {
  if (view.value?.configured) {
    form.api_url = view.value.api_url || '';
    form.email = view.value.email || '';
    form.enabled = view.value.enabled;
    form.password = '';
  }
  isEditing.value = true;
};

// Exits edit mode and resets transient errors and password buffer.
const cancelEdit = () => {
  isEditing.value = false;
  form.password = '';
  errors.api_url = '';
  errors.email = '';
};

// Saves the form to the server and exits edit mode on success.
const onSave = async () => {
  if (!studioId.value || !canSubmit.value) return;
  const payload = { api_url: form.api_url, email: form.email, password: form.password, enabled: form.enabled };
  const result = await integrationStore.saveStudioIntegrationConfig(studioId.value, integrationId, payload);
  if (result) {
    form.password = '';
    isEditing.value = false;
  }
};

// Runs a live connection check using the supplied credentials.
const onTest = async () => {
  if (!studioId.value || !canTest.value) return;
  const payload = { api_url: form.api_url, email: form.email, password: form.password, enabled: form.enabled };
  await integrationStore.testStudioIntegrationConfig(studioId.value, integrationId, payload);
};

// Removes the integration after user confirmation.
const onDelete = async () => {
  if (!studioId.value) return;
  const confirmed = window.confirm(t('settings.confirmDeleteIntegration'));
  if (!confirmed) return;
  const ok = await integrationStore.deleteStudioIntegrationConfig(studioId.value, integrationId);
  if (ok) {
    form.api_url = '';
    form.email = '';
    form.password = '';
    form.enabled = true;
    errors.api_url = '';
    errors.email = '';
    isEditing.value = true;
  }
};

// watchers
watch(studioId, async (id) => {
  if (id) await refresh();
});

// lifecycle hooks
onMounted(async () => {
  await refresh();
});
</script>

<style scoped>
@import "@/assets/desktop.css";

.settings-component-root {
  width: 100%;
  height: 100%;
  overflow: hidden;
  display: block;
  overflow-y: scroll;
  border-radius: var(--very-large-radius);
  box-sizing: border-box;
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

.card-section {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.card-header-row {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
}

.card-help {
  margin: 0;
  font-size: .85rem;
  line-height: 1.4;
  color: var(--white-half);
  flex: 1;
}

.error-banner {
  background-color: rgba(255, 80, 80, 0.12);
  color: #ff6b6b;
  border-radius: var(--small-radius);
  padding: .5rem .75rem;
  font-size: .85rem;
}

.display-fields {
  display: flex;
  flex-direction: column;
  gap: .75rem;
}

.display-field {
  display: flex;
  flex-direction: column;
  gap: .15rem;
}

.display-label {
  font-size: .75rem;
  text-transform: uppercase;
  letter-spacing: .05em;
  color: var(--white-half);
}

.display-value {
  font-size: .9rem;
  color: var(--white);
  word-break: break-all;
}

.edit-fields {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.enabled-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
}

.form-label {
  font-size: .85rem;
  color: var(--white);
}

.actions-row {
  display: flex;
  gap: .5rem;
  flex-wrap: wrap;
  margin-top: .5rem;
}

.status-badge {
  font-size: .7rem;
  padding: .25rem .6rem;
  border-radius: 999px;
  text-transform: uppercase;
  letter-spacing: .05em;
  white-space: nowrap;
  align-self: flex-start;
}

.status-running { background-color: rgba(36, 129, 30, .25); color: #6ee07a; }
.status-stopped { background-color: rgba(255, 255, 255, .08); color: var(--white-half); }
.status-error { background-color: rgba(255, 80, 80, .15); color: #ff6b6b; }
</style>
