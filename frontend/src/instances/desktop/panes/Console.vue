<template>
  <div class="general-pane-root">
    <div class="console-container">
      <div class="console-messages" ref="messagesContainer">
        <div v-for="(message, index) in messages" :key="index" :class="['console-message', message.type]">
          <div class="message-content">
            <div class="message-header">
              <span class="message-sender">{{ message.sender }}</span>
              <span class="message-time">{{ formatTime(message.timestamp) }}</span>
            </div>
            <div class="message-text">{{ message.content }}</div>
          </div>
        </div>

        <div v-if="!messages.length" class="console-empty">
          <div class="empty-text">{{ emptyStateTitle }}</div>
          <div class="empty-subtext">{{ emptyStateSubtext }}</div>
        </div>
      </div>

      <div class="console-input-container">
        <div class="console-input-wrapper">
          <textarea ref="textareaRef" v-model="currentMessage" class="console-input" type="text" placeholder="[Coming Soon]"
            spellcheck="false" @input="handleInput" @keydown.enter="sendMessage" />
          <ActionButton :icon="getAppIcon('send')" :showLabel="false" :isDisabled="!currentMessage.trim()"
            v-tooltip="'Send message'" :buttonFunction="sendMessage" />
        </div>

        <div class="console-tabs-container">
          <PaneHeaderTabs :iconsOnly="false" :useSelected="true" :selectedTab="selectedConsoleTab"
            :dataTypes="consoleTabs" @filter="handleConsoleTabClick" />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
// imports
import { computed, nextTick, onMounted, ref } from 'vue';

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import PaneHeaderTabs from '@/instances/common/components/PaneHeaderTabs.vue';

// stores
import { useAssetStore } from '@/stores/assets';
import { useCollectionStore } from '@/stores/collections';
import { useIconStore } from '@/stores/icons';
import { useStageStore } from '@/stores/stages';

const assetStore = useAssetStore();
const collectionStore = useCollectionStore();
const iconStore = useIconStore();
const stage = useStageStore();

// refs
const consoleTabs = ref([
  { name: "Agent", icon: "brain" },
  { name: "Bash", icon: "console" }
]);
const currentMessage = ref('');
const messages = ref([]);
const messagesContainer = ref(null);
const selectedConsoleTab = ref('Agent');
const textareaRef = ref(null);

// computed properties
const emptyStateSubtext = computed(() => {
  if (selectedConsoleTab.value === 'Bash') return `Execute terminal commands on this ${itemType.value}`;
  return `Perform an operation on this ${itemType.value} or get help with Clustta`;
});

const emptyStateTitle = computed(() => {
  if (selectedConsoleTab.value === 'Bash') return 'Terminal ready...';
  return 'Start a conversation...';
});

const itemType = computed(() => {
  if (assetStore.selectedAsset) return 'asset';
  if (collectionStore.selectedCollection) return 'collection';
  if (!stage.markedItems.length) return 'project';
  return 'item';
});

// methods

// Formats a timestamp into a readable time string.
const formatTime = (timestamp) => {
  const date = new Date(timestamp);
  return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
};

// Returns the app icon path for the given icon name.
const getAppIcon = (iconName) => iconStore.getAppIcon(iconName);

// Switches the active console tab mode.
const handleConsoleTabClick = (tabName) => {
  selectedConsoleTab.value = tabName;
};

// Auto-resizes the textarea based on content.
const handleInput = () => {
  if (!textareaRef.value) return;
  textareaRef.value.style.height = 'auto';
  textareaRef.value.style.height = textareaRef.value.scrollHeight + 'px';
};

// Scrolls the messages container to the bottom.
const scrollToBottom = () => {
  nextTick(() => {
    if (messagesContainer.value) messagesContainer.value.scrollTop = messagesContainer.value.scrollHeight;
  });
};

