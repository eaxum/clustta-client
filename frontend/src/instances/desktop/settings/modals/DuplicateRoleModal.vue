<template>
  <div class="modal-container" v-esc="closeModal">
    <HeaderArea :title="$t('modals.duplicateRole')" icon="copy" :showSearch="false" />

    <div class="general-container">
      <div class="input-section">
        <input v-model="roleName" class="input-short" type="text" :placeholder="$t('placeholders.roleName')"
          @keydown.enter="handleEnterKey" v-focus />
        <div v-if="roleNameInUse" class="horizontal-flex input-alert">
          {{ $t('modals.roleNameExists') }}
        </div>
      </div>

      <div class="pop-up-actions">
        <GeneralButton :label="$t('common.cancel')" :fullWidth="true" :buttonFunction="closeModal"
          :isActive="!isAwaitingResponse" :colored="false" />
        <GeneralButton :label="$t('common.duplicate')" :fullWidth="true" @click="duplicateRole"
          :isActive="canDuplicate" :loading="isAwaitingResponse" />
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed, onUnmounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { permissionGroups } from '@/lib/permissions';

import GeneralButton from '@/instances/common/components/GeneralButton.vue';
import HeaderArea from '@/instances/common/components/HeaderArea.vue';

import { UserService } from '@/services';

import { useDesktopModalStore } from '@/stores/desktopModals';
import { useNotificationStore } from '@/stores/notifications';
import { useProjectStore } from '@/stores/projects';
import { useUserStore } from '@/stores/users';

const modals = useDesktopModalStore();
const notificationStore = useNotificationStore();
const projectStore = useProjectStore();
const userStore = useUserStore();

const { t } = useI18n();

const sourceRole = userStore.selectedRole;
const isAwaitingResponse = ref(false);
const roleName = ref(`${sourceRole?.name || ''} copy`);

const normalizedRoleName = computed(() => roleName.value.trim());

const roleNameInUse = computed(() => {
  const candidateName = normalizedRoleName.value.toLowerCase();
  if (!candidateName) return false;
  return userStore.getProjectRoles.some(role => role.name.trim().toLowerCase() === candidateName);
});

const canDuplicate = computed(() => {
  return !!sourceRole
    && userStore.canDo('change_role')
    && normalizedRoleName.value !== ''
    && !roleNameInUse.value
    && !isAwaitingResponse.value;
});

const roleAttributes = () => {
  const permissions = Object.values(permissionGroups).flat();
  return Object.fromEntries(permissions.map(permission => [permission, sourceRole[permission] === true]));
};

const closeModal = () => {
  if (isAwaitingResponse.value) return;
  modals.setModalVisibility('duplicateRoleModal', false);
};

const duplicateRole = async () => {
  if (!canDuplicate.value) return;

  isAwaitingResponse.value = true;
  try {
    const duplicatedRole = await UserService.AddRole(
      projectStore.activeProject.uri,
      normalizedRoleName.value,
      roleAttributes(),
    );
    userStore.roles.push(duplicatedRole);
    notificationStore.addNotification(t('notifications.roleCreated'), '', 'success');
    modals.setModalVisibility('duplicateRoleModal', false);
  } catch (error) {
    notificationStore.errorNotification(t('notifications.errorCreatingRole'), error);
  } finally {
    isAwaitingResponse.value = false;
  }
};

const handleEnterKey = () => {
  if (canDuplicate.value) duplicateRole();
};

onUnmounted(() => {
  userStore.selectedRole = null;
});
</script>

<style scoped>
@import "@/assets/desktop.css";
@import "@/assets/modals.css";

.general-container {
  gap: 1rem;
}

.input-short {
  flex: 1;
  width: 100%;
}
</style>
