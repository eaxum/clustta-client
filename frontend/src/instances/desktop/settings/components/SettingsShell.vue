<template>
  <div class="settings-shell">
    <aside class="settings-shell-sidebar">
      <h1 class="settings-shell-title">{{ $t(titleKey) }}</h1>

      <select
        class="settings-shell-select input-short"
        :value="activePageId"
        :aria-label="$t(titleKey)"
        @change="selectPage($event.target.value)"
      >
        <optgroup v-for="group in visibleGroups" :key="group.id" :label="$t(group.nameKey)">
          <option v-for="page in group.pages" :key="page.id" :value="page.id">
            {{ $t(page.nameKey) }}
          </option>
        </optgroup>
      </select>

      <nav class="settings-shell-navigation" :aria-label="$t(titleKey)">
        <section v-for="group in visibleGroups" :key="group.id" class="settings-navigation-group">
          <h2 class="settings-navigation-group-title">{{ $t(group.nameKey) }}</h2>
          <button
            v-for="page in group.pages"
            :key="page.id"
            :ref="(element) => setNavigationItem(page.id, element)"
            class="settings-navigation-item"
            :class="{ 'settings-navigation-item-selected': page.id === activePageId }"
            :aria-current="page.id === activePageId ? 'page' : undefined"
            type="button"
            @click="selectPage(page.id)"
            @keydown="handleNavigationKeydown($event, page.id)"
          >
            <img class="small-icons settings-navigation-icon" :src="getAppIcon(page.icon)" alt="">
            <span>{{ $t(page.nameKey) }}</span>
          </button>
        </section>
      </nav>
    </aside>

    <main v-if="activePage" class="settings-shell-main">
      <div class="settings-shell-page">
        <header class="settings-page-header">
          <h2 class="settings-page-title">{{ $t(activePage.nameKey) }}</h2>
          <p v-if="activePage.descriptionKey" class="settings-page-description">
            {{ $t(activePage.descriptionKey) }}
          </p>
        </header>

        <div class="settings-page-content">
          <component :is="activePage.component" :key="activePage.id" />
        </div>
      </div>
    </main>
  </div>
</template>

<script setup>
import { computed, ref } from 'vue';
import { useIconStore } from '@/stores/icons';

const emit = defineEmits(['select']);

const props = defineProps({
  activePageId: {
    type: String,
    required: true,
  },
  groups: {
    type: Array,
    default: () => [],
  },
  pages: {
    type: Array,
    default: () => [],
  },
  titleKey: {
    type: String,
    required: true,
  },
});

const iconStore = useIconStore();
const navigationItems = ref({});

const activePage = computed(() => {
  return props.pages.find((page) => page.id === props.activePageId) || null;
});

const visibleGroups = computed(() => {
  return props.groups
    .map((group) => ({
      ...group,
      pages: props.pages.filter((page) => page.group === group.id),
    }))
    .filter((group) => group.pages.length);
});

const orderedPageIds = computed(() => {
  return visibleGroups.value.flatMap((group) => group.pages.map((page) => page.id));
});

const getAppIcon = (iconName) => iconStore.getAppIcon(iconName);

const selectPage = (pageId) => {
  emit('select', pageId);
};

const setNavigationItem = (pageId, element) => {
  if (element) {
    navigationItems.value[pageId] = element;
    return;
  }
  delete navigationItems.value[pageId];
};

const handleNavigationKeydown = (event, pageId) => {
  if (!['ArrowDown', 'ArrowUp', 'Home', 'End'].includes(event.key)) return;
  event.preventDefault();

  const currentIndex = orderedPageIds.value.indexOf(pageId);
  if (currentIndex === -1) return;

  let targetIndex = currentIndex;
  if (event.key === 'ArrowDown') targetIndex = (currentIndex + 1) % orderedPageIds.value.length;
  if (event.key === 'ArrowUp') targetIndex = (currentIndex - 1 + orderedPageIds.value.length) % orderedPageIds.value.length;
  if (event.key === 'Home') targetIndex = 0;
  if (event.key === 'End') targetIndex = orderedPageIds.value.length - 1;

  const targetPageId = orderedPageIds.value[targetIndex];
  navigationItems.value[targetPageId]?.focus();
  selectPage(targetPageId);
};
</script>

<style>
@import "@/assets/desktop.css";

.settings-shell {
  display: grid;
  grid-template-columns: 230px minmax(0, 1fr);
  width: 100%;
  height: 100%;
  min-width: 0;
  min-height: 0;
  color: var(--text);
  background-color: var(--surface-3);
}

