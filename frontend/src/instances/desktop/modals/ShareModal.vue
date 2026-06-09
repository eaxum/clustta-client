<template>
  <div ref="modalContainer" class="modal-container" v-stop-propagation>
    <HeaderArea :title="$t('modals.shareLink')" :icon="getAppIcon('send')" :showSearch="false" />
    <div class="general-container">

      <div v-if="!shareUrl" class="share-form">
        <div class="input-section">
          <div class="compound-input-section">
            <input v-model="label" class="input-short" type="text" :placeholder="$t('modals.shareLabelPlaceholder')" v-focus />
          </div>
        </div>

        <div class="input-section">
          <label class="input-label">{{ $t('modals.shareExpiry') }}</label>
          <DropDownBox :items="expiryOptions" :selectedItem="selectedExpiryLabel" :onSelect="selectExpiry" :useFilter="false" />
        </div>

        <div class="pop-up-actions">
          <GeneralButton :label="$t('common.cancel')" :fullWidth="true" :buttonFunction="closeModal" :colored="false" />
          <GeneralButton :label="$t('modals.generateLink')" :fullWidth="true" @click="generateLink" :isActive="isFormValid" :loading="isLoading" />
        </div>
      </div>

      <div v-else class="share-result">
        <div class="success-message">
          <div>{{ $t('modals.shareLinkReady') }}</div>
        </div>

        <div class="menu-divider"></div>

        <div class="share-link-container">
          <div class="share-link-header">
            <div class="share-info-label">{{ $t('modals.shareLink') }}</div>
            <ActionButton :icon="getAppIcon('copy')" :buttonFunction="copyLink" :label="linkCopied ? $t('common.copied') : $t('common.copy')" :showLabel="true" />
          </div>
          <input :value="shareUrl" readonly class="share-link-input" @click="selectShareUrl" />
        </div>

        <div class="share-expiry-info">
          <span class="share-info-label">{{ $t('modals.shareLinkExpires', { date: expiresAtFormatted }) }}</span>
        </div>

        <div class="pop-up-actions single-action">
          <GeneralButton :label="$t('common.close')" :fullWidth="true" :buttonFunction="closeModal" :colored="false" />
        </div>
      </div>

    </div>
  </div>
</template>

<script setup>
// imports
import { computed, onMounted, ref, watchEffect } from 'vue';
import { useI18n } from 'vue-i18n';

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import DropDownBox from '@/instances/common/components/DropDownBox.vue';
import GeneralButton from '@/instances/common/components/GeneralButton.vue';
import HeaderArea from '@/instances/common/components/HeaderArea.vue';

// services
import { ClipboardService, ShareService } from '@/services';

// stores
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useIconStore } from '@/stores/icons';
import { useMenu } from '@/stores/menu';
import { useNotificationStore } from '@/stores/notifications';
import { useProjectStore } from '@/stores/projects';
import { useTrayStates } from '@/stores/TrayStates';

const iconStore = useIconStore();
const modals = useDesktopModalStore();
const notificationStore = useNotificationStore();
const projectStore = useProjectStore();
const trayStates = useTrayStates();
const menu = useMenu();
const { t } = useI18n();

// refs
const expiresAt = ref('');
const expiresInHours = ref(168);
const isLoading = ref(false);
const label = ref('');
const linkCopied = ref(false);
const modalContainer = ref(null);
const shareUrl = ref('');

// computed properties
// Returns the expiry dropdown options.
const expiryOptions = computed(() => [
  { name: t('modals.shareExpiry1Day'), value: 24 },
  { name: t('modals.shareExpiry3Days'), value: 72 },
  { name: t('modals.shareExpiry7Days'), value: 168 },
  { name: t('modals.shareExpiry30Days'), value: 720 }
]);

// Returns the formatted expiry date string.
const expiresAtFormatted = computed(() => {
  if (!expiresAt.value) return '';
  return new Date(expiresAt.value).toLocaleDateString();
});

// Returns whether the form is valid for submission.
const isFormValid = computed(() => {
  return label.value.trim() !== '' && shareData.value && shareData.value.checkpointIds.length > 0;
});

// Returns the label of the currently selected expiry option.
const selectedExpiryLabel = computed(() => {
  const option = expiryOptions.value.find(o => o.value === expiresInHours.value);
  return option ? option.name : '';
});

// Returns the share modal data from tray states.
const shareData = computed(() => trayStates.shareModalData);

// methods
// Closes the modal and resets state.
const closeModal = () => {
  modals.setModalVisibility('shareModal', false);
  trayStates.shareModalData = null;
};

// Copies the share URL to clipboard.
const copyLink = async () => {
  try {
    await ClipboardService.WriteText(shareUrl.value);
    linkCopied.value = true;
    setTimeout(() => { linkCopied.value = false; }, 2000);
  } catch (error) {
    notificationStore.addNotification(t('modals.shareLinkCopyError'), error.message, 'error', false);
  }
};

// Generates a share link via the backend service.
const generateLink = async () => {
  if (!isFormValid.value) return;
  isLoading.value = true;
  try {
    const result = await ShareService.CreateShareLink(
      projectStore.selectedStudio.id,
      projectStore.activeProject.name,
      shareData.value.checkpointIds,
      label.value.trim(),
      expiresInHours.value
    );
    shareUrl.value = result.share_url;
    expiresAt.value = result.expires_at;
  } catch (error) {
    notificationStore.addNotification(t('modals.shareError'), error.message, 'error', false);
  } finally {
    isLoading.value = false;
  }
};

// Returns the app icon path for the given icon name.
const getAppIcon = (iconName) => {
  return iconStore.getAppIcon(iconName);
};

// Selects an expiry option from the dropdown.
const selectExpiry = (selectedName) => {
  const option = expiryOptions.value.find(o => o.name === selectedName);
  if (option) expiresInHours.value = option.value;
};

// Selects the share URL input text on click.
const selectShareUrl = (event) => {
  event.target.select();
};

// watchers
watchEffect(() => {
  if (modalContainer.value) {
    menu.clickOutsideMask = modalContainer.value;
  }
});

// lifecycle hooks
onMounted(() => {
  menu.clickOutsideMask = null;
  if (shareData.value && shareData.value.label) {
    label.value = shareData.value.label;
  }
});
</script>

<style scoped>
@import "@/assets/desktop.css";

.share-form,
.share-result {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  width: 100%;
}

.success-message {
  font-size: 0.875rem;
  font-weight: 400;
  color: hsl(var(--foreground));
}

.share-link-container {
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.share-link-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.share-info-label {
  font-size: 0.75rem;
  font-weight: 500;
  color: hsl(var(--muted-foreground));
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.share-link-input {
  width: 100%;
  background-color: transparent;
  color: hsl(var(--foreground));
  border: 1px solid hsl(var(--input));
  border-radius: calc(var(--radius) - 2px);
  padding: 0.5rem 0.75rem;
  font-family: 'Courier New', monospace;
  font-size: 0.875rem;
  font-weight: 500;
  outline: none;
  cursor: pointer;
  box-sizing: border-box;
  transition: border-color 0.15s ease;
}

.share-link-input:focus {
  border-color: hsl(var(--ring));
  box-shadow: 0 0 0 1px hsl(var(--ring));
}

.share-expiry-info {
  padding: 4px 0;
}

.single-action {
  justify-content: flex-end;
}
</style>
