<template>
  <div  class="scroll-list-container">
      <div v-for="(item, index) in items" :key="index" :index="index" class="scroll-list-item">
      <div v-if="useAvatar" class="profile-picture" :style="{ backgroundColor: item.avatarColor}">
        <img class="profile-img"  :src=" item.profile ? item.profile : generateAvatar(item.id)">
      </div>
      <img v-else-if="useIcons" class="small-icons" :src="item.icon">
      <div class="scroll-list-item-name"> {{ item.name }}</div>
      <div v-if="item.meta" class="scroll-list-item-meta"> {{ item.meta }}</div>

      <div v-if="useItemId" class="scroll-list-item-actions">
          <span v-if="editItems" v-stop-propagation class="single-action-button" @click="editListItem(item.id)" v-tooltip="$t('components.scrollList.edit')">
            <img class="small-icons" src="/icons/edit.svg">
          </span>
          <span v-if="deleteItems" v-stop-propagation class="single-action-button" :class="{'item-inactive' : buttonInactive(item.id)}" @click="deleteListItem(item.id)"  v-tooltip="$t('components.scrollList.delete')">
            <img class="small-icons" src="/icons/delete.svg">
          </span>
      </div>
      <div v-else class="scroll-list-item-actions">
          <span v-if="editItems" v-stop-propagation class="single-action-button" @click="editListItem(index)" v-tooltip="$t('components.scrollList.edit')">
            <img class="small-icons" src="/icons/edit.svg">
          </span>
          <span v-if="deleteItems" v-stop-propagation class="single-action-button" :class="{'item-inactive' : buttonInactive(index)}" @click="deleteListItem(index)"  v-tooltip="$t('components.scrollList.delete')">
            <img class="small-icons" src="/icons/delete.svg">
          </span>
      </div>

      </div>
  </div>
</template>


<script setup>
import { generateAvatar } from '@/lib/avatar';
const props = defineProps({
    items: Array,
    useAvatar: {
        type:Boolean, 
        default: false
    }, 
    avatarColor: {
        type:String, 
        default: '#7FFF00'
    }, 
    useIcons: {
        type:Boolean, 
        default: false
    }, 
    useItemId: {
        type:Boolean, 
        default: false
    }, 
    wrapItems: {
        type:Boolean, 
        default: false
    }, 
    useMeta: {
        type:Boolean, 
        default: false
    }, 
    editItems: {
        type:Boolean, 
        default: false
    }, 
    deleteItems: {
        type:Boolean, 
        default: false
    },
    buttonInactive: {
        type:Function, 
        default: () => {
          return false
        }
    },  
    editListItem: Function,   
    deleteListItem: Function,   
});
</script>

<style scoped>

.profile-picture {
  height: 24px;
  min-width: 24px;
  overflow: hidden;
  display: flex;
  align-items: center;
  border-radius: 24px;
}

.profile-img {
  width: 100%;
  height: 100%;
}

.scroll-list-item-name {
  font-family: 'Inter', sans-serif;
  font-weight: 400;
  color: hsl(var(--foreground));
  font-size: 0.875rem;
  display: flex;
  flex: 1;
  height: 100%;
  align-items: center;
  justify-content: flex-start;
}

.scroll-list-item-meta {
  color: hsl(var(--muted-foreground));
  background-color: hsl(var(--muted));
  padding: 0.25rem 0.5rem;
  border-radius: calc(var(--radius) - 4px);
  font-size: 0.75rem;
}

.scroll-list-container {
  box-sizing: border-box;
  padding: 0.5rem;
  align-items: center;
  flex-direction: column;
  gap: 0.125rem;
  background-color: hsl(var(--muted) / 0.5);
  overflow: hidden;
  overflow-y: scroll;
  width: 100%;
  border-radius: var(--radius);
  border: 1px solid hsl(var(--border));
}

.scroll-list-container::-webkit-scrollbar {
  width: 4px;
}

.scroll-list-container::-webkit-scrollbar-thumb {
  border-radius: var(--radius);
  background-color: hsl(var(--border));
}

.scroll-list-container::-webkit-scrollbar-track {
  border-radius: var(--radius);
}

.scroll-list-item {
  color: hsl(var(--foreground));
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.5rem;
  width: 100%;
  border-bottom: 1px solid hsl(var(--border));
  height: 40px;
  overflow: hidden;
  padding: 0.25rem;
  border-radius: calc(var(--radius) - 4px);
}

.scroll-list-item:last-child {
  border-bottom: 0px;
}

.scroll-list-item:hover>*:last-child {
  opacity: 1;
  visibility: visible;
  transition: opacity 0.2s ease-in-out;
  display: flex;
}

.scroll-list-item-actions {
  display: none;
  opacity: 0;
  visibility: hidden;
}
</style>

