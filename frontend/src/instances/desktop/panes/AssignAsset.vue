<template>
  <div class="details-pane-container absolute-pane">
    <div class="general-pane-header">
      <HeaderArea :title="collectionName" :icon="collectionIcon" />
    </div>
    <div class="general-pane-root">
      <div class="general-pane-container">
        <!-- Assignee:  -->
        <ScrollList v-if="assigneeList && assigneeList.length" :unassignListItem="unassignAsset" :isSingle="true"
          :useAvatar="true" :items="assigneeList" :unassignItems="true" />
        <!-- {{collection.assignee_id }} -->
        {{ $t('panes.assignToSomeoneElse') }}
        <ScrollList v-if="collaboratorsList && collaboratorsList.length" :items="collaboratorsList" :useAvatar="true"
          :deleteItems="false" :assignItems="true" :editListItem="assignAsset" :assignListItem="assignCollection" />

      </div>
    </div>
  </div>
</template>

<script setup>

import { useTrayStates } from '@/stores/TrayStates';
import { ref, onMounted, computed, nextTick } from 'vue';
import { useI18n } from 'vue-i18n';
import utils from "@/services/utils";

// store/state imports
import { useUserStore } from '@/stores/users';
import { useNotificationStore } from '@/stores/notifications';
import { useAssetStore } from '@/stores/assets';

// services
import { AssetService } from "@/services";

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

const { t } = useI18n();

// refs
const collectionName = ref('');
const collectionIcon = ref('');

// computed props
const assignTo = computed(() => { return trayStates.assignTo });
const asset = computed(() => { return assetStore.selectedAsset });

const collection = computed(() => {
  if (assignTo.value === 'asset') {
    collectionName.value = asset.value.name;
    collectionIcon.value = asset.value.icon;
    return asset.value
  } else return ''
});

const projectCollaborators = computed(() => {
  return userStore.getProjectCollaborators;
});

const assignee = computed(() => {
  if (!collection.value.assignee_id) {
    return
  };
  return userStore.getUserData(collection.value.assignee_id);
});

const assigneeList = computed(() => {
  if (!collection.value.assignee_id) {
    return
  };
  const allCollaborators = userStore.getProjectCollaborators;
  const filteredCollaborators = allCollaborators.filter((item) => item.id === assignee.value.id);
  return formatCollaborators(filteredCollaborators);
});

const collaboratorsList = computed(() => {

  const allCollaborators = projectCollaborators.value;
  if (!collection.value.assignee_id) {
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

const assignAsset = async (index) => {
  let asset = collection.value;
  let assetId = collection.value.id;
  let user = collaboratorsList.value[index];
  let userId = user ? user.id : "";
  await AssetService.AssignAsset(projectStore.activeProject.uri, assetId, userId)
    .then(async (result) => {
      assetStore.findAsset(assetId).assignee_id = userId;
      notificationStore.notifyMetadataUpdate(result, t('notifications.assetAssigned'))
    })
    .catch((error) => {
      console.log(error)
      notificationStore.errorNotification(t('notifications.errorAssigningAsset'), error)
    });
};

const unassignAsset = async (index) => {
  let asset = collection.value;
  let assetId = collection.value.id;
  await AssetService.UnassignAsset(projectStore.activeProject.uri, assetId)
    .then(async (result) => {
      assetStore.findAsset(assetId).assignee_id = ""
      notificationStore.notifyMetadataUpdate(result, t('notifications.assetUnassigned'))
    })
    .catch((error) => {
      console.log(error)
      notificationStore.errorNotification(t('notifications.errorUnassigningAsset'), error)
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




