<template>
  <div class="page-list-root absolute-pane">
    <div class="trash-list-stage-root">
      <div class="trash-list-stage-header">
        <HeaderTabs :useTooltip="false" :dataTypes="trashTypes" @filter="filterList" :fullWidth="true" />
      </div>
      <div class="trash-list-stage-body">
        <div class="trash-list-stage-body-root">
          <div class="trash-list-stage-body-container">

            <PageState v-if="!filteredtrashItems.length" :message="message()" :illustration="illustration()" />

            <TrashItem class="task-item" v-for="(trashItem, index) in filteredtrashItems" :key="index"
              :trashItem="trashItem" :trashItemIndex="index" :allTrash="filteredtrashItems"
              :style="{ animationDelay: index < 10 ? `${(index - 1) * 0.05}s` : '0s' }" />

          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
// imports
import { ref, computed, onMounted, onBeforeUnmount } from 'vue';

// services
import { TrashService } from "@/../bindings/clustta/services";

// state and store imports
import { useModalStore } from '@/stores/modals';
import { useTrayStates } from '@/stores/TrayStates';

// components
import TrashItem from '@/instances/desktop/components/TrashItem.vue';
import HeaderTabs from '@/instances/common/components/HeaderTabs.vue';
import PageState from '@/instances/common/components/PageState.vue';
import { useEntityStore } from '@/stores/entity';
import { useAssetStore } from '@/stores/assets';
import { onBeforeMount } from 'vue';
import { useProjectStore } from '@/stores/projects';

// states and stores
const trayStates = useTrayStates();
const modalStore = useModalStore();
const entityStore = useEntityStore();
const assetStore = useAssetStore();
const projectStore = useProjectStore();

// refs
const tasks = ref([]);
const resources = ref([]);
const filterIndex = ref(0);
const trashTypefilter = ref('all');

// computed
const trashTypes = computed(() => {
  const trashTypes = trayStates.trashTypes;
  // console.log(trashTypes);
  return trashTypes.filter(trashType => !trashType.name.includes('checkpoint'));
});

const sortedTrashItems = computed(() => {

  const trashables = trayStates.trashables;
  const allTrash = [];
  for (const key in trashables) {
    if (trashables.hasOwnProperty(key)) {
      const item = trashables[key];
      const entityName = getMeta(item.type, item.id, item.parent_id, item.data).name;
      allTrash.push({
        ...item,
        entity_name: entityName,
      });
    }
  };

  const ids = new Set(allTrash.map(item => item.id));
  const orphanItems = allTrash.filter(item => !ids.has(item.parent_id));
  const parentMap = {};

  orphanItems.forEach(item => {
    if (item.type.includes('checkpoint')) {
      if (!parentMap[item.parent_id]) {
        parentMap[item.parent_id] = {
          name: item.name.slice(0, -21),
          type: item.type.replace("_checkpoint", ""),
          id: item.parent_id,
          entity_name: item.entity_name,
          task_name: item.task_name,
          parent_id: item.parent_id,
          checkpoints: []
        };
      }
      parentMap[item.parent_id].checkpoints.push(item);
    }
    else {
      parentMap[item.id] = {
        ...item,
        checkpoints: []
      };
    }
  });

  const result = Object.values(parentMap).map(parent => {
    const originalItem = orphanItems.find(item => item.id === parent.id);
    if (originalItem) {
      return {
        ...originalItem,
        checkpoints: parent.checkpoints
      };
    }
    return parent;
  });

  // console.log(result);

  return sortByName(result);
});

const filteredtrashItems = computed(() => {
  const data = sortedTrashItems.value;
  if (!trayStates.showTraySearch || trayStates.trashSearchQuery === "") {
    if (trashTypefilter.value === 'all') {
      return data;
    } else {
      return data.filter((item) => item.type === trashTypefilter.value);
    }
  } else {
    if (trashTypefilter.value === 'all') {
      return data.filter((item) => item.name.toLowerCase().includes(trayStates.trashSearchQuery.toLowerCase()));
    } else {
      return data.filter((item) => item.type === trashTypefilter.value && item.name.toLowerCase().includes(trayStates.trashSearchQuery.toLowerCase()));
    }
  }
});


// methods
const message = () => {
  const type = trashTypefilter.value;
  if (type === 'all') { return 'You have not deleted any items' }
  else { return 'You have not deleted any ' + type + '.' };
};

const illustration = () => {
  const type = trashTypefilter.value;
  if(type === 'entity'){
    return  '/page-states/collections.png';
  } else if ( type === 'task'){
    return '/page-states/tasks.png'
  } else if ( type === 'resource'){
    return '/page-states/resources.png'
  } else if ( type === 'template'){
    return '/page-states/template.png'
  } else if ( type === 'all'){
    return '/page-states/resources.png'
  }
};

