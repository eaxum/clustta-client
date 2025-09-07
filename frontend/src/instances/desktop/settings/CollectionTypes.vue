<template>
  <div class="settings-component-root">
    <div class="settings-component-container">

      <ActionBar :itemType="'Add collection type'" :addFunction="addEntityType" />

      <ScrollList v-if="projectEntityTypes.length" :items="projectEntityTypes" :useIcons="true" :useItemId="true"
        :wrapItems="true" :editItems="true" :editListItem="prepEditEntityType" :deleteItems="true"
        :deleteListItem="deleteEntityType" />
      <PageState v-else :message="message()" :illustration="illustration()" :secondaryIcon="getAppIcon('plus-circle')"
        :secondaryActionMessage="secondaryActionMessage()" :secondaryActionFunction="secondaryActionFunction" />

    </div>
  </div>
</template>

<script setup>

// imports
import { onMounted, computed, ref } from 'vue';
import utils from '@/services/utils';

// store imports
import { useEntityStore } from '@/stores/entity';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useProjectStore } from '@/stores/projects';
import { useNotificationStore } from '@/stores/notifications';
import { useIconStore } from '@/stores/icons';

// components
import ScrollList from '@/instances/desktop/components/ScrollList.vue';
import ActionBar from '@/instances/desktop/components/ActionBar.vue';
import PageState from '@/instances/common/components/PageState.vue';
import { EntityService } from '@/../bindings/clustta/services/index';

// states
const entityStore = useEntityStore();
const modals = useDesktopModalStore();
const projectStore = useProjectStore();
const iconStore = useIconStore();
const notificationStore = useNotificationStore();

// computed props
const projectEntityTypes = computed(() => {
  
    let entityTypes = entityStore.getEntityTypes;
    let viewEntityTypeIds = [];
    let entities = entityStore.entities;

    for (const entity of entities){
      let entityTypeId = entity.entity_type_id;
      if(!viewEntityTypeIds.includes(entityTypeId)){
        viewEntityTypeIds.push(entityTypeId)
      }
    }

    const allTypes =  entityTypes.map(type => ({
      ...type,
      can_delete: !viewEntityTypeIds.includes(type.id),
      can_edit: type.name !== 'generic',
    }))


  return utils.sortAlphabetically(allTypes);

});


// methods
const getAppIcon = (iconName) => {
    const icon = iconStore.getAppIcon(iconName);
    return icon
};

const message = () => {
  return 'You have no collection types';
};

const illustration = () => {
  return '/page-states/resources.png';
};

const secondaryActionMessage = () => {
  return 'Add collection type'
};

const secondaryActionFunction = () => {
  addEntityType();
};

const addEntityType = () => {
  modals.setModalVisibility('addCollectionTypeModal', true);
};

const prepEditEntityType = (selectedEntityTypeId) => {

  console.log(selectedEntityTypeId)
  entityStore.selectedEntityType = entityStore.getEntityTypes.find((item) => item.id === selectedEntityTypeId)
  modals.setModalVisibility('editCollectionTypeModal', true);
};

const replaceSymbols = (name) => {
  return name.replace(/_/g, " ").toLowerCase().replace(/(^\w|\s\w)/g, match => match.toUpperCase());
};

const deleteEntityType = async (entityTypeId) => {
  EntityService.DeleteEntityType(projectStore.activeProject.uri, entityTypeId)
    .then((response) => {
      notificationStore.addNotification("Entity Type Deleted", "", "success");
      const index = entityStore.entityTypes.findIndex(entityType => entityType.id === entityTypeId);
      entityStore.entityTypes.splice(index, 1);
    })
    .catch((error) => {
      notificationStore.errorNotification("Error Deleting Entity Type", error);
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