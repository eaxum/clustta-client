<template>
  <div class="auth-loader">
    <img src="/icons/clustta.png" alt="Clustta" class="loader-logo" />

    <div class="loader-progress-container">
      <div class="loader-progress-bar"></div>
    </div>

    <div class="loader-status">{{ status }}</div>
  </div>
</template>

<script setup>
defineProps({
  status: { type: String, default: '' },
});
</script>

<style scoped>
.auth-loader {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 2rem 1rem;
  gap: 1rem;
}

.loader-logo {
  width: 64px;
  height: 64px;
}

.loader-progress-container {
  width: 200px;
  height: 4px;
  background-color: var(--thick-transparent-line);
  border-radius: 4px;
  overflow: hidden;
  position: relative;
}

.loader-progress-bar {
  position: absolute;
  top: 0;
  right: 100%;
  bottom: 0;
  left: 0;
  height: 100%;
  background: linear-gradient(to right, transparent, #643193, transparent);
  width: 0;
  animation: borealisBar 0.8s linear infinite;
}

.loader-status {
  font-size: 0.85rem;
  color: var(--surface-3);
  animation: pulse 1.5s ease-in-out infinite;
}

@keyframes borealisBar {
  0% { left: 0%; right: 100%; width: 0%; }
  25% { left: 0%; right: 75%; width: 40%; }
  50% { left: 0%; right: 50%; width: 90%; }
  75% { right: 0%; left: 75%; width: 40%; }
  100% { left: 100%; right: 0%; width: 0%; }
}

@keyframes pulse {
  0%, 100% { opacity: 0.5; }
  50% { opacity: 0.9; }
}
</style>
