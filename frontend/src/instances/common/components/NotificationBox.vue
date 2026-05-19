<template>
  <div :class="['notification-box', notificationClass, { clickable: isClickable }]" @click="handleClick">
    <div class="notification-icon">
      <img class="small-icons no-filter" :src="icon" :alt="iconAlt" />
    </div>
    <div class="notification-content">
      <div class="notification-title">{{ title }}</div>
      <div class="notification-message">{{ message }}</div>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue';

const props = defineProps({
  type: {
    type: String,
    default: 'info',
    validator: (value) => ['info', 'warning', 'invitation', 'alert'].includes(value)
  },
  icon: {
    type: String,
    required: true
  },
  iconAlt: {
    type: String,
    default: ''
  },
  title: {
    type: String,
    required: true
  },
  message: {
    type: String,
    required: true
  },
  clickable: {
    type: Boolean,
    default: false
  }
});

const emit = defineEmits(['click']);

const notificationClass = computed(() => {
  const typeMap = {
    info: 'info-notification',
    warning: 'studio-notification',
    invitation: 'new-user-notification',
    alert: 'studio-notification'
  };
  return typeMap[props.type] || 'info-notification';
});

const isClickable = computed(() => props.clickable);

const handleClick = () => {
  if (props.clickable) {
    emit('click');
  }
};
</script>

<style scoped>
.notification-box {
  display: flex;
  align-items: flex-start;
  gap: 0.75rem;
  width: 100%;
  padding: 0.75rem;
  border-radius: var(--small-radius);
}

.notification-box.clickable {
  cursor: pointer;
  transition: all 0.2s ease;
}

/* Info notification (blue) */
.info-notification {
  background-color: rgba(45, 156, 219, 0.1);
  border: 1px solid #2D9CDB;
}

.info-notification.clickable:hover {
  background-color: rgba(45, 156, 219, 0.15);
  border-color: #3AAFEF;
}

.info-notification .notification-icon img {
  filter: brightness(0) saturate(100%) invert(47%) sepia(74%) saturate(6614%) hue-rotate(201deg) brightness(95%) contrast(91%);
}

/* Studio/Warning notification (orange) */
.studio-notification {
  background-color: rgba(255, 165, 0, 0.1);
  border: 1px solid #FFA500;
}

.studio-notification.clickable:hover {
  background-color: rgba(255, 165, 0, 0.15);
  border-color: #FFB733;
}

.studio-notification .notification-icon img {
  filter: brightness(0) saturate(100%) invert(69%) sepia(100%) saturate(2582%) hue-rotate(14deg) brightness(101%) contrast(101%);
}

/* New user/Invitation notification (blue) */
.new-user-notification {
  background-color: rgba(45, 156, 219, 0.1);
  border: 1px solid #2D9CDB;
}

.new-user-notification.clickable:hover {
  background-color: rgba(45, 156, 219, 0.15);
  border-color: #3AAFEF;
}

.new-user-notification .notification-icon img {
  filter: brightness(0) saturate(100%) invert(47%) sepia(74%) saturate(6614%) hue-rotate(201deg) brightness(95%) contrast(91%);
}

.notification-icon {
  flex-shrink: 0;
  margin-top: 0.1rem;
}

.notification-icon img {
  width: 20px;
  height: 20px;
}

.notification-content {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  width: 100%;
}

.notification-title {
  font-weight: 500;
  color: var(--text);
  font-size: 14px;
}

.notification-message {
  font-size: 13px;
  color: var(--text);
  line-height: 1.4;
  font-weight: 400;
}
</style>
