<template>
  <div class="details-pane-container absolute-pane">
    <div class="general-pane-header">
      <HeaderArea :title="entityName" :icon="entityIcon" />
    </div>
    <div class="general-pane-root">
      <div class="general-pane-container">
        <!-- Assignee:  -->
        <ScrollList v-if="assigneeList && assigneeList.length" :unassignListItem="unassignTask" :isSingle="true"
          :useAvatar="true" :items="assigneeList" :unassignItems="true" />
        <!-- {{entity.assignee_id }} -->
        Assign to someone else
        <ScrollList v-if="collaboratorsList && collaboratorsList.length" :items="collaboratorsList" :useAvatar="true"
          :deleteItems="false" :assignItems="true" :editListItem="assignTask" :assignListItem="assignEntity" />

      </div>
    </div>
  </div>
</template>

<script setup>

import { useTrayStates } from '@/stores/TrayStates';
import { ref, onMounted, computed, nextTick } from 'vue';
import utils from "@/services/utils";

// store/state imports
import { useUserStore } from '@/stores/users';
import { useNotificationStore } from '@/stores/notifications';
import { useAssetStore } from '@/stores/assets';

// services
import { AssetService } from "@/../bindings/clustta/services";

// components
import HeaderArea from '@/instances/common/components/HeaderArea.vue';
import ScrollList from '@/instances/desktop/components/ScrollList.vue';
import { useProjectStore } from '@/stores/projects';

// stores/states
const trayStates = useTrayStates();
const userStore = useUserStore();
const notificationStore = useNotificationStore();
const assetStore = useAssetStore();
const projectStore = useProjectStore();

// refs
const entityName = ref('');
const entityIcon = ref('');

// computed props
const assignTo = computed(() => { return trayStates.assignTo });
const task = computed(() => { return assetStore.selectedAsset });

const entity = computed(() => {
  if (assignTo.value === 'task') {
    entityName.value = task.value.name;
    entityIcon.value = task.value.icon;
    return task.value
  } else return ''
});

const projectCollaborators = computed(() => {
  return userStore.getProjectCollaborators;
});

const assignee = computed(() => {
  if (!entity.value.assignee_id) {
    return
  };
  return userStore.getUserData(entity.value.assignee_id);
});

const assigneeList = computed(() => {
  if (!entity.value.assignee_id) {
    return
  };
  const allCollaborators = userStore.getProjectCollaborators;
  const filteredCollaborators = allCollaborators.filter((item) => item.id === assignee.value.id);
  return formatCollaborators(filteredCollaborators);
});

const collaboratorsList = computed(() => {

  const allCollaborators = projectCollaborators.value;
  if (!entity.value.assignee_id) {
    return utils.sortAlphabetically(formatCollaborators(allCollaborators))
  };
  const filteredCollaborators = allCollaborators.filter((item) => item.id !== assignee.value.id);
  const result = formatCollaborators(filteredCollaborators);
  return utils.sortAlphabetically(result);
});

// methods
const formatCollaborators = (arr) => {
  return arr.map((user, index) => ({
    name: `${user.first_name} ${user.last_name}` || user,
    icon: user.photo || "",
    avatarColor: userStore.userProfileColor(user.id),
    id: user.id,
    index: index.toString(),
  }));
};

const assignTask = async (index) => {
  let task = entity.value;
  let taskId = entity.value.id;
  let user = collaboratorsList.value[index];
  let userId = user ? user.id : "";
  await AssetService.AssignTask(projectStore.activeProject.uri, taskId, userId)
    .then(async (data) => {
      assetStore.findAsset(taskId).assignee_id = userId;
      notificationStore.addNotification("Task Assigned Successfully.", "", "success")
    })
    .catch((error) => {
      console.log(error)
      notificationStore.errorNotification("Error Assigning Task", error)
    });
};

const unassignTask = async (index) => {
  let task = entity.value;
  let taskId = entity.value.id;
  await AssetService.UnassignTask(projectStore.activeProject.uri, taskId)
    .then(async (data) => {
      assetStore.findAsset(taskId).assignee_id = ""
      notificationStore.addNotification("Task UnAssigned Successfully.", "", "success")
    })
    .catch((error) => {
      console.log(error)
      notificationStore.errorNotification("Error UnAssigning Task", error)
    });
};

onMounted(() => {

});

</script>

<style scoped>
@import "@/assets/desktop.css";

.boxer {
  width: 96%;
}

.details-pane-container {
  padding: 1rem;
  box-sizing: border-box;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  /* background-color: hotpink; */
  flex-direction: column;
  color: white;
  justify-content: flex-start;
}

.general-pane-container {
  display: flex;
  gap: 1rem;
  /* background-color: red; */
  align-items: flex-start;
  justify-content: flex-start;
}
</style>




