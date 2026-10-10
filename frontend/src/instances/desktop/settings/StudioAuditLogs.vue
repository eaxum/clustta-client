<template>
  <div class="settings-component-root">
    <div class="settings-component-container">
      <div class="audit-log-skeleton" aria-hidden="true">
        <div class="audit-log-toolbar">
          <div class="skeleton-block skeleton-filter"></div>
          <div class="skeleton-block skeleton-search"></div>
        </div>

        <div v-for="row in SKELETON_ROW_COUNT" :key="row" class="audit-log-row">
          <div class="skeleton-block skeleton-avatar"></div>
          <div class="audit-log-copy">
            <div class="skeleton-block skeleton-title"></div>
            <div class="skeleton-block skeleton-description"></div>
          </div>
          <div class="skeleton-block skeleton-time"></div>
        </div>
      </div>

    </div>
  </div>
</template>

<script setup>
const SKELETON_ROW_COUNT = 6;
</script>

<style scoped>
.settings-component-container {
  position: relative;
}

.audit-log-skeleton {
  display: flex;
  flex-direction: column;
  gap: .65rem;
  padding: 1rem;
  border-radius: var(--very-large-radius);
  outline: var(--transparent-line);
  outline-offset: -1px;
  background-color: var(--surface-1);
  opacity: .45;
}

.audit-log-toolbar,
.audit-log-row {
  display: flex;
  align-items: center;
  gap: .75rem;
}

.audit-log-toolbar {
  justify-content: space-between;
  padding-bottom: .75rem;
}

.audit-log-row {
  min-height: 54px;
  padding: .5rem .75rem;
  border-radius: var(--large-radius);
  background-color: var(--surface-2);
}

.audit-log-copy {
  display: flex;
  flex: 1;
  flex-direction: column;
  gap: .4rem;
}

.skeleton-block {
  border-radius: var(--small-radius);
  background-color: var(--surface-4);
}

.skeleton-filter {
  width: 130px;
  height: 32px;
}

.skeleton-search {
  width: 240px;
  height: 32px;
}

.skeleton-avatar {
  width: 30px;
  height: 30px;
  border-radius: 50%;
}

.skeleton-title {
  width: 42%;
  height: 12px;
}

.skeleton-description {
  width: 68%;
  height: 9px;
}

.skeleton-time {
  width: 70px;
  height: 10px;
}

</style>
