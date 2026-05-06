<template>
  <div class="settings-component-root">
    <div class="settings-component-container">

      <ActionBar :itemType="$t('settings.addAssetType')" :addFunction="addAssetType" />

      <ScrollList v-if="projectAssetTypes.length" :items="projectAssetTypes" :useIcons="true" :useItemId="true"
        :wrapItems="true" :editItems="true" :editListItem="prepEditAssetType" :deleteItems="true"
        :deleteListItem="deleteAssetType" />

      <PageState v-else :message="message()" :illustration="illustration()" :secondaryIcon="CiPlusCircle"
        :secondaryActionMessage="secondaryActionMessage()" :secondaryActionFunction="secondaryActionFunction" />

    </div>
  </div>
</template>

<script setup>
import { useIconStore } from '@/stores/icons';
const iconStore = useIconStore();

const getAppIcon = (iconName) => {
    const icon = iconStore.resolveIcon(iconName);
    return icon
};

// imports
import { onMounted, computed, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import utils from '@/services/utils';
import { CiPlusCircle } from '@clustta/icons-vue';
import { resolveIcon } from '@/lib/icon-map';

// services
import { AssetService } from "@/services";

// store imports
import { useAssetStore } from '@/stores/assets';
import { useNotificationStore } from '@/stores/notifications';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useProjectStore } from '@/stores/projects';


// components
import ScrollList from '@/instances/desktop/components/ScrollList.vue';
import ActionBar from '@/instances/desktop/components/ActionBar.vue';
import PageState from '@/instances/common/components/PageState.vue';

// states
const assetStore = useAssetStore();
const notificationStore = useNotificationStore();
const modals = useDesktopModalStore();
const projectStore = useProjectStore();
const { t } = useI18n();

const projectAssetTypes = computed(() => {
  
  let assetTypes = assetStore.getAssetTypes;
  console.log(assetTypes)
  let viewAssetTypeIds = [];
  let assets = assetStore.assets;

  for (const asset of assets){
    let assetTypeId = asset.asset_type_id;
    if(!viewAssetTypeIds.includes(assetTypeId)){
      viewAssetTypeIds.push(assetTypeId)
    }
  }
  const restrictedNames = ['generic', 'weblink']
  const allTypes =  assetTypes.map(type => ({
    ...type,
    can_delete: !restrictedNames.includes(type.name),
    can_edit: !restrictedNames.includes(type.name),
  }))


return allTypes

});

// methods
const message = () => {
  return t('settings.noAssetTypes');
};

const illustration = () => {
  return '/page-states/resources.png';
};

const secondaryActionMessage = () => {
  return t('settings.addAssetType')
};

const secondaryActionFunction = () => {
  addAssetType();
};

const addAssetType = () => {
  modals.setModalVisibility('addAssetTypeModal', true);
};

const prepEditAssetType = (selectedAssetTypeId) => {
  console.log(selectedAssetTypeId)
  assetStore.selectedAssetType = assetStore.getAssetTypes.find((item) => item.id === selectedAssetTypeId)
  modals.setModalVisibility('editAssetTypeModal', true);

};

const replaceSymbols = (name) => {
  return name.replace(/_/g, " ").toLowerCase().replace(/(^\w|\s\w)/g, match => match.toUpperCase());
};

const deleteAssetType = async (assetTypeId) => {
  AssetService.DeleteAssetType(projectStore.activeProject.uri, assetTypeId)
    .then((response) => {
      notificationStore.addNotification(t('notifications.assetTypeDeleted'), "", "success");
      const index = assetStore.assetTypes.findIndex(assetType => assetType.id === assetTypeId);
      assetStore.assetTypes.splice(index, 1);
    })
    .catch((error) => {
      notificationStore.errorNotification(t('notifications.errorDeletingAssetType'), error);
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
  border-radius: var(--very-large-radius);
}
</style>




