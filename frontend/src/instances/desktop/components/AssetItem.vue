<template>
  <div class="asset-item" :class="[`asset-item-${variant}`, { 'asset-item-clickable': clickable }]" @click="handleClick">
    <div class="asset-item-meta">
      <img class="asset-item-icon small-icons" :class="{ 'no-filter': hasResolvedIcon }" :src="displayIcon" @error="handleIconError" />

      <div class="asset-item-label">
        <div class="asset-item-name" v-tooltip="displayName">{{ displayName }}</div>
      </div>

      <span v-if="showBadge && normalizedKindLabel" class="asset-item-badge" :class="'badge-' + normalizedKind">
        {{ normalizedKindLabel }}
      </span>
    </div>

    <div v-if="showActions" class="asset-item-actions">
      <ActionButton
        v-if="showNavigate"
        :icon="getAppIcon(navigateIcon)"
        v-tooltip="navigateTooltip"
        :buttonFunction="handleNavigate"
        :isDisabled="isDisabled"
      />
      <slot name="actions" />
    </div>
  </div>
</template>

<script setup>
import { computed, ref, useSlots, watch } from 'vue';
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import { useIconStore } from '@/stores/icons';

const props = defineProps({
  item: {
    type: Object,
    default: () => ({}),
  },
  assetPath: {
    type: String,
    default: '',
  },
  name: {
    type: String,
    default: '',
  },
  fullPath: {
    type: String,
    default: '',
  },
  extension: {
    type: String,
    default: '',
  },
  icon: {
    type: String,
    default: '',
  },
  fallbackIcon: {
    type: String,
    default: '',
  },
  kind: {
    type: String,
    default: '',
  },
  kindLabel: {
    type: String,
    default: '',
  },
  showBadge: {
    type: Boolean,
    default: true,
  },
  hideExtension: {
    type: Boolean,
    default: false,
  },
  showFullPath: {
    type: Boolean,
    default: false,
  },
  showNavigate: {
    type: Boolean,
    default: false,
  },
  navigateIcon: {
    type: String,
    default: 'file-search',
  },
  navigateTooltip: {
    type: String,
    default: '',
  },
  isDisabled: {
    type: Boolean,
    default: false,
  },
  clickable: {
    type: Boolean,
    default: false,
  },
  variant: {
    type: String,
    default: 'default',
  },
});

const emit = defineEmits(['click', 'navigate']);

const iconStore = useIconStore();
const slots = useSlots();
const resolvedIcon = ref('');
const iconLoadKey = ref('');

const itemValue = computed(() => props.item || {});

const normalizePath = (filePath = '') => filePath.replace(/\\/g, '/');

const getFileName = (filePath = '') => {
  return normalizePath(filePath).split('/').filter(Boolean).pop() || filePath;
};

const getExtension = (filePath = '') => {
  const fileName = getFileName(filePath);
  const dotIndex = fileName.lastIndexOf('.');
  if (dotIndex <= 0) return '';
  return fileName.slice(dotIndex);
};

const stripExtension = (value = '', extension = '') => {
  if (!extension || !value.toLowerCase().endsWith(extension.toLowerCase())) return value;
  return value.slice(0, -extension.length);
};

const extension = computed(() => {
  return props.extension
    || itemValue.value.extension
    || getExtension(props.fullPath || itemValue.value.fullPath || itemValue.value.display_path || itemValue.value.asset_path || '');
});

const normalizedFullPath = computed(() => {
  const path = normalizePath(
    props.fullPath
    || itemValue.value.fullPath
    || itemValue.value.display_path
    || props.assetPath
    || itemValue.value.asset_path
    || ''
  );
  if (!path || !extension.value || path.toLowerCase().endsWith(extension.value.toLowerCase())) return path;
  return `${path}${extension.value}`;
});

const normalizedName = computed(() => {
  return props.name
    || itemValue.value.name
    || stripExtension(getFileName(normalizedFullPath.value), extension.value);
});

const displayName = computed(() => {
  if (props.showFullPath) {
    return props.hideExtension ? stripExtension(normalizedFullPath.value, extension.value) : normalizedFullPath.value;
  }
  if (props.hideExtension || !extension.value) return normalizedName.value;
  if (normalizedName.value.toLowerCase().endsWith(extension.value.toLowerCase())) return normalizedName.value;
  return `${normalizedName.value}${extension.value}`;
});

