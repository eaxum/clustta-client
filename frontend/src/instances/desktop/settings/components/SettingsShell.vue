<template>
  <div class="settings-shell">
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
import { computed } from 'vue';

const props = defineProps({
  activePageId: {
    type: String,
    required: true,
  },
  pages: {
    type: Array,
    default: () => [],
  },
});

const activePage = computed(() => {
  return props.pages.find((page) => page.id === props.activePageId) || null;
});

</script>

<style>
@import "@/assets/desktop.css";

.settings-shell {
  display: flex;
  width: 100%;
  height: 100%;
  min-width: 0;
  min-height: 0;
  color: var(--text);
  background-color: var(--surface-3);
}

.settings-shell-main {
  display: flex;
  justify-content: center;
  width: 100%;
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
  max-width: 1200px;
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
  font-size: .72rem;
  font-weight: 300;
  line-height: 1.4;
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

.settings-page-content .settings-component-root,
.settings-page-content .settings-component-container {
  width: 100% !important;
  max-width: none !important;
  margin-right: 0 !important;
  margin-left: 0 !important;
  box-sizing: border-box;
}

.settings-page-content .user-settings-page {
  height: 100%;
  min-height: 0;
  padding-right: .2rem;
  overflow-x: hidden;
  overflow-y: auto;
  border-radius: var(--very-large-radius);
}

.settings-page-content .user-settings-page > .settings-component-container {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
  min-height: 100%;
}

.settings-page-content,
.settings-page-content * {
  scrollbar-color: var(--surface-4) transparent;
  scrollbar-width: thin;
}

.settings-page-content::-webkit-scrollbar,
.settings-page-content *::-webkit-scrollbar {
  width: 6px !important;
  height: 6px !important;
}

.settings-page-content::-webkit-scrollbar-track,
.settings-page-content *::-webkit-scrollbar-track {
  background-color: transparent !important;
}

.settings-page-content::-webkit-scrollbar-thumb,
.settings-page-content *::-webkit-scrollbar-thumb {
  min-height: 36px;
  border-radius: 999px !important;
  background-color: var(--surface-4) !important;
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

.settings-page-content .settings-section-card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  margin-bottom: .75rem;
  padding-bottom: .65rem;
  border-bottom: var(--transparent-line);
}

.settings-page-content .settings-section-card-title {
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

.settings-page-content .settings-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  min-height: 50px;
  height: max-content;
  padding: .5rem 1rem;
  overflow: hidden;
  border-bottom: 1px solid var(--surface-4);
  color: var(--text);
  background-color: var(--surface-2);
  cursor: default;
  box-sizing: border-box;
  transition: none;
}

.settings-page-content .settings-item:hover {
  background-color: var(--surface-2);
}

.settings-page-content .location-item,
.settings-page-content .integration-settings-row,
.settings-page-content .project-item-main {
  cursor: default;
  transition: none;
}

.settings-page-content .location-item:hover,
.settings-page-content .integration-settings-row:hover,
.settings-page-content .project-item-main:hover {
  background-color: var(--surface-2);
}

.settings-page-content .hover-action {
  display: flex;
}

.settings-page-content .project-item-actions {
  min-width: 116px;
  max-width: 132px;
  opacity: 1;
  transform: translateX(0);
}

.settings-page-content .row-delete {
  opacity: 1;
}

.settings-page-content .settings-icon {
  display: flex;
  align-items: center;
  width: max-content;
  height: 100%;
  padding: .3rem;
  overflow: hidden;
  box-sizing: border-box;
}

.settings-page-content .settings-content {
  display: flex;
  flex: 1;
  flex-direction: column;
  align-items: flex-start;
  justify-content: center;
  height: 100%;
  padding: .4rem .2rem;
  overflow: hidden;
  box-sizing: border-box;
}

.settings-page-content .settings-header {
  padding: .1rem;
  font-size: 14px;
  font-weight: 400;
}

.settings-page-content .settings-body {
  padding: .1rem;
  color: var(--text-muted);
  font-size: 12px;
  opacity: .8;
}

.settings-page-content .settings-action {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: .5rem;
  width: max-content;
  height: 100%;
  overflow: hidden;
  box-sizing: border-box;
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
  .settings-shell-main {
    padding: 1rem;
  }
}
</style>
