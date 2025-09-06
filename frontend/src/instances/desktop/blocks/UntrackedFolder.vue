<template>
  <div ref="itemItem" class="item-main" v-stop-propagation
    :class="{ 'item-main-selected': stage.markedItems.includes(item.id) }">

    <div class="task-spacer" v-tooltip="'Untracted Item'" @click="console.log(item)">
      <span class="single-action-button single-action-button-disabled">
        <img class="small-icons entity-collapsed" src="/types-icons/folder.svg">
      </span>
    </div>

    <div class="main-item-root">

      <div class="item-container drop-zone">

        <div v-if="commonStore.showThumbs" class="item-preview-container">
          <div class="item-preview-image">
            <img v-if="item.preview" class="screenshot-thumb" :src="item.preview">
          </div>
        </div>

        <div class="item-icon-container">
          <img v-if="item.icon" class="large-icons" :src="item.icon">
          <span v-else class="app-ext">
          </span>
        </div>

        <div class="item-content">
          <div class="item-details">
            {{ itemName }}
            <!-- {{ item.id }} -->
          </div>
        </div>
        <ActionButton :icon="getAppIcon('plus-circle')" v-tooltip="'Add Folder'" @click="importItem" />

      </div>

    </div>
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
import { computed, ref, onMounted, onBeforeUnmount, nextTick, watch } from 'vue';

// states/store imports
import { useTrayStates } from '@/stores/TrayStates';
import { useTaskStore } from '@/stores/task';
import { useMenu } from '@/stores/menu';
import { usePaneStore } from '@/stores/panes';
import { useStageStore } from '@/stores/stages';
import { useUserStore } from '@/stores/users';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useNotificationStore } from '@/stores/notifications';
import { useCommonStore } from '@/stores/common';
import { useEntityStore } from '@/stores/entity';
import { useProjectStore } from '@/stores/projects';
import { useDndStore } from '@/stores/dnd';

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue'
import StatusMenu from '@/instances/desktop/menus/StatusMenu.vue'

// states/stores
const userStore = useUserStore();
const trayStates = useTrayStates();
const menu = useMenu();
const panes = usePaneStore();
const stage = useStageStore();
const modals = useDesktopModalStore();
const notificationStore = useNotificationStore();
const taskStore = useTaskStore();
const commonStore = useCommonStore();
const entityStore = useEntityStore();
const projectStore = useProjectStore();
const dndStore = useDndStore();

// emits
const emit = defineEmits(['update:height', 'expand']);

// props
const props = defineProps({
  item: Object,
  entityId: { type: String, default: '' },
  index: Number,
  isChild: { type: Boolean, default: false },
});

// refs
const itemItem = ref(null);
const isExpanded = ref(false);
const displayStatusMenu = ref(false);

// Update the item's height and emit the new height to the virtual list
const updateHeight = () => {
  nextTick(() => {
    emit('update:height', itemItem.value.offsetHeight)
  })
}

// Watch for changes in expansion and adjust height
watch(isExpanded, updateHeight)

