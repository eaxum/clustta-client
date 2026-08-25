<template>
  <div class="modal-container" v-esc="closeModal">
    <HeaderArea :title="title" icon="hook" :showSearch="false" />

    <div class="general-container hook-modal-content">
      <div class="hook-form-scroll">
        <div class="input-section">
          <input v-model="hook.name" class="input-short" type="text"
            :placeholder="$t('settings.hookNamePlaceholder')" v-focus />
        </div>

        <div class="input-section">
          <IgnoreListBox :selectedItems="hook.extensions" :placeholder="$t('settings.addHookExtension')"
            @itemAdded="addExtension" @itemRemoved="removeExtension" />
        </div>

        <div class="input-section">
          <input v-model="hook.application_version" class="input-short" type="text"
            :placeholder="$t('settings.hookApplicationVersionPlaceholder')" />
        </div>

        <div class="input-section">
          <DropDownBox :items="availableScriptOptions" :selectedItem="selectedScriptName"
            :onSelect="selectScript" :placeHolder="$t('settings.selectHookScript')" />
        </div>

        <div class="input-section">
          <DropDownBox :items="availableEnvironmentOptions" :selectedItem="''"
            :onSelect="selectEnvironmentVariable" :placeHolder="$t('settings.selectHookEnvironmentVariable')" />

          <div v-if="selectedEnvironmentVariables.length" class="environment-list">
            <div v-for="variable in selectedEnvironmentVariables" :key="variable.id" class="environment-row">
              <div class="selected-environment-details">
                <span>{{ variable.name }}</span>
                <span class="selected-environment-value">{{ variable.value }}</span>
              </div>
              <ActionButton :icon="getAppIcon('trash')"
                :buttonFunction="() => removeEnvironmentVariable(variable.id)" />
            </div>
          </div>
        </div>

        <div class="input-section failure-policy-section">
          <div class="policy-row">
            <span class="input-label">{{ $t('settings.hookFailurePolicy') }}</span>
            <div class="policy-action">
              <DropDownBox :items="failurePolicyNames" :selectedItem="selectedFailurePolicyName"
                :onSelect="selectFailurePolicy" :fixedWidth="true" />
            </div>
          </div>
        </div>
      </div>

      <div class="pop-up-actions">
        <GeneralButton :label="$t('common.close')" :fullWidth="true" :buttonFunction="closeModal"
          :isActive="!isAwaitingResponse" :colored="false" />
        <GeneralButton :label="saveLabel" :fullWidth="true" :buttonFunction="saveHook"
          :isActive="canSave" :loading="isAwaitingResponse" />
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { v4 as uuidv4 } from 'uuid';

import DropDownBox from '@/instances/common/components/DropDownBox.vue';
import GeneralButton from '@/instances/common/components/GeneralButton.vue';
import HeaderArea from '@/instances/common/components/HeaderArea.vue';
import IgnoreListBox from '@/instances/common/components/IgnoreListBox.vue';
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import { AgentService } from '@/services';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useIconStore } from '@/stores/icons';
import { useNotificationStore } from '@/stores/notifications';
import { useProjectStore } from '@/stores/projects';
import { useSettingsStore } from '@/stores/settings';

const modals = useDesktopModalStore();
const iconStore = useIconStore();
const notificationStore = useNotificationStore();
const projectStore = useProjectStore();
const settingsStore = useSettingsStore();
const { t } = useI18n();

const isAwaitingResponse = ref(false);
const availableScripts = ref([]);
const selectedScripts = ref([]);
const isEditing = computed(() => Boolean(settingsStore.selectedPreLaunchHook));
const failurePolicyNames = computed(() => [t('settings.hookBlockLaunch'), t('settings.hookWarnAndContinue')]);
const availableScriptOptions = computed(() => availableScripts.value.map((script) => ({
  id: script.id,
  name: script.name,
  icon: getAppIcon('code-bracket'),
})));
const selectedScriptName = computed(() => selectedScripts.value[0]?.name || '');

const createDefaultHook = () => ({
  id: uuidv4(),
  name: '',
  enabled: true,
  extensions: [],
  application_version: '',
  script_asset_ids: [],
  environment_variable_ids: [],
  failure_policy: 'block',
});

const hook = ref(settingsStore.selectedPreLaunchHook
  ? JSON.parse(JSON.stringify(settingsStore.selectedPreLaunchHook))
  : createDefaultHook());

const title = computed(() => isEditing.value ? t('settings.editLaunchHook') : t('settings.addLaunchHook'));
const saveLabel = computed(() => isEditing.value ? t('common.update') : t('common.add'));
const selectedFailurePolicyName = computed(() => hook.value.failure_policy === 'warn'
  ? t('settings.hookWarnAndContinue')
  : t('settings.hookBlockLaunch'));
