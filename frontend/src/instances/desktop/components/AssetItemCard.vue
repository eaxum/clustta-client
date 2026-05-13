<template>
  <div class="collection-item-wrapper" @click="console.log(asset)">
    <div class="collection-item-main">
      <!-- thumbnail -->
      <div v-if="commonStore.showThumbs" class="collection-item-preview-container">
        <div class="collection-item-preview-image">
          <img v-if="displayThumbnail" class="screenshot-thumb" :class="{ 'fallback-icon': isFallbackIcon, 'no-filter': isFallbackIcon }" :src="displayThumbnail">
        </div>
      </div>
      <!-- name and app icon -->
      <div class="collection-item-data">
        <div class="collection-item-content">
          <div class="collection-item-info">
            <div class="collection-item-text">
              {{ utils.capitalizeStr(asset.name) }}
            </div>
          </div>
          <div v-if="asset.asset_type_icon" class="collection-item-icon-container">
            <img class="small-icons" :src="getAppIcon(asset.asset_type_icon)">
          </div>
        </div>
        <!-- assignee -->
        <div v-if="asset.assignee_id" class="collection-item-assignee" v-stop-propagation>
          <div class="single-action-button" v-tooltip="userFullName" @click="prepAssignAsset($event)">
            <div class="profile-picture" :style="{ backgroundColor: profileColor(asset.assignee_id) }">
                <img v-if="userPhoto" class="profile-img" :src="userPhoto">
                <img v-else class="profile-img" :src="generateAvatar(asset.assignee_id)">
            </div>
          </div>
        </div>
        <div v-else-if="userStore.canDo('assign_asset')" class="collection-item-assignee collection-item-assignee-button" v-stop-propagation >
          <ActionButton :icon="getAppIcon('person-plus')" v-tooltip="$t('blocks.assignAsset')" @click="prepAssignAsset($event)" />
        </div>
      </div>
    </div>
    <!-- tags -->
    <div v-if="visibleTags.length" class="collection-item-tags" v-stop-propagation>
      <div v-for="tag in visibleTags" :key="tag" class="collection-item-tag-chip" :class="{ 'is-active': isTagFilterActive(tag) }" v-tooltip="tag" @click="toggleTagFilter(tag)">
        {{ tag }}
      </div>
      <div v-if="overflowTags.length" class="collection-item-tag-chip overflow" v-tooltip="overflowTags.join(', ')">
        +{{ overflowTags.length }}
      </div>
    </div>
  </div>
</template>

<script setup>
// imports
import { computed } from 'vue';
import utils from '@/services/utils';
import { generateAvatar } from '@/lib/avatar';
import emitter from '@/lib/mitt';

// composables
import { useAssetThumbnail } from '@/composables/useAssetThumbnail';

// states/store imports
import { useTrayStates } from '@/stores/TrayStates';
import { useMenu } from '@/stores/menu';
import { usePaneStore } from '@/stores/panes';
import { useStageStore } from '@/stores/stages';
import { useUserStore } from '@/stores/users';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useNotificationStore } from '@/stores/notifications';
import { useCommonStore } from '@/stores/common';
import { useIconStore } from '@/stores/icons';

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue'
import { useAssetStore } from '@/stores/assets';

// states/stores
const userStore = useUserStore();
const trayStates = useTrayStates();
const menu = useMenu();
const panes = usePaneStore();
const stage = useStageStore();
const assetStore = useAssetStore();
const commonStore = useCommonStore();
const iconStore = useIconStore();

const getAppIcon = (iconName) => {
  const icon = iconStore.getAppIcon(iconName);
  return icon
};

// props
const props = defineProps({
  asset: Object,
  index: Number,
});

// thumbnail
const { displayThumbnail, isFallbackIcon } = useAssetThumbnail(
  () => props.asset,
  { enabled: () => commonStore.showThumbs, includeAssetIcon: true },
);

const userFullName = computed(() => {
  let user = userStore.getUserData(props.asset.assignee_id);
  if (!user) {
    return 'No User'
  } else {
    return `${user.first_name} ${user.last_name}`;
  }
});

const userPhoto = computed(() => {
  const userPhoto = userStore.userProfilePhoto(props.asset.assignee_id);
  return userPhoto
});

const MAX_VISIBLE_TAGS = 3;

const assetTags = computed(() => Array.isArray(props.asset.tags) ? props.asset.tags : []);

