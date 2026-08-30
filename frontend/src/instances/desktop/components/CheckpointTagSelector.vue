<template>
  <div class="checkpoint-tag-selector">
    <RenameInput v-if="isCreatingTag" v-model="newTagName" originalValue="" placeholder="Tag name"
      @confirm="confirmNewTag" @cancel="cancelNewTag" />
    <DropDownBox v-else :items="tagOptions" :selectedItem="selectedItem" :onSelect="selectTag"
      :useFilter="false" placeHolder="No tag">
      <template #footer="{ close }">
        <div class="checkpoint-tag-dropdown-divider"></div>
        <button class="checkpoint-tag-create-action" type="button" @click="startNewTag(close)">
          <img class="small-icons" :src="getAppIcon('plus-circle')">
          <span>Add new tag</span>
        </button>
      </template>
    </DropDownBox>
  </div>
</template>

<script setup>
import { computed, ref, watch } from 'vue';

import DropDownBox from '@/instances/common/components/DropDownBox.vue';
import RenameInput from '@/instances/desktop/components/RenameInput.vue';
import { CheckpointService } from '@/services';
import { useIconStore } from '@/stores/icons';
import { useNotificationStore } from '@/stores/notifications';
import { useProjectStore } from '@/stores/projects';

const NO_TAG = 'checkpoint-tag:none';
const TAG_PREFIX = 'checkpoint-tag:';

const props = defineProps({
  assetIds: { type: Array, default: () => [] },
  modelValue: { type: String, default: '' },
});

const emit = defineEmits(['update:modelValue']);
const iconStore = useIconStore();
const notificationStore = useNotificationStore();
const projectStore = useProjectStore();

const availableTags = ref([]);
const isCreatingTag = ref(false);
const newTagName = ref('');

const normalizedTagName = (name) => name.trim().toLowerCase();
const tagOptionId = (name) => `${TAG_PREFIX}${normalizedTagName(name)}`;

const tagOptions = computed(() => {
  const tags = [...availableTags.value];
  if (props.modelValue && !tags.some(tag => normalizedTagName(tag.name) === normalizedTagName(props.modelValue))) {
    tags.push({ id: tagOptionId(props.modelValue), name: props.modelValue });
  }
  return [
    { id: NO_TAG, name: 'No tag', selectionValue: NO_TAG, icon: getAppIcon('tag') },
    ...tags.map(tag => ({
      ...tag,
      id: tagOptionId(tag.name),
      selectionValue: tagOptionId(tag.name),
      icon: getAppIcon('tag'),
    })),
  ];
});

const selectedItem = computed(() => props.modelValue ? tagOptionId(props.modelValue) : NO_TAG);

const loadTags = async () => {
  const assetIds = [...new Set(props.assetIds.filter(Boolean))];
  if (!assetIds.length) {
    availableTags.value = [];
    return;
  }
  try {
    const tagLists = await Promise.all(assetIds.map(assetId => (
      CheckpointService.GetCheckpointTags(projectStore.activeProject.uri, assetId)
    )));
    const tagsByName = new Map();
    tagLists.flat().forEach(tag => {
      const name = normalizedTagName(tag.name);
      if (!tagsByName.has(name)) tagsByName.set(name, tag);
    });
    availableTags.value = [...tagsByName.values()];
  } catch (error) {
    notificationStore.errorNotification('Unable to load checkpoint tags', error);
  }
};

const selectTag = (optionId) => {
  if (optionId === NO_TAG) {
    emit('update:modelValue', '');
    return;
  }
  const tag = tagOptions.value.find(option => option.id === optionId);
  emit('update:modelValue', tag?.name || '');
};

const startNewTag = (closeDropdown) => {
  closeDropdown();
  newTagName.value = '';
  isCreatingTag.value = true;
};

const confirmNewTag = (name) => {
  const trimmedName = name.trim();
  if (trimmedName) emit('update:modelValue', trimmedName);
  isCreatingTag.value = false;
  newTagName.value = '';
};

const cancelNewTag = () => {
  isCreatingTag.value = false;
  newTagName.value = '';
};

const getAppIcon = (iconName) => iconStore.getAppIcon(iconName);

watch(() => props.assetIds, loadTags, { deep: true, immediate: true });
</script>

<style scoped>
.checkpoint-tag-selector {
  width: 100%;
}

.checkpoint-tag-dropdown-divider {
  border-top: var(--transparent-line);
  margin: .3rem .1rem;
}

.checkpoint-tag-create-action {
  display: flex;
  align-items: center;
  gap: .5rem;
  width: 100%;
  padding: .3rem .5rem;
  border: 0;
  border-radius: var(--normal-radius);
  color: var(--text);
  background: transparent;
  font: inherit;
  font-size: 14px;
  text-align: left;
  cursor: pointer;
}

.checkpoint-tag-create-action:hover {
  background-color: var(--surface-4);
}
</style>
