<template>
  <div class="scroll-list-root" :class="{ 'scroll-list-root-single': isSingle }">

    <div v-if="isSingle" class="scroll-list-container">
      <div v-for="(item, index) in items" :key="index" :index="index" @click="selectItem(index)"
        class="scroll-list-item" :class="{ 'scroll-list-item-wrap': wrapItems, 'selected-item': selectedItem(index) }">
        <div v-if="useAvatar" class="profile-picture" :style="{ backgroundColor: item.avatarColor }">
          <img class="profile-img" :src="item.profile ? item.profile : generateAvatar(item.id)">
        </div>
        <div v-else-if="useIcons" class="asset-item-icon-container">
          <img class="large-icons" :src="getAppIcon(item.icon)">
        </div>
        <div v-else-if="customIcons" class="asset-item-icon-container">
          <img class="large-icons no-filter" :src="item.icon">
        </div>


        <div class="asset-item-content">
          <div class="asset-item-details">
            {{ item.name }}
          </div>
        </div>
        <div v-if="item.meta" class="scroll-list-item-meta-container">
          <div class="scroll-list-item-meta"> {{ utils.capitalizeStr(item.meta) }}</div>
        </div>
        <div class="asset-item-container-footer">
          <div v-if="useItemId" class="asset-item-actions">
            <ActionButton v-if="item.can_edit" :icon="getAppIcon('edit')" v-tooltip="$t('components.scrollList.edit')"
              @click="editListItem(item.id)" />
            <ActionButton v-if="item.can_delete" :icon="getAppIcon('trash')"
              :class="{ 'item-inactive': buttonInactive(item.id) }" @click="deleteListItem(item.id)"
              v-tooltip="$t('components.scrollList.delete')" />
            <ActionButton v-if="assignItems" :icon="getAppIcon('person-plus')"
              :class="{ 'item-inactive': buttonInactive(item.id) }" @click="assignListItem(item.id)" :label="$t('components.scrollList.assign')"
              v-tooltip="$t('components.scrollList.assign')" />
            <ActionButton v-if="unassignItems" :icon="getAppIcon('person-minus')"
              :class="{ 'item-inactive': buttonInactive(item.id) }" @click="unassignListItem(item.id)"
              :label="$t('components.scrollList.unassign')" v-tooltip="$t('components.scrollList.unassign')" />
          </div>
          <div v-else class="asset-item-actions">
            <ActionButton v-if="item.can_edit" :icon="getAppIcon('edit')" v-tooltip="$t('components.scrollList.edit')"
              @click="editListItem(index)" />
            <ActionButton v-if="item.can_delete" :icon="getAppIcon('trash')"
              :class="{ 'item-inactive': buttonInactive(index) }" @click="deleteListItem(index)" v-tooltip="$t('components.scrollList.delete')" />
            <ActionButton v-if="assignItems" :icon="getAppIcon('person-plus')"
              :class="{ 'item-inactive': buttonInactive(index) }" @click="assignListItem(index)" :label="$t('components.scrollList.assign')"
              v-tooltip="$t('components.scrollList.assign')" />
            <ActionButton v-if="unassignItems" :icon="getAppIcon('person-minus')"
              :class="{ 'item-inactive': buttonInactive(index) }" @click="unassignListItem(index)" :label="$t('components.scrollList.unassign')"
              v-tooltip="$t('components.scrollList.unassign')" />
          </div>
        </div>
      </div>
    </div>

    <div v-else class="scroll-list-container" :class="{ 'scroll-list-container-wrap': wrapItems }">

      <div v-for="(item, index) in items" :key="index" :index="index" @click="selectItem(index)"
        class="scroll-list-item" :class="{ 'scroll-list-item-wrap': wrapItems, 'selected-item': selectedItem(index) }">
   
        <div class="asset-item-content">
          <div v-if="useAvatar" class="profile-picture" :style="{ backgroundColor: item.avatarColor }">
            <img class="profile-img" :src="item.profile ? item.profile : generateAvatar(item.id)">
          </div>
          <div v-else-if="useIcons" class="asset-item-icon-container">
            <img class="large-icons" :src="getAppIcon(item.icon)">
          </div>
          <div v-else-if="customIcons" class="asset-item-icon-container">
            <img class="large-icons no-filter" :src="item.icon">
          </div>
            <div class="asset-item-details">
              {{ item.name }}
            </div>
        </div>

        <div v-if="item.meta" class="scroll-list-item-meta-container">
          <div class="scroll-list-item-meta"> {{ utils.capitalizeStr(item.meta) }}</div>
        </div>

        <div class="asset-item-container-footer">
          <div v-if="useItemId" class="asset-item-actions">
            <ActionButton v-if="item.can_edit" :icon="getAppIcon('edit')" v-tooltip="$t('components.scrollList.edit')"
              @click="editListItem(item.id)" />
            <ActionButton v-if="item.can_delete" :icon="forCollab ? getAppIcon('person-minus') : getAppIcon('trash')"
              :class="{ 'item-inactive': buttonInactive(item.id) }" @click="deleteListItem(item.id)"
              v-tooltip="$t('components.scrollList.delete')" />
            <ActionButton v-if="assignItems" :icon="getAppIcon('person-plus')"
              :class="{ 'item-inactive': buttonInactive(item.id) }" @click="assignListItem(item.id)" :label="$t('components.scrollList.assign')"
              v-tooltip="$t('components.scrollList.assign')" />
            <ActionButton v-if="unassignItems" :icon="getAppIcon('person-minus')"
              :class="{ 'item-inactive': buttonInactive(item.id) }" @click="unassignListItem(item.id)"
              :label="$t('components.scrollList.unassign')" v-tooltip="$t('components.scrollList.unassign')" />
          </div>
          <div v-else-if="useItemName" class="asset-item-actions">
            <ActionButton v-if="item.can_edit" :icon="getAppIcon('edit')" v-tooltip="$t('components.scrollList.edit')"
              @click="editListItem(item.name)" />
            <ActionButton v-if="item.can_delete" :icon="forCollab ? getAppIcon('person-minus') : getAppIcon('trash')"
              :class="{ 'item-inactive': buttonInactive(item.name) }" @click="deleteListItem(item.name)"
              v-tooltip="$t('components.scrollList.delete')" />
            <ActionButton v-if="assignItems" :icon="getAppIcon('person-plus')"
              :class="{ 'item-inactive': buttonInactive(item.name) }" @click="assignListItem(item.name)" :label="$t('components.scrollList.assign')"
              v-tooltip="$t('components.scrollList.assign')" />
            <ActionButton v-if="unassignItems" :icon="getAppIcon('person-minus')"
              :class="{ 'item-inactive': buttonInactive(item.name) }" @click="unassignListItem(item.name)"
              :label="$t('components.scrollList.unassign')" v-tooltip="$t('components.scrollList.unassign')" />
          </div>
          <div v-else class="asset-item-actions">
            <ActionButton v-if="item.can_edit" :icon="getAppIcon('edit')" v-tooltip="$t('components.scrollList.edit')"
              @click="editListItem(index)" />
            <ActionButton v-if="item.can_delete" :icon="forCollab ? getAppIcon('person-minus') : getAppIcon('trash')"
              :class="{ 'item-inactive': buttonInactive(index) }" @click="deleteListItem(index)" v-tooltip="$t('components.scrollList.delete')" />
            <ActionButton v-if="assignItems" :icon="getAppIcon('person-plus')"
              :class="{ 'item-inactive': buttonInactive(index) }" @click="assignListItem(index)" :label="$t('components.scrollList.assign')"
              v-tooltip="$t('components.scrollList.assign')" />
            <ActionButton v-if="unassignItems" :icon="getAppIcon('person-minus')"
              :class="{ 'item-inactive': buttonInactive(index) }" @click="unassignListItem(index)" :label="$t('components.scrollList.unassign')"
              v-tooltip="$t('components.scrollList.unassign')" />
          </div>
        </div>
      </div>
    </div>
  </div>

