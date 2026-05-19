<template>
  <div class="share-download-root">
    <div class="share-download-header">
      <NavigationBar />
    </div>

    <!-- Loading State -->
    <div v-if="loading" class="state-container">
      <div class="loading-spinner"></div>
      <p>{{ $t('share.loading') }}</p>
    </div>

    <!-- Error State -->
    <div v-else-if="error" class="state-container">
      <img :src="getAppIcon('alert-circle')" alt="Error" class="state-icon" />
      <h2>{{ errorTitle }}</h2>
      <p class="state-message">{{ errorMessage }}</p>
    </div>

    <!-- Download Content -->
    <div v-else class="share-download-body">
      <div class="share-download-container">

        <div class="share-info-section">
          <img :src="getAppIcon('send')" alt="Share" class="share-icon" />
          <h1 class="share-label">{{ shareInfo.label }}</h1>
          <p class="share-meta">{{ shareInfo.project_name }} &middot; {{ $t('share.fileCount', { count: files.length }) }} &middot; {{ formattedTotalSize }}</p>
          <p class="share-expiry">{{ $t('share.expiresOn', { date: formattedExpiry }) }}</p>
        </div>

        <div class="files-section">
          <div class="files-header">
            <span class="files-title">{{ $t('share.files') }}</span>
            <button class="download-all-btn" @click="downloadAll" :disabled="downloadingAll || !!downloadingFileId">
              <div v-if="downloadingAll" class="btn-spinner"></div>
              <img v-else :src="getAppIcon('download')" class="btn-icon" />
              <span>{{ downloadingAll ? $t('share.downloading') : $t('share.downloadAll') }}</span>
            </button>
          </div>

          <div class="files-list">
            <div v-for="file in files" :key="file.checkpoint_id" class="file-item">
              <div class="file-info">
                <img :src="file.icon || getAppIcon('file')" class="file-icon" />
                <span class="file-name">{{ file.file_name }}</span>
                <span v-if="downloadingFileId === file.checkpoint_id" class="file-size progress">{{ downloadProgress }}</span>
                <span v-else class="file-size">{{ formatFileSize(file.file_size) }}</span>
              </div>
              <button class="file-download-btn" @click="downloadFile(file)" :disabled="downloadingAll || !!downloadingFileId">
                <div v-if="downloadingFileId === file.checkpoint_id" class="btn-spinner small"></div>
                <img v-else :src="getAppIcon('download')" class="btn-icon" />
              </button>
            </div>
          </div>
        </div>

      </div>
    </div>
  </div>
</template>

<script setup>
// imports
import { computed, onMounted, ref } from 'vue';
import { useRoute } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { GLOBAL_API } from '@/services/adapters/config.js';
import { streamTLVAndDecompress, triggerBrowserDownload } from '@/services/adapters/checkpointservice.js';

// components
import NavigationBar from '@/instances/web/components/NavigationBar.vue';

// stores
import { useIconStore } from '@/stores/icons';

const route = useRoute();
const iconStore = useIconStore();
const { t } = useI18n();

// refs
const downloadingAll = ref(false);
const downloadingFileId = ref(null);
const downloadProgress = ref('');
const error = ref(false);
const errorMessage = ref('');
const errorTitle = ref('');
const files = ref([]);
const loading = ref(true);
const shareInfo = ref({});

// computed properties
// Returns formatted expiry date string.
const formattedExpiry = computed(() => {
  if (!shareInfo.value.expires_at) return '';
  return new Date(shareInfo.value.expires_at).toLocaleDateString();
});

// Returns the total file size formatted as a human-readable string.
const formattedTotalSize = computed(() => {
  const total = files.value.reduce((sum, f) => sum + (f.file_size || 0), 0);
  return formatFileSize(total);
});

