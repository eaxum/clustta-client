<template>
    <div class="settings-component-root">
        <div class="settings-component-container">

            <div class="settings-component-header">

                <div v-if="projectTemplateStore.projectTemplates.length" class="settings-component-left">
                    <DropDownBox :items="projectTemplateStore.projectTemplateNames"
                        :onSelect="selectActiveProjectTemplate"
                        :selectedItem="projectTemplateStore.activeProjectTemplateName" :placeHolder="'None'"
                        :fixedWidth="true" />

                    <ActionButton :icon="getAppIcon('edit')" @click="editProjectTemplate" v-tooltip="'Edit Template'" />

                    <ActionButton :icon="getAppIcon('duplicate')" @click="duplicateProjectTemplate"
                        v-tooltip="'Duplicate Template'" />

                    <ActionButton :icon="getAppIcon('trash')" @click="prepDeletePopUpModal"
                        v-tooltip="'Delete Template'" />

                </div>

                <ActionButton :icon="getAppIcon('plus-circle')" @click="addNewProjectTemplate"
                    :label="'New Project Template'" v-tooltip="'New Template'" />

            </div>

            <div v-if="projectTemplateStore.projectTemplates.length" class="settings-component-tabs">
                <div class="menu-divider"></div>
                <HeaderTabs :dataTypes="templateContexts" @filter="filterList" :fullWidth="true" />
            </div>

            <div v-if="projectTemplateStore.projectTemplates.length" class="settings-component-body">

                <ScrollList v-if="projectTemplateStore.taskTemplates.length && activeTemplateContext === 'Templates'"
                    :items="projectTemplateStore.taskTemplates" :customIcons="true" :useItemId="true" :wrapItems="true"
                    :editItems="true" :editListItem="prepEditTemplate" :deleteItems="true"
                    :deleteListItem="deleteTaskTemplate" />

                <ScrollList v-else-if="projectTemplateStore.taskTypes.length && activeTemplateContext === 'Task types'"
                    :items="projectTemplateStore.taskTypes" :useIcons="true" :useItemId="true" :wrapItems="true"
                    :editItems="true" :editListItem="prepEditTaskType" :deleteItems="true"
                    :deleteListItem="deleteTaskType" />

                <ScrollList
                    v-else-if="projectTemplateStore.entityTypes.length && activeTemplateContext === 'Collection types'"
                    :items="projectTemplateStore.entityTypes" :useIcons="true" :useItemId="true" :wrapItems="true"
                    :editItems="true" :editListItem="prepEditEntityType" :deleteItems="true"
                    :deleteListItem="deleteEntityType" />

                <IgnoreListBox v-else-if="activeTemplateContext === 'Ignore List'" :placeholder="'Add item'"
                    :selectedItems="ignoreList" @itemAdded="addIgnoredItem" @itemRemoved="removeIgnoredItem" />

                <PageState v-else :message="message()" :illustration="illustration()" />

            </div>

            <div v-else class="settings-component-body">

                <PageState  :message="'You have no project templates'" :illustration="illustration()" />

                </div>

            <div v-if="projectTemplateStore.projectTemplates.length && activeTemplateContext !== 'Ignore List'" class="settings-component-footer">
                <ActionButton :icon="getAppIcon('plus-circle')" @click="contextAddFunction"
                    :label="contextPropmtMessage()" v-tooltip="'New Template'" />
            </div>


        </div>
    </div>
</template>

<script setup>
import { onMounted, computed, ref } from 'vue';
import utils from '@/services/utils';
import { TaskService } from "@/../bindings/clustta/services";

// store imports
import { useIconStore } from '@/stores/icons';
import { useNotificationStore } from '@/stores/notifications';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useSettingsStore } from '@/stores/settings';
import { useTrayStates } from '@/stores/TrayStates';

// components
import ScrollList from '@/instances/desktop/components/ScrollList.vue';
import ActionBar from '@/instances/desktop/components/ActionBar.vue';
import PageState from '@/instances/common/components/PageState.vue';
import DropDownBox from '@/instances/common/components/DropDownBox.vue';
import HeaderTabs from '@/instances/common/components/HeaderTabs.vue';
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import IgnoreListBox from '@/instances/common/components/IgnoreListBox.vue';