</template>

<script setup>
import { useIconStore } from '@/stores/icons';
import { generateAvatar } from '@/lib/avatar';
const iconStore = useIconStore();

const getAppIcon = (iconName) => {
  const formattedIconName = getIconName(iconName)
  const icon = iconStore.getAppIcon(formattedIconName);
  return icon
};

const getIconName = (path) => {
  if (!path.includes('/') && !path.includes('.svg')) {
    return path;
  }
  return path.split('/').pop().replace('.svg', '');
};

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue'
import utils from "@/services/utils";

const props = defineProps({
  items: Array,
  isSingle: {
    type: Boolean,
    default: false
  },
  useAvatar: {
    type: Boolean,
    default: false
  },
  avatarColor: {
    type: String,
    default: '#7FFF00'
  },
  useItemId: {
    type: Boolean,
    default: false
  },
  useItemName: {
    type: Boolean,
    default: false
  },
  useIcons: {
    type: Boolean,
    default: false
  },
  customIcons: {
    type: Boolean,
    default: false
  },
  wrapItems: {
    type: Boolean,
    default: false
  },
  useMeta: {
    type: Boolean,
    default: false
  },
  editItems: {
    type: Boolean,
    default: false
  },
  deleteItems: {
    type: Boolean,
    default: false
  },
  assignItems: {
    type: Boolean,
    default: false
  },
  unassignItems: {
    type: Boolean,
    default: false
  },
  forCollab: {
    type: Boolean,
    default: false
  },
  buttonInactive: {
    type: Function,
    default: () => {
      return false
    }
  },
  selectItem: {
    type: Function,
    default: () => {
      console.log('selected')
    }
  },
  selectedItem: {
    type: Function,
    default: () => {
      return false
    }
  },
  editListItem: {
    type: Function,
    default: () => {
    }
  },
  deleteListItem: {
    type: Function,
    default: () => {
    }
  },
  assignListItem: {
    type: Function,
    default: () => {
    }
  },
  unassignListItem: {
    type: Function,
    default: () => {
    }
  },
});