const normalizedKind = computed(() => props.kind || itemValue.value.kind || 'default');
const normalizedKindLabel = computed(() => props.kindLabel || itemValue.value.kindLabel || '');

const explicitIcon = computed(() => props.icon || itemValue.value.icon || itemValue.value.fallbackIcon || props.fallbackIcon || '');
const displayIcon = computed(() => resolvedIcon.value || explicitIcon.value || getAppIcon('generic'));
const hasResolvedIcon = computed(() => !!(resolvedIcon.value || explicitIcon.value));
const showActions = computed(() => props.showNavigate || !!slots.actions);

const getAppIcon = (iconName) => iconStore.getAppIcon(iconName);

const handleClick = () => {
  if (props.clickable) emit('click', itemValue.value);
};

const handleNavigate = () => {
  emit('navigate', itemValue.value);
};

const handleIconError = (event) => {
  event.target.src = '/file-icons/default.svg';
};

const loadIcon = async () => {
  const nextKey = `${extension.value}:${explicitIcon.value}`;
  iconLoadKey.value = nextKey;
  resolvedIcon.value = '';
  if (explicitIcon.value || !extension.value) return;

  const ext = extension.value.toLowerCase().replace(/^\./, '');
  const iconPath = await iconStore.getIcon(ext);
  if (iconLoadKey.value === nextKey) {
    resolvedIcon.value = iconPath || '/file-icons/default.svg';
  }
};

watch(() => `${extension.value}:${explicitIcon.value}`, loadIcon, { immediate: true });

</script>

<style scoped>
@import "@/assets/desktop.css";

.asset-item {
  position: relative;
  cursor: auto;
  box-sizing: border-box;
  display: flex;
  flex-direction: row;
  align-items: center;
  justify-content: space-between;
  gap: .5rem;
  width: 100%;
  min-height: 40px;
  padding: 0 .5rem;
  background-color: var(--surface-3);
  border-radius: var(--large-radius);
  overflow: hidden;
  outline: var(--transparent-line);
  outline-offset: -1px;
  transition: all .2s ease-in-out;
}

.asset-item:hover {
  border-radius: var(--small-radius);
  background-color: var(--surface-3);
}

.asset-item-clickable {
  cursor: pointer;
}

.asset-item-compact {
  min-height: 34px;
  background-color: transparent;
}

.asset-item-compact:hover {
  background-color: var(--surface-3);
}

.asset-item-actions {
  display: flex;
  flex: 0 0 auto;
  align-items: center;
  gap: .25rem;
  max-width: 0;
  opacity: 0;
  overflow: hidden;
  transform: translateX(.5rem);
  transition: max-width .2s ease-in-out, opacity .2s ease-out, transform .2s ease-out;
}

.asset-item:hover .asset-item-actions {
  max-width: none;
  width: max-content;
  align-items: center;
  justify-content: center;
  box-sizing: border-box;
  overflow: visible;
  opacity: 1;
  transform: translateX(0);
}

.asset-item-badge {
  border-radius: 4px;
  flex-shrink: 0;
  font-size: 10px;
  font-weight: 500;
  padding: 1px 5px;
  text-transform: uppercase;
  white-space: nowrap;
  margin-left: auto;
}

.asset-item-icon {
  width: 20px;
  height: 20px;
  min-width: 20px;
  object-fit: contain;
}

.asset-item-label {
  overflow: hidden;
  width: 100%;
  display: flex;
  white-space: nowrap;
}

.asset-item-meta {
  box-sizing: border-box;
  overflow: hidden;
  display: flex;
  flex-direction: row;
  align-items: center;
  gap: .5rem;
  flex: 1 1 auto;
  width: auto;
  min-height: 40px;
  min-width: 0;
}

.asset-item-name {
  font-size: 13px;
  font-weight: 300;
  color: var(--text);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.badge-modified {
  background-color: rgba(59, 130, 246, 0.15);
  color: #60a5fa;
}

.badge-untracked {
  background-color: rgba(34, 197, 94, 0.15);
  color: #4ade80;
}
</style>
