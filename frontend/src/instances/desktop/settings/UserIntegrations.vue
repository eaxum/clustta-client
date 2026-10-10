<template>
  <div class="settings-component-root">
    <div class="settings-component-scroll">
      <div class="settings-component-container">
        <div class="settings-section-card">
          <div class="settings-section-card-header">
            <h2 class="settings-section-card-title">{{ $t('settings.integrations') }}</h2>
          </div>

          <div class="settings-section-card-content">
            <div
              v-for="integration in connectedIntegrations"
              :key="integration.id"
              class="settings-item"
              @click="openIntegrationAuth"
            >
              <div class="settings-icon">
                <img class="small-icons" :src="getAppIcon(integration.icon)" alt="">
              </div>
              <div class="settings-content">
                <div class="settings-header">{{ integration.name }}</div>
                <div class="settings-body">{{ $t('settings.connected') }}</div>
              </div>
              <div class="settings-action">
                <span class="connected-badge">{{ $t('common.connected') }}</span>
              </div>
            </div>

            <div class="settings-item" @click="openIntegrationAuth">
              <div class="settings-icon">
                <img class="small-icons" :src="getAppIcon('plug')" alt="">
              </div>
              <div class="settings-content">
                <div class="settings-header">{{ $t('settings.connectIntegration') }}</div>
                <div class="settings-body">{{ $t('settings.connectIntegrationDescription') }}</div>
              </div>
              <div class="settings-action">
                <img class="small-icons" :src="getAppIcon('chevron-right')" alt="">
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed, onMounted } from 'vue';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useIconStore } from '@/stores/icons';
import { useIntegrationStore } from '@/stores/integrations';

const desktopModals = useDesktopModalStore();
const iconStore = useIconStore();
const integrationStore = useIntegrationStore();

const connectedIntegrations = computed(() => {
  return integrationStore.availableIntegrations.filter((integration) => {
    return integrationStore.isAuthenticated(integration.id);
  });
});

const getAppIcon = (iconName) => iconStore.getAppIcon(iconName);

const openIntegrationAuth = () => {
  desktopModals.setModalVisibility('integrationAuthModal', true);
};

onMounted(async () => {
  try {
    await integrationStore.initialize();
  } catch (error) {
    console.error('Unable to load user integrations:', error);
  }
});
</script>

<style scoped>
@import "@/assets/desktop.css";

.settings-component-root {
  display: block;
  width: 100%;
  height: 100%;
  overflow-y: auto;
  border-radius: var(--very-large-radius);
  box-sizing: border-box;
}

.settings-component-scroll {
  display: flex;
  justify-content: center;
}

.settings-component-container {
  display: flex;
  flex-direction: column;
  width: 100%;
  gap: 1.5rem;
  padding-right: .2rem;
  box-sizing: border-box;
}

.settings-item {
  display: flex;
  align-items: center;
  width: 100%;
  min-height: 50px;
  padding: .5rem 1rem;
  border: 0;
  border-bottom: 1px solid var(--surface-4);
  color: var(--text);
  background-color: var(--surface-2);
  cursor: pointer;
  box-sizing: border-box;
}

.settings-item:hover {
  background-color: #ffffff15;
}

.settings-icon {
  display: flex;
  padding: .3rem;
}

.settings-content {
  display: flex;
  flex: 1;
  flex-direction: column;
  padding: .4rem .2rem;
}

.settings-header {
  padding: .1rem;
  font-size: 14px;
  font-weight: 400;
}

.settings-body {
  padding: .1rem;
  color: var(--text-muted);
  font-size: 12px;
}

.settings-action {
  display: flex;
  align-items: center;
  justify-content: flex-end;
}

.connected-badge {
  padding: 4px 8px;
  border-radius: var(--small-radius);
  color: var(--accent-primary);
  background-color: rgba(var(--accent-primary-rgb), .15);
  font-size: 12px;
  font-weight: 500;
}
</style>
