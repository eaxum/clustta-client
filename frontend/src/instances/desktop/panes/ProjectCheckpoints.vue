<template>
  <div class="general-pane-header">
    <SearchBar v-model="searchQuery" placeholder="Search by message or author" @clear="clearSearch" />
  </div>

  <div class="general-pane-root">
    <CheckpointListSkeleton v-if="projectCheckpointsLoading" />

    <div v-else-if="filteredCheckpoints.length" ref="checkpointList" id="checkpointList" class="checkpoint-list-container"
      v-stop-propagation>
      <TimelineItem class="task-item" v-for="(timelineItem, index) in filteredCheckpoints" :key="index"
      @updateExpanded="updateExpanded" :isExpanded="isExpanded" :timelineItem="timelineItem" :timelineItemIndex="index"
        :style="{ animationDelay: index < 10 ? `${(index - 1) * 0.05}s` : '0s' }" />
    </div>

    <PageState v-else :message="message()" :illustration="illustration()" />

  </div>
</template>

<script setup>
// imports
import { computed, onMounted, onUnmounted, ref, watchEffect } from 'vue';
import { v4 as uuidv4 } from 'uuid';

// services
import { CheckpointService } from '@/services';

// components
import CheckpointListSkeleton from '@/instances/common/components/CheckpointListSkeleton.vue';
import PageState from '@/instances/common/components/PageState.vue';
import SearchBar from '@/instances/desktop/components/SearchBar.vue';
import TimelineItem from '@/instances/desktop/components/TimelineItem.vue';

// store imports
import { useNotificationStore } from '@/stores/notifications';
import { useProjectStore } from '@/stores/projects';
import { useUserStore } from '@/stores/users';

// stores
const notificationStore = useNotificationStore();
const projectStore = useProjectStore();
const userStore = useUserStore();

// refs
const checkpointList = ref(null);
const isExpanded = ref(-1);
const modalContainer = ref(null);
const projectCheckpoints = ref([]);
const projectCheckpointsLoading = ref(true);
const searchQuery = ref('');

// computed properties
const filteredCheckpoints = computed(() => {
  const query = searchQuery.value.toLowerCase();
  return projectCheckpoints.value.filter((checkpoint) =>
    checkpoint.comment?.toLowerCase().includes(query) ||
    checkpoint.author?.toLowerCase().includes(query)
  );
});

// methods

// Clears the search query.
const clearSearch = () => {
  searchQuery.value = '';
};

// Returns the illustration path for the empty state.
const illustration = () => '/page-states/resources.png';

// Returns the message for the empty state.
const message = () => {
  if (searchQuery.value) return 'No checkpoints match your search';
  return 'No checkpoints in this project';
};

// Updates the expanded checkpoint index.
const updateExpanded = (index) => {
  isExpanded.value = index;
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
      projectCheckpoints.value = timelineData;
      projectCheckpointsLoading.value = false;
      console.log(timelineData)
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



