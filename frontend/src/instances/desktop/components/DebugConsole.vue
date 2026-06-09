<template>
  <div class="debug-console">
    <div class="debug-console-header">
      <div class="debug-console-title">
        <img class="small-icons" :src="getAppIcon('console')">
        <span>{{ $t('components.debugConsole.title') }}</span>
        <span class="log-count">({{ searchQuery ? `${filteredLogs.length}/${logs.length}` : logs.length }})</span>
      </div>
      <div class="debug-console-actions">
        <SearchBar v-model="searchQuery" :placeholder="$t('components.debugConsole.filterPlaceholder')" />
        <ActionButton :icon="getAppIcon('copy')" v-tooltip="$t('components.debugConsole.copyLogs')" :buttonFunction="copyLogs" />
        <ActionButton :icon="getAppIcon('folder-arrow-up-right')" v-tooltip="$t('components.debugConsole.openLogsFolder')" :buttonFunction="openLogsFolder" />
        <ActionButton :icon="getAppIcon('megaphone')" v-tooltip="$t('components.debugConsole.submitDiagnostics')" :buttonFunction="openDiagnosticsModal" />
        <ActionButton :icon="getAppIcon('broom')" v-tooltip="$t('components.debugConsole.clear')" :buttonFunction="clearLogs" />
        <ActionButton :icon="getAppIcon('close')" v-tooltip="$t('components.debugConsole.close')" :buttonFunction="closeConsole" />
      </div>
    </div>

    <div ref="logsContainer" class="debug-console-logs">
      <div v-for="(log, index) in filteredLogs" :key="index" class="log-entry" :class="'log-' + log.type">
        <span class="log-time">{{ log.time }}</span>
        <span class="log-type">{{ log.type.toUpperCase() }}</span>
        <div class="log-content">
          <span class="log-message">{{ formatMessage(log.args) }}</span>
        </div>
      </div>
      <div v-if="filteredLogs.length === 0" class="no-logs">{{ logs.length === 0 ? $t('components.debugConsole.noLogs') : $t('components.debugConsole.noMatchingLogs') }}</div>
    </div>
  </div>
</template>

<script setup>
// imports
import { computed, nextTick, onBeforeUnmount, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { Events } from "@wailsio/runtime";

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import SearchBar from '@/instances/desktop/components/SearchBar.vue';

// services
import { FSService, SettingsService } from '@/services';

// stores
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useIconStore } from '@/stores/icons';
import { useNotificationStore } from '@/stores/notifications';

const iconStore = useIconStore();
const modals = useDesktopModalStore();
const notificationStore = useNotificationStore();

const { t } = useI18n();

const emit = defineEmits(['close']);

// refs
const logs = ref([]);
const logsContainer = ref(null);
const maxLogs = ref(100);
const originalConsole = ref({});
const searchQuery = ref('');

// computed

// Returns logs filtered by search query.
const filteredLogs = computed(() => {
  if (!searchQuery.value.trim()) return logs.value;
  const query = searchQuery.value.toLowerCase();
  return logs.value.filter(log => {
    const message = formatMessage(log.args).toLowerCase();
    const type = log.type.toLowerCase();
    return message.includes(query) || type.includes(query);
  });
});

// methods

// Clears all captured logs.
const clearLogs = () => {
  logs.value = [];
};

// Emits close event to hide the debug console.
const closeConsole = () => {
  emit('close');
};

// Copies the filtered logs to clipboard in a formatted text format.
const copyLogs = async () => {
  const logsToExport = filteredLogs.value;
  if (logsToExport.length === 0) return;
  
  const formattedLogs = logsToExport.map(log => {
    return `[${log.time}] [${log.type.toUpperCase()}] ${formatMessage(log.args)}`;
  }).join('\n');
  
  try {
    await navigator.clipboard.writeText(formattedLogs);
    notificationStore.addNotification(t('components.debugConsole.copiedLogEntries', { count: logsToExport.length }), t('components.debugConsole.copiedLogEntriesToClipboard', { count: logsToExport.length }), 'success');
  } catch (e) {
    console.error('Failed to copy logs:', e);
  }
};

// Formats log arguments into a displayable string.
const formatMessage = (args) => {
  return args.map(arg => {
    if (arg === null) return 'null';
    if (arg === undefined) return 'undefined';
    if (typeof arg === 'object') {
      try {
        return JSON.stringify(arg, null, 2);
      } catch (e) {
        return String(arg);
      }
    }
    return String(arg);
  }).join(' ');
};

// Returns the app icon path for the given icon name.
const getAppIcon = (iconName) => iconStore.getAppIcon(iconName);

