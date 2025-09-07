<template>


    <div class="general-pane-header">
      <HeaderArea :title="title" :icon="'clock'" :showSearch="true" />
    </div>
    <div class="general-container general-container-wide">
      <TaskListSkeleton v-if="timelineLoading" :forModal="true" />
      <div class="trash-list-stage-body">
        <div class="trash-list-stage-body-container">
          <TimelineItem class="task-item" v-for="(timelineItem, index) in timeline" :key="index"
            :timelineItem="timelineItem" :timelineIndex="index"
            :style="{ animationDelay: index < 10 ? `${(index - 1) * 0.05}s` : '0s' }" />
        </div>
      </div>


    </div>


</template>

<script setup>
import { useIconStore } from '@/stores/icons';
import { v4 as uuidv4 } from 'uuid'
const iconStore = useIconStore();

const getAppIcon = (iconName) => {
  const icon = iconStore.getAppIcon(iconName);
  return icon
};
// imports
import { onMounted, watchEffect, onUnmounted, ref, computed } from 'vue';

// state imports
import { useTrayStates } from '@/stores/TrayStates';

// store imports
import { useNotificationStore } from '@/stores/notifications';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useStageStore } from '@/stores/stages';
import { useAssetStore } from '@/stores/assets';
import { useCollectionStore } from '@/stores/collections';
import { useStatusStore } from '@/stores/status';
import { useProjectStore } from '@/stores/projects';
import { useDndStore } from '@/stores/dnd';
import { useMenu } from '@/stores/menu';

// components
import HeaderArea from '@/instances/common/components/HeaderArea.vue';
import GeneralButton from '@/instances/common/components/GeneralButton.vue';
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import TaskListSkeleton from '@/instances/desktop/components/TaskListSkeleton.vue'
import TimelineItem from '@/instances/desktop/components/TimelineItem.vue';
import { CheckpointService } from '@/../bindings/clustta/services';
import { useUserStore } from '@/stores/users';

// states
const trayStates = useTrayStates();
const stage = useStageStore();
const menu = useMenu();
const notificationStore = useNotificationStore();
const modals = useDesktopModalStore();
const statusStore = useStatusStore();
const projectStore = useProjectStore();
const collectionStore = useCollectionStore();
const assetStore = useAssetStore();
const dndStore = useDndStore();
const userStore = useUserStore();

// vars
let showSearch = false;
let title = 'Timeline';

// refs
// type Timeline struct {
// CreatedAt string   `db:"created_at" json:"created_at"`
// TaskPaths []string `db:"task_paths" json:"task_paths"`
// Comment   string   `db:"comment" json:"comment"`
// AuthorUID string   `db:"author_id" json:"author_id"`
// Preview   []byte   `db:"preview" json:"preview"`
// }
const timeline = ref([
]);
const timelineLoading = ref(true);
const modalContainer = ref(null);
const popUpActions = ref(null);
const isAwaitingResponse = ref(false);

// computed
const storeHasData = computed(() => {
  const rawData = dndStore.previewData
  return Object.values(rawData).some(arr => arr.length > 0);
});

function getPathParent(path) {
  // If there's no slash, it's a root item, return empty string
  if (!path.includes('/')) {
    return "";
  }

  // Find the last slash and return everything before it
  const lastSlashIndex = path.lastIndexOf('/');
  return path.substring(0, lastSlashIndex);
}

const refresh = async () => {
  assetStore.tasksLoaded = false;
  await projectStore.refreshActiveProject()
  await statusStore.reloadStatuses();
  await collectionStore.reloadEntities();
  await assetStore.reloadTasks();
  projectStore.getUntrackedItems()
  assetStore.tasksLoaded = true;
};

const handleEnterKey = (event) => {
  importItems();
};

const escape = () => {
  modals.setModalVisibility('timelineModal', false);
};

const closeModal = () => {
  modals.setModalVisibility("timelineModal", false);
};

// watchers
watchEffect(() => {
  if (modalContainer.value) {
    menu.clickOutsideMask = modalContainer.value;
  }
});

onMounted(async () => {
  CheckpointService.GetTimeline(projectStore.activeProject.uri)
    .then(async (response) => {
      let timelineData = []
      let userCache = {}
      for (const item of response) {
        let authorId = item.author_id
        if (!userCache[authorId]) {
          userCache[authorId] = await userStore.getUserData(authorId);
        }
        let author = userCache[authorId]
        if (!author) {
          // console.log("Author not found for ID:", authorId);
          continue; // Skip this item if author is not found
        }
        let authorName = `${author.first_name} ${author.last_name}`;
        let authorProfile = author.photo || "";

        let timelineItem = {
          created_at: item.created_at,
          task_paths: item.task_paths,
          comment: item.comment,
          author_id: item.author_id,
          preview: item.preview,
          author_name: authorName,
          author_profile: authorProfile,
          avatarColor: userStore.userProfileColor(authorId),
        }
        timelineData.push(timelineItem)
      }
      timeline.value = timelineData;
      timelineLoading.value = false;
    })
    .catch((error) => {
      notificationStore.errorNotification(
        "Error loading Timeline", error
      )
      console.log(error)
      closeModal()
    });
});

onUnmounted(() => {
})


</script>

<style scoped>
@import "@/assets/desktop.css";

.pop-up-actions {
  /* width: min-content; */
  align-items: center;
  /* background-color: forestgreen; */
  /* justify-content: space-around; */
  box-sizing: border-box;
}

