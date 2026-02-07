<template>
  <div class="debug-console">
    <div class="debug-console-header">
      <div class="debug-console-title">
        <img class="small-icons" :src="getAppIcon('bug')">
        <span>Debug Console</span>
        <span class="log-count">({{ logs.length }})</span>
      </div>
      <div class="debug-console-actions">
        <ActionButton :icon="getAppIcon('code-bracket')" :isActive="showSource" v-tooltip="'Show source location'" :buttonFunction="toggleShowSource" />
        <ActionButton :icon="getAppIcon('broom')" v-tooltip="'Clear'" :buttonFunction="clearLogs" />
      </div>
    </div>

    <div ref="logsContainer" class="debug-console-logs">
      <div v-for="(log, index) in logs" :key="index" class="log-entry" :class="'log-' + log.type">
        <span class="log-time">{{ log.time }}</span>
        <span class="log-type">{{ log.type.toUpperCase() }}</span>
        <div class="log-content">
          <span class="log-message">{{ formatMessage(log.args) }}</span>
          <span v-if="showSource && log.source" class="log-source">{{ log.source }}</span>
        </div>
      </div>
      <div v-if="logs.length === 0" class="no-logs">No logs to display</div>
    </div>
  </div>
</template>

<script setup>
// imports
import { nextTick, onBeforeUnmount, onMounted, ref } from 'vue';

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue';

// stores
import { useIconStore } from '@/stores/icons';

const iconStore = useIconStore();

// refs
const logs = ref([]);
const logsContainer = ref(null);
const maxLogs = ref(500);
const originalConsole = ref({});
const showSource = ref(false);

// methods

// Clears all captured logs.
const clearLogs = () => {
  logs.value = [];
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

// Extracts source location from stack trace.
const getSourceLocation = () => {
  try {
    const stack = new Error().stack;
    if (!stack) return null;
    
    const lines = stack.split('\n');
    // Find the first line that's not from this file or console internals
    for (let i = 0; i < lines.length; i++) {
      const line = lines[i];
      if (line.includes('DebugConsole.vue')) continue;
      if (line.includes('getSourceLocation')) continue;
      if (line.includes('interceptConsole')) continue;
      if (line.includes('at console.')) continue;
      if (line.includes('at Object.')) continue;
      
      // Extract file:line:column from the stack trace
      const match = line.match(/(?:at\s+)?(?:.*?\s+\()?(.+?):(\d+):(\d+)\)?$/);
      if (match) {
        let filePath = match[1];
        const lineNum = match[2];
        const colNum = match[3];
        
        // Clean up the file path - extract just the filename
        const fileMatch = filePath.match(/([^/\\]+\.(?:vue|js|ts|jsx|tsx))(?:\?.*)?$/);
        if (fileMatch) {
          return `${fileMatch[1]}:${lineNum}:${colNum}`;
        }
        return `${filePath.split('/').pop()}:${lineNum}`;
      }
    }
    return null;
  } catch (e) {
    return null;
  }
};

// Intercepts console methods and captures logs.
const interceptConsole = () => {
  const methods = ['info', 'warn', 'error'];
  
  methods.forEach(method => {
    originalConsole.value[method] = console[method];
    
    console[method] = (...args) => {
      // Get source location before calling original (to get correct stack)
      const source = getSourceLocation();
      
      // Call original console method
      originalConsole.value[method].apply(console, args);
      
      // Capture the log
      const now = new Date();
      const time = now.toLocaleTimeString('en-US', { hour12: false }) + '.' + String(now.getMilliseconds()).padStart(3, '0');
      
      logs.value.push({
        type: method,
        time: time,
        args: args,
        source: source,
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

// Toggles source location display.
const toggleShowSource = () => {
  showSource.value = !showSource.value;
};

// lifecycle hooks
onMounted(() => {
  interceptConsole();
});

onBeforeUnmount(() => {
  restoreConsole();
});
</script>

<style scoped>
.debug-console {
  position: relative;
  width: 100%;
  height: 250px;
  background-color: var(--black-steel);
  display: flex;
  flex-direction: column;
  overflow: hidden;
  color: var(--white);
  box-sizing: border-box;
  border-radius: var(--large-radius);
  /* padding: 1rem; */
}

.debug-console-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.5rem;
  background-color: var(--midnight-steel);
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
  color: var(--silver);
  font-weight: 400;
}

.debug-console-actions {
  display: flex;
  gap: 0.25rem;
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
  background-color: var(--light-steel);
}

.debug-console-logs::-webkit-scrollbar-track {
  border-radius: var(--small-radius);
}

.log-entry {
  display: flex;
  gap: 0.5rem;
  padding: 0.3rem 0.5rem;
  border-radius: var(--small-radius);
  margin-bottom: 2px;
  word-break: break-word;
}

.log-entry:hover {
  background-color: var(--hover);
}

.log-info { border-left: 2px solid var(--info); }
.log-warn { border-left: 2px solid var(--alert); background-color: rgba(255, 193, 7, 0.1); }
.log-error { border-left: 2px solid var(--red); background-color: rgba(220, 53, 69, 0.1); }

.log-time {
  color: var(--silver);
  font-size: 10px;
  min-width: 85px;
}

.log-type {
  font-weight: 600;
  font-size: 10px;
  min-width: 45px;
}

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

.log-source {
  color: var(--silver);
  font-size: 10px;
  font-style: italic;
  white-space: nowrap;
}

.no-logs {
  text-align: center;
  color: var(--silver);
  padding: 2rem;
}
</style>
