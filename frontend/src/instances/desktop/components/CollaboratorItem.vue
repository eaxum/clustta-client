<template>
  <div ref="collaboratorItem" class="collaborator-item-main" :class="{ 'compact-mode': compact, 'compact-editing': compact && isEditing }" v-esc="handleEscKey" v-stop-propagation>
    <div class="collaborator-item-spacer">
      <ProfilePhoto :assigneeId="collaborator.id" :userPhoto="collaborator.photo" />
    </div>

    <div class="collaborator-item-root">
      <div class="collaborator-item-container">

        <div v-if="!compact || !isEditing" class="collaborator-item-content">
          <div class="collaborator-item-details">
            <div class="collaborator-item-name">{{ userFullName }}</div>
            <div class="collaborator-item-email">{{ collaborator.email }}</div>
          </div>
        </div>

        <div v-if="!compact && canEditRole" class="collaborator-item-dropdown">
          <DropDownBox :selectedItem="collaborator.role_name || collaborator.role?.name" :items="collaboratorRoles" :onSelect="selectRole" />
        </div>

        <div v-if="compact && isEditing && canEditRole" class="collaborator-item-dropdown compact-dropdown">
          <DropDownBox :selectedItem="collaborator.role_name || collaborator.role?.name" :items="collaboratorRoles" :onSelect="selectRole" />
        </div>

        <div v-if="!compact" class="collaborator-item-actions">
          <ActionButton v-if="canDeleteUser" :icon="getAppIcon('person-minus')" @click="deleteCollaborator(collaborator.id)" v-tooltip="$t('components.collaboratorItem.remove')" />
          <ActionButton :icon="getAppIcon('person-search')" @click="openUserProfile" v-tooltip="$t('components.collaboratorItem.viewProfile')" />
        </div>

        <div v-if="compact && !isEditing && isProjectMember" class="collaborator-item-actions compact-actions-container">
          <div class="compact-role-meta" v-tooltip="displayRole">{{ displayRole }}</div>
          <div class="compact-hover-actions">
            <ActionButton v-if="canEditRole" :icon="getAppIcon('edit')" @click="startEditing" v-tooltip="$t('components.collaboratorItem.editRole')" />
            <ActionButton v-if="canDeleteUser" :icon="getAppIcon(isLoading ? 'loading' : 'person-minus')" :isLoading="isLoading" @click="deleteCollaborator(collaborator.id)" v-tooltip="isLoading ? $t('components.collaboratorItem.removing') : $t('components.collaboratorItem.remove')" />
          </div>
        </div>

        <div v-if="compact && !isProjectMember" class="collaborator-item-actions compact-actions-container">
          <ActionButton v-if="canEditRole" :icon="getAppIcon(isLoading ? 'loading' : 'person-plus')" :isLoading="isLoading" @click="addToProject" v-tooltip="isLoading ? $t('components.collaboratorItem.adding') : $t('components.collaboratorItem.addToProject')" />
        </div>

      </div>
    </div>
  </div>
</template>

<script setup>
// imports
import { computed, ref, watch, onBeforeUnmount } from 'vue';
import { useI18n } from 'vue-i18n';
import { Browser } from "@wailsio/runtime";

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import DropDownBox from '@/instances/common/components/DropDownBox.vue';
import ProfilePhoto from '@/instances/common/components/ProfilePhoto.vue';

// services
import { StudioService } from "@/services";

// stores
import { useIconStore } from '@/stores/icons';
import { useNotificationStore } from '@/stores/notifications';
import { useProjectStore } from '@/stores/projects';
import { useStudioStore } from '@/stores/studio';
import { useUserStore } from '@/stores/users';

const iconStore = useIconStore();
const notificationStore = useNotificationStore();
const projectStore = useProjectStore();
const studioStore = useStudioStore();
const userStore = useUserStore();

const { t } = useI18n();

