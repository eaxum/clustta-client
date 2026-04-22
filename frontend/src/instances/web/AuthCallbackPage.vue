<template>
  <div class="sso-callback">
    <div class="content">
      <div v-if="errorMessage" class="error">
        <h2>Sign-in failed</h2>
        <p>{{ errorMessage }}</p>
        <button class="btn" @click="$router.push('/auth/login')">Back to login</button>
      </div>
      <div v-else class="loading">
        <div class="spinner" />
        <p>Completing sign-in&hellip;</p>
      </div>
    </div>
  </div>
</template>

<script setup>
import { onMounted, ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { addAccountToStorage, STORAGE_KEYS } from '@clustta/web-adapters';

const route = useRoute();
const router = useRouter();
const errorMessage = ref('');

const decodeUser = (encoded) => {
  if (!encoded) return null;
  try {
    const padded = encoded.replace(/-/g, '+').replace(/_/g, '/');
    const json = atob(padded + '=='.slice((padded.length + 2) % 4));
    return JSON.parse(json);
  } catch (e) {
    console.warn('Failed to decode user payload', e);
    return null;
  }
};

onMounted(() => {
  const sessionId = route.query.session_id;
  const userParam = route.query.user;

  if (!sessionId) {
    errorMessage.value = 'Missing session_id in callback URL';
    return;
  }

  const userData = decodeUser(userParam) || {};
  const user = {
    id: userData.id || userData.Id || '',
    username: userData.username || userData.user_name || userData.UserName || '',
    email: userData.email || userData.Email || '',
    first_name: userData.first_name || userData.FirstName || '',
    last_name: userData.last_name || userData.LastName || '',
    photo: userData.photo || '',
  };

  localStorage.setItem(STORAGE_KEYS.SESSION_ID, sessionId);
  localStorage.setItem(STORAGE_KEYS.USER, JSON.stringify(user));
  addAccountToStorage({ session_id: sessionId, user });

  router.replace('/');
});
</script>

<style scoped>
.sso-callback {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 100vh;
  background: var(--app-background, #1a1a1a);
  color: var(--text-color, #fff);
}

.content {
  text-align: center;
  max-width: 360px;
  padding: 2rem;
}

.spinner {
  width: 32px;
  height: 32px;
  margin: 0 auto 1rem;
  border: 3px solid rgba(255, 255, 255, 0.15);
  border-top-color: var(--accent, #4a9eff);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.error h2 {
  margin: 0 0 0.5rem;
  font-size: 1.1rem;
}

.error p {
  margin: 0 0 1rem;
  color: var(--muted, #999);
}

.btn {
  padding: 0.5rem 1rem;
  border-radius: 6px;
  border: 1px solid var(--border, #333);
  background: transparent;
  color: inherit;
  cursor: pointer;
}

.btn:hover {
  background: rgba(255, 255, 255, 0.05);
}
</style>
