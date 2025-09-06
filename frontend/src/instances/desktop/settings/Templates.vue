<template>
  <div class="settings-component-root">
    <div class="settings-component-container">

      <ActionBar v-if="userStore.canDo('create_template')" :itemType="'Add template'" :addFunction="addTemplate" />

      <ScrollList v-if="projectTemplates.length" :items="projectTemplates" :customIcons="true" :useItemId="true"
        :wrapItems="true" :editItems="true" :editListItem="prepEditTemplate" :deleteItems="true"
        :deleteListItem="deleteTemplate" />

      <PageState v-else :message="message()" :illustration="illustration()" :secondaryIcon="getAppIcon('plus-circle')"
        :secondaryActionMessage="secondaryActionMessage()" :secondaryActionFunction="secondaryActionFunction" />

    </div>
    
  </div>
</template>

<script setup>
import { useIconStore } from '@/stores/icons';
import { useUserStore } from '@/stores/users';
const iconStore = useIconStore();
const userStore = useUserStore();

const getAppIcon = (iconName) => {
    const icon = iconStore.getAppIcon(iconName);
    return icon
};

// imports
import { onMounted, computed, ref } from 'vue';
import utils from '@/services/utils';

// services
import { TemplateService } from "@/../bindings/clustta/services";
import { TrashService } from "@/../bindings/clustta/services";

// state imports
import { useTrayStates } from '@/stores/TrayStates';

// store imports
import { useNotificationStore } from '@/stores/notifications';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useTemplateStore } from '@/stores/template';
import { useProjectStore } from '@/stores/projects';


// components
import ScrollList from '@/instances/desktop/components/ScrollList.vue';
import ActionBar from '@/instances/desktop/components/ActionBar.vue';
import PageState from '@/instances/common/components/PageState.vue';

// states
const trayStates = useTrayStates();
const notificationStore = useNotificationStore();
const templateStore = useTemplateStore();
const modals = useDesktopModalStore();
const projectStore = useProjectStore();

// refs

const projectTemplates = computed(() => {
  const templates = templateStore.getTemplates;
  return templates.map((template) => ({
    ...template,
    can_delete: userStore.canDo('delete_template'),
    can_edit: userStore.canDo('update_template'),
  }))
});

// methods
const message = () => {
  return 'You have no templates';
};

const illustration = () => {
  return '/page-states/template.png';
};

const secondaryActionMessage = () => {
  return 'Add Template'
};

const secondaryActionFunction = () => {
  addTemplate();
};

const addTemplate = () => {
  modals.setModalVisibility('addTemplateModal', true);
};

const prepEditTemplate = (selectedTemplateId) => {
  const selectedTemplate = templateStore.getTemplates.find((template) => template.id === selectedTemplateId);
  templateStore.selectedTemplate = selectedTemplate;
  modals.setModalVisibility('editTemplateModal', true);
};

const replaceSymbols = (name) => {
  return name.replace(/_/g, " ").toLowerCase().replace(/(^\w|\s\w)/g, match => match.toUpperCase());
};

const deleteTemplate = async (selectedTemplateId) => {
  let template = templateStore.getTemplates.find((template) => template.id === selectedTemplateId);
  await TemplateService.DeleteTemplate(projectStore.activeProject.uri, template.name)
    .then((response) => {
      templateStore.markTemplateAsDeleted(template.id);
      trayStates.undoItemId = template.id;
      trayStates.undoFunction = undoTemplateDelete;
    })
    .catch((error) => {
      notificationStore.errorNotification('Error deleting template', error);
    });
  let longMessage = `Template of name: ${template.name} was moved to Trash.`
  notificationStore.addNotification("Template moved to Trash.", longMessage, "success", true)
};

const undoTemplateDelete = async () => {
  TrashService.Restore(projectStore.activeProject.uri, trayStates.undoItemId, "task")
    .then(async (response) => {
      templateStore.unmarkTemplateAsDeleted(trayStates.undoItemId)
    })
    .catch((error) => {
      notificationStore.errorNotification("Error Restoring Item", error)
    });
};

// onMounted hook
onMounted(async () => {

});
</script>


<style scoped>
.input-short {
  flex: 1;
  width: 100%;
}

.settings-component-root{
  width:100%;
  height:100%;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  gap: 5px;
  box-sizing: border-box;
  align-items: center;
  justify-content: center;
}

.settings-component-container{
  display: flex;
  flex-direction: column;
  box-sizing: border-box;
  height: 100%;
  overflow: hidden;
  box-sizing: border-box;
  width: 96%;
  gap: .5rem;
  align-items: center;
  color: white;
  justify-content: space-between;
  border-radius: var(--large-radius);
  padding: 1rem;
  background-color: crimson;
  background-color: var(--black-steel);
}
</style>