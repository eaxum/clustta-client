<template>

  <div class="general-pane-header">
    <SearchBar v-model="searchQuery" :placeholder="$t('placeholders.searchByMessageOrAuthor')" @input="updateSearch" @clear="clearSearch" />
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
import { useI18n } from 'vue-i18n';
import emitter from '@/lib/mitt';
import { FSService, CheckpointService } from '@/services';
import utils from '@/services/utils';

// components
import CheckpointItem from '@/instances/desktop/components/CheckpointItem.vue';
import CheckpointListSkeleton from '@/instances/common/components/CheckpointListSkeleton.vue';
import PageState from '@/instances/common/components/PageState.vue';
import SearchBar from '@/instances/desktop/components/SearchBar.vue';

// store/state imports
import { useAssetStore } from '@/stores/assets';
import { useNotificationStore } from '@/stores/notifications';
import { useProjectStore } from '@/stores/projects';
import { useTrayStates } from '@/stores/TrayStates';
import { useUserStore } from '@/stores/users';

// stores/states
const assetStore = useAssetStore();
const notificationStore = useNotificationStore();
const projectStore = useProjectStore();
const trayStates = useTrayStates();
const userStore = useUserStore();

const { t } = useI18n();

// refs
const checkpointItem = ref(null);
const checkpointList = ref(null);
const checkpoints = ref([]);
const isExpanded = ref(-1);
const searchQuery = ref('');
const taskHash = ref('');

// methods

// Clears the search and refreshes checkpoints.
const clearSearch = () => {
  searchQuery.value = '';
  refreshCheckpoints();
};

// Returns the illustration path for the empty state.
const illustration = () => '/page-states/resources.png';

// Returns the message for the empty state.
const message = () => t('panes.noCheckpointsMatchSearch');

// Filters checkpoints based on the search query.
const updateSearch = () => {
  if (!searchQuery.value) {
    refreshCheckpoints(); 
    return;
  }
  const query = searchQuery.value?.toLowerCase();
  checkpoints.value = checkpoints.value.filter(checkpoint => {
    return checkpoint.comment.toLowerCase().includes(query) ||
      checkpoint.author.toLowerCase().includes(query);
  });
};

// Updates the expanded checkpoint index.
const updateExpanded = (index) => {
  isExpanded.value = index;
};

// computed properties
const checkpointEntity = computed(() => assetStore.selectedAsset);

// watchers
watch(checkpointEntity, () => {
  refreshCheckpoints();
});

// Refreshes the checkpoints list from the server.
const refreshCheckpoints = async () => {

  taskHash.value = "";
  if (assetStore.selectedAsset && await FSService.Exists(assetStore.selectedAsset.file_path)) {
    taskHash.value = await FSService.FileHash(assetStore.selectedAsset.file_path);
  }

  trayStates.checkpointsLoaded = false;
  checkpoints.value = [];

  if (!assetStore.selectedAsset) {
    trayStates.checkpointsLoaded = true;
    return;
  }

  let task = assetStore.selectedAsset;
  let taskCheckpoints = await CheckpointService.GetCheckpoints(projectStore.activeProject.uri, task.id)
    .then((data) => {
      return data;
    })
    .catch((error) => {
      notificationStore.addNotification(
        t('notifications.errorLoadingCheckpoints'),
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
      // Try fetching from global server before skipping
      try {
        console.log(taskCheckpoints)
        author = await userStore.fetchUserData(authorId);
        if (author) {
          userCache[authorId] = author;
        }
      } catch (error) {
        console.error(`Failed to fetch user data for ${authorId}:`, error);
      }
      
      if (!author) {
        // Generate a placeholder user for removed/deleted users
        author = {
          id: authorId,
          first_name: 'Removed',
          last_name: 'User',
          photo: '',
          email: ''
        };
        userCache[authorId] = author;
      }
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
  if (assetStore.selectedAsset && await FSService.Exists(assetStore.selectedAsset.file_path)) {
    taskHash.value = await FSService.FileHash(assetStore.selectedAsset.file_path);
  }
};

const updateCheckpoints = async () => {
  await refreshCheckpoints();
  await updateTaskHash();
}

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
  await updateCheckpoints();
  // Add keyboard navigation listener
  // window.addEventListener('keydown', handleKeyDown);
  emitter.on('update-checkpoints', updateCheckpoints);
});

onBeforeUnmount(() => {
  // Remove keyboard navigation listener
  // window.removeEventListener('keydown', handleKeyDown);
  emitter.off('update-checkpoints', updateCheckpoints);
});

</script>

<style scoped>
@import "@/assets/desktop.css";

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



