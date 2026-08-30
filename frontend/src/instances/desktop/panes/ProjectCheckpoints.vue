<template>
  <div class="general-pane-header">
    <SearchBar v-model="searchQuery" :placeholder="$t('placeholders.searchByMessageOrAuthor')" @clear="clearSearch" />
  </div>

  <div class="general-pane-root">
    <CheckpointListSkeleton v-if="projectCheckpointsLoading" />

    <div v-else-if="groupedCheckpoints.length" ref="checkpointList" id="checkpointList" class="checkpoint-list-container"
      v-stop-propagation>
      <TimelineGroup v-for="(group, groupIndex) in groupedCheckpoints" :key="group.key" :group="group"
        :expandedId="expandedId" :isFirstGroup="groupIndex === 0"
        :isLastGroup="groupIndex === groupedCheckpoints.length - 1"
        @updateExpanded="updateExpanded" />
    </div>

    <PageState v-else :message="message()" :illustration="illustration()" />

  </div>
</template>

<script setup>
// imports
import { computed, onMounted, onUnmounted, ref, watchEffect } from 'vue';
import { useI18n } from 'vue-i18n';
import { v4 as uuidv4 } from 'uuid';

// services
import { CheckpointService } from '@/services';

// components
import CheckpointListSkeleton from '@/instances/common/components/CheckpointListSkeleton.vue';
import PageState from '@/instances/common/components/PageState.vue';
import SearchBar from '@/instances/desktop/components/SearchBar.vue';
import TimelineGroup from '@/instances/desktop/components/TimelineGroup.vue';

// store imports
import { useNotificationStore } from '@/stores/notifications';
import { useProjectStore } from '@/stores/projects';
import { useUserStore } from '@/stores/users';

// stores
const notificationStore = useNotificationStore();
const projectStore = useProjectStore();
const userStore = useUserStore();

const { t, locale } = useI18n();

// refs
const checkpointList = ref(null);
const expandedId = ref('');
const modalContainer = ref(null);
const projectCheckpoints = ref([]);
const projectCheckpointsLoading = ref(true);
const searchQuery = ref('');

// methods

// Clears the search query.
const clearSearch = () => {
  searchQuery.value = '';
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

// Returns the illustration path for the empty state.
const illustration = () => '/page-states/resources.png';

// Returns the message for the empty state.
const message = () => {
  if (searchQuery.value) return t('panes.noCheckpointsMatchSearch');
  return t('panes.noCheckpointsInProject');
};

// Updates the expanded checkpoint ID.
const updateExpanded = (id) => {
  expandedId.value = id;
};

// computed properties

// Filters and groups checkpoints by date tier.
const groupedCheckpoints = computed(() => {
  const now = new Date();
  let items = projectCheckpoints.value;

  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase();
    items = items.filter(checkpoint => {
      const { label } = getGroupKey(checkpoint.created_at, now);
      return checkpoint.comment?.toLowerCase().includes(query) ||
        checkpoint.author_name?.toLowerCase().includes(query) ||
        label.toLowerCase().includes(query);
    });
  }

  const groups = new Map();
  for (const cp of items) {
    const { key, label } = getGroupKey(cp.created_at, now);
    if (!groups.has(key)) {
      groups.set(key, { key, label, items: [] });
    }
    groups.get(key).items.push(cp);
  }
  return [...groups.values()];
});

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
          asset_paths: item.asset_paths ?? [],
          extensions: item.extensions ?? [],
          group_id: item.group_id,
          tags: item.tags ?? [],
          follower_count: item.follower_count ?? 0,
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
    })
    .catch((error) => {
      notificationStore.errorNotification(
        t('notifications.errorLoadingTimeline'), error
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
  gap: 0;
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


