<template>
  <div class="settings-component-root">
    <div class="settings-component-container">
      <div class="asset-header">
        <ActionButton
          :icon="getAppIcon(refreshing ? 'loading' : 'refresh')"
          :isLoading="refreshing"
          :label="$t('common.refresh')"
          :buttonFunction="refresh"
          v-tooltip="$t('common.refresh')"
        />
        <div class="search-bar">
          <SearchBar v-model="searchQuery" :placeholder="$t('settings.searchProjects')" @clear="searchQuery = ''" />
        </div>
      </div>

      <AssetListSkeleton v-if="refreshing && !projects.length" />
      <div v-else-if="filteredProjects.length" class="project-list-container">
        <div class="project-list">
          <div
            v-for="(project, index) in filteredProjects"
            :key="project.id || project.uri || project.name"
            class="project-item-main"
            :style="{ animationDelay: index < 12 ? `${index * 0.03}s` : '0s' }"
          >
            <div class="project-item-spacer">
              <img class="project-icon" :src="getAppIcon('briefcase')" alt="">
            </div>
            <div class="project-item-root">
              <div class="project-item-container">
                <div class="project-item-content">
                  <div class="project-item-details">
                    <div class="project-item-name">{{ project.name }}</div>
                    <div class="project-item-meta">
                      <template v-if="conversionFor(project)">
                        {{ $t('settings.currentStorageMode', { mode: storageModeLabel(conversionFor(project).current_mode) }) }}
                        <span v-if="conversionFor(project).required_bytes > 0">
                          · {{ $t('settings.storageRequired', { size: utils.formatBytes(conversionFor(project).required_bytes, 1) }) }}
                        </span>
                      </template>
                      <template v-else>{{ project.is_closed ? $t('settings.archiveProject') : project.id }}</template>
                    </div>
                    <div v-if="conversionFor(project)?.status === 'running'" class="storage-progress-track">
                      <div class="storage-progress-value" :style="{ width: storageProgress(conversionFor(project)) + '%' }"></div>
                    </div>
                    <div v-if="conversionFor(project)?.error" class="storage-error">{{ conversionFor(project).error }}</div>
                  </div>
                </div>

                <div class="project-item-actions">
                  <ActionButton
                    v-if="canShowConversion(project)"
                    :icon="getAppIcon('refresh')"
                    :isDisabled="!canConvertStorage(conversionFor(project))"
                    @click="confirmStorageConversion(conversionFor(project))"
                    v-tooltip="storageActionLabel(conversionFor(project))"
                  />
                  <ActionButton
                    v-if="studioStore.canManageProject"
                    :icon="getAppIcon(project.is_closed ? 'unarchive' : 'archive')"
                    @click="confirmArchive(project)"
                    v-tooltip="project.is_closed ? $t('settings.unarchiveProject') : $t('settings.archiveProject')"
                  />
                  <ActionButton
                    v-if="studioStore.canManageProject"
                    :icon="getAppIcon('trash')"
                    @click="confirmDelete(project)"
                    v-tooltip="$t('settings.deleteProject')"
                  />
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
      <PageState v-else :message="searchQuery ? $t('settings.noProjectsMatch') : $t('settings.noStorageConversions')" :illustration="'/page-states/resources.png'" />
    </div>
  </div>
</template>

<script setup>
import { computed, onBeforeUnmount, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { ProjectService, StudioService } from '@/services';
import utils from '@/services/utils';
import { refreshEntitlements } from '@/lib/sync';
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import AssetListSkeleton from '@/instances/desktop/components/AssetListSkeleton.vue';
import SearchBar from '@/instances/desktop/components/SearchBar.vue';
import PageState from '@/instances/common/components/PageState.vue';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useIconStore } from '@/stores/icons';
import { useNotificationStore } from '@/stores/notifications';
import { useProjectStore } from '@/stores/projects';
import { useStudioStore } from '@/stores/studio';
import { useTrayStates } from '@/stores/TrayStates';

const { t } = useI18n();
const iconStore = useIconStore();
const modals = useDesktopModalStore();
const notificationStore = useNotificationStore();
const projectStore = useProjectStore();
const studioStore = useStudioStore();
const trayStates = useTrayStates();

const searchQuery = ref('');
const refreshing = ref(false);
const storageConversions = ref([]);
let storagePoll = null;
let pollingEnabled = true;

const projects = computed(() => projectStore.projects || []);
const filteredProjects = computed(() => {
  const query = searchQuery.value.trim().toLowerCase();
  if (!query) return projects.value;
  return projects.value.filter((project) =>
    [project.name, project.id, project.uri].some((value) => String(value || '').toLowerCase().includes(query))
  );
});
const storageCapabilities = computed(() => projectStore.selectedStudio?.capabilities?.project_storage || null);
const conversionSupported = computed(() =>
  projectStore.selectedStudio?.hosting_mode !== 'cloud' &&
  projectStore.selectedStudio?.name !== 'Personal' &&
  studioStore.isStudioAdmin &&
  storageCapabilities.value?.conversion_supported === true
);
const deflatedAvailable = computed(() => storageCapabilities.value?.available_modes?.includes('deflated'));