const secondaryActionMessage = () => {
  const type = trashTypefilter.value;
  if (type === 'all') { return 'Import resource' }
  else { return 'Import ' + type };
};

const highlightFilter = (index) => {
  filterIndex.value = index;
};

const sortByName = (arr) => {
  return arr.sort((a, b) => a.name.localeCompare(b.name));
};

const getMeta = (type, id, parent_id, data) => {

  if (type === 'task') {
    const entityName = tasks.value.find(item => item.id === id)?.entity_name;
    return { name: entityName, task: '' };

  } else if (type === 'task_checkpoint') {
    const entityName = tasks.value.find(item => item.id === parent_id)?.entity_name;
    console.log(entityName)
    return { name: entityName, task: '' };
  }
  else {
    return ''
  }

};

const findParentID = (id) => {
  const entityId = tasks.value.filter(item => item.id === id)[0].entity_id;
  const entityName = entityStore.entities.filter(item => item.id === entityId)[0].name;
  // console.log(entityName);
  return entityName;
};


const filterList = (trashType) => {
  let trashTypeName
  if(trashType === 'collections'){
    trashTypeName = 'entity'
  } else {
    trashTypeName = trashType
  }
  trashTypefilter.value = trashTypeName;
};

const editParams = (itemType) => {
  modalStore.setModalVisibility(itemType, true);
};

// onMounted hook
onBeforeMount(async () => {
  trayStates.showMeta = false;
  tasks.value = assetStore.tasks;
  trayStates.trashables = await TrashService.GetTrashs(projectStore.activeProject.uri);
});

onBeforeUnmount(async () => {
  trayStates.trashables = [];
});



</script>

<style scoped>
@import "@/assets/desktop.css";

.page-list-root {
  box-sizing: border-box;
  padding: .4rem;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  /* background-color: darkblue; */
}

.trash-list-stage-root {
  display: flex;
  flex-direction: column;
  box-sizing: border-box;
  /* background-color: firebrick; */
  width: 100%;
  height: 100%;
  gap: .5rem;

}

.trash-list-stage-header {
  width: 100%;
  display: flex;
  box-sizing: border-box;
  /* background-color: rebeccapurple; */
  align-items: flex-start;
  justify-content: flex-start;
}

.trash-list-stage-body {
  width: 100%;
  /* max-width: 1200px; */
  height: 100%;
  display: flex;
  box-sizing: border-box;
  /* background-color: teal; */
  align-items: flex-start;
  justify-content: flex-start;
  justify-content: center;
  overflow: hidden;
  padding: .5rem;
}

.trash-list-stage-body-root {

  /* background-color: tomato; */
  max-width: 960px;
  position: relative;
  display: flex;
  padding-right: .4rem;
  overflow: hidden;
  overflow-y: scroll;
  height: 100%;
  width: 100%;
  box-sizing: border-box;
  /* max-width: 600px; */
}

.trash-list-stage-body-root::-webkit-scrollbar {
  width: 8px;
}

.trash-list-stage-body-root::-webkit-scrollbar-thumb {
  border-radius: 10px;
  background-color: var(--dark-steel);
}

.trash-list-stage-body-root::-webkit-scrollbar-track {
  border-radius: 10px;
}

.trash-list-stage-body-container {
  width: 100%;
  gap: .5rem;
  max-width: 960px;
  /* height: 100%; */
  height: max-content;
  display: flex;
  box-sizing: border-box;
  flex-direction: column;
  /* background-color: tomato; */
  align-items: flex-start;
  justify-content: flex-start;
  overflow: hidden;
  padding: .5rem;
}

.page-list {
  background-color: rgb(125, 192, 59);
  padding: .4rem;
  display: flex;
  gap: .4rem;
  flex-direction: column;
  box-sizing: border-box;
  height: 100%;
  width: 100%;
  /* overflow: hidden; */
  height: max-content;
}

.page-list-container {
  display: flex;
  padding-right: .4rem;
  overflow: hidden;
  overflow-y: scroll;
  height: 100%;
  width: 100%;
  box-sizing: border-box;
  /* max-width: 600px; */
  min-width: 300px;
}

.page-list-container::-webkit-scrollbar {
  width: 8px;
}

.page-list-container::-webkit-scrollbar-thumb {
  border-radius: 10px;
  background-color: rgb(36, 49, 59);
}

.page-list-container::-webkit-scrollbar-track {
  border-radius: 10px;
}

.page-header {
  position: relative;
  display: flex;
  width: 100%;
  align-items: center;
  height: max-content;
  gap: 1rem;
  justify-content: space-between;
  background-color: khaki;
  padding: .2rem;
  box-sizing: border-box;
  min-width: max-content;
}
</style>