.settings-shell-sidebar {
  display: flex;
  flex-direction: column;
  min-width: 0;
  min-height: 0;
  padding: 1.25rem .75rem;
  border-right: var(--transparent-line);
  background-color: var(--surface-2);
  box-sizing: border-box;
}

.settings-shell-title {
  margin: 0 .5rem 1.25rem;
  color: var(--text);
  font-size: 1.35rem;
  font-weight: 600;
  line-height: 1.2;
}

.settings-shell-navigation {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  min-height: 0;
  overflow-y: auto;
  padding-right: .15rem;
}

.settings-navigation-group {
  display: flex;
  flex-direction: column;
  gap: .2rem;
}

.settings-navigation-group-title {
  margin: 0 .55rem .25rem;
  color: var(--text-muted);
  font-size: .68rem;
  font-weight: 600;
  letter-spacing: .06em;
  line-height: 1.4;
  text-transform: uppercase;
}

.settings-navigation-item {
  display: flex;
  align-items: center;
  width: 100%;
  min-height: 36px;
  gap: .65rem;
  padding: .45rem .6rem;
  border: 0;
  border-radius: var(--large-radius);
  color: var(--text-muted);
  background-color: transparent;
  font: inherit;
  font-size: .82rem;
  font-weight: 400;
  text-align: left;
  cursor: pointer;
  box-sizing: border-box;
}

.settings-navigation-item:hover {
  color: var(--text);
  background-color: var(--surface-3);
}

.settings-navigation-item:focus-visible {
  outline: 2px solid var(--accent);
  outline-offset: -2px;
}

.settings-navigation-item-selected {
  color: var(--text);
  background-color: var(--surface-4);
  font-weight: 500;
}

.settings-navigation-icon {
  flex: 0 0 auto;
  opacity: .8;
}

.settings-navigation-item-selected .settings-navigation-icon {
  opacity: 1;
}

.settings-shell-select {
  display: none;
}

.settings-shell-main {
  display: flex;
  justify-content: center;
  min-width: 0;
  min-height: 0;
  padding: 1.5rem;
  overflow: hidden;
  box-sizing: border-box;
}

.settings-shell-page {
  display: flex;
  flex-direction: column;
  width: 100%;
  max-width: 960px;
  min-width: 0;
  min-height: 0;
}

.settings-page-header {
  flex: 0 0 auto;
  padding: .1rem .5rem 1rem;
}

.settings-page-title {
  margin: 0;
  color: var(--text);
  font-size: 1.5rem;
  font-weight: 600;
  line-height: 1.25;
}

.settings-page-description {
  margin: .35rem 0 0;
  color: var(--text-muted);
  font-size: .82rem;
  line-height: 1.45;
}

.settings-page-content {
  display: flex;
  flex: 1;
  min-width: 0;
  min-height: 0;
  overflow: hidden;
}

.settings-page-content > * {
  width: 100%;
  min-width: 0;
  min-height: 0;
}

.settings-component-container {
  border-radius: var(--gigantic-radius) !important;
}

.settings-section-card {
  display: flex;
  flex-direction: column;
  width: 100%;
  height: min-content;
  overflow: hidden;
  outline: var(--transparent-line);
  outline-offset: -1px;
  border-radius: var(--very-large-radius);
  background-color: var(--surface-1);
  box-sizing: border-box;
}

.settings-section-card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  margin-bottom: .75rem;
  padding-bottom: .65rem;
  border-bottom: var(--transparent-line);
}

.settings-section-card-title {
  flex: 1;
  margin: 0;
  color: var(--text);
  font-size: .9rem;
  font-weight: 500;
}

.settings-section-card-content {
  display: flex;
  flex-direction: column;
  height: min-content;
  overflow: hidden;
  border-radius: var(--normal-radius);
  background-color: var(--surface-2);
}

.actions-divider {
  display: flex;
  width: 1.5px;
  height: 16px;
  background-color: var(--surface-4);
}

.horizontal-flex {
  display: flex;
  align-items: center;
  flex: 1;
  gap: .5rem;
}

@media (max-width: 760px) {
  .settings-shell {
    display: flex;
    flex-direction: column;
  }

  .settings-shell-sidebar {
    flex: 0 0 auto;
    padding: .8rem 1rem;
    border-right: 0;
    border-bottom: var(--transparent-line);
  }

  .settings-shell-title {
    margin: 0 0 .65rem;
    font-size: 1.1rem;
  }

  .settings-shell-navigation {
    display: none;
  }

  .settings-shell-select {
    display: block;
    width: 100%;
  }

  .settings-shell-main {
    padding: 1rem;
  }
}
</style>
