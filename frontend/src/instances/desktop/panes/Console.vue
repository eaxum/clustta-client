<template>
  <div class="general-pane-root">
    <div class="console-container">
      <!-- Chat messages area -->

      <div class="console-messages" ref="messagesContainer">
        
        <div 
          v-for="(message, index) in messages" 
          :key="index"
          :class="['console-message', message.type]" >

          <div class="message-content">
            <div class="message-header">
              <span class="message-sender">{{ message.sender }}</span>
              <span class="message-time">{{ formatTime(message.timestamp) }}</span>
            </div>
            <div class="message-text">{{ message.content }}</div>
          </div>
        </div>
        
        <!-- Placeholder when no messages -->
        <div v-if="!messages.length" class="console-empty">
          <div class="empty-icon">
            <i :class="selectedConsoleTab === 'Bash' ? 'console-icon' : 'console-icon'"></i>
          </div>
          <div class="empty-text">{{ emptyStateTitle }}</div>
          <div class="empty-subtext">{{ emptyStateSubtext }}</div>
        </div>
      </div>

      <!-- Input area -->
      <div class="console-input-container">
        <div class="console-input-wrapper">
            
          <textarea ref="textareaRef" v-model="currentMessage" class="console-input" type="text" placeholder="..."
            spellcheck="false" @input="handleInput" @keydown.enter="sendMessage" />

          <ActionButton 
            :icon="getAppIcon('send')" 
            :showLabel="false"
            :isDisabled="!currentMessage.trim()"
            v-tooltip="'Send message'"
            :buttonFunction="sendMessage" 
          />
        </div>
        
        <!-- Console mode tabs -->
        <div class="console-tabs-container">
          <PaneHeaderTabs 
            :iconsOnly="false" 
            :useSelected="true" 
            :selectedTab="selectedConsoleTab" 
            :dataTypes="consoleTabs" 
            @filter="handleConsoleTabClick" 
          />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
// imports
import { ref, computed, nextTick, onMounted, onUnmounted } from 'vue';
import emitter from '@/lib/mitt';

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import PaneHeaderTabs from '@/instances/common/components/PaneHeaderTabs.vue';

// stores
import { useAssetStore } from '@/stores/assets';
import { useCollectionStore } from '@/stores/collections';
import { useIconStore } from '@/stores/icons';
import { useUserStore } from '@/stores/users';
import { useStageStore } from '@/stores/stages';

const assetStore = useAssetStore();
const collectionStore = useCollectionStore();
const iconStore = useIconStore();
const userStore = useUserStore();
const stage = useStageStore();

// refs
const messagesContainer = ref(null);
const messageInput = ref(null);
const textareaRef = ref(null);
const currentMessage = ref('');
const selectedConsoleTab = ref('Agent');
const messages = ref([
]);

// console tabs data
const consoleTabs = ref([
  { name: "Agent", icon: "brain" },
  { name: "Bash", icon: "console" }
]);

// computed properties
const selectedItemName = computed(() => {
  if (assetStore.selectedAsset) {
    return assetStore.selectedAsset.name || 'this asset';
  } else if (collectionStore.selectedCollection) {
    return collectionStore.selectedCollection.name || 'this collection';
  }
  return 'this item';
});

const itemType = computed(() => {
  if (assetStore.selectedAsset) {
    return 'asset';
  } else if (collectionStore.selectedCollection) {
    return 'collection';
  } else if(!stage.markedItems.length){
    return 'project'
  }
  return 'item';
});

const emptyStateTitle = computed(() => {
  if (selectedConsoleTab.value === 'Bash') {
    return 'Terminal ready...';
  } else {
    return 'Start a conversation...';
  }
});

const emptyStateSubtext = computed(() => {
  if (selectedConsoleTab.value === 'Bash') {
    return `Execute terminal commands on this ${itemType.value}`;
  } else {
    return `Perform an operation on this ${itemType.value} or get help with Clustta`;
  }
});

// methods
const getAppIcon = (iconName) => {
  return iconStore.getAppIcon(iconName);
};

const getMessageIcon = (messageType) => {
  return messageType === 'user' ? 'person' : 'robot';
};

const formatTime = (timestamp) => {
  const date = new Date(timestamp);
  return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
};

const handleInput = (event) => {
  if (textareaRef.value) {
    textareaRef.value.style.height = 'auto';
    
    const newHeight = textareaRef.value.scrollHeight;
    textareaRef.value.style.height = newHeight + 'px';
  }
};

const handleConsoleTabClick = (tabName) => {
  selectedConsoleTab.value = tabName;
  // Handle different console modes
  switch (tabName) {
    case 'Agent':
      break;
    case 'Bash':
      break;
  }
};

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
  
  // Simulate AI response (replace with actual API call)
  setTimeout(() => {
    const aiResponse = {
      id: messages.value.length + 1,
      type: 'assistant',
      sender: 'Clustta',
      content: `I understand you want to "${currentMessage.value}". Let me help you with that ...`,
      timestamp: Date.now()
    };
    messages.value.push(aiResponse);
    scrollToBottom();
  }, 1000);

  currentMessage.value = '';
  
  // Reset textarea height after sending message
  if (textareaRef.value) {
    textareaRef.value.style.height = 'auto';
  }
  
  scrollToBottom();
};

