<template>
    <div class="general-pane-root">
      <div v-if="taskDependencies.length" class="sidebar-scroll">
        <DependencyVirtualScroll :itemComponent="VirtualNode" :items="taskDependencies" :useRef="false" :isDependency="true" :showAdd="false" />
        
        <div class="bottom-bar">
          <ActionButton :icon="getAppIcon('square-arrow-right-up')" :showLabel="true" :iconAfter="true" :fullWidth="false" label="View in Graph"
          :buttonFunction="goToDependencyGraph" />
        </div>
      </div>

      <PageState v-else :message="message()" :illustration="illustration()" />

    </div>

</template>

<script setup>
// imports
import { computed, ref, watch, onMounted, onUnmounted } from 'vue';
import { TaskService } from "@/../bindings/clustta/services";
import emitter from '@/lib/mitt';
import utils from '@/services/utils';
import { isValidWeblink } from '@/lib/pointer';

// services
import { EntityService } from "@/../bindings/clustta/services";

// states/store imports
import { useStageStore } from '@/stores/stages';
import { useNotificationStore } from '@/stores/notifications';
import { useTaskStore } from '@/stores/task';
import { useIconStore } from '@/stores/icons';
import { useProjectStore } from '@/stores/projects';

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue'
import DependencyVirtualScroll from '@/instances/desktop/components/DependencyVirtualScroll.vue';
import VirtualNode from '@/instances/desktop/components/VirtualNode.vue';
import PageState from '@/instances/common/components/PageState.vue';

// states/stores
const stage = useStageStore();
const notificationStore = useNotificationStore();
const taskStore = useTaskStore();
const projectStore = useProjectStore();
const iconStore = useIconStore();

const message = () => {
  return 'This asset has no dependencies';
};

const illustration = () => {
  return '/page-states/resources.png';
};

const getAppIcon = (iconName) => {
  const icon = iconStore.getAppIcon(iconName);
  return icon
};

const selectedTask = computed(() => {
  return taskStore.selectedTask
});

const goToDependencyGraph = () => {
  stage.setStageVisibility('dependencies', true);
};

// refs
const taskDependencies = ref([]);

// Helper function to emit task data updates
const emitTaskUpdates = (taskId, updates) => {
  const updateData = { itemId: taskId, updates };
  
  // Emit to both Browser and VirtuaItem components
  emitter.emit('update-root-data', updateData);
  emitter.emit('update-children', updateData);
};

watch(() => taskStore.selectedTask, async () => {
    getTaskDependencies();
}, { deep: true });

const getTaskDependencies = async() => {
	let project = projectStore.activeProject
  let allDependencies;
  const selectedTaskDependencies = taskStore.selectedTask?.dependencies;
  const selectedTaskEntityDependencies = taskStore.selectedTask?.entity_dependencies;
  allDependencies = [ ...selectedTaskDependencies, ...selectedTaskEntityDependencies];
  const children = await TaskService.GetTaskDependencies(project.uri, allDependencies);

  for (let i = 0; i < children.length; i++) {
      let item = children[i];
      let extension = "";
      if (item.pointer) {
        item.file_path = item.pointer;
      }
      if (item.is_link && !isValidWeblink(item.pointer)) {
        extension = utils.getFileExtension(item.pointer).substring(1);
      } else if (!item.is_link) {
        extension = children[i].extension?.toLowerCase().substring(1);
        if (!taskStore.projectExtensionsFlat.includes(extension)) {
          taskStore.projectExtensionsFlat.push(extension);
          let extensionData = {
            name: extension,
            type: "extension",
            extension: "." + extension,
            icon: (await iconStore.getIcon(extension)) || "",
          };
          taskStore.projectExtensions.push(extensionData);
        }
      }
      let iconPath = "";
      if(item.type === "task"){
        if (item.is_link && isValidWeblink(item.pointer)) {
          iconPath = await iconStore.getWebIcon(item.pointer);
        } else {
          iconPath = (await iconStore.getIcon(extension)) || "";
        }
      } else{
          // iconPath = getAppIcon('folder')
      }
      children[i].icon = iconPath;
      let preview = null;
      if (item.preview) {
        preview = "data:image/png;base64," + item.preview;
      }
      children[i].preview = preview;
    }

    taskDependencies.value = children;
    console.log(children);
}

const handleRemoveDependency = (payload) => {
  removeDependency(payload.id, payload.itemType);
};

const removeDependency = async (dependencyId, itemType) => {
  const task = taskStore.selectedTask;
  let selectedTaskDependencies;
  console.log(itemType)

  if (itemType === "task") {
    selectedTaskDependencies = task.dependencies;
    await TaskService.RemoveTaskDependency(projectStore.activeProject.uri, task.id, dependencyId)
      .then((response) => {
        notificationStore.addNotification("Dependency Removed", "", "success");
        taskDependencies.value = taskDependencies.value.filter((task) => task.id !== dependencyId)
        emitTaskUpdates(task.id, [
          { property: 'dependencies', value: selectedTaskDependencies.filter(dep => dep !== dependencyId) }
        ])
      })
      .catch((error) => {
        notificationStore.errorNotification("Error removing dependencies", error);
      });
  } else {
    selectedTaskDependencies = task.entity_dependencies;
    await TaskService.RemoveEntityDependency(projectStore.activeProject.uri, task.id, dependencyId)
      .then((response) => {
        notificationStore.addNotification("Dependency Removed", "", "success");
        taskDependencies.value = taskDependencies.value.filter((task) => task.id !== dependencyId)
        emitTaskUpdates(task.id, [
          { property: 'entity_dependencies', value: selectedTaskDependencies.filter(dep => dep !== dependencyId) }
        ])
      })
      .catch((error) => {
        notificationStore.errorNotification("Error removing dependencies", error);
      });
  }
  
};


// onMounted hook
onMounted( async () => {
  await getTaskDependencies();
  emitter.on('removeDependency', handleRemoveDependency);
});


onUnmounted(() => {
  emitter.off('removeDependency', handleRemoveDependency)
});


</script>
<style scoped>
@import "@/assets/desktop.css";

.bottom-bar{
  /* background-color: crimson; */
  display: flex;
  justify-content: flex-end;
}
.sidebar-scroll {
  display: flex;
  flex-direction: column;
  overflow: hidden;
  /* overflow-y: scroll; */
  box-sizing: border-box;
  gap: .4rem;
  padding: .5rem;
  position: relative;
  height: 100%;
  width: 100%;
  justify-content: flex-start;
  padding: 5px;
}

.sidebar-scroll::-webkit-scrollbar {
  width: 4px;
}

.sidebar-scroll::-webkit-scrollbar-thumb {
  border-radius: 10px;
  background-color: var(--white);
}

.sidebar::-webkit-scrollbar-track {
  border-radius: 10px;
}

.action-bar {
  position: relative;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: .6rem;
  width: max-content;
  width: 100%;
  /* justify-content: space-around; */
  height: max-content;
  padding: .2rem;
  /* background-color: black; */
  /* background-color: tomato; */
  align-items: flex-start;
}
</style>