// Opens the diagnostics modal for submitting logs via email.
const openDiagnosticsModal = () => {
  modals.setModalVisibility('submitDiagnosticsModal', true);
};

// Parses Go slog format logs: time=2026-02-07T17:59:31.150+01:00 level=ERROR msg="..."
const parseGoLog = (message) => {
  // Try to parse slog format
  const timeMatch = message.match(/time=(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:\.\d+)?(?:[+-]\d{2}:\d{2})?)/);
  const levelMatch = message.match(/level=(\w+)/);
  const msgMatch = message.match(/msg="([^"]*(?:\\.[^"]*)*)"/);
  
  if (timeMatch && levelMatch) {
    // Parse the ISO timestamp
    const isoTime = timeMatch[1];
    const date = new Date(isoTime);
    const time = date.toLocaleTimeString('en-US', { hour12: false }) + '.' + String(date.getMilliseconds()).padStart(3, '0');
    
    // Get level
    const level = levelMatch[1].toLowerCase();
    
    // Get message or use remaining text
    let msg = msgMatch ? msgMatch[1] : message.replace(/time=[^\s]+\s*/, '').replace(/level=\w+\s*/, '').replace(/msg=/, '');
    
    // Unescape the message
    msg = msg.replace(/\\n/g, '\n').replace(/\\"/g, '"').replace(/\\\\/g, '\\');
    
    return { time, level, message: msg, parsed: true };
  }
  
  return { time: null, level: null, message, parsed: false };
};

// Loads previous logs from the log file.
const loadLogsFromFile = async () => {
  try {
    const logLines = await SettingsService.GetLogContents(maxLogs.value);
    if (!logLines || logLines.length === 0) return;
    
    for (const line of logLines) {
      if (!line.trim()) continue;
      
      // Try to parse slog format first
      const parsed = parseGoLog(line);
      
      if (parsed.parsed) {
        logs.value.push({
          type: parsed.level === 'error' ? 'error' : parsed.level === 'warn' ? 'warn' : parsed.level === 'info' ? 'info' : 'log',
          time: parsed.time,
          args: [`[Go] ${parsed.message}`],
          source: 'backend',
          timestamp: 0
        });
      } else {
        // Try old format: [timestamp] message
        const match = line.match(/^\[(\d{4}-\d{2}-\d{2}\s+\d{2}:\d{2}:\d{2})\]\s*(.*)$/);
        
        let time = '';
        let message = line;
        
        if (match) {
          const datePart = match[1];
          time = datePart.split(' ')[1] || datePart;
          message = match[2];
        }
        
        // Determine log level from message content
        let level = 'log';
        const lowerMsg = message.toLowerCase();
        if (lowerMsg.includes('error') || lowerMsg.includes('fatal')) {
          level = 'error';
        } else if (lowerMsg.includes('warn')) {
          level = 'warn';
        } else if (lowerMsg.includes('info')) {
          level = 'info';
        }
        
        logs.value.push({
          type: level,
          time: time,
          args: [`[Go] ${message}`],
          source: 'backend',
          timestamp: 0
        });
      }
    }
    
    // Scroll to bottom after loading
    nextTick(() => {
      if (logsContainer.value) {
        logsContainer.value.scrollTop = logsContainer.value.scrollHeight;
      }
    });
  } catch (e) {
    console.error('Failed to load logs from file:', e);
  }
};

// Opens the folder containing Clustta logs in the system file explorer.
const openLogsFolder = async () => {
  try {
    const logPath = await SettingsService.GetLogPath();
    if (logPath) {
      FSService.RevealInExplorer(logPath);
    }
  } catch (e) {
    console.error('Failed to open logs folder:', e);
  }
};

// Handles backend log events from Go.
const handleBackendLog = (event) => {
  const data = event.data;
  if (!data) return;
  
  // Try to parse slog format from the message
  const parsed = parseGoLog(data.message);
  
  let time, level, message;
  if (parsed.parsed) {
    time = parsed.time;
    level = parsed.level === 'error' ? 'error' : parsed.level === 'warn' ? 'warn' : parsed.level === 'info' ? 'info' : 'log';
    message = parsed.message;
  } else {
    const now = new Date();
    time = now.toLocaleTimeString('en-US', { hour12: false }) + '.' + String(now.getMilliseconds()).padStart(3, '0');
    level = data.level || 'log';
    message = data.message;
  }
  
  logs.value.push({
    type: level,
    time: time,
    args: [`[Go] ${message}`],
    timestamp: Date.now()
  });
  
  if (logs.value.length > maxLogs.value) {
    logs.value = logs.value.slice(-maxLogs.value);
  }
  
  nextTick(() => {
    if (logsContainer.value) {
      logsContainer.value.scrollTop = logsContainer.value.scrollHeight;
    }
  });
};

// Intercepts console methods and captures logs.
const interceptConsole = () => {
  const methods = ['info', 'warn', 'error'];
  
  methods.forEach(method => {
    originalConsole.value[method] = console[method];
    
    console[method] = (...args) => {
      // Call original console method
      originalConsole.value[method].apply(console, args);
      
      // Capture the log
      const now = new Date();
      const time = now.toLocaleTimeString('en-US', { hour12: false }) + '.' + String(now.getMilliseconds()).padStart(3, '0');
      
      logs.value.push({
        type: method,
        time: time,
        args: args,
        timestamp: now.getTime()
      });
      
      // Trim logs if exceeding max
      if (logs.value.length > maxLogs.value) {
        logs.value = logs.value.slice(-maxLogs.value);
      }
      
      // Auto-scroll to bottom
      nextTick(() => {
        if (logsContainer.value) {
          logsContainer.value.scrollTop = logsContainer.value.scrollHeight;
        }
      });
    };
  });
};

// Restores original console methods.
const restoreConsole = () => {
  const methods = ['info', 'warn', 'error'];
  methods.forEach(method => {
    if (originalConsole.value[method]) {
      console[method] = originalConsole.value[method];
    }
  });
};

// lifecycle hooks
onMounted(async () => {
  // Load previous logs from file first
  await loadLogsFromFile();
  
  interceptConsole();
  Events.On("backend-log", handleBackendLog);  
  // Signal to backend that debug console is ready to receive logs
  Events.Emit("debug-console-ready");
});

onBeforeUnmount(() => {
  restoreConsole();
  Events.Off("backend-log");
});

</script>

<style scoped>
.debug-console {
  position: relative;
  width: 100%;
  height: 250px;
  background-color: hsl(var(--background));
  display: flex;
  flex-direction: column;
  overflow: hidden;
  color: hsl(var(--foreground));
  box-sizing: border-box;
  border-radius: var(--large-radius);
  /* padding: 1rem; */
}

.debug-console-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.5rem;
  background-color: hsl(var(--card));
  user-select: none;
  border-radius: var(--normal-radius);

}