const sendQuickMessage = (action) => {
  let message = '';
  switch (action) {
    case 'explain':
      message = `Explain what ${selectedItemName.value} does`;
      break;
    case 'suggestions':
      message = `Give me suggestions for improving ${selectedItemName.value}`;
      break;
    case 'debug':
      message = `Help me debug issues with ${selectedItemName.value}`;
      break;
  }
  
  currentMessage.value = message;
  sendMessage();
};

const scrollToBottom = () => {
  nextTick(() => {
    if (messagesContainer.value) {
      messagesContainer.value.scrollTop = messagesContainer.value.scrollHeight;
    }
  });
};

const clearMessages = () => {
  messages.value = [];
};

// lifecycle
onMounted(() => {
  scrollToBottom();
});

onUnmounted(() => {
  // Cleanup if needed
});
</script>

<style scoped>
@import "@/assets/desktop.css";

.console-container {
  display: flex;
  flex-direction: column;
  height: 100%;
  /* background-color: var(--black); */
  border-radius: var(--medium-radius);
  overflow: hidden;
  width: 100%;
  box-sizing: border-box;
  /* background-color: crimson; */
}

.console-messages {
  flex: 1;
  overflow-y: auto;
  padding: 1rem .5rem;
  display: flex;
  flex-direction: column;
  gap: 1rem;
  min-height: 0;
}

.console-message {
  display: flex;
  gap: 0.75rem;
  align-items: flex-start;
  /* max-width: 90%; */
  background-color: red;
  border-radius: 8px;
  background-color: var(--steel);
  color: var(--white);
  font-size: 14px;
}

.console-message.user {
  margin-left: auto;
  flex-direction: row-reverse;
}

.console-message.assistant {
  margin-right: auto;
  flex-direction: row;
}

.message-avatar {
  flex-shrink: 0;
}

.message-content {
  flex: 1;
  min-width: 0;
  background-color: var(--gray-800);
  border-radius: var(--small-radius);
  padding: 0.75rem;
}

.console-message.user .message-content {
  background-color: var(--accent-color);
}

.console-message.assistant .message-content {
  background-color: var(--gray-800);
}

.message-header {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin-bottom: 0.25rem;
}

.console-message.user .message-header {
  justify-content: flex-end;
}

.console-message.assistant .message-header {
  justify-content: flex-start;
}

.message-sender {
  font-weight: 600;
  font-size: 0.875rem;
  color: var(--white);
}

.message-time {
  font-size: 0.75rem;
  color: var(--gray-500);
}

.message-text {
  color: var(--gray-200);
  line-height: 1.5;
  word-wrap: break-word;
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

.empty-icon {
  font-size: 3rem;
  color: var(--gray-600);
  margin-bottom: 1rem;
}

.empty-text {
  font-size: 1.125rem;
  font-weight: 500;
  color: var(--gray-400);
}

.empty-subtext {
  font-size: 0.875rem;
  max-width: 250px;
}

.console-input-container {
  border-top: 1px solid var(--gray-800);
  /* padding: 1rem; */
  background-color: var(--black-steel);
  height: min-content;
}

.console-input-wrapper {
  display: flex;
  gap: 0.5rem;
  align-items: flex-end;
  margin-bottom: 0.75rem;
  background-color: var(--steel);
  border-radius: 8px;
  padding: .5rem;
  height: min-content;
  max-height: 30vh;
}

.console-input-wrapper:hover {
    outline: var(--transparent-line);
    outline-offset: -1px;
}

.console-input{
  font-family: 'Inter', sans-serif;
  font-weight: 300;
  box-sizing: border-box;
  font-size: 14px;
  border-width: 0px;
  width: 100%;
  min-height: 60px;
  height: auto;
  max-height: 30vh;
  margin: auto;
  outline: none;
  resize: none;
  background-color: transparent;
  color: var(--white);
  overflow-y: auto;
  scrollbar-width: none; /* Firefox */
  -ms-overflow-style: none; /* Internet Explorer 10+ */
  pointer-events: none;
}

/* Hide scrollbar for WebKit browsers */
.console-input::-webkit-scrollbar {
  display: none;
}

.console-tabs-container {
  padding: 0.5rem 0;
  border-top: 1px solid var(--gray-800);
}

.general-pane-root{
    padding: .5rem 0;
    box-sizing: border-box;
}

/* Scrollbar styling */
.console-messages::-webkit-scrollbar {
  width: 6px;
}

.console-messages::-webkit-scrollbar-track {
  background: var(--gray-800);
  border-radius: 3px;
}

.console-messages::-webkit-scrollbar-thumb {
  background: var(--gray-600);
  border-radius: 3px;
}

.console-messages::-webkit-scrollbar-thumb:hover {
  background: var(--gray-500);
}
</style>