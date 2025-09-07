<template>

  <div class="general-pane-header">
    <HeaderArea :title="utils.capitalizeStr(itemName)" :icon="'layers'" :placeholder="placeholder" :showSearch="true"
      @updateSearch="updateSearch" @clearSearch="clearSearch" />
  </div>

  <div class="general-pane-root">

    <CheckpointListSkeleton v-if="!trayStates.checkpointsLoaded" />

    <div v-else-if="checkpoints.length" ref="checkpointList" class="checkpoint-list-container" v-stop-propagation>
      <CheckpointItem v-for="(checkpoint, index) in checkpoints" ref="checkpointItem" :checkpoint="checkpoint"
        :index="index" :taskHash="taskHash" :isExpanded="isExpanded" @refreshCheckpoints="refreshCheckpoints"
        @updateTaskHash="updateTaskHash" @updateExpanded="updateExpanded" />
    </div>

    <PageState v-else :message="message()" :illustration="illustration()" />

  </div>
  <!-- </div> -->
</template>

<script setup>
// imports
import { ref, onMounted, computed, watch, onBeforeUnmount } from 'vue';
import { FSService, CheckpointService } from '@/../bindings/clustta/services/index';
import utils from '@/services/utils';

// components
import CheckpointItem from '@/instances/desktop/components/CheckpointItem.vue';
import HeaderArea from '@/instances/common/components/HeaderArea.vue';
import PageState from '@/instances/common/components/PageState.vue';
import CheckpointListSkeleton from '@/instances/common/components/CheckpointListSkeleton.vue';

// store/state imports
import { useTrayStates } from '@/stores/TrayStates';
import { useAssetStore } from '@/stores/assets';
import { useNotificationStore } from '@/stores/notifications';
import { useUserStore } from '@/stores/users';
import { useProjectStore } from '@/stores/projects';

// stores/states
const assetStore = useAssetStore();
const trayStates = useTrayStates();
const userStore = useUserStore();
const notificationStore = useNotificationStore();
const projectStore = useProjectStore();

// vars
let placeholder = 'Search Checkpoints...';

// refs
const checkpointList = ref(null);
const checkpointItem = ref(null);
const checkpoints = ref([]);
const taskHash = ref('');
const isExpanded = ref(-1);

const updateSearch = (searchQuery) => {
  if (!searchQuery) {
    refreshCheckpoints(); 
    return;
  }

  const query = searchQuery.toLowerCase();
  // Filter the existing checkpoints without needing a separate array
  checkpoints.value = checkpoints.value.filter(checkpoint => {
    return checkpoint.comment.toLowerCase().includes(query) ||
      checkpoint.author.toLowerCase().includes(query);
  });
}

const message = () => {
  return 'No checkpoints match your search';
};

const illustration = () => {
  return '/page-states/resources.png';
};

const clearSearch = () => {
  refreshCheckpoints();
}

const updateExpanded = (index) => {
  isExpanded.value = index;
}

// computed props
const itemName = computed(() => {
  if (!assetStore.selectedTask) {
    return 'No task Selected'
  }
  return assetStore.selectedTask.name;
});

const checkpointEntity = computed(() => {
  return assetStore.selectedTask;
});

watch(checkpointEntity, () => {
  refreshCheckpoints();
});

