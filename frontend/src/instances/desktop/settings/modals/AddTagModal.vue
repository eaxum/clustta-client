<template>
  <div class="modal-container" v-stop-propagation v-esc="closeModal">
    <HeaderArea :title="$t('modals.addTag')" icon="tag" />
    <div class="general-container">
      <TagForm mode="create" @created="closeModal" @cancel="closeModal" />
    </div>
  </div>
</template>

<script setup>
import HeaderArea from '@/instances/common/components/HeaderArea.vue';
import TagForm from '@/instances/common/components/TagForm.vue';
import { useDesktopModalStore } from '@/stores/desktopModals';

const modals = useDesktopModalStore();
const closeModal = () => modals.setModalVisibility('addTagModal', false);
</script>

<style scoped>
@import "@/assets/desktop.css";
</style>