const selectedEnvironmentVariables = computed(() => settingsStore.projectEnvironmentVariables.filter(
  (variable) => hook.value.environment_variable_ids.includes(variable.id),
));
const availableEnvironmentOptions = computed(() => settingsStore.projectEnvironmentVariables
  .filter((variable) => !hook.value.environment_variable_ids.includes(variable.id))
  .map((variable) => ({ id: variable.id, name: variable.name, icon: getAppIcon('code-bracket') })));
const canSave = computed(() => (
  Boolean(hook.value.name.trim())
  && hook.value.extensions.length > 0
  && (hook.value.script_asset_ids.length > 0
    || hook.value.environment_variable_ids.length > 0
    || Boolean(hook.value.application_version.trim()))
  && !isAwaitingResponse.value
));

const getAppIcon = (iconName) => iconStore.getAppIcon(iconName);
const normalizeExtension = (extension) => {
  let value = `${extension || ''}`.trim().toLowerCase().replace(/^\*/, '');
  if (value && !value.startsWith('.')) value = `.${value}`;
  return value;
};
const addExtension = (extension) => {
  const value = normalizeExtension(extension);
  if (value && !hook.value.extensions.includes(value)) hook.value.extensions.push(value);
};
const removeExtension = (extension) => {
  hook.value.extensions = hook.value.extensions.filter((value) => value !== extension);
};
const selectFailurePolicy = (name) => {
  hook.value.failure_policy = name === t('settings.hookWarnAndContinue') ? 'warn' : 'block';
};
const selectScript = (scriptName) => {
  const script = availableScripts.value.find((item) => item.name === scriptName);
  selectedScripts.value = script ? [script] : [];
  hook.value.script_asset_ids = script ? [script.id] : [];
};
const selectEnvironmentVariable = (variableName) => {
  const variable = settingsStore.projectEnvironmentVariables.find((item) => item.name === variableName);
  if (variable && !hook.value.environment_variable_ids.includes(variable.id)) {
    hook.value.environment_variable_ids.push(variable.id);
  }
};
const removeEnvironmentVariable = (variableID) => {
  hook.value.environment_variable_ids = hook.value.environment_variable_ids.filter((id) => id !== variableID);
};
const closeModal = () => modals.setModalVisibility('preLaunchHookModal', false);

const saveHook = async () => {
  if (!canSave.value) return;
  isAwaitingResponse.value = true;
  const existingHooks = settingsStore.preLaunchHooks.filter((item) => item.id !== hook.value.id);
  try {
    await settingsStore.savePreLaunchHooks(projectStore.activeProject.uri, [...existingHooks, hook.value]);
    notificationStore.addNotification(
      isEditing.value ? t('notifications.launchHookUpdated') : t('notifications.launchHookCreated'),
      '',
      'success',
    );
    closeModal();
  } catch (error) {
    notificationStore.errorNotification(t('notifications.errorSavingLaunchHook'), error);
  } finally {
    isAwaitingResponse.value = false;
  }
};

onMounted(async () => {
  try {
    const scripts = await AgentService.ListScriptReferences(projectStore.activeProject.uri) || [];
    availableScripts.value = scripts.filter((script) => script.tracked).map((script) => ({
      ...script,
      category: script.path,
    }));
    selectedScripts.value = hook.value.script_asset_ids.slice(0, 1)
      .map((id) => availableScripts.value.find((script) => script.id === id))
      .filter(Boolean);
  } catch (error) {
    notificationStore.errorNotification(t('notifications.errorLoadingHookScripts'), error);
  }
});
</script>

<style scoped>
@import "@/assets/desktop.css";

.general-container {
  width: 500px;
  max-width: calc(100vw - 2rem);
  gap: 1rem;
}

.hook-modal-content {
  max-height: 80vh;
  overflow: hidden;
}

.hook-form-scroll {
  display: flex;
  width: 100%;
  box-sizing: border-box;
  flex-direction: column;
  flex: 1;
  gap: 1rem;
  max-height: 64vh;
  min-height: 0;
  overflow-y: auto;
  padding: .1rem .25rem .25rem;
}

.input-section {
  width: 100%;
  box-sizing: border-box;
  display: flex;
  flex-direction: column;
  gap: .45rem;
  color: var(--text);
}

.policy-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  min-height: 48px;
  gap: 1rem;
  padding: .45rem 0;
}

.policy-action {
  display: flex;
  justify-content: flex-end;
  width: 190px;
}

.input-label {
  color: var(--text);
  font-size: 12px;
  font-weight: 500;
}

.environment-list {
  display: flex;
  flex-direction: column;
  gap: .35rem;
}

.environment-row {
  display: flex;
  align-items: center;
  gap: .4rem;
  padding: .35rem;
  background-color: var(--surface-2);
  border-radius: var(--small-radius);
}

.selected-environment-details {
  display: flex;
  min-width: 0;
  flex: 1;
  flex-direction: column;
  gap: .15rem;
}

.selected-environment-value {
  overflow: hidden;
  color: var(--text-muted);
  font-size: 11px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.input-short {
  width: 100%;
  box-sizing: border-box;
}
</style>