.modal-container {
  justify-content: flex-start;
  align-items: flex-start;
  max-height: 90vh;
}

.general-container-wide {
  display: flex;
  flex-direction: column;
  /* background-color: firebrick; */
  overflow: hidden;
  width: 50vw;
  min-width: 600px !important;
  max-width: 1000px;
  max-height: 80vh;

  box-sizing: border-box;
  align-items: center;
  justify-content: center;
}

.folder-path {
  width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
}

.rules-toggle {
  display: flex;
  /* background-color: red; */
  gap: .5rem;
  align-items: center;
  min-width: max-content
}

.selected-folder {
  /* background-color: darkblue; */
  width: 100%;
  padding: .2rem;
  overflow: hidden;
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
  padding: 10px 20px;
  box-sizing: border-box;
}

.selected-folder-container {
  display: flex;
  /* background-color: firebrick; */
  width: 100%;
  gap: .2rem;
  overflow: hidden;
  display: flex;
  align-items: center;
}

.regex-instances {
  width: 100%;
  display: flex;
  max-height: 50vh;
  flex-direction: column;
  gap: 10px;
  /* background-color: green; */
  /* padding-right: 5px; */
  overflow: hidden;
  /* overflow-y: scroll; */
  box-sizing: border-box;
}

.regex-instances-scroll {
  box-sizing: border-box;
  width: 100%;
  display: flex;
  height: min-content;
  flex-direction: column;
  gap: 10px;
  /* background-color: green; */
}

.regex-instances::-webkit-scrollbar {
  width: 8px;
}

.regex-instances::-webkit-scrollbar-thumb {
  border-radius: 10px;
  background-color: var(--light-steel);
}

.regex-instances::-webkit-scrollbar-track {
  border-radius: 10px;
}

.attachment-area {
  box-sizing: border-box;
  /* margin-top: 1rem; */
  display: flex;
  align-items: center;
  flex-direction: column;
  /* justify-content: space-between; */
  gap: 1rem;
  /* background-color: darkkhaki; */
  width: 98%;
}

.attachment-list {
  box-sizing: border-box;
  /* margin-top: 1rem; */
  display: flex;
  padding: .5rem;
  align-items: center;
  flex-direction: column;
  /* justify-content: space-between; */
  gap: .2rem;
  /* background-color: rgb(57, 122, 108); */

  background-color: rgba(0, 0, 0, 0.144);
  max-height: 150px;
  overflow: hidden;
  overflow-y: scroll;
  width: 100%;
  border-radius: 10px;
}

.attachment-list::-webkit-scrollbar {
  width: 4px;
}

.attachment-list::-webkit-scrollbar-thumb {
  border-radius: 10px;
  background-color: rgba(255, 255, 255, 0.295);
}

.attachment-list::-webkit-scrollbar-track {
  border-radius: 10px;
  /* background-color: rgba(0, 0, 0, 0.295); */
}

.attachment {
  /* margin-top: 1rem; */
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: .5rem;
  width: 100%;
  border-bottom: 1px solid rgba(255, 255, 255, 0.096);
  height: max-content;
  padding: .2rem;

  /* background-color: greenyellow; */
}

.compound-input-section {
  display: flex;
  flex-direction: row;
  align-items: center;
  gap: .4rem;
}

.task-options-container {
  position: relative;
  box-sizing: border-box;

  width: 100%;
  height: 0px;
  /* height: 80px; */
  overflow: hidden;
  transition-property: height;
  transition-duration: 0.2s;
  transition-timing-function: ease-in-out;
  transition: opacity .5s ease-in-out;
  /* transition: all .1s ease-in-out; */
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  margin: 0;
  opacity: 1;
}

.task-options-container-closed {
  transition: all .2s ease-in-out;
  opacity: 0;
  height: 0px;
  padding: 0;
  overflow: hidden;
  /* margin-bottom: -1.5rem; */
}


.input-short {
  flex: 1;
  width: 100%;
}

.listbox-short {

  flex: 1;
  width: 130px;
}

.input-label {

  font-family: Inter, sans-serif;
  color: white;
  font-size: 16px;
  white-space: nowrap;
  flex: 1;

}

.pop-up-prompt {
  gap: 10px;
  /* background-color: bisque; */
  align-items: center;
  /* justify-content: center; */
  max-height: 400px;
}

.trash-list-stage-body {
  width: 100%;
  /* max-width: 1200px; */
  height: 100%;
  display: flex;
  box-sizing: border-box;
  /* background-color: teal; */
  align-items: flex-start;
  justify-content: flex-start;
  justify-content: center;
  overflow: hidden;
  overflow-y: scroll;
  padding: .5rem;
}

.trash-list-stage-body::-webkit-scrollbar {
  width: 8px;
}

.trash-list-stage-body::-webkit-scrollbar-thumb {
  border-radius: 10px;
  background-color: var(--dark-steel);
}

.trash-list-stage-body::-webkit-scrollbar-track {
  border-radius: 10px;
}

.trash-list-stage-body-container {
  width: 100%;
  gap: .5rem;
  max-width: 960px;

  /* height: 100%; */
  height: max-content;
  display: flex;
  box-sizing: border-box;
  flex-direction: column;
  /* background-color: tomato; */
  align-items: flex-start;
  justify-content: flex-start;
  overflow: hidden;
  padding: .5rem;
}
</style>