const visibleTags = computed(() => assetTags.value.slice(0, MAX_VISIBLE_TAGS));

const overflowTags = computed(() => assetTags.value.slice(MAX_VISIBLE_TAGS));

// Toggles the tag in the global asset filter when its chip is clicked.
const toggleTagFilter = (tag) => {
  const filters = commonStore.assetFilters;
  const idx = filters.findIndex((f) => f.type === 'tags' && f.name === tag);
  if (idx >= 0) {
    filters.splice(idx, 1);
  } else {
    filters.push({ type: 'tags', name: tag });
  }
  emitter.emit('refresh-browser');
};

// Returns true when the tag is currently active in the asset filters.
const isTagFilterActive = (tag) => {
  return commonStore.assetFilters.some((f) => f.type === 'tags' && f.name === tag);
};

const profileColor = (uuid) => {
  const parts = uuid.split('-');
  return '#' + parts[0];
};

// Opens the assign menu anchored at the click event for this asset.
const prepAssignAsset = (event) => {
  if (!userStore.canDo('assign_asset')) return;
  const asset = props.asset;
  assetStore.selectAsset(asset);
  stage.markedAssets = [asset.id];
  menu.showContextMenu(event, 'assignMenu', true);
};

</script>

<style scoped>
@import "@/assets/desktop.css";

.collection-item-wrapper {
  display: flex;
  flex-direction: column;
  gap: .2rem;
  color: var(--white);
  box-sizing: border-box;
  min-height: 80px;
  width: 100%;
  width: 250px;
  overflow: hidden;
  border-radius: var(--small-radius);
  background-color: var(--dark-steel);
  /* background-color: darkorange; */
  padding: .2rem;
  border-radius: var(--large-radius);
  outline: var(--transparent-line);
  outline-offset: -1px;
  transition: all .2s ease-out;
}

.collection-item-main {
  display: flex;
  gap: .2rem;
  align-items: center;
  width: 100%;
  height: 76px;
  overflow: hidden;
}


.collection-item-wrapper:hover {
  background-color: var(--steel);
  border-radius: var(--small-radius);
  outline: 1px solid var(--light-steel);
  outline-offset: -1px;
}

.collection-item-root {
  display: flex;
  flex-direction: column;
  gap: .5rem;
  color: var(--white);
  align-items: center;
  box-sizing: border-box;
  width: 100%;
  /* flex: 1; */
  overflow: hidden;
  /* max-width: 360px; */
  height: min-content;
  justify-content: space-between;
  /* transition: all .3s ease-out; */
}

/* .collection-item-root:hover>*:last-child {
  opacity: 0;
  transition: opacity 0.2s ease-in-out;
} */

.collection-item-root-minimized {
  width: 100%;
}

.collection-item-container {
  display: flex;
  flex-direction: column;
  gap: .5rem;
  color: var(--white);
  align-items: center;
  padding: .4rem;
  box-sizing: border-box;
  width: 100%;
  height: min-content;
  justify-content: space-between;
  transition: all .3s ease-out;
  /* background-color: violet; */
  border-radius: var(--normal-radius);
}

.collection-item-container-minimized {
  justify-content: center;
}

.collection-item-container-selected {
  outline: 1.5px solid rgb(255, 255, 255);
  outline-offset: -1.5px;
}

.collection-item-container-selected:hover {
  outline: 1.5px solid rgb(255, 255, 255);
  outline-offset: -1.5px;
}

.collection-item-preview-container {

  display: flex;
  box-sizing: border-box;
  align-items: center;
  justify-content: center;
  padding: .1rem;
  overflow: hidden;
  min-width: 60px;
  /* background-color: firebrick; */
  /* width: 100%; */
  height: 100%;
  aspect-ratio: 4 / 3;
  border-radius: var(--large-radius);
  transition: border-radius .2s ease-out;
}

.collection-item-wrapper:hover .collection-item-preview-container {
  border-radius: var(--tiny-radius);
}

.collection-item-preview-image {
  display: flex;
  position: relative;
  box-sizing: border-box;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  height: 100%;
  width: 100%;
  /* aspect-ratio: 4 / 3; */
  background-color: var(--black-steel);
  border-radius: 5px;
  pointer-events: none;
}

.collection-item-preview-image .screenshot-thumb.fallback-icon {
  width: 40%;
  height: 40%;
  max-width: 32px;
  max-height: 32px;
  object-fit: contain;
  /* opacity: 0.6; */
}

