<template>
  <div class="settings-component-root">
    <div class="settings-component-container">

      <ActionBar :itemType="$t('settings.addCollectionType')" :addFunction="addCollectionType" />

      <ScrollList v-if="projectCollectionTypes.length" :items="projectCollectionTypes" :useIcons="true" :useItemId="true"
        :wrapItems="true" :editItems="true" :editListItem="prepEditCollectionType" :deleteItems="true"
        :deleteListItem="deleteCollectionType" />
      <PageState v-else :message="message()" :illustration="illustration()" :secondaryIcon="getAppIcon('plus-circle')"
        :secondaryActionMessage="secondaryActionMessage()" :secondaryActionFunction="secondaryActionFunction" />

    </div>
  </div>
</template>

<script setup>

// imports
import { onMounted, computed, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import utils from '@/services/utils';

// store imports
import { useCollectionStore } from '@/stores/collections';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useProjectStore } from '@/stores/projects';
import { useNotificationStore } from '@/stores/notifications';
import { useIconStore } from '@/stores/icons';

// components
import ScrollList from '@/instances/desktop/components/ScrollList.vue';
import ActionBar from '@/instances/desktop/components/ActionBar.vue';
import PageState from '@/instances/common/components/PageState.vue';
import { CollectionService } from '@/services';

// states
const collectionStore = useCollectionStore();
const modals = useDesktopModalStore();
const projectStore = useProjectStore();
const iconStore = useIconStore();
const notificationStore = useNotificationStore();
const { t } = useI18n();

// computed props
const projectCollectionTypes = computed(() => {
  
    let collectionTypes = collectionStore.getCollectionTypes;
    let viewCollectionTypeIds = [];
    let collections = collectionStore.collections;

    for (const collection of collections){
      let collectionTypeId = collection.collection_type_id;
      if(!viewCollectionTypeIds.includes(collectionTypeId)){
        viewCollectionTypeIds.push(collectionTypeId)
      }
    }

    const allTypes =  collectionTypes.map(type => ({
      ...type,
      can_delete: !viewCollectionTypeIds.includes(type.id),
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
  return t('settings.noCollectionTypes');
};

const illustration = () => {
  return '/page-states/resources.png';
};

const secondaryActionMessage = () => {
  return t('settings.addCollectionType')
};

const secondaryActionFunction = () => {
  addCollectionType();
};

const addCollectionType = () => {
  modals.setModalVisibility('addCollectionTypeModal', true);
};

const prepEditCollectionType = (selectedCollectionTypeId) => {

  console.log(selectedCollectionTypeId)
  collectionStore.selectedCollectionType = collectionStore.getCollectionTypes.find((item) => item.id === selectedCollectionTypeId)
  modals.setModalVisibility('editCollectionTypeModal', true);
};

const replaceSymbols = (name) => {
  return name.replace(/_/g, " ").toLowerCase().replace(/(^\w|\s\w)/g, match => match.toUpperCase());
};

const deleteCollectionType = async (collectionTypeId) => {
  CollectionService.DeleteCollectionType(projectStore.activeProject.uri, collectionTypeId)
    .then((response) => {
      notificationStore.addNotification(t('notifications.collectionTypeDeleted'), "", "success");
      const index = collectionStore.collectionTypes.findIndex(collectionType => collectionType.id === collectionTypeId);
      collectionStore.collectionTypes.splice(index, 1);
    })
    .catch((error) => {
      notificationStore.errorNotification(t('notifications.errorDeletingCollectionType'), error);
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
  color: hsl(var(--foreground));
  justify-content: space-between;
  
  padding: 1rem;
  background-color: hsl(var(--destructive));
  background-color: hsl(var(--background));
  
}
</style>