// computed properties
const itemName = computed(() => {
  const item = props.item;
  const itemName = item.name;
  const isDirectParent = props.entityId === item.entity_id;
  const itemPath = item.item_path.replace(/\//g, ' / ');

  if (commonStore.showFullPath) {
    return itemPath
  }
  if (props.isChild) {
    if (commonStore.showChildEntities) {
      return itemName
    } else {
      return isDirectParent ? itemName : itemPath
    }
  } else {
    if (commonStore.viewSearchQuery) {
      return itemPath
    } else {
      return itemName
    }
  }
});

// methods
const closeStatusMenu = () => {
  displayStatusMenu.value = false;
};

const importItem = () => {
  const parentId = entityStore.entities.find((item) => item.entity_path === props.item.entity_path)?.id;
  const folderPath = props.item.file_path;

  dndStore.droppedFolders = []
  dndStore.targetItemId = parentId;
  dndStore.droppedFolders.push(folderPath);
  modals.setModalVisibility('importItemsModal', true);
};

const launchTaskCommand = async (index, item, event) => {
  if (!userStore.canDo('pull_chunk')) {
    return
  }

  let file_path = item.file_path
  if (await FSService.Exists(file_path)) {
    FSService.LaunchFile(file_path)
  } else {
    notificationStore.addNotification("File Not On Disk, Rebuild", "File not found on disk, rebuild task", "error")
  }
};

const prepTaskMenu = (index, item, event) => {
  menu.showContextMenu(event, 'itemMenu', true);
};

const openMenu = (event) => {
  console.log('item');
  const index = props.index;
  const item = props.item;
  prepTaskMenu(index, item, event);
};

const handleClickOutside = (event) => {
  if (displayStatusMenu.value) {
    displayStatusMenu.value = false;
  }
};

onMounted(() => {
  document.addEventListener('click', handleClickOutside);
});

onBeforeUnmount(() => {
  document.removeEventListener('click', handleClickOutside);
});

</script>

<style scoped>
@import "@/assets/desktop.css";

.single-action-button-disabled {
  pointer-events: none;
}

.task-collapsed {
  transform: rotate(-90deg);
}

.task-expanded {
  transform: rotate(0deg);
}

.chevron-inactive {
  opacity: .2;
}

.item-main {
  display: flex;
  gap: .2rem;
  color: white;
  align-items: center;
  padding-left: .5rem;
  box-sizing: border-box;
  width: 100%;
  height: min-content;
  justify-content: flex-end;
  align-items: flex-start;
  background-color: var(--dark-steel);
  border-radius: 10px;
  overflow: hidden;
  padding-right: 0px;

}

.item-main:hover {
  outline: var(--transparent-line);
  outline: 1px solid rgb(255, 255, 255);
  outline-offset: -1.5px;
}

.item-main-selected {
  outline: 1px solid rgb(255, 255, 255);
  outline: var(--transparent-line);
  outline-offset: -1px;
  background-color: var(--blue-steel);
}

.item-main-selected:hover {
  outline: 1px solid rgb(255, 255, 255);
  outline-offset: -1px;
}

.item-child {
  padding-left: 0px;
}

.main-item-root {
  display: flex;
  flex-direction: column;
  gap: .2rem;
  color: white;
  align-items: center;
  padding: .3rem;
  box-sizing: border-box;
  width: 100%;
  height: min-content;
  justify-content: flex-end;
  /* background-color: blue; */
  border-radius: 10px;
  overflow: hidden;
  padding-right: 0px;

}

.item-container {
  display: flex;
  gap: .5rem;
  color: white;
  align-items: center;
  padding: .4rem;
  box-sizing: border-box;
  width: 100%;
  height: 50px;
  justify-content: space-between;
  transition: all .3s ease-out;
}

.task-spacer {
  /* display: flex;
  align-items: center;
  justify-content: center;
  max-width: 40px;
  min-width: 40px;
  aspect-ratio: 1/1;
  height: 100%;
  overflow: hidden;
  box-sizing: border-box; */
  position: relative;
  width: min-content;
  width: 25px;
  height: 60px;
  display: flex;
  box-sizing: border-box;
  align-items: center;
  /* background-color: purple; */
  /* flex: 1; */
}

.item-preview-container {
  display: flex;
  box-sizing: border-box;
  align-items: center;
  justify-content: center;
  padding: .1rem;
  overflow: hidden;
  min-width: 60px;
  height: 100%;
  aspect-ratio: 16 / 9;
  /* background-color: firebrick; */
}

.item-preview-image {
  display: flex;
  box-sizing: border-box;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  height: 100%;
  aspect-ratio: 16 / 9;
  background-color: var(--light-steel);
  border-radius: 5px;
}

.item-icon-container {
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

.item-content {
  gap: .4rem;
  /* flex-direction: column; */
  display: flex;
  box-sizing: border-box;
  align-items: center;
  justify-content: space-between;
  height: 100%;
  /* background-color: indigo; */
  width: 100%;
  overflow: hidden;
}

.item-details {
  /* display: flex; */
  padding: .2rem;
  flex-wrap: nowrap;
  overflow: hidden;
  box-sizing: border-box;
  align-items: center;
  justify-content: space-between;
  /* width: 50%; */
  height: 100%;
  height: min-content;
  white-space: nowrap;
  text-overflow: ellipsis;

  /* background-color: forestgreen; */
}
</style>