// methods
// Downloads all shared files sequentially.
const downloadAll = async () => {
  if (downloadingAll.value || downloadingFileId.value) return;
  downloadingAll.value = true;
  try {
    for (const file of files.value) {
      await downloadSingleFile(file);
    }
  } finally {
    downloadingAll.value = false;
  }
};

// Downloads a single file when clicked individually.
const downloadFile = async (file) => {
  if (downloadingFileId.value || downloadingAll.value) return;
  await downloadSingleFile(file);
};

// Fetches, streams TLV, decompresses, and triggers browser download for a single file.
const downloadSingleFile = async (file) => {
  downloadingFileId.value = file.checkpoint_id;
  downloadProgress.value = '';
  try {
    const url = GLOBAL_API + '/share/' + route.params.token + '/download/' + file.checkpoint_id;
    const response = await fetch(url);
    if (!response.ok) throw new Error('Download failed');
    const onProgress = (receivedBytes, totalBytes) => {
      const received = formatFileSize(receivedBytes);
      const total = totalBytes > 0 ? formatFileSize(totalBytes) : '';
      downloadProgress.value = total ? `${received} / ${total}` : received;
    };
    const blob = await streamTLVAndDecompress(response, onProgress);
    triggerBrowserDownload(blob, file.file_name);
  } finally {
    downloadingFileId.value = null;
    downloadProgress.value = '';
  }
};

// Formats a byte count into a human-readable size string.
const formatFileSize = (bytes) => {
  if (!bytes || bytes === 0) return '0 B';
  const units = ['B', 'KB', 'MB', 'GB'];
  let i = 0;
  let size = bytes;
  while (size >= 1024 && i < units.length - 1) {
    size /= 1024;
    i++;
  }
  return size.toFixed(i === 0 ? 0 : 1) + ' ' + units[i];
};

// Returns the app icon path for the given icon name.
const getAppIcon = (iconName) => {
  return iconStore.getAppIcon(iconName);
};

// Resolves file icons for all files using the icon store.
const processFileIcons = async (fileList) => {
  for (const file of fileList) {
    const ext = (file.file_name || '').split('.').pop()?.toLowerCase() || '';
    file.icon = await iconStore.getIcon(ext) || '';
  }
};

// Fetches share link info from global server, then file metadata from studio server.
const loadShareData = async () => {
  const token = route.params.token;
  if (!token) {
    setError(t('share.invalidLink'), t('share.noToken'));
    return;
  }

  try {
    // Fetch share link info from global server
    const shareResp = await fetch(GLOBAL_API + '/share/' + token);

    if (shareResp.status === 410) {
      const body = await shareResp.json().catch(() => ({}));
      const msg = body.error || body.message || '';
      if (msg.toLowerCase().includes('revoked')) {
        setError(t('share.linkRevoked'), t('share.linkRevokedMessage'));
      } else {
        setError(t('share.linkExpired'), t('share.linkExpiredMessage'));
      }
      return;
    }

    if (!shareResp.ok) {
      setError(t('share.linkNotFound'), t('share.linkNotFoundMessage'));
      return;
    }

    const data = await shareResp.json();
    shareInfo.value = data;

    // Use files from the share link response
    files.value = data.files || [];
    await processFileIcons(files.value);
  } catch (err) {
    setError(t('share.loadError'), t('share.loadErrorMessage'));
  } finally {
    loading.value = false;
  }
};

// Sets error state with title and message.
const setError = (title, message) => {
  error.value = true;
  errorTitle.value = title;
  errorMessage.value = message;
  loading.value = false;
};

// lifecycle hooks
onMounted(() => {
  loadShareData();
});
</script>

<style scoped>
@import "@/assets/desktop.css";

.share-download-root {
  width: 100%;
  min-height: 100vh;
  max-height: 100vh;
  background-color: var(--bg);
  display: flex;
  flex-direction: column;
  align-items: center;
  overflow: hidden;
  overflow-y: auto;
  color: var(--text);
  box-sizing: border-box;
}