import {
    SettingsService,
    ProjectService,
    SyncService,
    FSService,
} from "@/../bindings/clustta/services";
import { useProjectTemplateStore } from '@/stores/project_template';

// states
const trayStates = useTrayStates();
const iconStore = useIconStore();
const notificationStore = useNotificationStore();
const modals = useDesktopModalStore();
const settings = useSettingsStore();
const projectTemplateStore = useProjectTemplateStore();


const ignoreList = ref([])

// computed
// const projectTemplates = computed(() => {
//     return settings.userProjectTemplates.project_templates;
// });

const templateContexts = computed(() => {
    return settings.templateContexts
});

const templateContextNames = computed(() => {
    return settings.templateContexts.map((context) => context.name)
});

const activeTemplateContext = ref(templateContextNames.value[0]);

const selectActiveProjectTemplate = (templateName) => {
    projectTemplateStore.selectActiveProjectTemplate(templateName)
    projectTemplateStore.reloadProjectTemplate()

    ignoreList.value = projectTemplateStore.activeProjectTemplate?.ignore_list
};

const filterList = (selectedContext) => {
    console.log(selectedContext)
    activeTemplateContext.value = selectedContext;
};

// methods
const message = () => {
    const templateContext = activeTemplateContext.value.toLowerCase();
    return `You have no ${templateContext}`;
};

const illustration = () => {
    return '/page-states/resources.png';
};

const contextPropmtMessage = () => {
    let templateContext = activeTemplateContext.value.toLowerCase();
    templateContext = templateContext.slice(0, -1); // Removes the last character
    return `Add a ${templateContext}`
};

const contextAddFunction = () => {
    let templateContext = activeTemplateContext.value;
    if (templateContext === 'Templates') {
        addTaskTemplate();
    } else if (templateContext === 'Asset types') {
        addTaskType();
    } else if (templateContext === 'Collection types') {
        addEntityType();
    } else {
        return
    }
};

const getAppIcon = (iconName) => {
    const icon = iconStore.getAppIcon(iconName);
    return icon
};

const addNewProjectTemplate = () => {
    modals.setModalVisibility('addProjectTemplateModal', true);
};

const editProjectTemplate = () => {
    modals.setModalVisibility('editProjectTemplateModal', true);
};

const duplicateProjectTemplate = () => {
    modals.setModalVisibility('duplicateProjectTemplateModal', true);

};

const deleteProjectTemplate = async () => {
    let projectTemplatesFolder = await FSService.UserProjectTemplatesPath()

    let templateName = projectTemplateStore.activeProjectTemplate.name + ".clst"

    let templatePath = await FSService.JoinPath(projectTemplatesFolder, templateName)
    await FSService.DeleteFile(templatePath).then(async (project) => {
        await projectTemplateStore.loadProjectTemplates()
    }).catch((error) => {
        console.log(error)
        notificationStore.errorNotification('Error creating project', error);
    });
    modals.disableAllModals()
};

const addIgnoredItem = (item) => {
    if (!ignoreList.value.includes(item)) {
        ignoreList.value.push(item);
        let projectUri = projectTemplateStore.activeProjectTemplate.uri
        ProjectService.SetIgnoreList(projectUri, "", ignoreList.value)
            .then((response) => {
                let projectTemplate = projectTemplateStore.activeProjectTemplate;
                let projectIndex = projectTemplateStore.projectTemplates.findIndex((p) => p.id === projectTemplate.id);
                projectTemplateStore.projectTemplates[projectIndex].ignore_list = ignoreList.value;
                projectTemplateStore.activeProjectTemplate.ignore_list = ignoreList.value;
            })
            .catch((error) => {
                notificationStore.addNotification('Failed to update ignore list', 'error');
            });
    }
};

