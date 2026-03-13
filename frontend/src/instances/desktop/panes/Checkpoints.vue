<template>

  <div class="general-pane-header">
    <SearchBar v-model="searchQuery" :placeholder="$t('placeholders.searchByMessageOrAuthor')" @input="updateSearch" @clear="clearSearch" />
  </div>

  <div class="general-pane-root">

    <CheckpointListSkeleton v-if="!trayStates.checkpointsLoaded" />

    <div v-else-if="checkpoints.length" ref="checkpointList" class="checkpoint-list-container" v-stop-propagation>
      <CheckpointGroup v-for="(group, groupIndex) in groupedCheckpoints" :key="group.key" :group="group"
        :taskHash="taskHash" :expandedId="expandedId" :isFirstGroup="groupIndex === 0"
        :isLastGroup="groupIndex === groupedCheckpoints.length - 1"
        @refreshCheckpoints="refreshCheckpoints" @updateTaskHash="updateTaskHash" @updateExpanded="updateExpanded" />
    </div>

    <PageState v-else :message="message()" :illustration="illustration()" />

  </div>
</template>

<script setup>
// imports
import { ref, onMounted, computed, watch, onBeforeUnmount } from 'vue';
import { useI18n } from 'vue-i18n';
import emitter from '@/lib/mitt';
import { FSService, CheckpointService } from '@/services';
import utils from '@/services/utils';

// components
import CheckpointGroup from '@/instances/desktop/components/CheckpointGroup.vue';
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

const { t, locale } = useI18n();

// refs
const checkpointList = ref(null);
const checkpoints = ref([]);
const expandedId = ref('');
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
  const now = new Date();
  checkpoints.value = checkpoints.value.filter(checkpoint => {
    const { label } = getGroupKey(checkpoint.created_at, now);
    return checkpoint.comment.toLowerCase().includes(query) ||
      checkpoint.author.toLowerCase().includes(query) ||
      label.toLowerCase().includes(query);
  });
};

// Updates the expanded checkpoint ID.
const updateExpanded = (checkpointId) => {
  expandedId.value = checkpointId;
};

// Returns the group key label for a given checkpoint date.
const getGroupKey = (dateStr, now) => {
  const date = new Date(dateStr);
  const today = new Date(now.getFullYear(), now.getMonth(), now.getDate());
  const diffDays = Math.floor((today - new Date(date.getFullYear(), date.getMonth(), date.getDate())) / (1000 * 60 * 60 * 24));

  if (diffDays === 0) return { key: 'today', label: t('checkpointGroups.today') };
  if (diffDays === 1) return { key: 'yesterday', label: t('checkpointGroups.yesterday') };
  if (diffDays < 5) {
    const dayName = date.toLocaleDateString(locale.value, { weekday: 'long' });
    return { key: `day-${diffDays}`, label: dayName };
  }
  if (diffDays < 12) return { key: 'last-week', label: t('checkpointGroups.lastWeek') };
  if (diffDays < 19) return { key: '2-weeks-ago', label: t('checkpointGroups.twoWeeksAgo') };
  if (diffDays < 26) return { key: '3-weeks-ago', label: t('checkpointGroups.threeWeeksAgo') };

  const monthYear = date.toLocaleDateString(locale.value, { month: 'long', year: 'numeric' });
  return { key: `month-${date.getFullYear()}-${date.getMonth()}`, label: monthYear };
};

// computed properties
const checkpointEntity = computed(() => assetStore.selectedAsset);

// Groups checkpoints by date tier.
const groupedCheckpoints = computed(() => {
  const now = new Date();
  const groups = new Map();

  for (const cp of checkpoints.value) {
    const { key, label } = getGroupKey(cp.created_at, now);
    if (!groups.has(key)) {
      groups.set(key, { key, label, items: [] });
    }
    groups.get(key).items.push(cp);
  }
  return [...groups.values()];
});

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
  if (!taskCheckpoints || !taskCheckpoints.length) return;
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
      author_id: authorId,
      created_at: checkpoint.created_at,
      preview: preview,
      ownerId: checkpoint.task_id,
      checkpoint_id: checkpoint.id,
      is_downloaded: checkpoint.is_downloaded,
      hash: checkpoint.xxhash_checksum,
      author_profile: authorProfile,
      avatarColor: userStore.userProfileColor(authorId),
      synced: !projectStore.activeProject?.has_remote || checkpoint.synced,
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

  const allCheckpointIds = checkpoints.value.map(cp => cp.checkpoint_id);
  const currentIndex = allCheckpointIds.indexOf(expandedId.value);

  if (event.key === 'ArrowDown') {
    const newIndex = currentIndex === -1 ? 0 : Math.min(currentIndex + 1, allCheckpointIds.length - 1);
    expandedId.value = allCheckpointIds[newIndex];
    setTimeout(() => {
      const element = document.querySelectorAll('.checkpoint-item')[newIndex];
      if (element) element.scrollIntoView({ behavior: 'smooth', block: 'start' });
    }, 10);
  } else if (event.key === 'ArrowUp') {
    if (currentIndex > 0) {
      const newIndex = currentIndex - 1;
      expandedId.value = allCheckpointIds[newIndex];
      setTimeout(() => {
        const element = document.querySelectorAll('.checkpoint-item')[newIndex];
        if (element) element.scrollIntoView({ behavior: 'smooth', block: 'start' });
      }, 10);
    }
  } else if (event.key === 'Escape') {
    expandedId.value = '';
  }
};

// lifecycle hooks
onMounted(async () => {
  await updateCheckpoints();
  emitter.on('update-checkpoints', updateCheckpoints);
});

onBeforeUnmount(() => {
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
  gap: 0;
  overflow: hidden;
  overflow-y: scroll;
  border-radius: 10px;
  padding-right: 5px;
  padding-bottom: 1rem;
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



