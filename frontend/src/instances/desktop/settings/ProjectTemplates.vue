<template>
    <div class="settings-component-root">
        <div class="settings-component-container">

            <div class="settings-component-header">

                <div v-if="projectTemplateStore.projectTemplates.length" class="settings-component-left">
                    <DropDownBox :items="projectTemplateStore.projectTemplateNames"
                        :onSelect="selectActiveProjectTemplate"
                        :selectedItem="projectTemplateStore.activeProjectTemplateName" :placeHolder="'None'"
                        :fixedWidth="true" />

                    <ActionButton :icon="getAppIcon('edit')" @click="editProjectTemplate" v-tooltip="$t('settings.editTemplate')" />

                    <ActionButton :icon="getAppIcon('duplicate')" @click="duplicateProjectTemplate"
                        v-tooltip="$t('settings.duplicateTemplate')" />

                    <ActionButton :icon="getAppIcon('trash')" @click="prepDeletePopUpModal"
                        v-tooltip="$t('settings.deleteTemplate')" />

                </div>

                <ActionButton :icon="getAppIcon('plus-circle')" @click="addNewProjectTemplate"
                    :label="$t('settings.newProjectTemplate')" v-tooltip="$t('settings.newTemplate')" />

            </div>

            <div v-if="projectTemplateStore.projectTemplates.length" class="settings-component-tabs">
                <div class="menu-divider"></div>
                <HeaderTabs :dataTypes="templateContexts" @filter="filterList" :fullWidth="true" />
            </div>

            <div v-if="projectTemplateStore.projectTemplates.length" class="settings-component-body">

                <ScrollList v-if="projectTemplateStore.assetTemplates.length && activeTemplateContext === 'templates'"
                    :items="assetTemplates" :customIcons="true" :useItemId="true" :wrapItems="true"
                    :editItems="true" :editListItem="prepEditTemplate" :deleteItems="true"
                    :deleteListItem="deleteAssetTemplate" />

                <ScrollList v-else-if="projectTemplateStore.assetTypes.length && activeTemplateContext === 'assettypes'"
                    :items="assetTypes" :useIcons="true" :useItemId="true" :wrapItems="true"
                    :editItems="true" :editListItem="prepEditAssetType" :deleteItems="true"
                    :deleteListItem="deleteAssetType" />

                <ScrollList
                    v-else-if="projectTemplateStore.collectionTypes.length && activeTemplateContext === 'collectiontypes'"
                    :items="collectionTypes" :useIcons="true" :useItemId="true" :wrapItems="true"
                    :editItems="true" :editListItem="prepEditCollectionType" :deleteItems="true"
                    :deleteListItem="deleteCollectionType" />

                <IgnoreListBox v-else-if="activeTemplateContext === 'ignorelist'" :placeholder="$t('placeholders.addItem')"
                    :selectedItems="ignoreList" @itemAdded="addIgnoredItem" @itemRemoved="removeIgnoredItem" />

                <PageState v-else :message="message()" :illustration="illustration()" />

            </div>

            <div v-else class="settings-component-body">

                <PageState  :message="$t('settings.noProjectTemplates')" :illustration="illustration()" />

                </div>

            <div v-if="projectTemplateStore.projectTemplates.length && activeTemplateContext !== 'ignorelist'" class="settings-component-footer">
                <ActionButton :icon="getAppIcon('plus-circle')" @click="contextAddFunction"
                    :label="contextPropmtMessage()" v-tooltip="$t('settings.newTemplate')" />
            </div>


        </div>
    </div>
</template>

<script setup>
import { onMounted, computed, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import utils from '@/services/utils';
import { AssetService } from "@/services";

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
} from "@/services";
import { useProjectTemplateStore } from '@/stores/project_template';

// states
const trayStates = useTrayStates();
const iconStore = useIconStore();
const notificationStore = useNotificationStore();
const modals = useDesktopModalStore();
const settings = useSettingsStore();
const projectTemplateStore = useProjectTemplateStore();
const { t } = useI18n();


