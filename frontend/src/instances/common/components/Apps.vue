<template>
  <div class="apps-container-full">
    <span class="arrow-buttons" @click="scrollElement(-1)"><img class="small-icons" src="/icons/arrow_left.svg"></span>
    <div class="apps-container" @wheel="handleWheel">
      <div ref="scrollableElement" class="apps-flex">
        <span v-for="(template) in iconStore.appsWithIcons" :key="template.id" class="apps-flex-item"
          @click="selectApp($event, template)" v-tooltip="replaceSymbols(template.name)"
          :class="{ 'apps-flex-item-selected': templateStore.selectedTemplateName === template.name }">
          <img v-if="template.icon" class="app-icons" :src="template.icon">
          <span v-else class="app-ext">
            <p class="app-ext-text">{{ (template.extension.slice(1)) }}</p>
          </span>
        </span>
      </div>
    </div>
    <span class="arrow-buttons" @click="scrollElement(1)"><img class="small-icons" src="/icons/arrow_right.svg"></span>
  </div>

</template>

<script setup>
import { ref, computed, onMounted, onBeforeUnmount } from 'vue';
import { useI18n } from 'vue-i18n';
import { useTrayStates } from '@/stores/TrayStates';
import { useIconStore } from '@/stores/icons';
import { useModalStore } from '@/stores/modals';
import { useTemplateStore } from '@/stores/template';

const trayStates = useTrayStates();
const iconStore = useIconStore();
const modalStore = useModalStore();
const templateStore = useTemplateStore();
const { t } = useI18n();
const scrollableElement = ref(null);

const emit = defineEmits([ 'template-selected'])

const props = defineProps({
  selectedTemplateId: { type: String, default: '' } 
});

const selectApp = (event, template) => {
  const templateName = template.name
  const templateIcon = template.icon;
  templateStore.selectedTemplateName = templateName;
  trayStates.popUpModalIcon = templateIcon;
  trayStates.popUpModalTitle = t('components.apps.newAssetTitle', { name: templateName.replace(/_/g, ' ') });

  const selectedIcon = event.target;
  const appsCenter = scrollableElement.value.offsetWidth / 2;
  const iconCenter = selectedIcon.offsetWidth / 2;
  const scrollPosition = selectedIcon.offsetLeft - (appsCenter - iconCenter);
  scrollableElement.value.scrollTo({
    left: scrollPosition,
    behavior: 'smooth'
  });
  emit('template-selected', template)
};

const replaceSymbols = (name) => {
  return name.replace(/_/g, " ").toLowerCase().replace(/(^\w|\s\w)/g, match => match.toUpperCase());
}
const handleWheel = (event) => {
  const element = scrollableElement.value;
  const delta = Math.sign(event.deltaY);
  element.scrollLeft += delta * 50; // Adjust scroll speed as needed
  event.preventDefault(); // Prevent the default vertical scrolling behavior
}


onMounted(() => {
  iconStore.fetchAppIcons();
  trayStates.appsContainer = scrollableElement.value;
  if(props.selectedTemplateId){
    const selectedTemplate = templateStore.getTemplates.find((template) => template.id === props.selectedTemplateId);
    templateStore.selectedTemplateName = selectedTemplate?.name
    console.log('yooopi')
  }
});

const scrollElement = (direction) => {
  if (scrollableElement.value) {
    const elementWidth = scrollableElement.value.clientWidth;
    canScrollRef();
    scrollableElement.value.scrollBy({
      left: elementWidth * direction,
      behavior: 'smooth',
    });
  }
};

const canScrollRightRef = ref(true);
const canScrollLeftRef = ref(false);

const canScrollRef = () => {
  if (document.querySelector('.apps-flex')) {
    const appsContainer = document.querySelector('.apps-flex');
    const { scrollWidth, clientWidth, scrollLeft } = appsContainer;
    canScrollLeftRef.value = scrollLeft > 0;
    canScrollRightRef.value = scrollWidth > clientWidth + scrollLeft;

  }
};

</script>

<style scoped>

.small-icons {
  min-width: 20px;
  min-height: 20px;
}

.apps-container-full {
  box-sizing: border-box;
  width: 100%;
  display: flex;
  position: relative;
  display: flex;
  justify-content: space-between;
  align-items: center;
  overflow: hidden;
  scroll-behavior: smooth;
}

.apps-container {
  box-sizing: border-box;
  overflow: hidden;
  padding: .2rem;
  flex: 1;
  display: flex;
  position: relative;
  display: flex;
  justify-content: space-between;
}


.arrow-buttons {
  position: relative;
  border-radius: var(--normal-radius);
  box-sizing: border-box;
  cursor: pointer;
  display: flex;
  gap: 10px;
  align-items: center;
  padding: 5px 5px;
  width: 20px;
  overflow: hidden;
  align-items: center;
  justify-content: center;
}

.arrow-buttons-disabled {
  /* background-color: hsl(var(--destructive)); */
  opacity: .2;
  cursor: not-allowed;
}

.apps-flex {
  /* background-color: deeppink; */
  box-sizing: border-box;
  overflow: hidden;
  display: flex;
  flex-direction: row;
  align-items: center;
  gap: .5rem;
  flex-wrap: nowrap;
}

.apps-flex-item {
  padding: .2rem;
  border-radius: var(--normal-radius);
}

/* .app-icons{
  background-color: aqua;
} */

.app-ext {
  box-sizing: border-box;
  background-color: rgb(6, 86, 117);
  border-radius: var(--normal-radius);
  padding: 2px;
  height: 40px;
  width: 40px;
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: center;
}

.app-ext-text {
  font-size: 16px;
  overflow: hidden;
  overflow-wrap: break-word;

}

.apps-flex-item:hover {
  background-color: rgb(121, 121, 121);
  background-color: hsl(var(--accent));
}

.apps-flex-item:active {
  background-color: rgb(70, 70, 70);
  background-color: hsl(var(--accent));
}

.apps-flex-item-selected {
  box-sizing: border-box;
  background-color: hsl(var(--muted));
  /* border: 1px solid white; */
  outline: solid 1px white;
  

}

.apps-flex-item-selected:hover {
  background-color: hsl(var(--muted));

}
</style>