.collection-item-type-container {
  display: flex;
  box-sizing: border-box;
  align-items: center;
  justify-content: center;
  min-width: min-content;
  padding: .1rem;
  overflow: hidden;
  height: 100%;
  /* background-color: firebrick; */
}

.collection-item-icon-container {
  display: flex;
  box-sizing: border-box;
  align-items: center;
  justify-content: center;
  min-width: min-content;
  padding: .1rem;
  overflow: hidden;
  /* height: 100%; */
  /* background-color: firebrick; */
}

.collection-item-content {
  display: flex;
  box-sizing: border-box;
  align-items: center;
  /* background-color: goldenrod; */
  justify-content: space-between;
  width: 100%;
  gap: .3rem;
  padding: 0px .3rem;
  overflow: hidden;
  flex-shrink: 0;
}

.collection-item-data {
  display: flex;
  flex-direction: column;
  box-sizing: border-box;
  align-items: flex-start;
  height: 100%;
  /* background-color: goldenrod; */
  justify-content: space-between;
  justify-content: flex-start;
  /* align-items: flex-start; */
  width: 100%;
  gap: .3rem;
  padding: 0px .3rem;
  overflow: hidden;
}

.collection-item-info {
  display: flex;
  box-sizing: border-box;
  align-items: center;
  height: 100%;
  /* background-color: black; */
  justify-content: space-between;
  width: 100%;
  gap: .3rem;
  overflow: hidden;
}

.collection-item-text {
  padding: .2rem;
  flex-wrap: nowrap;
  overflow: hidden;
  box-sizing: border-box;
  align-items: center;
  height: 100%;
  height: min-content;
  white-space: nowrap;
  text-overflow: ellipsis;
  /* background-color: forestgreen; */
  font-size: 14px;
  /* font-weight: 200; */
}

.collection-item-meta {
  display: flex;
  padding: .2rem;
  box-sizing: border-box;
  align-items: center;
  gap: .4rem;
  width: 100%;
  height: 100%;
  overflow: hidden;
  /* background-color: rosybrown; */
}

.collection-item-type-indicator {
  color: rgb(219, 219, 219);
  color: var(--white);
  background-color: rgba(0, 0, 0, 0.216);
  padding: .3rem;
  border-radius: 5px;
  font-size: 12px;
}

.collection-item-tag {
  display: flex;
  box-sizing: border-box;
  overflow: hidden;
  padding: .1rem .4rem;
  font-size: 12px;
  background-color: black;
  border-radius: 20px;
}

.collection-item-tags {
  display: flex;
  flex-wrap: nowrap;
  align-items: center;
  gap: .2rem;
  width: 100%;
  padding: 0 .3rem;
  overflow: hidden;
}

.collection-item-tag-chip {
  display: inline-flex;
  align-items: center;
  max-width: 108px;
  padding: 2px 7px;
  font-size: 12px;
  line-height: 1.4;
  border-radius: 12px;
  border: 1px solid var(--steel);
  background-color: var(--black-steel);
  color: var(--white);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  cursor: pointer;
  transition: filter .15s ease-out, transform .15s ease-out;
}

.collection-item-tag-chip:hover {
  filter: brightness(1.25);
}

.collection-item-tag-chip.is-active {
  outline: 1px solid var(--white);
  outline-offset: -1px;
}

.collection-item-tag-chip.overflow {
  background-color: var(--black-steel);
  color: var(--light-steel);
  border-color: var(--steel);
}


.collection-item-assignee {
  display: flex;
  box-sizing: border-box;
  align-items: center;
  justify-content: flex-end;
  width: min-content;
  width: 100%;
  min-width: max-content;
  gap: .7rem;
  /* padding: 0px .3rem; */
  height: 100%;
  /* background-color: tomato; */
  /* flex: 1; */
}

.collection-item-assignee-name {
  font-size: 14px;
}

.collection-item-assignee-button {
  display: none;
}

.collection-item-wrapper:hover .collection-item-assignee-button {
  display: flex;
}

.collection-item-unassigned {
  font-size: 14px;
  opacity: .5;
  /* font-weight: 100; */
  font-style: italic;
}

.collection-item-actions {
  display: flex;
  box-sizing: border-box;
  align-items: center;
  justify-content: space-between;
  width: min-content;
  min-width: max-content;
  gap: .7rem;
  height: 100%;
  background-color: darkcyan;
  /* flex: 1; */
}
</style>

