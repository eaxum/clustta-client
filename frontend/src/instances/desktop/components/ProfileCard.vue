<template>
  <div class="profile-card">
    <div v-if="title" class="profile-card-header">
      <h2 class="profile-card-title">{{ title }}</h2>
    </div>
    <div class="profile-card-content">
      <slot></slot>
    </div>
  </div>
</template>

<script setup>
const props = defineProps({
  title: {
    type: String,
    default: ''
  }
});
</script>

<style scoped>
.profile-card {
  background-color: var(--black-steel);
  border-radius: var(--large-radius);
  border-radius: 24px;
  padding: 1.5rem;
  box-sizing: border-box;
  width: 100%;
  outline: var(--transparent-line);
  outline-offset: -1px;
}

.profile-card-header {
  margin-bottom: 1rem;
  padding-bottom: 0.75rem;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
  border-bottom: var(--transparent-line);
}

.profile-card-title {
  font-size: 1rem;
  font-weight: 300;
  color: var(--white);
  margin: 0;
}

.profile-card-content {
  color: var(--white);
}
</style>
