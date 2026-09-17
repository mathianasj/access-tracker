<script setup lang="ts">
import { useAuthStore } from './stores/auth';
import { useRouter } from 'vue-router';

const authStore = useAuthStore();
const router = useRouter();

function logout() {
  authStore.logout();
  router.push('/login');
}
</script>

<template>
  <div id="app">
    <nav v-if="authStore.isAuthenticated" class="navbar">
      <span>Welcome, {{ authStore.username }}</span>
      <button @click="logout" class="logout-btn">Logout</button>
    </nav>
    <RouterView />
  </div>
</template>

<style scoped>
.navbar {
  display: flex;
  justify-content: flex-end;
  align-items: center;
  padding: 1rem;
  background: #333;
  color: white;
  gap: 1rem;
}

.logout-btn {
  padding: 0.5rem 1rem;
  background: #dc3545;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
}

.logout-btn:hover {
  background: #c82333;
}
</style>