const getAppIcon = (name) => iconStore.getAppIcon(name);
const conversionFor = (project) => storageConversions.value.find((item) =>
  item.project_name === project.name || item.project_id === project.id
);
const canShowConversion = (project) => conversionSupported.value && !!conversionFor(project);
const storageModeLabel = (mode) => mode === 'deflated' ? t('settings.deflatedMode') : t('settings.compactMode');
const storageTargetMode = (conversion) =>
  ['failed', 'cleanup_failed'].includes(conversion.status)
    ? conversion.target_mode
    : conversion.current_mode === 'compact' ? 'deflated' : 'compact';
const canConvertStorage = (conversion) => {
  if (!conversion || conversion.status === 'running') return false;
  const target = storageTargetMode(conversion);
  return target === 'compact' || deflatedAvailable.value;
};
const storageActionLabel = (conversion) => {
  if (conversion.status === 'running') return t('settings.storageConverting');
  if (['failed', 'cleanup_failed'].includes(conversion.status)) return t('common.retry');
  return t('settings.convertStorageTo', { mode: storageModeLabel(storageTargetMode(conversion)) });
};
const storageProgress = (conversion) =>
  conversion?.total_chunks ? Math.min(100, (conversion.processed_chunks / conversion.total_chunks) * 100) : 0;

const projectUrl = async (project) => {
  if (project.remote) return project.remote;
  const studioUrl = await projectStore.resolveStudioUrl();
  return `${studioUrl}/${project.name}`;
};

const fetchStorageConversions = async () => {
  if (!conversionSupported.value) {
    storageConversions.value = [];
    return;
  }
  try {
    storageConversions.value = await StudioService.GetStorageConversions(await projectStore.resolveStudioUrl()) || [];
  } catch (error) {
    console.error('Failed to load project storage conversions:', error);
  } finally {
    scheduleStoragePoll();
  }
};

const refresh = async () => {
  if (refreshing.value) return;
  refreshing.value = true;
  try {
    await projectStore.loadProjects();
    await fetchStorageConversions();
  } catch (error) {
    notificationStore.errorNotification(t('common.error'), error);
  } finally {
    refreshing.value = false;
  }
};

const startStorageConversion = async (conversion) => {
  await StudioService.StartStorageConversion(
    await projectStore.resolveStudioUrl(),
    conversion.project_name,
    storageTargetMode(conversion),
    '',
  );
  await fetchStorageConversions();
};

const confirmStorageConversion = (conversion) => {
  if (!canConvertStorage(conversion)) return;
  const target = storageTargetMode(conversion);
  trayStates.dangerousActionTitle = t('settings.storageConversionTitle', { project: conversion.project_name });
  trayStates.dangerousActionMessage = t('settings.confirmStorageConversion', {
    project: conversion.project_name,
    source: storageModeLabel(conversion.current_mode),
    target: storageModeLabel(target),
  });
  trayStates.dangerousActionIcon = 'refresh';
  trayStates.dangerousActionConfirmText = conversion.project_name;
  trayStates.dangerousActionShowInput = true;
  trayStates.dangerousActionInputSecret = false;
  trayStates.dangerousActionRequireExactInput = true;
  trayStates.dangerousActionShowToggle = false;
  trayStates.dangerousActionFunction = () => startStorageConversion(conversion);
  modals.setModalVisibility('confirmDangerousActionModal', true);
};

const toggleArchive = async (project) => {
  await ProjectService.ToggleCloseProject(await projectUrl(project), projectStore.selectedStudio.name);
  project.is_closed = !project.is_closed;
  modals.setModalVisibility('popUpModal', false);
};

const confirmArchive = (project) => {
  if (project.is_closed) {
    toggleArchive(project);
    return;
  }
  trayStates.popUpModalTitle = t('menus.archiveProjectTitle', { name: project.name });
  trayStates.popUpModalMessage = t('confirmations.archiveProject');
  trayStates.popUpModalFunction = () => toggleArchive(project);
  trayStates.popUpModalIcon = 'archive';
  modals.setModalVisibility('popUpModal', true);
};

const deleteProject = async (project) => {
  await ProjectService.DeleteRemoteProject(await projectUrl(project), projectStore.selectedStudio.name);
  if (project.uri) {
    projectStore.removeProjectFromList(project.uri, { force: true });
  } else {
    projectStore.projects = projectStore.projects.filter((item) => item.id !== project.id);
  }
  await refreshEntitlements();
  notificationStore.addNotification(
    t('notifications.projectDeleted'),
    t('notifications.projectDeletedDesc', { name: project.name }),
    'success',
    false,
  );
};

