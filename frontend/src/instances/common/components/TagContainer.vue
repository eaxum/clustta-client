<template>
<!-- <div class="task-options-container" :class="{ 'task-options-container-closed': showTaskOptions === true }"> -->
  <div class="tag-chip-wrapper-parent" :class="{ 'tag-chip-wrapper-parent-closed': !emptyData }">
    <div ref="tagChipWrapper" class="tag-chip-wrapper" @wheel="handleWheel">
      <div v-if="queryTag" @click="removeQueryTag()" v-stop-propagation class="tag-chip chip query-tag">
        {{ queryTag }}
        <img class="small-icons" src="/icons/close.svg">
      </div>
      <div v-for="(tag, index) in tags" :key="tag" @click="removeTag(tag, index)"  v-stop-propagation :ref="setItemRefs" class="tag-chip chip">
        <!-- {{ utils.capitalizeStr(tag) }} -->
        {{ tag.toLowerCase().replace(/_/g, " ") }}
        <img class="small-icons remove-chip-icon" src="/icons/remove.svg">
      </div>
    </div>
  </div>

</template>

<script setup>
import { ref, computed, onMounted, watch, nextTick, onBeforeUnmount } from 'vue';
import { useTrayStates } from '@/stores/TrayStates';

const props = defineProps({
  tags: { type: Array, default: () => [] },
  queryTag: { type: String, default: '' },
});
const trayStates = useTrayStates();
const expandDataItems = ref(true);
const emit = defineEmits(['tagRemoved', 'queryTagRemoved']);
const emptyData = computed(()=>{return props.tags.length || props.queryTag});
const tagChipWrapper = ref(null);
const itemRefs = ref([]);

const removeTag = (tag) => {
  emit('tagRemoved', tag);
};

const removeQueryTag = () => {
  emit('queryTagRemoved');
};

const toggleSearch = () => {
  emit('toggleSearch')
}

const setItemRefs = el => {
  if (el) {
    itemRefs.value.push(el);
  }
};

watch(() => props.tags.length, (newLength, oldLength) => {
    if (newLength > oldLength) {
      onItemAdded();
    }
  }
);

const onItemAdded = () => {
  nextTick(() => {
    if (itemRefs.value.length > 0) {
      const lastItem = itemRefs.value[itemRefs.value.length - 1];
      if (lastItem) {
        lastItem.scrollIntoView({ behavior: 'smooth' });
      }
    }
  });
};

const handleWheel = (event) => {
  const element = tagChipWrapper.value;
  const delta = Math.sign(event.deltaY);
  element.scrollLeft += delta * 50; 
  event.preventDefault();
};

const scrollElement = (direction) => {
  if (tagChipWrapper.value) {
    const elementWidth = tagChipWrapper.value.clientWidth;
    tagChipWrapper.value.scrollBy({
      left: elementWidth * direction,
      behavior: 'smooth',
    });
  }
};

onMounted(()=>{
  tagChipWrapper.value.scrollLeft = tagChipWrapper.value.scrollWidth;
})
</script>

<style scoped>

/* @import "@/assets/tray.css"; */

.tag-chip-wrapper-parent {
  box-sizing: border-box;
  position: relative;
  /* min-height: 40px; */
  overflow: hidden;
  overflow-x: hidden;
  height: 40px;
  display: flex;
  flex-direction: row;
  gap: .2rem;
  border-right: 1px solid rgb(255, 255, 255);
  width: 100%;
  /* background-color: aqua; */
  transition: all .1s ease-in-out;
  /* background-color: darkslateblue; */
}

.tag-chip-wrapper-parent-closed{
  height: 0px;
  padding: 0;
  /* background-color: red; */

}

.tag-chip-wrapper {
  box-sizing: border-box;
  /* position: absolute; */
  /* border-radius: 8px; */
  min-height: 32px;
  line-height: 1.4 !important;
  /* background-color: #fff; */
  /* overflow: unset; */
  overflow-x: scroll;
  text-align: left;
  height: max-content;
  display: flex;
  /* flex-wrap: wrap; */
  flex-direction: row;
  gap: .2rem;
  padding: .2rem;
}

.tag-chip-wrapper::-webkit-scrollbar {
  display: none;
}

.tag-chip {
  background-color: #2e2e2e;
  color: rgb(205, 205, 205);
  outline: solid 1px white;
  
  /* pointer-events: none; */
}

.tag-chip:hover {
  background-color: red;
  background-color: rgb(110, 163, 241);
  background-color: #ffffff17;
}

.selected-tag {

  background-color: darkmagenta;
}

.query-tag {
  background-color: rgb(55, 128, 46);
}
</style>