// Sends the current message and simulates an AI response.
const sendMessage = () => {
  if (!currentMessage.value.trim()) return;
  const userMessage = {
    id: messages.value.length + 1,
    type: 'user',
    sender: 'You',
    content: currentMessage.value.trim(),
    timestamp: Date.now()
  };
  messages.value.push(userMessage);
  const messageContent = currentMessage.value;
  currentMessage.value = '';
  if (textareaRef.value) textareaRef.value.style.height = 'auto';
  scrollToBottom();
  setTimeout(() => {
    const aiResponse = {
      id: messages.value.length + 1,
      type: 'assistant',
      sender: 'Clustta',
      content: `I understand you want to "${messageContent}". Let me help you with that ...`,
      timestamp: Date.now()
    };
    messages.value.push(aiResponse);
    scrollToBottom();
  }, 1000);
};

// lifecycle hooks
onMounted(() => {
  scrollToBottom();
});
</script>

<style scoped>
@import "@/assets/desktop.css";

.console-container {
  display: flex;
  flex-direction: column;
  height: 100%;
  width: 100%;
  box-sizing: border-box;
  border-radius: var(--medium-radius);
  overflow: hidden;
}

.console-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100%;
  text-align: center;
  color: var(--white);
  gap: 0.5rem;
}

.console-input {
  font-family: 'Inter', sans-serif;
  font-weight: 300;
  font-size: 14px;
  box-sizing: border-box;
  width: 100%;
  min-height: 60px;
  height: auto;
  max-height: 30vh;
  margin: auto;
  border-width: 0px;
  outline: none;
  resize: none;
  background-color: transparent;
  color: var(--white);
  overflow-y: auto;
  scrollbar-width: none;
  -ms-overflow-style: none;
}

.console-input::-webkit-scrollbar {
  display: none;
}

.console-input-container {
  height: min-content;
  border-top: 1px solid var(--gray-800);
  background-color: var(--black-steel);
}

.console-input-wrapper {
  display: flex;
  align-items: flex-end;
  gap: 0.5rem;
  height: min-content;
  max-height: 30vh;
  margin-bottom: 0.75rem;
  padding: .5rem;
  box-sizing: border-box;
  background-color: var(--steel);
  border-radius: 8px;
}

.console-input-wrapper:hover {
  outline: var(--transparent-line);
  outline-offset: -1px;
}

.console-message {
  display: flex;
  align-items: flex-start;
  gap: 0.75rem;
  font-size: 14px;
  color: var(--white);
  background-color: var(--steel);
  border-radius: 8px;
}

.console-message.assistant {
  flex-direction: row;
  margin-right: auto;
}

.console-message.assistant .message-content {
  background-color: var(--gray-800);
}

.console-message.assistant .message-header {
  justify-content: flex-start;
}

.console-message.user {
  flex-direction: row-reverse;
  margin-left: auto;
}

.console-message.user .message-content {
  background-color: var(--accent-color);
}

.console-message.user .message-header {
  justify-content: flex-end;
}

.console-messages {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 1rem;
  min-height: 0;
  padding: 1rem .5rem;
  overflow-y: auto;
}

.console-messages::-webkit-scrollbar {
  width: 4px;
}

.console-messages::-webkit-scrollbar-thumb {
  border-radius: var(--small-radius);
  background-color: var(--light-steel);
}

.console-messages::-webkit-scrollbar-track {
  border-radius: var(--small-radius);
}

.console-tabs-container {
  padding: 0.5rem 0;
  border-top: 1px solid var(--gray-800);
}

.empty-subtext {
  font-size: 0.875rem;
  max-width: 250px;
}

.empty-text {
  font-size: 1.125rem;
  font-weight: 500;
  color: var(--gray-400);
}

.general-pane-root {
  padding: .5rem 0;
  box-sizing: border-box;
}

.message-content {
  flex: 1;
  min-width: 0;
  padding: 0.75rem;
  background-color: var(--gray-800);
  border-radius: var(--small-radius);
}

.message-header {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin-bottom: 0.25rem;
}

.message-sender {
  font-size: 0.875rem;
  font-weight: 600;
  color: var(--white);
}

.message-text {
  line-height: 1.5;
  color: var(--gray-200);
  word-wrap: break-word;
}

.message-time {
  font-size: 0.75rem;
  color: var(--gray-500);
}
</style>