// props
const props = defineProps({
  canDelete: { type: Boolean, default: true },
  canEdit: { type: Boolean, default: true },
  collaborator: Object,
  compact: { type: Boolean, default: false },
  index: Number,
  isLoading: { type: Boolean, default: false },
  isProjectMember: { type: Boolean, default: true },
  onAdd: { type: Function, default: null },
  onDelete: { type: Function, default: null },
  onRoleChange: { type: Function, default: null },
  roles: { type: Array, default: () => ['Admin', 'User'] },
});

// refs
const collaboratorItem = ref(null);
const isEditing = ref(false);

// computed props
const canDeleteUser = computed(() => {
  return !isCurrentUser.value && props.canDelete;
});

const canEditRole = computed(() => {
  return !isCurrentUser.value && props.canEdit;
});

const collaboratorRoles = computed(() => props.roles);

// Returns the formatted role name for display in compact mode.
const displayRole = computed(() => {
  const roleName = props.collaborator.role_name || props.collaborator.role?.name || '';
  return roleName.charAt(0).toUpperCase() + roleName.slice(1).toLowerCase();
});

const isCurrentUser = computed(() => {
  return userStore.user?.id === props.collaborator.id;
});

const userFullName = computed(() => {
  return `${props.collaborator.first_name} ${props.collaborator.last_name}`;
});

// methods
// Adds the collaborator to the project.
const addToProject = () => {
  if (props.onAdd) {
    props.onAdd(props.collaborator.id);
  }
};

// Deletes a collaborator from the project or studio.
const deleteCollaborator = (collaboratorId) => {
  if (props.onDelete) {
    props.onDelete(collaboratorId);
    return;
  }

  const userId = props.collaborator.id;

  StudioService.RemoveCollaborator(collaboratorId, projectStore.selectedStudio.id)
    .then(() => {
      studioStore.studioUsers = studioStore.studioUsers.filter((user) => user.id !== userId);
    })
    .catch((error) => {
      notificationStore.errorNotification(t('components.collaboratorItem.errorRemovingCollaborator'), error.response.data);
    });
};

// Returns the icon path for the given icon name.
const getAppIcon = (iconName) => {
  return iconStore.getAppIcon(iconName);
};

// Handles the escape key press to exit editing mode.
const handleEscKey = () => {
  if (props.compact && isEditing.value) {
    stopEditing();
  }
};

// Opens the collaborator's profile in a browser.
const openUserProfile = () => {
  const profileUrl = `https://app.clustta.com/user/${props.collaborator.username}`;
  Browser.OpenURL(profileUrl);
};

// Handles role selection and updates the collaborator's role.
const selectRole = (role) => {
  if (props.onRoleChange) {
    props.onRoleChange(props.collaborator.id, role);
    if (props.compact) {
      stopEditing();
    }
    return;
  }

  const userId = props.collaborator.id;
  const selectedUser = studioStore.studioUsers.find((user) => user.id === userId);

  StudioService.ChangeCollaboratorRole(props.collaborator.id, projectStore.selectedStudio.id, role)
    .then(() => {
      selectedUser.role_name = role;
      if (props.compact) {
        stopEditing();
      }
    })
    .catch((error) => {
      notificationStore.errorNotification(t('components.collaboratorItem.errorUpdatingRole'), error.response.data);
    });
};

// Starts editing mode in compact view.
const startEditing = () => {
  isEditing.value = true;
};

// Stops editing mode in compact view.
const stopEditing = () => {
  isEditing.value = false;
};

// Handles clicks outside the component to exit editing mode.
// Ignores clicks on teleported dropdown elements.
const handleClickOutside = (event) => {
  const isInsideComponent = collaboratorItem.value && collaboratorItem.value.contains(event.target);
  const isInsideDropdown = event.target.closest('.listbox-list-items-root');
  
  if (!isInsideComponent && !isInsideDropdown) {
    stopEditing();
  }
};

// watchers
// Manages document click listener based on editing state.
watch(isEditing, (editing) => {
  if (editing) {
    document.addEventListener('mousedown', handleClickOutside, { capture: true });
  } else {
    document.removeEventListener('mousedown', handleClickOutside, { capture: true });
  }
});