.debug-console-title {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-weight: 500;
  font-size: 13px;
}

.log-count {
  color: hsl(var(--muted-foreground));
  font-weight: 400;
}

.debug-console-actions {
  display: flex;
  align-items: center;
  gap: 0.25rem;
}

.debug-console-actions :deep(.searchbar-container) {
  height: 28px;
  min-height: 28px;
  width: 180px;
}

.debug-console-actions :deep(.searchbar-input) {
  font-size: 12px;
  padding: 6px 8px;
}

.debug-console-logs {
  flex: 1;
  overflow-y: auto;
  padding: 0.5rem;
  font-family: 'Consolas', 'Monaco', monospace;
  font-size: 12px;
}

.debug-console-logs::-webkit-scrollbar {
  width: 4px;
}

.debug-console-logs::-webkit-scrollbar-thumb {
  border-radius: var(--small-radius);
  background-color: hsl(var(--border));
}

.debug-console-logs::-webkit-scrollbar-track {
  border-radius: var(--small-radius);
}

.log-entry {
  display: flex;
  gap: 0.5rem;
  padding: 0.3rem 0.5rem;
  border-radius: 3px;
  margin-bottom: 2px;
  word-break: break-word;
}

.log-entry:hover {
  background-color: hsl(var(--accent));
}

.log-log { border-left: 2px solid hsl(var(--muted-foreground)); }
.log-info { border-left: 2px solid var(--info); }
.log-warn { border-left: 2px solid var(--alert); background-color: rgba(255, 193, 7, 0.1); }
.log-error { border-left: 2px solid var(--red); background-color: rgba(220, 53, 69, 0.1); }

.log-time {
  color: hsl(var(--muted-foreground));
  font-size: 10px;
  min-width: 85px;
}

.log-type {
  font-weight: 600;
  font-size: 10px;
  min-width: 45px;
}

.log-log .log-type { color: hsl(var(--muted-foreground)); }
.log-info .log-type { color: var(--info); }
.log-warn .log-type { color: var(--alert); }
.log-error .log-type { color: var(--red); }

.log-content {
  flex: 1;
  display: flex;
  gap: 0.5rem;
  align-items: flex-start;
}

.log-message {
  flex: 1;
  white-space: pre-wrap;
}

.no-logs {
  text-align: center;
  color: hsl(var(--muted-foreground));
  padding: 2rem;
}
</style>
