<template>
  <div class="general-pane-header">

  </div>

</template>

<script setup>
// imports
import { computed, ref, onMounted, onBeforeUnmount } from 'vue';
import { SettingsService, ProjectService, SyncService } from "@/../bindings/clustta/services";
import { ClipboardService, FSService } from '@/../bindings/clustta/services/index';

// services
import { EntityService } from "@/../bindings/clustta/services";

// states/store imports
import { useTrayStates } from '@/stores/TrayStates';
import { useMenu } from '@/stores/menu';
import { usePaneStore } from '@/stores/panes';
import { useStageStore } from '@/stores/stages';
import { useNotificationStore } from '@/stores/notifications';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useUserStore } from '@/stores/users';
import { useModalStore } from '@/stores/modals';
import { useEntityStore } from '@/stores/entity';
import { useAssetStore } from '@/stores/assets';
import { useCommonStore } from '@/stores/common';
import { useIconStore } from '@/stores/icons';
import { useProjectStore } from '@/stores/projects';

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue'
import HeaderArea from '@/instances/common/components/HeaderArea.vue';

// states/stores
const trayStates = useTrayStates();
const userStore = useUserStore();
const menu = useMenu();
const panes = usePaneStore();
const stage = useStageStore();
const modals = useDesktopModalStore();
const notificationStore = useNotificationStore();
const entityStore = useEntityStore();
const assetStore = useAssetStore();
const projectStore = useProjectStore();
const commonStore = useCommonStore();
const iconStore = useIconStore();



const getAppIcon = (iconName) => {
  const icon = iconStore.getAppIcon(iconName);
  return icon
};


// onMounted hook
onMounted(() => {
});

onBeforeUnmount(() => {

});



</script>
<style scoped>
@import "@/assets/desktop.css";

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