// lifecycle hooks
onBeforeUnmount(() => {
  document.removeEventListener('mousedown', handleClickOutside, { capture: true });
});

</script>

<style scoped>
@import "@/assets/desktop.css";

.collaborator-item-main {
  display: flex;
  gap: .2rem;
  color: var(--text);
  align-items: center;
  padding-left: .5rem;
  box-sizing: border-box;
  width: 100%;
  height: min-content;
  justify-content: flex-end;
  align-items: flex-start;
  background-color: var(--surface-2);
  border-radius: var(--large-radius);
  overflow: hidden;
  padding-right: 0px;
  outline: var(--transparent-line);
  outline-offset: -1px;
  transition: all .2s ease-out;
}

.collaborator-item-main:hover {
  background-color: var(--surface-3);
  border-radius: var(--small-radius);
  outline: 1px solid var(--surface-4);
}

.collaborator-item-root {
  display: flex;
  flex-direction: column;
  gap: .2rem;
  color: var(--text);
  align-items: center;
  padding: .3rem;
  box-sizing: border-box;
  width: 100%;
  height: min-content;
  justify-content: flex-end;
  border-radius: 10px;
  overflow: hidden;
  padding-right: 0px;
}

.collaborator-item-container {
  display: flex;
  gap: .5rem;
  color: var(--text);
  align-items: center;
  padding: .2rem .4rem;
  box-sizing: border-box;
  width: 100%;
  height: 50px;
  justify-content: space-between;
  transition: all .3s ease-out;
}

.collaborator-item-spacer {
  position: relative;
  width: 36px;
  height: 60px;
  display: flex;
  box-sizing: border-box;
  align-items: center;
  justify-content: center;
  overflow: hidden;
}

.collaborator-item-content {
  gap: .4rem;
  display: flex;
  box-sizing: border-box;
  align-items: center;
  justify-content: space-between;
  height: 100%;
  width: 100%;
  overflow: hidden;
}

.collaborator-item-details {
  padding: .2rem;
  flex-wrap: nowrap;
  overflow: hidden;
  box-sizing: border-box;
  align-items: center;
  justify-content: space-between;
  height: min-content;
  white-space: nowrap;
  text-overflow: ellipsis;
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: .1rem;
}

.collaborator-item-name {
  font-size: 14px;
  font-weight: 400;
}

.collaborator-item-email {
  font-size: 12px;
  font-weight: 300;
  color: var(--text);
  opacity: 0.5;
}

.collaborator-item-dropdown {
  display: flex;
  box-sizing: border-box;
  align-items: center;
  width: 200px;
  min-width: 200px;
  overflow: hidden;
}

.collaborator-item-actions {
  display: flex;
  box-sizing: border-box;
  align-items: center;
  justify-content: flex-end;
  width: min-content;
  min-width: max-content;
  gap: .5rem;
  height: 100%;
}

/* Compact mode styles */
.compact-actions-container {
  display: flex;
  align-items: center;
  gap: .5rem;
  height: 100%;
}

.compact-hover-actions {
  display: none;
  opacity: 0;
  transition: all 0.2s ease-out;
}

.compact-mode:hover .compact-hover-actions {
  display: flex;
  box-sizing: border-box;
  align-items: center;
  justify-content: space-between;
  width: min-content;
  min-width: max-content;
  gap: .5rem;
  height: 100%;
  opacity: 1;
}

.compact-role-meta {
  color: var(--text);
  background-color: rgba(0, 0, 0, 0.216);
  padding: .3rem .5rem;
  border-radius: 5px;
  white-space: nowrap;
  font-size: 12px;
  max-width: 100px;
  overflow: hidden;
  text-overflow: ellipsis;
}

.compact-mode:hover .compact-role-meta {
  opacity: 0;
  width: 0;
  max-width: 0;
  padding: 0;
  transition: all 0.2s ease-out;
}

.compact-dropdown {
  width: 200px;
  min-width: 200px;
}

.compact-editing .collaborator-item-container {
  justify-content: flex-end;
}
</style>