.share-download-header {
  width: 100%;
  display: flex;
  align-items: center;
  background-color: rgba(255, 255, 255, 0.05);
  position: sticky;
  top: 0;
  z-index: 99999;
  backdrop-filter: blur(30px);
}

.share-download-body {
  flex: 1;
  width: 100%;
  max-width: 700px;
  padding: 2rem;
  box-sizing: border-box;
  display: flex;
  flex-direction: column;
  align-items: center;
}

.share-download-container {
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: 24px;
}

/* States */
.state-container {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 12px;
  color: var(--surface-5);
}

.state-icon {
  width: 48px;
  height: 48px;
  opacity: 0.6;
}

.state-message {
  color: var(--surface-3);
  max-width: 400px;
  text-align: center;
}

.loading-spinner {
  width: 32px;
  height: 32px;
  border: 3px solid var(--surface-3);
  border-top-color: var(--grape);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

/* Share Info */
.share-info-section {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  padding: 24px 0;
  text-align: center;
}

.share-icon {
  width: 40px;
  height: 40px;
  opacity: 0.8;
}

.share-label {
  font-size: 1.5rem;
  font-weight: 600;
  margin: 0;
}

.share-meta {
  color: var(--surface-5);
  font-size: 0.9rem;
  margin: 0;
}

.share-expiry {
  color: var(--surface-5);
  font-size: 0.85rem;
  margin: 0;
}

/* Files */
.files-section {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.files-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.files-title {
  font-size: 0.9rem;
  font-weight: 500;
  color: var(--surface-5);
}

.download-all-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  background-color: var(--grape);
  border: none;
  border-radius: var(--small-radius);
  color: var(--text);
  cursor: pointer;
  font-size: 0.85rem;
  font-weight: 500;
}

.download-all-btn:hover {
  opacity: 0.9;
}

.download-all-btn:disabled, .file-download-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-spinner {
  width: 16px;
  height: 16px;
  border: 2px solid rgba(255, 255, 255, 0.3);
  border-top-color: var(--text);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

.btn-spinner.small {
  width: 14px;
  height: 14px;
}

.btn-icon {
  width: 16px;
  height: 16px;
  filter: brightness(10);
}

.files-list {
  display: flex;
  flex-direction: column;
  gap: 2px;
  background-color: rgba(255, 255, 255, 0.03);
  border-radius: var(--small-radius);
  overflow: hidden;
}

.file-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 14px;
  background-color: rgba(255, 255, 255, 0.03);
}

.file-item:hover {
  background-color: rgba(255, 255, 255, 0.06);
}

.file-info {
  display: flex;
  align-items: center;
  gap: 10px;
  min-width: 0;
  flex: 1;
}

.file-icon {
  width: 20px;
  height: 20px;
  flex-shrink: 0;
  opacity: 0.7;
}

.file-name {
  font-size: 0.9rem;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.file-size {
  color: var(--surface-5);
  font-size: 0.8rem;
  flex-shrink: 0;
  margin-left: auto;
  padding-right: 12px;
}

.file-size.progress {
  color: var(--grape);
}

.file-download-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  background: none;
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: var(--small-radius);
  cursor: pointer;
  flex-shrink: 0;
}

.file-download-btn:hover {
  background-color: rgba(255, 255, 255, 0.1);
}

.file-download-btn .btn-icon {
  filter: brightness(0.7);
}

/* Scrollbar */
.share-download-root::-webkit-scrollbar {
  width: 10px;
  height: 10px;
}

.share-download-root::-webkit-scrollbar-track {
  background: var(--bg);
  border-radius: 5px;
}

.share-download-root::-webkit-scrollbar-thumb {
  background: var(--surface-3);
  border-radius: 5px;
}

.share-download-root::-webkit-scrollbar-thumb:hover {
  background-color: rgba(255, 255, 255, 0.3);
}

/* Responsive */
@media (max-width: 768px) {
  .share-download-body {
    padding: 1rem;
  }
}
</style>
