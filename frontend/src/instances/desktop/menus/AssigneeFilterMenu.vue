<template>
  <div ref="collectionMenu" class="filter-menu-container" v-stop-propagation>

    <span class="filter-menu-item" @click="toggleHasAssignees()">
      <img class="small-icons" :src="getAppIcon('person-plus')">
      <div class="horizontal-flex">
        <div class="menu-item-text" >Is assigned</div>
        <ToggleSwitch :switchValueProp="commonStore.hasAssignees" />
      </div>
    </span>

    <span class="filter-menu-item" @click="toggleNoAssignees()">
      <img class="small-icons" :src="getAppIcon('person-minus')">
      <div class="horizontal-flex">
        <div class="menu-item-text" >Is not assigned</div>
        <ToggleSwitch :switchValueProp="commonStore.noAssignees" />
      </div>
    </span>

    <span v-if="!assigneeFilterActive && allCollaborators.length" class="menu-divider"></span>

    <span v-if="!assigneeFilterActive" v-for="collaborator in allCollaborators" class="filter-menu-item" @click="toggleFilter(collaborator)">
      <div class="profile-picture" :style="{ backgroundColor: collaborator.avatarColor}">
            <img class="profile-img"  :src=" collaborator.photo ? collaborator.photo : '/icons/default_profile_picture.svg'">
        </div>
      <div class="horizontal-flex">
        <div class="menu-item-text"> {{  utils.capitalizeStr(collaborator.name) }} </div>
        <ToggleSwitch :switchValueProp="isFilterActive(collaborator)"  />
      </div>
    </span>

  </div>

</template>

<script setup>
import { useIconStore } from '@/stores/icons';
const iconStore = useIconStore();

const getAppIcon = (iconName) => {
    const icon = iconStore.getAppIcon(iconName);
    return icon
};

// imports
import { computed, ref, onMounted, onBeforeUnmount } from 'vue';
import utils from '@/services/utils';
import emitter from '@/lib/mitt';

// states/store imports
import { useMenu } from '@/stores/menu';
import { useCommonStore } from '@/stores/common';
import { useUserStore } from '@/stores/users';

// components
import ToggleSwitch from '@/instances/common/components/ToggleSwitch.vue';

// states/stores
const menu = useMenu();
const commonStore = useCommonStore();
const userStore = useUserStore();

// refs
const collectionMenu = ref(null);



const assigneeFilterActive = computed(() => {
    return commonStore.hasAssignees || commonStore.noAssignees;
});

const allCollaborators = computed(() => {

  let collaborators = userStore.getProjectCollaborators;

  for(let i = 0; i < collaborators.length; i++){
    collaborators[i].name = collaborators[i].first_name + ' ' + collaborators[i].last_name;
    collaborators[i].id = collaborators[i].id;
    collaborators[i].type = 'assignation';
    collaborators[i].avatarColor = userStore.userProfileColor(collaborators[i].id)
  }

  const availableCollaborators = collaborators;

  return availableCollaborators
});


// methods
const isFilterActive = (filter) => {
    return commonStore.taskFilters.includes(filter);
};

const toggleNoAssignees = () => {
  commonStore.noAssignees = !commonStore.noAssignees;
  if(commonStore.hasAssignees){
    commonStore.hasAssignees = false;
  }
  emitter.emit('refresh-browser');
};

const toggleHasAssignees = () => {
  commonStore.hasAssignees = !commonStore.hasAssignees;
  if(commonStore.noAssignees){
    commonStore.noAssignees = false;
  }
  emitter.emit('refresh-browser');
};

const toggleFilter = (filter) => {
    if(commonStore.taskFilters.includes(filter)){
        removeFilter(filter);
    } else {
        addFilter(filter);
    }
    emitter.emit('refresh-browser');

};

const addFilter = (filter) => {
  commonStore.taskFilters.push(filter);
  console.log(commonStore.taskFilters)
};

const removeFilter = (filter) => {
	commonStore.taskFilters = commonStore.taskFilters.filter((item) => item !== filter );
};



// methods

// onMounted hook
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
</style>