<template>
  <div class="modal-container image-viewer-modal-container" v-stop-propagation>
    <div class="image-viewer-modal-header">
      <div class="image-viewer-modal-title">
        <img class="image-viewer-modal-icon" :src="getAppIcon('image')">
        <span>{{ title }}</span>
      </div>
      <ActionButton :icon="getAppIcon('close')" :showLabel="false" v-tooltip="$t('common.close')" :buttonFunction="closeModal" />
    </div>

    <div class="image-viewer-modal-body">
      <img v-if="displaySrc" class="image-viewer-image" :src="displaySrc" :alt="title">
    </div>
  </div>
</template>

<script setup>
// imports
import { computed, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue';

// services
import { FSService } from '@/services';

// stores
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useIconStore } from '@/stores/icons';

const iconStore = useIconStore();
const modals = useDesktopModalStore();
const { t } = useI18n();

// MIME types for image formats the webview can render natively.
const renderableMimeTypes = {
  '.png': 'image/png',
  '.jpg': 'image/jpeg',
  '.jpeg': 'image/jpeg',
  '.gif': 'image/gif',
  '.bmp': 'image/bmp',
  '.webp': 'image/webp',
  '.svg': 'image/svg+xml',
};

// refs
const fullResSrc = ref('');

// computed properties
const thumbnailSrc = computed(() => modals.imageViewer.src);

const displaySrc = computed(() => fullResSrc.value || thumbnailSrc.value);

const title = computed(() => modals.imageViewer.title || t('menus.imageViewer'));

// methods

// Closes the image viewer modal.
const closeModal = () => {
  modals.setModalVisibility('imageViewerModal', false);
};

// Returns the app icon path for the given icon name.
const getAppIcon = (iconName) => iconStore.getAppIcon(iconName);

// Loads the full-resolution image from disk for web-renderable formats.
const loadFullResImage = async () => {
  fullResSrc.value = '';
  const { filePath, extension } = modals.imageViewer;
  const mimeType = renderableMimeTypes[(extension || '').toLowerCase()];
  if (!filePath || !mimeType) return;

  try {
    if (!(await FSService.Exists(filePath))) return;
    const base64 = await FSService.ReadFile(filePath);
    if (base64 && base64.length > 0) {
      fullResSrc.value = `data:${mimeType};base64,${base64}`;
    }
  } catch (error) {
    console.debug('Full-resolution image load failed:', error);
  }
};

// watchers
watch(
  () => modals.imageViewer.src,
  () => loadFullResImage(),
  { immediate: true },
);
</script>

<style scoped>
@import "@/assets/desktop.css";

.image-viewer-modal-container {
  padding: 0 .5rem;
  display: flex;
  flex-direction: column;
}

.image-viewer-modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  padding: 0.5rem;
  border-radius: var(--small-radius);
  background-color: var(--bg);
  outline: var(--transparent-line);
}

.image-viewer-modal-title {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 18px;
  font-weight: 500;
  color: var(--text);
}

.image-viewer-modal-icon {
  width: 16px;
  height: 16px;
}

.image-viewer-modal-body {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 80vw;
  height: 80vh;
  min-width: 480px;
  min-height: 480px;
  max-width: 1400px;
  max-height: 85vh;
  min-height: 0;
  overflow: hidden;
  padding: .5rem 0;
}

.image-viewer-image {
  width: 100%;
  height: 100%;
  object-fit: contain;
  border-radius: var(--normal-radius);
}
</style>
