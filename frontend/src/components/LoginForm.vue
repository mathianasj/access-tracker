<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { useMutation } from '@vue/apollo-composable';
import { useRouter, useRoute } from 'vue-router';
import { useAuthStore } from '../stores/auth';
import { gql } from '@apollo/client/core';

const LOGIN_MUTATION = gql`
  mutation Login($input: LoginInput!) {
    login(input: $input) {
      token
      username
    }
  }
`;

const router = useRouter();
const route = useRoute();
const authStore = useAuthStore();

const username = ref('');
const password = ref('');
const error = ref('');
const loading = ref(false);

const { mutate: login } = useMutation(LOGIN_MUTATION);

onMounted(async () => {
  const errorParam = route.query.error as string;
  if (errorParam) {
    const errorMessages: Record<string, string> = {
      oauth_state_missing: 'OAuth state missing. Please try again.',
      oauth_state_mismatch: 'OAuth state mismatch. Please try again.',
      oauth_code_missing: 'OAuth code missing. Please try again.',
      token_exchange_failed: 'Failed to exchange OAuth token.',
      get_user_failed: 'Failed to get user info from GitHub.',
      jwt_generation_failed: 'Failed to generate authentication token.',
    };
    error.value = errorMessages[errorParam] || 'OAuth login failed.';
  }

  const oauthSuccess = route.query.oauth as string;
  if (oauthSuccess === 'success') {
    const authenticated = await authStore.checkAuth();
    if (authenticated) {
      router.push('/requests');
    } else {
      error.value = 'OAuth login failed. Please try again.';
    }
    return;
  }

  if (authStore.isAuthenticated) {
    router.push('/requests');
    return;
  }

  const hasToken = await authStore.checkAuth();
  if (hasToken) {
    router.push('/requests');
  }
});

async function handleSubmit() {
  error.value = '';
  loading.value = true;

  try {
    const result = await login({
      input: {
        username: username.value,
        password: password.value
      }
    });

    if (result?.data?.login) {
      authStore.setToken(result.data.login.token);
      authStore.setUser(result.data.login.username);
      router.push('/requests');
    } else {
      error.value = 'Invalid response from server';
    }
  } catch (e: any) {
    error.value = e.message || 'Login failed';
  } finally {
    loading.value = false;
  }
}

function loginWithGithub() {
  window.location.href = '/oauth/github';
}
</script>

<template>
  <div class="login-container">
    <form @submit.prevent="handleSubmit" class="login-form">
      <h1>Access Tracker Login</h1>

      <div v-if="error" class="error">{{ error }}</div>

      <div class="form-group">
        <label for="username">Username</label>
        <input
          id="username"
          v-model="username"
          type="text"
          required
          autocomplete="username"
        />
      </div>

      <div class="form-group">
        <label for="password">Password</label>
        <input
          id="password"
          v-model="password"
          type="password"
          required
          autocomplete="current-password"
        />
      </div>

      <button type="submit" :disabled="loading">
        {{ loading ? 'Logging in...' : 'Login' }}
      </button>

      <div class="divider">
        <span>or</span>
      </div>

      <button type="button" @click="loginWithGithub" class="github-btn">
        <svg height="20" width="20" viewBox="0 0 24 24" fill="currentColor">
          <path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z"/>
        </svg>
        Continue with GitHub
      </button>
    </form>
  </div>
</template>

<style scoped>
.login-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  background: #f5f5f5;
}

.login-form {
  background: white;
  padding: 2rem;
  border-radius: 8px;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);
  width: 100%;
  max-width: 400px;
}

h1 {
  margin: 0 0 1.5rem;
  color: #333;
  text-align: center;
}

.form-group {
  margin-bottom: 1rem;
}

label {
  display: block;
  margin-bottom: 0.5rem;
  color: #666;
}

input {
  width: 100%;
  padding: 0.75rem;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 1rem;
  box-sizing: border-box;
}

input:focus {
  outline: none;
  border-color: #4a90d9;
}

button {
  width: 100%;
  padding: 0.75rem;
  background: #4a90d9;
  color: white;
  border: none;
  border-radius: 4px;
  font-size: 1rem;
  cursor: pointer;
  margin-top: 0.5rem;
}

button:hover:not(:disabled) {
  background: #3a7bc8;
}

button:disabled {
  background: #ccc;
  cursor: not-allowed;
}

.error {
  background: #fee;
  border: 1px solid #fcc;
  color: #c00;
  padding: 0.75rem;
  border-radius: 4px;
  margin-bottom: 1rem;
}

.divider {
  display: flex;
  align-items: center;
  margin: 1.5rem 0;
  color: #666;
}

.divider::before,
.divider::after {
  content: "";
  flex: 1;
  border-bottom: 1px solid #ddd;
}

.divider span {
  padding: 0 1rem;
  font-size: 0.875rem;
}

.github-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  background: #24292e;
  color: white;
}

.github-btn:hover:not(:disabled) {
  background: #1a1e22;
}
</style>
