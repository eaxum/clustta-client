import { computed, onMounted, ref, unref, watch } from 'vue';
import { FSService } from '@/services';
import { useIconStore } from '@/stores/icons';

// Module-scoped cache shared across all callers (key = file_path, value = base64 string).
const thumbnailCache = new Map();

// Returns the appropriate file-type icon name for an asset based on extension.
export const getFileTypeIcon = (asset) => {
  const extension = asset?.extension?.toLowerCase() || '';

  const imageFormats = ['.png', '.exr', '.jpg', '.jpeg', '.gif', '.bmp', '.tiff', '.webp', '.svg'];
  const videoFormats = ['.mp4', '.mkv', '.avi', '.mov', '.wmv', '.flv', '.webm'];
  const audioFormats = ['.mp3', '.wav', '.flac', '.aac', '.ogg'];
  const archiveFormats = ['.zip', '.rar', '.7z', '.tar', '.gz', '.bz2'];
  const textFormats = ['.txt', '.md', '.rtf'];
  const codeFormats = ['.js', '.ts', '.css', '.html', '.vue', '.py', '.java', '.cpp', '.c', '.go', '.rs'];
  const spreadsheetFormats = ['.xls', '.xlsx', '.csv'];
  const presentationFormats = ['.ppt', '.pptx'];
  const wordFormats = ['.doc', '.docx'];

  if (imageFormats.includes(extension)) return 'image';
  if (videoFormats.includes(extension)) return 'video-camera';
  if (audioFormats.includes(extension)) return 'music';
  if (extension === '.pdf') return 'file-pdf';
  if (archiveFormats.includes(extension)) return 'file-zip';
  if (textFormats.includes(extension)) return 'file-text';
  if (codeFormats.includes(extension)) return 'file-code';
  if (spreadsheetFormats.includes(extension)) return 'file-excel';
  if (presentationFormats.includes(extension)) return 'file-powerpoint';
  if (wordFormats.includes(extension)) return 'file-word';
  return 'file';
};

// Resolves a function or ref-like value to its current value.
const resolve = (source) => (typeof source === 'function' ? source() : unref(source));

// useAssetThumbnail returns reactive thumbnail state for an asset.
// - assetSource: a getter (() => asset) or a ref of an asset object.
// - options.enabled: getter or ref controlling whether OS thumbnails are loaded.
// - options.includeAssetIcon: when true, asset.icon is included in the displayThumbnail priority chain.
export function useAssetThumbnail(assetSource, options = {}) {
  const { enabled = () => true, includeAssetIcon = false } = options;
  const iconStore = useIconStore();

  const osThumbnail = ref('');
  const thumbnailLoading = ref(false);

  const displayThumbnail = computed(() => {
    const asset = resolve(assetSource);
    if (!asset) return '';
    if (asset.preview) return asset.preview;
    if (osThumbnail.value) return `data:image/png;base64,${osThumbnail.value}`;
    if (includeAssetIcon && asset.icon) return asset.icon;
    return iconStore.getAppIcon(getFileTypeIcon(asset));
  });

  const isFallbackIcon = computed(() => {
    const asset = resolve(assetSource);
    if (!asset) return true;
    if (asset.preview || osThumbnail.value) return false;
    if (includeAssetIcon && asset.icon) return false;
    return true;
  });

  // Loads OS-generated thumbnail for the asset's file when no embedded preview exists.
  const loadOSThumbnail = async () => {
    const asset = resolve(assetSource);
    if (!asset) return;
    const filePath = asset.file_path;
    if (!resolve(enabled) || asset.preview || !filePath || asset.is_link || thumbnailLoading.value) {
      return;
    }

    if (thumbnailCache.has(filePath)) {
      osThumbnail.value = thumbnailCache.get(filePath);
      return;
    }

    let fileExists = false;
    try {
      fileExists = await FSService.Exists(filePath);
    } catch (error) {
      return;
    }
    if (!fileExists) return;

    thumbnailLoading.value = true;

    try {
      const size = 512;
      let thumbnail = await FSService.GetCachedOSThumbnail(filePath, size);

      if (thumbnail && thumbnail.length > 0) {
        osThumbnail.value = thumbnail;
        thumbnailCache.set(filePath, thumbnail);
        thumbnailLoading.value = false;
        return;
      }

      setTimeout(async () => {
        try {
          thumbnail = await FSService.GetOSThumbnail(filePath, size);
          if (thumbnail && thumbnail.length > 0) {
            osThumbnail.value = thumbnail;
            thumbnailCache.set(filePath, thumbnail);
          }
        } catch (error) {
          console.debug('Thumbnail generation failed:', error);
        } finally {
          thumbnailLoading.value = false;
        }
      }, 0);
    } catch (error) {
      console.debug('Thumbnail loading failed:', error);
      thumbnailLoading.value = false;
    }
  };

  watch(() => resolve(enabled), (isEnabled) => {
    if (isEnabled) loadOSThumbnail();
  });

  watch(() => {
    const asset = resolve(assetSource);
    return [asset?.file_path, asset?.preview];
  }, () => {
    osThumbnail.value = '';
    loadOSThumbnail();
  });

  onMounted(() => {
    loadOSThumbnail();
  });

  return {
    osThumbnail,
    thumbnailLoading,
    displayThumbnail,
    isFallbackIcon,
    loadOSThumbnail,
    getFileTypeIcon,
  };
}
