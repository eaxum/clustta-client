<template>
  <div ref="collectionMenu" class="filter-menu-container" v-stop-propagation>
    <div class="input-section">
      <div class="horizontal-flex">
        <input ref="searchUserInput" v-stop-propagation v-model="searchUserTerm" class="input-short" type="text"
          :placeholder="$t('placeholders.searchUser')" />
      </div>
    </div>

    <span class="filter-menu-item" @click="toggleHasAssignees()">
      <img class="small-icons" :src="getAppIcon('person-plus')">
      <div class="horizontal-flex">
        <div class="menu-item-text" >{{ $t('menus.isAssigned') }}</div>
        <ToggleSwitch :switchValueProp="commonStore.hasAssignees" />
      </div>
    </span>

    <span class="filter-menu-item" @click="toggleNoAssignees()">
      <img class="small-icons" :src="getAppIcon('person-minus')">
      <div class="horizontal-flex">
        <div class="menu-item-text" >{{ $t('menus.isNotAssigned') }}</div>
        <ToggleSwitch :switchValueProp="commonStore.noAssignees" />
      </div>
    </span>

    <div ref="assigneeScrollContainer" class="assignee-scroll-container">
      <span v-if="!assigneeFilterActive && filteredCollaborators.length" class="menu-divider"></span>

      <span v-if="!assigneeFilterActive" v-for="collaborator in filteredCollaborators" :key="collaborator.id" class="filter-menu-item" @click="toggleFilter(collaborator)">
        <div class="profile-picture" :style="{ backgroundColor: collaborator.avatarColor}">
              <img class="profile-img"  :src=" collaborator.photo ? collaborator.photo : generateAvatar(collaborator.id)">
          </div>
        <div class="horizontal-flex">
          <div class="menu-item-text"> {{  utils.capitalizeStr(collaborator.name) }} </div>
          <ToggleSwitch :switchValueProp="isFilterActive(collaborator)"  />
        </div>
      </span>

      <div v-if="!assigneeFilterActive && searchUserTerm && !filteredCollaborators.length" class="no-results">
        {{ $t('menus.noResults') }}
      </div>
    </div>

  </div>

</template>

<script setup>
// imports
import { computed, onBeforeUnmount, onMounted, ref } from 'vue';
import emitter from '@/lib/mitt';
import utils from '@/services/utils';
import { generateAvatar } from '@/lib/avatar';

// components
import ToggleSwitch from '@/instances/common/components/ToggleSwitch.vue';

// stores
import { useCommonStore } from '@/stores/common';
import { useIconStore } from '@/stores/icons';
import { useMenu } from '@/stores/menu';
import { useUserStore } from '@/stores/users';

const commonStore = useCommonStore();
const iconStore = useIconStore();
const menu = useMenu();
const userStore = useUserStore();

// refs
const collectionMenu = ref(null);
const assigneeScrollContainer = ref(null);
const searchUserInput = ref(null);
const searchUserTerm = ref('');

// computed properties
// Returns list of collaborators with formatted properties.
const allCollaborators = computed(() => {
  return userStore.getProjectCollaborators.map((collaborator) => ({
    ...collaborator,
    name: `${collaborator.first_name} ${collaborator.last_name}`,
    type: 'assignation',
    avatarColor: userStore.userProfileColor(collaborator.id),
  }));
});

// Checks if assignee filter is active (has or no assignees).
const assigneeFilterActive = computed(() => {
  return commonStore.hasAssignees || commonStore.noAssignees;
});

// Returns collaborators matching the search term.
const filteredCollaborators = computed(() => {
  const query = searchUserTerm.value.trim().toLowerCase();

  if (!query) return allCollaborators.value;

  return allCollaborators.value.filter((collaborator) => {
    const name = collaborator.name?.toLowerCase() || '';
    const username = collaborator.username?.toLowerCase() || '';
    const email = collaborator.email?.toLowerCase() || '';

    return name.includes(query) || username.includes(query) || email.includes(query);
  });
});

// methods
// Adds a filter to the asset filters list.
const addFilter = (filter) => {
  commonStore.assetFilters.push(filter);
};

// Returns the icon path for a given icon name.
const getAppIcon = (iconName) => {
  return iconStore.getAppIcon(iconName);
};

// Checks if two assignee filters represent the same user.
const isSameAssigneeFilter = (item, filter) => {
  return item.type === filter.type && String(item.id) === String(filter.id);
};

// Checks if a filter is currently active.
const isFilterActive = (filter) => {
  return commonStore.assetFilters.some((item) => isSameAssigneeFilter(item, filter));
};

// Removes a filter from the asset filters list.
const removeFilter = (filter) => {
  commonStore.assetFilters = commonStore.assetFilters.filter((item) => !isSameAssigneeFilter(item, filter));
};

// Toggles a filter on or off and refreshes browser.
const toggleFilter = (filter) => {
  if (isFilterActive(filter)) {
    removeFilter(filter);
  } else {
    addFilter(filter);
  }
  emitter.emit('refresh-browser');
};

// Toggles filter for items that have assignees.
const toggleHasAssignees = () => {
  commonStore.hasAssignees = !commonStore.hasAssignees;
  if (commonStore.noAssignees) {
    commonStore.noAssignees = false;
  }
  emitter.emit('refresh-browser');
};

// Toggles filter for items without assignees.
const toggleNoAssignees = () => {
  commonStore.noAssignees = !commonStore.noAssignees;
  if (commonStore.hasAssignees) {
    commonStore.hasAssignees = false;
  }
  emitter.emit('refresh-browser');
};

// lifecycle hooks
onMounted(() => {
  searchUserInput.value?.focus();
  if (assigneeScrollContainer.value) {
    assigneeScrollContainer.value.scrollTop = 0;
  }
  menu.assetMenuWidth = collectionMenu.value.getBoundingClientRect().width;
  menu.collectionMenu = collectionMenu.value;
});

onBeforeUnmount(() => {
  menu.assetMenuWidth = collectionMenu.value.getBoundingClientRect().width;
  menu.assetMenuHeight = collectionMenu.value.getBoundingClientRect().height;
});
</script>

<style scoped>
@import "@/assets/desktop.css";
@import "@/assets/menu.css";

.input-section {
  min-height: min-content;
}

.assignee-scroll-container {
  flex-direction: column;
  gap: .3rem;
  max-height: 50vh;
  overflow: hidden;
  overflow-y: auto;
  width: 100%;
}

.assignee-scroll-container::-webkit-scrollbar {
  width: 4px;
}

.assignee-scroll-container::-webkit-scrollbar-thumb {
  border-radius: var(--small-radius);
  background-color: var(--surface-4);
}

.assignee-scroll-container::-webkit-scrollbar-track {
  border-radius: var(--small-radius);
}

.profile-picture{
    background-color: red;
    height: 24px;
    min-width: 24px;
    overflow: hidden;
    display: flex;
    align-items: center;
    border-radius: 24px;
    /* padding: 5px; */
}

.profile-img{
    width: 100%;
    height: 100%;
}

.filter-menu-item{
  min-height: 35px;
}

.no-results {
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 1rem;
  font-size: 12px;
  color: var(--text);
  opacity: 0.5;
}
</style>

