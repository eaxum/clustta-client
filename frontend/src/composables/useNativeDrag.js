import { computed, onMounted, ref } from 'vue';
import { usePlatformStore } from '@/stores/platform';
import { useProjectStore } from '@/stores/projects';
import { useStageStore } from '@/stores/stages';
import { useUserStore } from '@/stores/users';
import { useNotificationStore } from '@/stores/notifications';
import { canDragFiles, dragSelection } from '@/lib/nativeDrag';

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

export function useNativeDrag(asset) {
  const platform = usePlatformStore();
  const projects = useProjectStore();
  const stage = useStageStore();
  const users = useUserStore();
  const notifications = useNotificationStore();
  const previewEnabled = import.meta.env.VITE_NATIVE_DRAG_OUT === undefined
    ? import.meta.env.DEV : import.meta.env.VITE_NATIVE_DRAG_OUT === 'true';
  const visible = computed(() => previewEnabled && platform.isDesktop && platform.isWindows);
  const selected = computed(() => dragSelection(asset.value, stage.selectedItems));
  const enabled = computed(() => available.value && !dragging.value && !stage.operationActive
    && users.canDo('pull_chunk') && canDragFiles(selected.value));

  onMounted(async () => {
    if (import.meta.env.VITE_PLATFORM === 'web' || !previewEnabled) return;
    try {
      const service = await loadService();
      capabilityPromise ??= service.Available();
      available.value = await capabilityPromise;
    } catch (error) {
      console.error('Native drag service unavailable:', error);
    }
  });

  async function startDrag(event) {
    event.preventDefault();
    if (!enabled.value) return;
    const project = projects.getActiveProject;
    if (!project) return;
    const request = {
      project_path: project.uri,
      project_id: project.id,
      asset_ids: selected.value.map((item) => item.id),
    };
    dragging.value = true;
    try {
      const service = await loadService();
      await service.StartDrag(request);
    } catch (error) {
      notifications.errorNotification('Could not drag the selected files', error);
    } finally {
      dragging.value = false;
    }
  }

  return { visible, enabled, startDrag };
}