const ignoreList = ref([])

// computed
// const projectTemplates = computed(() => {
//     return settings.userProjectTemplates.project_templates;
// });

const templateContexts = computed(() => {
    return settings.templateContexts
});

const assetTemplates = computed(() => {

    let assetTemplates = projectTemplateStore.assetTemplates;
    const allTypes =  assetTemplates.map(type => ({
    ...type,
    can_delete: true,
    can_edit: true,
  }));

  return allTypes
})

const assetTypes = computed(() => {

    let assetTypes = projectTemplateStore.assetTypes;
    const allTypes =  assetTypes.map(type => ({
    ...type,
    can_delete: true,
    can_edit: true,
  }));

  return allTypes
})

const collectionTypes = computed(() => {

    let collectionTypes = projectTemplateStore.collectionTypes;
    const allTypes =  collectionTypes.map(type => ({
    ...type,
    can_delete: true,
    can_edit: true,
  }));

  return allTypes
})

const templateContextNames = computed(() => {
    return settings.templateContexts.map((context) => context.id)
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
    const contextMessages = {
        templates: t('settings.noTemplates'),
        assettypes: t('settings.noAssetTypes'),
        collectiontypes: t('settings.noCollectionTypes'),
        ignorelist: t('settings.noIgnoreItems'),
    };
    return contextMessages[activeTemplateContext.value] || '';
};

const illustration = () => {
    return '/page-states/template.png';
};

const contextPropmtMessage = () => {
    const contextLabels = {
        templates: t('settings.addATemplate'),
        assettypes: t('settings.addAnAssetType'),
        collectiontypes: t('settings.addACollectionType'),
    };
    return contextLabels[activeTemplateContext.value] || '';
};

const contextAddFunction = () => {
    let templateContext = activeTemplateContext.value;
    if (templateContext === 'templates') {
        addAssetTemplate();
    } else if (templateContext === 'assettypes') {
        addAssetType();
    } else if (templateContext === 'collectiontypes') {
        addCollectionType();
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

const addAssetTemplate = () => {
    modals.setModalVisibility('addUserTemplateModal', true);
};

const prepEditTemplate = (selectedTemplateId) => {
    console.log(selectedTemplateId);
    modals.setModalVisibility('editUserTemplateModal', true);
};

const deleteAssetTemplate = async (assetTemplateId) => {
    projectTemplateStore.assetTemplates = projectTemplateStore.assetTemplates.filter((assetTemplate) => assetTemplate.id !== assetTemplateId);
};

// asset types
const addAssetType = () => {
    modals.setModalVisibility('addUserAssetTypeModal', true);
};

const prepEditAssetType = (selectedAssetTypeId) => {
    projectTemplateStore.selectedAssetTypeId = selectedAssetTypeId;
    modals.setModalVisibility('editUserAssetTypeModal', true);
};

const deleteAssetType = async (assetTypeId) => {
    projectTemplateStore.assetTypes = projectTemplateStore.assetTypes.filter((assetType) => assetType.id !== assetTypeId);
};

// collection types
const addCollectionType = () => {
    modals.setModalVisibility('addUserCollectionTypeModal', true);
};

const prepEditCollectionType = (selectedCollectionTypeId) => {
    projectTemplateStore.selectedCollectionTypeId = selectedCollectionTypeId;
    modals.setModalVisibility('editUserCollectionTypeModal', true);
};

const deleteCollectionType = async (collectionTypeId) => {
    projectTemplateStore.collectionTypes = projectTemplateStore.collectionTypes.filter((collectionType) => collectionType.id !== collectionTypeId);
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
    width: 100%;
    gap: .5rem;
    /* align-items: center; */
    color: white;
    /* justify-content: space-between; */
    padding: 1rem;
    background-color: var(--surface-1);
    border-radius: var(--very-large-radius);
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

