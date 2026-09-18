import { defineStore } from 'pinia';
import { ref, computed } from 'vue';

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string | null>(localStorage.getItem('auth_token'));
  const username = ref<string | null>(localStorage.getItem('auth_username'));

  const isAuthenticated = computed(() => !!token.value);

  async function checkAuth() {
    try {
      const response = await fetch('/token');
      if (response.ok) {
        const data = await response.json();
        setToken(data.token);
        setUser(data.username);
        return true;
      }
    } catch (e) {
      console.error('Failed to check auth:', e);
    }
    return false;
  }

  function setToken(newToken: string) {
    token.value = newToken;
    localStorage.setItem('auth_token', newToken);
  }

  function setUser(newUsername: string) {
    username.value = newUsername;
    localStorage.setItem('auth_username', newUsername);
  }

  function logout() {
    token.value = null;
    username.value = null;
    localStorage.removeItem('auth_token');
    localStorage.removeItem('auth_username');
  }

  function getAuthHeader() {
    return token.value ? `Bearer ${token.value}` : '';
  }

  return {
    token,
    username,
    isAuthenticated,
    checkAuth,
    setToken,
    setUser,
    logout,
    getAuthHeader
  };
});