</script>

<style scoped>
@import "@/assets/desktop.css";

.profile-picture {
  background-color: hsl(var(--destructive));
  height: 24px;
  min-width: 24px;
  overflow: hidden;
  display: flex;
  align-items: center;
  border-radius: 24px;
  /* padding: 5px; */
}

.profile-img {
  width: 100%;
  height: 100%;
}


.scroll-list-root {
  position: relative;
  display: flex;
  flex-direction: column;
  padding-right: .4rem;
  overflow: hidden;
  overflow-y: scroll;
  height: 100%;
  width: 100%;
  box-sizing: border-box;
  align-items: flex-start;
  min-width: 300px;
}

.scroll-list-root-single {
  box-sizing: border-box;
  height: min-content;
  height: 80px;
  overflow: hidden;
  /* background-color: hsl(var(--destructive)); */
}

.scroll-list-root::-webkit-scrollbar {
  width: 8px;
}

.scroll-list-root::-webkit-scrollbar-thumb {
  border-radius: var(--large-radius);
  background-color: hsl(var(--muted));
}

.scroll-list-root::-webkit-scrollbar-track {
  border-radius: var(--large-radius);
}

.scroll-list-item-meta {
  color: hsl(var(--foreground));
  background-color: hsl(var(--muted));
  padding: .3rem;
  border-radius: 5px;
  white-space: nowrap;
  font-size: 12px;
}

.scroll-list-item-meta-container {
  display: flex;
  height: 100%;
  box-sizing: border-box;
  align-items: center;
}

.scroll-list-container {
  box-sizing: border-box;
  display: flex;
  height: max-content;
  height: 100%;
  flex-direction: column;
  min-height: max-content;
  /* overflow: hidden; */
  width: 100%;
  border-radius: var(--large-radius);
  padding: 1%;
  height: 100px;
  height: max-content;
  gap: .5rem;
}

.scroll-list-container-wrap {
	display: grid;
	grid-template-columns: repeat(auto-fill, minmax(270px, 1fr));
	gap: 10px;
	width: 100%;

}

.scroll-list-container::-webkit-scrollbar {
  width: 4px;
}

.scroll-list-container::-webkit-scrollbar-thumb {
  border-radius: var(--large-radius);
  background-color: hsl(var(--border));
  /* background-color: rgb(236, 0, 0); */
}

.scroll-list-container::-webkit-scrollbar-track {
  border-radius: var(--large-radius);
}


.scroll-list-item {
  display: flex;
  gap: .2rem;
  color: hsl(var(--foreground));
  align-items: center;
  padding: .3rem 1rem;
  box-sizing: border-box;
  width: 100%;
  background-color: hsl(var(--muted));
  border-radius: var(--large-radius);
  overflow: hidden;
  height: 60px;
  border: 1px solid hsl(var(--border));
  
  /* background-color: hsl(var(--destructive)); */
  justify-content: space-between;
  border-radius: var(--large-radius);
  transition: all .2s ease-out;
}

.scroll-list-item-wrap {
  width: 100%;
}


.scroll-list-item:hover {
  background-color: hsl(var(--accent));
  border-radius: var(--small-radius);
}

.asset-item-container {
  display: flex;
  gap: .5rem;
  color: hsl(var(--foreground));
  align-items: center;
  padding: .4rem;
  box-sizing: border-box;
  width: 100%;
  height: 50px;
  justify-content: space-between;
}

.asset-item-container-cards {
  height: 100%;
  flex-direction: column;
}

.asset-item-container-selected {
  outline: 1px solid hsl(var(--primary));
  
}