const refreshCheckpoints = async () => {

  taskHash.value = "";
  if (assetStore.selectedTask && await FSService.Exists(assetStore.selectedTask.file_path)) {
    taskHash.value = await FSService.FileHash(assetStore.selectedTask.file_path);
  }

  trayStates.checkpointsLoaded = false;
  checkpoints.value = [];

  if (!assetStore.selectedTask) {
    trayStates.checkpointsLoaded = true;
    return;
  }

  let task = assetStore.selectedTask;
  let taskCheckpoints = await CheckpointService.GetCheckpoints(projectStore.activeProject.uri, task.id)
    .then((data) => {
      return data;
    })
    .catch((error) => {
      notificationStore.addNotification(
        "Error Loading Checkpoints",
        error.message,
        "error",
        false
      );
    });

  trayStates.checkpointsLoaded = true;
  let userCache = {}
  for (let i = 0; i < taskCheckpoints.length; i++) {
    let checkpoint = taskCheckpoints[i];
    let authorId = checkpoint.author_id
    if (!userCache[authorId]) {
      userCache[authorId] = await userStore.getUserData(authorId);
    }
    let author = userCache[authorId];
    if (!author) {
      continue; // Skip this item if author is not found
    }
    let author_fullname = `${author.first_name} ${author.last_name}`;
    let preview = null;
    if (checkpoint.preview) {
      preview = "data:image/png;base64," + checkpoint.preview;
    }

    let authorProfile = author.photo || "";


    const checkpointObj = {
      comment: checkpoint.comment,
      author: author_fullname,
      created_at: checkpoint.created_at,
      preview: preview,
      ownerId: checkpoint.task_id,
      checkpoint_id: checkpoint.id,
      is_downloaded: checkpoint.is_downloaded,
      hash: checkpoint.xxhash_checksum,
      author_profile: authorProfile,
      avatarColor: userStore.userProfileColor(authorId),
    };

    const existingCheckpoint = checkpoints.value.find(cp => cp.checkpoint_id === checkpoint.id);
    if (!existingCheckpoint) {
      checkpoints.value.push(checkpointObj);
    }
  }
};

const updateTaskHash = async () => {
  taskHash.value = "";
  if (assetStore.selectedTask && await FSService.Exists(assetStore.selectedTask.file_path)) {
    taskHash.value = await FSService.FileHash(assetStore.selectedTask.file_path);
  }
};

// Add keyboard navigation handler
const handleKeyDown = (event) => {
  if (!checkpoints.value.length) return;

  if (event.key === 'ArrowDown') {
    if (isExpanded.value < checkpoints.value.length - 1) {
      const newIndex = isExpanded.value === -1 ? 0 : isExpanded.value + 1;
      isExpanded.value = newIndex;

      // Scroll to the expanded item
      setTimeout(() => {
        const element = document.querySelectorAll('.checkpoint-item')[newIndex];
        if (element) {
          element.scrollIntoView({ behavior: 'smooth', block: 'start' });
        }
      }, 10);
    }
  } else if (event.key === 'ArrowUp') {
    if (isExpanded.value > 0) {
      const newIndex = isExpanded.value - 1;
      isExpanded.value = newIndex;

      // Scroll to the expanded item
      setTimeout(() => {
        const element = document.querySelectorAll('.checkpoint-item')[newIndex];
        if (element) {
          element.scrollIntoView({ behavior: 'smooth', block: 'start' });
        }
      }, 10);
    }
  } else if (event.key === 'Escape') {
    isExpanded.value = -1;
  }
};

onMounted(async () => {
  await refreshCheckpoints();

  // Get task hash
  await updateTaskHash();
  console.log(taskHash.value);

  // Add keyboard navigation listener
  // window.addEventListener('keydown', handleKeyDown);
});

onBeforeUnmount(() => {
  // Remove keyboard navigation listener
  // window.removeEventListener('keydown', handleKeyDown);
});

</script>

<style scoped>
/* @import "@/assets/tray.css"; */
@import "@/assets/desktop.css";

.checkpoint-task-item {
  background-color: green;
  background-color: transparent;
  height: 40px;
  width: 100%;
  box-sizing: border-box;
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  justify-content: center;
  padding: 0px 12px;
  color: white;
}

.checkpoint-task-item-text {
  font-family: 'Inter', sans-serif;
  box-sizing: border-box;
  font-size: 20px;

}

.checkpoint-list-container {
  width: 100%;
  height: 100%;
  box-sizing: border-box;
  display: flex;
  flex-direction: column;
  gap: .5rem;
  overflow: hidden;
  overflow-y: scroll;
  border-radius: 10px;
  padding-right: 5px;
  padding-bottom: 1rem;
  /* background-color: red; */
}

.checkpoint-list-container::-webkit-scrollbar {
  width: 4px;
}

.checkpoint-list-container::-webkit-scrollbar-thumb {
  border-radius: 10px;
  background-color: rgba(255, 255, 255, 0.295);
}

.checkpoint-list-container::-webkit-scrollbar-track {
  border-radius: 10px;
  /* background-color: rgba(0, 0, 0, 0.295); */
}
</style>



