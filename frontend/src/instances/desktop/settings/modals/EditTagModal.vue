<template>
  <div class="modal-container" v-stop-propagation v-esc="closeModal">
    <HeaderArea :title="$t('modals.editTag')" icon="tag" />
    <div class="general-container">
      <TagForm mode="edit" :initialName="selectedTag.name" :tagId="selectedTag.id" @updated="closeModal" @cancel="closeModal" />
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue';

import HeaderArea from '@/instances/common/components/HeaderArea.vue';
import TagForm from '@/instances/common/components/TagForm.vue';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useTagStore } from '@/stores/tags';

const modals = useDesktopModalStore();
const tagStore = useTagStore();

const selectedTag = computed(() => tagStore.selectedTag || { id: '', name: '' });
const closeModal = () => modals.setModalVisibility('editTagModal', false);
</script>

<style scoped>
@import "@/assets/desktop.css";
</style>
