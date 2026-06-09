<template>
  <div ref="collectionMenu" class="filter-menu-container" v-stop-propagation>

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

    <span v-if="!assigneeFilterActive && allCollaborators.length" class="menu-divider"></span>

    <span v-if="!assigneeFilterActive" v-for="collaborator in allCollaborators" class="filter-menu-item" @click="toggleFilter(collaborator)">
      <div class="profile-picture" :style="{ backgroundColor: collaborator.avatarColor}">
            <img class="profile-img"  :src=" collaborator.photo ? collaborator.photo : generateAvatar(collaborator.id)">
        </div>
      <div class="horizontal-flex">
        <div class="menu-item-text"> {{  utils.capitalizeStr(collaborator.name) }} </div>
        <ToggleSwitch :switchValueProp="isFilterActive(collaborator)"  />
      </div>
    </span>

  </div>

</template>

<script setup>
// imports
import { computed, onBeforeUnmount, onMounted, ref } from 'vue';
import emitter from '@/lib/mitt';
import { useI18n } from 'vue-i18n';
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

const { t } = useI18n();

// refs
const collectionMenu = ref(null);

// computed properties
// Returns list of collaborators with formatted properties.
const allCollaborators = computed(() => {
  let collaborators = userStore.getProjectCollaborators;

  for (let i = 0; i < collaborators.length; i++) {
    collaborators[i].name = collaborators[i].first_name + ' ' + collaborators[i].last_name;
    collaborators[i].id = collaborators[i].id;
    collaborators[i].type = 'assignation';
    collaborators[i].avatarColor = userStore.userProfileColor(collaborators[i].id);
  }

  return collaborators;
});

// Checks if assignee filter is active (has or no assignees).
const assigneeFilterActive = computed(() => {
  return commonStore.hasAssignees || commonStore.noAssignees;
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

// Checks if a filter is currently active.
const isFilterActive = (filter) => {
  return commonStore.assetFilters.includes(filter);
};

// Removes a filter from the asset filters list.
const removeFilter = (filter) => {
  commonStore.assetFilters = commonStore.assetFilters.filter((item) => item !== filter);
};

// Toggles a filter on or off and refreshes browser.
const toggleFilter = (filter) => {
  if (commonStore.assetFilters.includes(filter)) {
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

.profile-picture{
    background-color: hsl(var(--destructive));
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
</style>

