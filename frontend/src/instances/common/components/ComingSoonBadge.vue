<template>
  <span class="coming-soon-badge">{{ $t('settings.comingSoon') }}</span>
</template>

<style scoped>
.coming-soon-badge {
  display: inline-flex;
  align-items: center;
  border-radius: var(--small-radius);
  padding: 2px 6px;
  background-color: color-mix(in srgb, var(--warning) 16%, transparent);
  color: var(--warning);
  font-size: 10px;
  font-weight: 600;
  line-height: 1.2;
  white-space: nowrap;
}
</style>
