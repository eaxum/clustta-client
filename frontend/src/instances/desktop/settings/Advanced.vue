<template>
  <div class="settings-component-root">
    <div class="settings-component-scroll">
    <div class="settings-component-container">

      <SettingsCard
        v-if="userStore.canDo('manage_project_settings')"
        :title="$t('settings.scripts')"
      >
        <div class="script-settings">
          <label class="script-setting-label" for="agent-script-directory">{{ $t('settings.scriptDirectoryFromProjectRoot') }}</label>
          <input id="agent-script-directory" v-model="scriptDirectory" class="input-short script-directory-input"
            placeholder="Scripts" @keydown.enter.prevent="saveScriptSettings" />
          <div class="settings-body script-setting-help">
            {{ $t('settings.scriptDirectoryHelp') }}
          </div>

          <label class="script-setting-label">{{ $t('settings.allowedScriptExtensions') }}</label>
          <IgnoreListBox :selectedItems="scriptExtensions" :placeholder="$t('settings.addScriptExtension')"
            @itemAdded="addScriptExtension" @itemRemoved="removeScriptExtension" />
          <div class="script-settings-actions">
            <ActionButton :icon="getAppIcon('floppy-disk')" :label="$t('common.save')"
              :buttonFunction="saveScriptSettings" :isDisabled="!scriptSettingsChanged" />
          </div>
        </div>
      </SettingsCard>

      <SettingsCard
        v-if="userStore.canDo('manage_project_settings')"
        :title="$t('settings.environmentVariables')"
        :description="$t('settings.environmentVariablesHelp', { projectRoot: '<ProjectRoot>' })"
        :showContent="!!(environmentVariables.length || environmentVariablesChanged)"
      >
        <template #header-action>
          <ActionButton :icon="getAppIcon('plus-circle')" :label="$t('common.add')"
            :buttonFunction="addEnvironmentVariable" />
        </template>

        <div class="environment-settings">
          <div class="environment-variable-list">
            <div v-for="variable in environmentVariables" :key="variable.id" class="environment-variable-row">
              <input v-model="variable.name" class="input-short environment-variable-name" type="text"
                :placeholder="$t('settings.environmentVariableNamePlaceholder')" />
              <input v-model="variable.value" class="input-short" type="text"
                :placeholder="$t('settings.environmentVariableValuePlaceholder', { projectRoot: '<ProjectRoot>' })" />
              <ActionButton :icon="getAppIcon('trash')"
                :buttonFunction="() => removeEnvironmentVariable(variable.id)" />
            </div>
          </div>
          <div class="environment-settings-actions">
            <ActionButton :icon="getAppIcon('floppy-disk')" :label="$t('common.save')"
              :buttonFunction="saveEnvironmentVariables" :isDisabled="!environmentVariablesChanged" />
          </div>
        </div>
      </SettingsCard>

    </div>
    </div>
  </div>
</template>

