<template>
  <div class="profile-page-root">
    <div class="profile-header">
      <NavigationBar />
    </div>
    <div class="profile-body">
      <UserProfile />
    </div>
  </div>
</template>

<script setup>
import NavigationBar from '@/instances/web/components/NavigationBar.vue';
import UserProfile from '@/instances/desktop/stages/UserProfile.vue';
</script>

<style scoped>
@import "@/assets/desktop.css";

.profile-page-root {
  width: 100%;
  min-height: 100vh;
  max-height: 100vh;
  background-color: var(--midnight-steel);
  display: flex;
  flex-direction: column;
  align-items: center;
  overflow: hidden;
  overflow-y: auto;
  color: var(--white);
}

.profile-header {
  width: 100%;
  display: flex;
  align-items: center;
  background-color: rgba(255, 255, 255, 0.05);
  position: sticky;
  top: 0;
  z-index: 99999;
  backdrop-filter: blur(30px);
}

.profile-body {
  flex: 1;
  width: 100%;
  max-width: 1000px;
  padding: 2rem;
  box-sizing: border-box;
  display: flex;
  flex-direction: column;
  align-items: center;
}

/* Scrollbar */
.profile-page-root::-webkit-scrollbar {
  width: 10px;
  height: 10px;
}

.profile-page-root::-webkit-scrollbar-track {
  background: var(--midnight-steel);
  border-radius: 5px;
}

.profile-page-root::-webkit-scrollbar-thumb {
  background: var(--steel);
  border-radius: 5px;
}

.profile-page-root::-webkit-scrollbar-thumb:hover {
  background-color: rgba(255, 255, 255, 0.3);
}

/* Responsive */
@media (max-width: 768px) {
  .profile-body {
    padding: 1rem;
  }
}
</style>
