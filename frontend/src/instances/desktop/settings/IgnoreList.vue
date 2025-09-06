<template>
  <div class="settings-component-root">
    <div class="settings-component-container">

      <IgnoreListBox :placeholder="'Add item'" :selectedItems="ignoreList" @itemAdded="addIgnoredItem"
        @itemRemoved="removeIgnoredItem" />


    </div>
  </div>
</template>

<script setup>
import { useIconStore } from '@/stores/icons';
const iconStore = useIconStore();

// imports
import { onMounted, computed, ref } from 'vue';
import utils from '@/services/utils';
import { UserService } from '@/../bindings/clustta/services/index';
import { ProjectService } from "@/../bindings/clustta/services";

// state imports
import { useTrayStates } from '@/stores/TrayStates';

// store imports
import { useNotificationStore } from '@/stores/notifications';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useUserStore } from '@/stores/users';
import { useProjectStore } from '@/stores/projects';

// components
import IgnoreListBox from '@/instances/common/components/IgnoreListBox.vue';

// states
const userStore = useUserStore();
const modals = useDesktopModalStore();
const notificationStore = useNotificationStore();
const projectStore = useProjectStore();

// computed props
const getRoleTypeIcon = (icon) => {
  return '/icons/person.svg'
}

const ignoreList = ref([])

const addIgnoredItem = (item) => {
  if (!ignoreList.value.includes(item)) {
    ignoreList.value.push(item);
    let projectUri
    if (projectStore.activeProject.has_remote) {
      projectUri = projectStore.getActiveProjectUrl
    } else {
      projectUri = projectStore.activeProject.uri
    }
    ProjectService.SetIgnoreList(projectUri, projectStore.selectedStudio.name, ignoreList.value)
      .then((response) => {
        let project = projectStore.activeProject;
        let projectIndex = projectStore.projects.findIndex((p) => p.id === project.id);
        projectStore.projects[projectIndex].ignore_list = ignoreList.value;
        projectStore.activeProject.ignore_list = ignoreList.value;
        // notificationStore.addNotification('Ignore list updated', 'success');
        projectStore.refreshUntrackedItems()
      })
      .catch((error) => {
        console.log(error)
        notificationStore.addNotification('Failed to update ignore list', 'error');
      });
  }
};

const removeIgnoredItem = (item) => {
  ignoreList.value = ignoreList.value.filter((removedItem) => removedItem !== item)
  let projectUri
  if (projectStore.activeProject.has_remote) {
    projectUri = projectStore.getActiveProjectUrl
  } else {
    projectUri = projectStore.activeProject.uri
  }
  // console.log(ignoreList.value)
  ProjectService.SetIgnoreList(projectUri, projectStore.selectedStudio.name, ignoreList.value)
    .then((response) => {
      let project = projectStore.activeProject;
      let projectIndex = projectStore.projects.findIndex((p) => p.id === project.id);
      projectStore.projects[projectIndex].ignore_list = ignoreList.value;
      projectStore.activeProject.ignore_list = ignoreList.value;
      // notificationStore.addNotification('Ignore list updated', 'success');
      // projectStore.reloadUntrackedItems()
    })
    .catch((error) => {
      notificationStore.addNotification('Failed to update ignore list', 'error');
    });
};

const getAppIcon = (iconName) => {
  const icon = iconStore.getAppIcon(iconName);
  return icon
};

// onMounted hook
onMounted(async () => {
  ignoreList.value = projectStore.activeProject.ignore_list;
  console.log(projectStore.activeProject)

});
</script>


<style scoped>
.input-short {
  flex: 1;
  width: 100%;
}

.settings-component-root {
  width: 100%;
  height: 100%;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  gap: 5px;
  box-sizing: border-box;
  align-items: center;
  justify-content: center;
}

.settings-component-container {
  display: flex;
  flex-direction: column;
  box-sizing: border-box;
  height: 100%;
  overflow: hidden;
  box-sizing: border-box;
  width: 96%;
  gap: .5rem;
  align-items: center;
  color: white;
  justify-content: space-between;
  border-radius: var(--large-radius);
  padding: 1rem;
  background-color: crimson;
  background-color: var(--black-steel);
}
</style>