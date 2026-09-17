<script setup lang="ts">
import { ref } from 'vue';
import { useMutation } from '@vue/apollo-composable';
import { useRouter } from 'vue-router';
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
const authStore = useAuthStore();

const username = ref('');
const password = ref('');
const error = ref('');
const loading = ref(false);

const { mutate: login } = useMutation(LOGIN_MUTATION);

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
</style>