const removeIgnoredItem = (item) => {
    ignoreList.value = ignoreList.value.filter((removedItem) => removedItem !== item)
    let projectUri = projectTemplateStore.activeProjectTemplate.uri
    // console.log(ignoreList.value)
    ProjectService.SetIgnoreList(projectUri, "", ignoreList.value)
        .then((response) => {
            let projectTemplate = projectTemplateStore.activeProjectTemplate;
            let projectIndex = projectTemplateStore.projectTemplates.findIndex((p) => p.id === projectTemplate.id);
            projectTemplateStore.projectTemplates[projectIndex].ignore_list = ignoreList.value;
            projectTemplateStore.activeProjectTemplate.ignore_list = ignoreList.value;
        })
        .catch((error) => {
            notificationStore.addNotification('Failed to update ignore list', 'error');
        });
};

const addTaskTemplate = () => {
    modals.setModalVisibility('addUserTemplateModal', true);
};

const prepEditTemplate = (selectedTemplateId) => {
    console.log(selectedTemplateId);
    modals.setModalVisibility('editUserTemplateModal', true);
};

const deleteTaskTemplate = async (taskTemplateId) => {
    projectTemplateStore.taskTemplates = projectTemplateStore.taskTemplates.filter((taskTemplate) => taskTemplate.id !== taskTemplateId);
};

// task types
const addTaskType = () => {
    modals.setModalVisibility('addUserAssetTypeModal', true);
};

const prepEditTaskType = (selectedTaskTypeId) => {

    console.log(selectedTaskTypeId);
    modals.setModalVisibility('editUserAssetTypeModal', true);

};

const deleteTaskType = async (taskTypeId) => {
    projectTemplateStore.taskTypes = projectTemplateStore.taskTypes.filter((taskType) => taskType.id !== taskTypeId);
};

// entity types
const addEntityType = () => {
    modals.setModalVisibility('addUserCollectionTypeModal', true);
};

const prepEditEntityType = (selectedEntityTypeId) => {
    console.log(selectedEntityTypeId);
    modals.setModalVisibility('editUserCollectionTypeModal', true);
};

const deleteEntityType = async (entityTypeId) => {
    projectTemplateStore.entityTypes = projectTemplateStore.entityTypes.filter((entityType) => entityType.id !== entityTypeId);
};

const prepDeletePopUpModal = () => {
    trayStates.popUpModalTitle = `Delete \"${projectTemplateStore.activeProjectTemplate.name}\"? `;
    trayStates.popUpModalMessage = "New projects will no longer display this template on creation.";
    trayStates.popUpModalFunction = deleteProjectTemplate;
    trayStates.popUpModalIcon = 'trash';
    modals.setModalVisibility('popUpModal', true);
};


// onMounted hook
onMounted(async () => {
    await projectTemplateStore.loadProjectTemplates()
    ignoreList.value = projectTemplateStore.activeProjectTemplate?.ignore_list
});
</script>


<style scoped>
@import "@/assets/desktop.css";

.menu-divider {
    margin-top: 20px;
    margin-bottom: 20px;
    width: 100%;
}

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
    /* align-items: center; */
    color: white;
    /* justify-content: space-between; */
    border-radius: var(--large-radius);
    padding: 1rem;
    background-color: crimson;
    background-color: var(--black-steel);
}

.settings-component-header {
    width: 100%;
    display: flex;
    gap: 10px;
    align-items: center;
    justify-content: space-between;
}

.settings-component-tabs {
    width: 100%;
    display: flex;
    flex-direction: column;
    /* background-color: darkgoldenrod; */
    align-items: flex-start;
    justify-content: space-between;
}

.settings-component-body {
    width: 100%;
    display: flex;
    gap: 10px;
    height: 100%;
    /* background-color: royalblue; */
    overflow: hidden;
}

.settings-component-footer {
    width: 100%;
    display: flex;
    gap: 10px;
    align-items: center;
    justify-content: center;
    /* background-color: brown; */
    height: min-content;
}

.settings-component-left {
    width: 100%;
    display: flex;
    gap: 10px;
    align-items: center;
}
</style>