<script setup>
// imports
import { computed, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { v4 as uuidv4 } from 'uuid';

// stores
import { useIconStore } from '@/stores/icons';
import { useProjectStore } from '@/stores/projects';
import { useSettingsStore } from '@/stores/settings';
import { useUserStore } from '@/stores/users';
import { useNotificationStore } from '@/stores/notifications';

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import IgnoreListBox from '@/instances/common/components/IgnoreListBox.vue';
import SettingsCard from '@/instances/desktop/settings/components/SettingsCard.vue';

// services
import { SettingsService } from '@/services';

// refs
const iconStore = useIconStore();
const notificationStore = useNotificationStore();
const projectStore = useProjectStore();
const settingsStore = useSettingsStore();
const userStore = useUserStore();
const { t } = useI18n();
const scriptDirectory = ref('Scripts');
const scriptExtensions = ref(['.py']);
const environmentVariables = ref([]);
const savedScriptSettings = ref({ directory: 'Scripts', extensions: ['.py'] });
const savedEnvironmentVariables = ref([]);

// computed
const scriptSettingsChanged = computed(() => {
  return JSON.stringify(normalizedScriptSettings()) !== JSON.stringify(savedScriptSettings.value);
});
const environmentVariablesChanged = computed(() => {
  return JSON.stringify(normalizedEnvironmentVariables(environmentVariables.value)) !==
    JSON.stringify(savedEnvironmentVariables.value);
});

// methods
const getAppIcon = (iconName) => {
  return iconStore.getAppIcon(iconName);
};

const normalizeScriptExtension = (extension) => {
  let normalized = `${extension || ''}`.trim().toLowerCase().replace(/^\*/, '');
  if (!normalized) return '';
  if (!normalized.startsWith('.')) normalized = `.${normalized}`;
  return normalized;
};

const normalizedScriptSettings = () => ({
  directory: scriptDirectory.value.trim(),
  extensions: scriptExtensions.value.map(normalizeScriptExtension).filter(Boolean).sort(),
});

const normalizedEnvironmentVariables = (variables) => {
  return variables.map((variable) => ({
    id: variable.id,
    name: `${variable.name || ''}`.trim(),
    value: `${variable.value || ''}`.trim(),
  }));
};

const addScriptExtension = (extension) => {
  const normalized = normalizeScriptExtension(extension);
  if (normalized && !scriptExtensions.value.includes(normalized)) {
    scriptExtensions.value.push(normalized);
    scriptExtensions.value.sort();
  }
};

const removeScriptExtension = (extension) => {
  scriptExtensions.value = scriptExtensions.value.filter((item) => item !== extension);
};

const loadScriptSettings = async () => {
  const projectPath = projectStore.activeProject?.uri;
  if (!projectPath) return;
  const settings = await SettingsService.GetProjectScriptSettings(projectPath);
  scriptDirectory.value = settings?.directory || 'Scripts';
  scriptExtensions.value = Array.isArray(settings?.extensions) ? settings.extensions : ['.py'];
  savedScriptSettings.value = normalizedScriptSettings();
};

const saveScriptSettings = async () => {
  const projectPath = projectStore.activeProject?.uri;
  if (!projectPath || !scriptSettingsChanged.value) return;
  try {
    await SettingsService.SetProjectScriptSettings(projectPath, scriptDirectory.value, scriptExtensions.value);
    await loadScriptSettings();
    notificationStore.addNotification(t('settings.scriptSettingsSaved'), '', 'success');
  } catch (error) {
    notificationStore.errorNotification(t('settings.scriptSettingsSaveFailed'), error);
  }
};

const loadEnvironmentVariables = async () => {
  const projectPath = projectStore.activeProject?.uri;
  if (!projectPath) return;
  await settingsStore.loadPreLaunchHooks(projectPath);
  environmentVariables.value = JSON.parse(JSON.stringify(settingsStore.projectEnvironmentVariables));
  savedEnvironmentVariables.value = normalizedEnvironmentVariables(environmentVariables.value);
};

const addEnvironmentVariable = () => {
  environmentVariables.value.push({ id: uuidv4(), name: '', value: '' });
};

const removeEnvironmentVariable = (variableID) => {
  environmentVariables.value = environmentVariables.value.filter((variable) => variable.id !== variableID);
};

const saveEnvironmentVariables = async () => {
  const projectPath = projectStore.activeProject?.uri;
  if (!projectPath || !environmentVariablesChanged.value) return;
  try {
    await settingsStore.saveProjectEnvironmentVariables(projectPath, environmentVariables.value);
    environmentVariables.value = JSON.parse(JSON.stringify(settingsStore.projectEnvironmentVariables));
    savedEnvironmentVariables.value = normalizedEnvironmentVariables(environmentVariables.value);
    notificationStore.addNotification(t('settings.environmentVariablesSaved'), '', 'success');
  } catch (error) {
    notificationStore.errorNotification(t('settings.environmentVariablesSaveFailed'), error);
  }
};

// lifecycle hooks
onMounted(async () => {
  try {
    const projectUri = projectStore.activeProject?.uri;
    if (projectUri) {
      await Promise.all([loadScriptSettings(), loadEnvironmentVariables()]);
    }
  } catch (error) {
    console.error(error);
  }
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
  overflow: hidden;
  display: block;
  overflow-y: scroll;
  border-radius: var(--very-large-radius);
}

.script-settings {
  display: flex;
  flex-direction: column;
  gap: .75rem;
  padding: 1rem 1.5rem 1.25rem;
}

.script-setting-label {
  color: var(--text);
  font-size: 12px;
  font-weight: 500;
}

.script-directory-input {
  width: 100%;
  box-sizing: border-box;
}

.script-setting-help {
  padding: 0;
}

.script-settings-actions {
  display: flex;
  justify-content: flex-end;
}

.environment-settings {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  padding: 1rem 1.5rem 1.25rem;
}

.environment-variable-list {
  display: flex;
  flex-direction: column;
  gap: .75rem;
}

.environment-variable-row {
  display: grid;
  grid-template-columns: minmax(140px, 0.4fr) minmax(240px, 1fr) auto;
  align-items: center;
  gap: .75rem;
}

.environment-variable-row .input-short {
  width: 100%;
}

.environment-variable-name {
  min-width: 0;
}

.environment-settings-actions {
  display: flex;
  justify-content: flex-end;
  padding-top: .25rem;
}

@media (max-width: 700px) {
  .environment-variable-row {
    grid-template-columns: 1fr auto;
  }

  .environment-variable-row .input-short:not(.environment-variable-name) {
    grid-column: 1;
  }

  .environment-variable-row > :last-child {
    grid-column: 2;
    grid-row: 1 / span 2;
  }
}

.settings-component-root::-webkit-scrollbar {
  width: 6px;
}

.settings-component-root::-webkit-scrollbar-thumb {
  background-color: var(--bg);
  border-radius: 3px;
}

.settings-component-root::-webkit-scrollbar-track {
  background-color: var(--surface-4);
  border-radius: 3px;
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

</style>