.asset-item-container-selected:hover {
  outline: 1px solid hsl(var(--primary));
  
}

.selected-item {
  border: 1px solid hsl(var(--border));
  
}

.selected-item:hover {
  border: 1px solid hsl(var(--border));
  
}

.asset-spacer {
  display: flex;
  align-items: center;
  justify-content: center;
  max-width: 40px;
  min-width: 40px;
  aspect-ratio: 1/1;
  height: 100%;
  overflow: hidden;
  /* background-color: purple; */
  /* flex: 1; */
}

.asset-spacer-empty {
}

.checkboxes {
  width: 18px;
  height: 18px;
  border-radius: var(--tiny-radius);
  border: 2px solid yellow;
  background: #FFF;
  padding: 10px;
}

.action-column {
  text-align: center;
  padding: 2px;
  box-sizing: border-box;
  transition: all .3s ease-in;
}

.checkbox-container {
  display: flex;
  box-sizing: border-box;
  align-items: center;
  justify-content: center;
  padding: .1rem;
  overflow: hidden;
  min-width: min-content;
  height: 100%;
}


.asset-item-icon-container {
  display: flex;
  box-sizing: border-box;
  align-items: center;
  justify-content: center;
  min-width: min-content;
  padding: .1rem;
  overflow: hidden;
  height: 100%;
  /* background-color: hsl(var(--destructive)); */
}

.asset-item-content {
  gap: .4rem;
  display: flex;
  box-sizing: border-box;
  align-items: center;
  /* justify-content: space-between; */
  height: 100%;
  /* background-color: indigo; */
  width: 100%;
  overflow: hidden;
  /* max-width: 40%; */
}

.asset-item-content-cards {
  height: max-content;
}

.asset-item-details {
  /* display: flex; */
  padding: .2rem;
  flex-wrap: nowrap;
  overflow: hidden;
  box-sizing: border-box;
  align-items: center;
  justify-content: flex-start;
  /* width: 50%; */
  height: 100%;
  height: min-content;
  white-space: nowrap;
  text-overflow: ellipsis;

  /* background-color: hsl(var(--success)); */
}

.asset-item-meta {
  display: flex;
  padding: .2rem;
  box-sizing: border-box;
  align-items: center;
  gap: .4rem;
  width: 100%;
  height: 100%;
  overflow: hidden;
}

.asset-item-tag {
  display: flex;
  box-sizing: border-box;
  overflow: hidden;
  padding: .1rem .4rem;
  font-size: 12px;
  background-color: hsl(var(--background));
  border-radius: 20px;
}


.asset-item-status-container {
  display: flex;
  box-sizing: border-box;
  align-items: center;
  justify-content: center;
  width: min-content;
  padding: .2rem;
  height: 100%;
  /* flex: 1; */
}

.asset-item-container-footer {
  align-items: center;
  display: none;
  width: min-content;
  box-sizing: border-box;
  padding: .2rem;
  /* opacity: 0; */
  /* display: none; */
  /* width: 100%; */
  justify-content: flex-end;
}

.asset-item-container-footer-cards {
  width: 100%;
  justify-content: space-between;

}

.asset-item-status {
  display: flex;
  border-radius: var(--normal-radius);
  box-sizing: border-box;
  align-items: center;
  justify-content: center;
  width: min-content;
  width: 80px;
  padding: .4rem .4rem;
  height: max-content;
  /* background-color: hsl(var(--destructive)); */
  font-size: 12px;
  text-transform: uppercase;
  font-weight: 700;
  color: hsl(var(--background));
}

.asset-item-actions {
  display: flex;
  box-sizing: border-box;
  align-items: center;
  justify-content: space-between;
  width: max-content;
  min-width: max-content;
  padding: .2rem;
  /* opacity: 0; */
  /* gap: .7rem; */
  /* height: 100%; */
  /* background-color: darkcyan; */
  /* flex: 1; */
}

.asset-item-assignee {
  display: flex;
  box-sizing: border-box;
  align-items: center;
  justify-content: space-between;
  width: min-content;
  min-width: max-content;
  gap: .7rem;
  height: 100%;
  /* background-color: darkcyan; */
  /* flex: 1; */
}

.scroll-list-item:hover>*::nth-last-child(2) {
  background-color: hsl(var(--foreground));
  display: none;
  opacity: 0;
  transition: opacity 0.2s ease-in-out;
}

.scroll-list-item:hover>*:last-child {
  display: flex;
  opacity: 1;
  transition: opacity 0.2s ease-in-out;
}

</style>

