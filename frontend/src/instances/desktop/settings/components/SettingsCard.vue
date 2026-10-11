<template>
  <section class="settings-section-card settings-card">
    <div class="settings-section-card-header">
      <div class="settings-card-title-group">
        <h2 class="settings-section-card-title">{{ title }}</h2>
        <div v-if="description" class="settings-card-header-description">
          {{ description }}
        </div>
      </div>
      <slot name="header-action"></slot>
    </div>

    <div
      v-if="showContent && (items.length || $slots.default)"
      class="settings-section-card-content settings-card-content"
    >
      <template v-for="(item, index) in items" :key="item.id">
        <div class="settings-card-item">
          <div v-if="item.icon" class="settings-card-icon">
            <slot name="icon" :item="item">
              <img
                class="small-icons"
                :src="item.icon"
                alt=""
                v-tooltip="item.iconTooltip"
              >
            </slot>
          </div>

          <div class="settings-card-copy">
            <div class="settings-card-label">
              <slot name="label" :item="item">{{ item.label }}</slot>
            </div>
            <div v-if="item.description" class="settings-card-description">
              <slot name="description" :item="item">{{ item.description }}</slot>
            </div>
            <slot name="details" :item="item"></slot>
          </div>

          <div class="settings-card-action">
            <slot name="action" :item="item"></slot>
          </div>
        </div>

        <div v-if="index < items.length - 1" class="menu-divider"></div>
      </template>

      <slot></slot>
    </div>
  </section>
</template>

<script setup>
defineProps({
  description: {
    type: String,
    default: '',
  },
  items: {
    type: Array,
    default: () => [],
  },
  showContent: {
    type: Boolean,
    default: true,
  },
  title: {
    type: String,
    required: true,
  },
});
</script>

<style scoped>
.settings-card {
  max-width: 900px;
  margin-right: auto;
  margin-left: auto;
}

.settings-card-title-group {
  flex: 1;
  min-width: 0;
}

.settings-card-header-description {
  padding-top: .15rem;
  color: var(--text-muted);
  font-size: 12px;
  font-weight: 400;
  line-height: 1.35;
}

.settings-card-content {
  padding: .2rem 0;
}

.settings-card-item {
  display: flex;
  align-items: center;
  gap: .65rem;
  min-height: 54px;
  padding: .45rem 1rem;
  color: var(--text);
  box-sizing: border-box;
}

.settings-card-icon {
  display: flex;
  flex: 0 0 20px;
  align-items: center;
  justify-content: center;
}

.settings-card-copy {
  display: flex;
  flex: 1;
  min-width: 0;
  flex-direction: column;
  gap: .15rem;
}

.settings-card-label {
  font-size: 13px;
  font-weight: 500;
  line-height: 1.3;
}

.settings-card-description {
  color: var(--text-muted);
  font-size: 12px;
  font-weight: 400;
  line-height: 1.35;
}

.settings-card-action {
  display: flex;
  flex: 0 0 auto;
  align-items: center;
  justify-content: flex-end;
  min-width: 150px;
}

.settings-card .menu-divider {
  width: auto;
  margin: 0 1rem;
  border-bottom: var(--transparent-line);
}

@media (max-width: 680px) {
  .settings-card-item {
    align-items: flex-start;
    flex-wrap: wrap;
  }

  .settings-card-action {
    width: 100%;
    padding-left: 1.8rem;
  }
}
</style>
