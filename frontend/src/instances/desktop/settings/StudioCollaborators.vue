<template>
  <div class="settings-component-root">
    <div class="settings-component-container">

        <div class="task-header">
			<div class="create-menu">
				<ActionButton :icon="getAppIcon('person-plus')" label="Add Collaborator" :showLabel="true"
					@click="addCollaborator" v-tooltip="'Add Collaborator'" />
				<ActionButton :icon="getAppIcon('refresh')" :label="'Refresh'" v-tooltip="'Refresh'"
					:buttonFunction="refresh" />
			</div>
		</div>


      <TaskListSkeleton v-if="!studioStore.studioUsers.length" />

        <div v-else class="project-list-container" ref="openProjectsContainer">
            <div class="project-list" >
                <CollaboratorItem v-for="collaborator, index in studioStore.studioUsers" :key="collaborator.id"
                    :collaborator="collaborator"
                    :style="{ animationDelay: index < 12 ? `${(index - 1) * 0.03}s` : '0s' }" />
            </div>
        </div>


    </div>
  </div>
</template>

<script setup>

// imports
import { ref, computed, onMounted } from 'vue';
import { useStudioStore } from '@/stores/studio';
import { useIconStore } from '@/stores/icons';

// store imports
import { useProjectStore } from '@/stores/projects';

// components
import CollaboratorItem from '@/instances/desktop/components/CollaboratorItem.vue';
import ActionButton from '@/instances/desktop/components/ActionButton.vue'
import TaskListSkeleton from '@/instances/desktop/components/TaskListSkeleton.vue'
import { useUserStore } from '@/stores/users';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { StudioService } from "@/../bindings/clustta/services";


// states
const studioStore = useStudioStore();
const iconStore = useIconStore();

const modals = useDesktopModalStore();

const addCollaborator = () => {
  modals.setModalVisibility('addCollaboratorModal', true);
};

const refresh = () => {
    studioStore.studioUsers = []
    studioStore.getStudioUsers();
};

const getAppIcon = (iconName) => {
	const icon = iconStore.getAppIcon(iconName);
	return icon
};

// computed props
onMounted( async () => {
    await studioStore.getStudioUsers();
})

</script>

<style scoped>
.input-short {
  flex: 1;
  width: 100%;
}

.settings-component-root {
  width: 100%;
  height: 100%;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  gap: 5px;
  box-sizing: border-box;
  align-items: center;
  justify-content: center;
}

.settings-component-container {
  display: flex;
  flex-direction: column;
  box-sizing: border-box;
  height: 100%;
  overflow: hidden;
  box-sizing: border-box;
  width: 96%;
  gap: .5rem;
  align-items: center;
  color: var(--white);
  /* justify-content: space-between; */
  border-radius: var(--large-radius);
  padding: 1rem;
  background-color: crimson;
  background-color: var(--black-steel);
}

.project-list {
	box-sizing: border-box;
	height: max-content;
	display: grid;
	grid-template-columns: repeat(auto-fill, minmax(500px, 1fr));
	gap: 10px;
	width: 100%;
}

.project-list-root {
	/* z-index: 5; */
	position: relative;
	padding: .5rem;
	overflow: hidden;
	/* overflow-y: scroll; */
	height: 100%;
	background-color: tomato;
	border-radius: var(--large-radius);
	background-color: var(--black-steel);
	width: 100%;
	box-sizing: border-box;
	min-width: 300px;
}

.project-list-root::-webkit-scrollbar {
	width: 8px;
}

.project-list-root::-webkit-scrollbar-thumb {
	border-radius: 10px;
	background-color: var(--dark-steel);
}

.project-list-root::-webkit-scrollbar-track {
	border-radius: 10px;
}

.project-list-root {
	z-index: 1;
	position: relative;
	display: flex;
	flex-direction: column;
	/* background-color: firebrick; */
	color: white;
	width: 100%;
	/* min-width: max-content; */
	/* max-width: 300px; */
	height: 100%;
	box-sizing: border-box;
	overflow: hidden;
	padding: .4rem;
	gap: .4rem;
	/* flex: 1 1 50%; */
}

.project-list-container {
	z-index: 1;
	position: relative;
	display: flex;
	flex-direction: column;
	/* background-color: firebrick; */
	color: white;
	width: 100%;
	/* height: 100%; */
	box-sizing: border-box;
	overflow: hidden;
	padding: .4rem;
	gap: .4rem;
	overflow: hidden;
	overflow-y: scroll;
}

.project-list-container::-webkit-scrollbar {
	width: 8px;
}

.project-list-container::-webkit-scrollbar-thumb {
	border-radius: 10px;
	background-color: var(--dark-steel);
}

.project-list-container::-webkit-scrollbar-track {
	border-radius: 10px;
}

.task-header {
	position: relative;
	display: flex;
	width: 100%;
	align-items: center;
	height: max-content;
	gap: 1rem;
	justify-content: space-between;
	/* background-color: khaki; */
	padding: .2rem;
	box-sizing: border-box;
	min-width: max-content;
}

.create-menu {
	position: relative;
	display: flex;
	align-items: center;
	gap: .4rem;
	width: max-content;
	height: max-content;
	padding: .2rem;
	/* background-color: black; */
}

.action-bar {
	position: relative;
	display: flex;
	align-items: center;
	gap: .4rem;
	width: max-content;
	height: max-content;
	padding: .2rem;
	/* background-color: black; */
}

.view-options {
	display: flex;
	gap: .4rem;
	align-items: center;
	padding: .2rem;
	width: max-content;
	height: max-content;
	/* background-color: darkorange; */
}

</style>

