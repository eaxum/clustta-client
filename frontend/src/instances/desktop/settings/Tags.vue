<template>
  <div class="settings-component-root">
    <div class="settings-component-container">
      <ActionBar :itemType="$t('settings.addTag')" :addFunction="addTag" />

      <ScrollList v-if="projectTags.length" :items="projectTags" :useIcons="true" :useItemId="true"
        :wrapItems="true" :editItems="true" :editListItem="prepEditTag" :deleteItems="true"
        :deleteListItem="confirmDeleteTag" />

      <PageState v-else :message="$t('settings.noTags')" illustration="/page-states/resources.png"
        :secondaryIcon="getAppIcon('plus-circle')" :secondaryActionMessage="$t('settings.addTag')"
        :secondaryActionFunction="addTag" />
    </div>
  </div>
</template>

<script setup>
import { computed, onMounted } from 'vue';
import { useI18n } from 'vue-i18n';

import PageState from '@/instances/common/components/PageState.vue';
import ActionBar from '@/instances/desktop/components/ActionBar.vue';
import ScrollList from '@/instances/desktop/components/ScrollList.vue';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useIconStore } from '@/stores/icons';
import { useNotificationStore } from '@/stores/notifications';
import { useTagStore } from '@/stores/tags';
import { useTrayStates } from '@/stores/TrayStates';

const iconStore = useIconStore();
const modals = useDesktopModalStore();
const notificationStore = useNotificationStore();
const tagStore = useTagStore();
const trayStates = useTrayStates();
const { t } = useI18n();

const projectTags = computed(() => [...tagStore.tags]
  .sort((firstTag, secondTag) => firstTag.name.localeCompare(secondTag.name))
  .map((tag) => ({ ...tag, icon: 'tag', can_delete: true, can_edit: true })));

const getAppIcon = (iconName) => iconStore.getAppIcon(iconName);
const addTag = () => modals.setModalVisibility('addTagModal', true);
const prepEditTag = (tagId) => {
  tagStore.selectedTag = tagStore.tags.find((tag) => tag.id === tagId) || null;
  if (tagStore.selectedTag) modals.setModalVisibility('editTagModal', true);
};

const deleteTag = async (tagId) => {
  try {
    await tagStore.deleteTag(tagId);
    notificationStore.addNotification(t('notifications.tagDeleted'), '', 'success');
  } catch (error) {
    notificationStore.errorNotification(t('notifications.errorDeletingTag'), error);
    throw error;
  }
};

const confirmDeleteTag = async (tagId) => {
  const tag = tagStore.tags.find((item) => item.id === tagId);
  if (!tag) return;

  try {
    const usageCount = await tagStore.getTagUsageCount(tagId);
    trayStates.dangerousActionTitle = t('settings.deleteTagTitle', { name: tag.name });
    trayStates.dangerousActionMessage = t('settings.deleteTagMessage', { count: usageCount });
    trayStates.dangerousActionIcon = 'trash';
    trayStates.dangerousActionConfirmLabel = t('common.delete');
    trayStates.dangerousActionShowInput = false;
    trayStates.dangerousActionFunction = () => deleteTag(tagId);
    modals.setModalVisibility('confirmDangerousActionModal', true);
  } catch (error) {
    notificationStore.errorNotification(t('notifications.errorLoadingTagUsage'), error);
  }
};

onMounted(async () => {
  try {
    await tagStore.reloadTags();
  } catch (error) {
    notificationStore.errorNotification(t('notifications.errorLoadingTags'), error);
  }
});
</script>

<style scoped>
.settings-component-root {
  width: 100%;
  height: 100%;
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
  width: 96%;
  overflow: hidden;
  gap: .5rem;
  align-items: center;
  justify-content: space-between;
  padding: 1rem;
  color: white;
  background-color: var(--surface-1);
  border-radius: var(--very-large-radius);
  border-radius: var(--gigantic-radius);
}
</style>