const confirmDelete = (project) => {
  trayStates.dangerousActionTitle = t('menus.deleteProjectTitle', { name: project.name });
  trayStates.dangerousActionMessage = t('confirmations.deleteRemoteProject', { name: project.name });
  trayStates.dangerousActionIcon = 'trash';
  trayStates.dangerousActionConfirmText = project.name;
  trayStates.dangerousActionShowInput = true;
  trayStates.dangerousActionInputSecret = false;
  trayStates.dangerousActionRequireExactInput = true;
  trayStates.dangerousActionShowToggle = false;
  trayStates.dangerousActionFunction = () => deleteProject(project);
  modals.setModalVisibility('confirmDangerousActionModal', true);
};

const scheduleStoragePoll = () => {
  if (!pollingEnabled || !storageConversions.value.some(({ status }) => status === 'running')) return;
  if (storagePoll) window.clearTimeout(storagePoll);
  storagePoll = window.setTimeout(fetchStorageConversions, 1000);
};

onMounted(async () => {
  if (projectStore.selectedStudio?.hosting_mode !== 'cloud' && projectStore.selectedStudio?.name !== 'Personal') {
    try {
      await projectStore.ensureStudioCapabilities();
    } catch (error) {
      console.warn('Could not load Studio capabilities:', error);
    }
  }
  if (!projects.value.length) await refresh();
  else await fetchStorageConversions();
});

onBeforeUnmount(() => {
  pollingEnabled = false;
  if (storagePoll) window.clearTimeout(storagePoll);
});
</script>

<style scoped>
@import "@/assets/desktop.css";

.settings-component-root { width: 100%; height: 100%; overflow: hidden; display: flex; }
.settings-component-container { display: flex; flex-direction: column; width: 96%; height: 100%; margin: auto; padding: 1rem; gap: .5rem; box-sizing: border-box; overflow: hidden; color: var(--text); border-radius: var(--gigantic-radius); background: var(--surface-1); }
.asset-header { display: flex; width: 100%; align-items: center; justify-content: space-between; gap: 1rem; padding: .2rem; box-sizing: border-box; }
.search-bar { display: flex; flex: 1; max-width: 40%; min-width: 240px; padding: .2rem; }
.project-list-container { width: 100%; height: 100%; overflow-y: auto; padding: .4rem; box-sizing: border-box; }
.project-list-container::-webkit-scrollbar { width: 8px; }
.project-list-container::-webkit-scrollbar-thumb { border-radius: 10px; background: var(--surface-2); }
.project-list { display: grid; grid-template-columns: repeat(auto-fill, minmax(500px, 1fr)); gap: 10px; width: 100%; }
.project-item-main { display: flex; align-items: flex-start; width: 100%; min-width: 0; padding-left: .5rem; overflow: hidden; box-sizing: border-box; color: var(--text); background: var(--surface-2); border-radius: var(--large-radius); outline: var(--transparent-line); outline-offset: -1px; transition: all .2s ease-out; }
.project-item-main:hover { background: var(--surface-3); border-radius: var(--small-radius); outline: 1px solid var(--surface-4); }
.project-item-spacer { width: 36px; height: 60px; flex: 0 0 36px; display: flex; align-items: center; justify-content: center; }
.project-icon { width: 24px; height: 24px; object-fit: contain; }
.project-item-root { width: 100%; min-width: 0; padding: .3rem 0 .3rem .3rem; box-sizing: border-box; }
.project-item-container { display: flex; align-items: center; justify-content: space-between; width: 100%; min-height: 50px; gap: .5rem; padding: .2rem .4rem; box-sizing: border-box; }
.project-item-content { flex: 1; min-width: 0; overflow: hidden; }
.project-item-details { display: flex; flex-direction: column; gap: .15rem; min-width: 0; padding: .2rem; }
.project-item-name { overflow: hidden; font-size: 14px; font-weight: 400; white-space: nowrap; text-overflow: ellipsis; }
.project-item-meta { overflow: hidden; color: var(--text); opacity: .5; font-size: 12px; white-space: nowrap; text-overflow: ellipsis; }
.project-item-actions { display: flex; align-items: center; justify-content: flex-end; gap: .25rem; max-width: 0; opacity: 0; overflow: hidden; transform: translateX(.5rem); transition: max-width .2s ease-in-out, opacity .2s ease-out, transform .2s ease-out; }
.project-item-main:hover .project-item-actions { max-width: 132px; min-width: 116px; opacity: 1; transform: translateX(0); }
.storage-progress-track { height: 4px; margin-top: .35rem; overflow: hidden; border-radius: 4px; background: var(--surface-4); }
.storage-progress-value { height: 100%; background: var(--accent); transition: width .2s ease; }
.storage-error { margin-top: .2rem; color: var(--warning); font-size: .72rem; }
</style>
