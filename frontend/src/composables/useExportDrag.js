import { computed, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { usePlatformStore } from '@/stores/platform';
import { useProjectStore } from '@/stores/projects';
import { useStageStore } from '@/stores/stages';
import { useNotificationStore } from '@/stores/notifications';

const available = ref(false);
const dragging = ref(false);
let servicePromise;
let capabilityPromise;

function loadService() {
  if (!servicePromise) {
    servicePromise = import('../../bindings/clustta/services/dragoutservice.js');
  }
  return servicePromise;
}

export function useExportDrag(assets, exportable) {
  const { t } = useI18n();
  const platform = usePlatformStore();
  const projects = useProjectStore();
  const stage = useStageStore();
  const notifications = useNotificationStore();
  const visible = computed(() => platform.isDesktop && available.value);
  const enabled = computed(() => available.value && !dragging.value && !stage.operationActive
    && exportable.value);

  onMounted(async () => {
    if (import.meta.env.VITE_PLATFORM === 'web' || !platform.isDesktop) return;
    try {
      const service = await loadService();
      capabilityPromise ??= service.Available();
      available.value = await capabilityPromise;
    } catch (error) {
      console.error('Native drag service unavailable:', error);
    }
  });

  async function startExportDrag(event) {
    event.preventDefault();
    if (!enabled.value) return;
    const project = projects.getActiveProject;
    if (!project) return;
    const request = {
      project_path: project.uri,
      project_id: project.id,
      asset_ids: assets.value.map((item) => item.id),
    };
    dragging.value = true;
    try {
      const service = await loadService();
      await service.StartDrag(request);
    } catch (error) {
      notifications.errorNotification(t('blocks.exportDragFailed'), error);
    } finally {
      dragging.value = false;
    }
  }

  return { visible, enabled, startExportDrag };
}
