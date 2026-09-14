<script setup lang="ts">
import { ref } from 'vue';
import { useMutation } from '@vue/apollo-composable';
import { gql } from '@apollo/client/core';

const CREATE_ACCESS_REQUEST = gql`
  mutation CreateAccessRequest($input: CreateAccessRequestInput!) {
    createAccessRequest(input: $input) {
      id
      requestId
      requester
      systemResource
      accessLevel
      status
      createdAt
    }
  }
`;

const form = ref({
  requester: '',
  systemResource: '',
  accessLevel: 'READ' as 'READ' | 'WRITE' | 'ADMIN',
  justification: ''
});

const errors = ref({
  requester: '',
  systemResource: '',
  justification: ''
});

const isSubmitting = ref(false);
const toastMessage = ref('');
const toastType = ref<'success' | 'error'>('success');
const showToast = ref(false);

function validateForm(): boolean {
  errors.value = {
    requester: '',
    systemResource: '',
    justification: ''
  };

  if (!form.value.requester.trim()) {
    errors.value.requester = 'Requester is required';
  }

  if (!form.value.systemResource.trim()) {
    errors.value.systemResource = 'System/Resource is required';
  }

  if (!form.value.justification.trim()) {
    errors.value.justification = 'Justification is required';
  }

  return !errors.value.requester && !errors.value.systemResource && !errors.value.justification;
}

const { mutate: createAccessRequest } = useMutation(CREATE_ACCESS_REQUEST);

async function handleSubmit() {
  if (!validateForm()) {
    return;
  }

  isSubmitting.value = true;

  try {
    await createAccessRequest({
      input: {
        requester: form.value.requester,
        systemResource: form.value.systemResource,
        accessLevel: form.value.accessLevel,
        justification: form.value.justification
      }
    });

    toastMessage.value = 'Access request submitted successfully!';
    toastType.value = 'success';
    showToast.value = true;

    form.value = {
      requester: '',
      systemResource: '',
      accessLevel: 'READ',
      justification: ''
    };

    setTimeout(() => {
      showToast.value = false;
    }, 3000);
  } catch (error) {
    toastMessage.value = 'Failed to submit access request. Please try again.';
    toastType.value = 'error';
    showToast.value = true;

    setTimeout(() => {
      showToast.value = false;
    }, 5000);
  } finally {
    isSubmitting.value = false;
  }
}
</script>

<template>
  <div class="form-container">
    <h2>Create Access Request</h2>

    <form @submit.prevent="handleSubmit" class="access-request-form">
      <div class="form-group">
        <label for="requester">Requester</label>
        <input
          id="requester"
          v-model="form.requester"
          type="text"
          placeholder="Enter your name"
          :class="{ 'error': errors.requester }"
        />
        <span v-if="errors.requester" class="error-message">{{ errors.requester }}</span>
      </div>

      <div class="form-group">
        <label for="systemResource">System/Resource</label>
        <input
          id="systemResource"
          v-model="form.systemResource"
          type="text"
          placeholder="Enter system or resource name"
          :class="{ 'error': errors.systemResource }"
        />
        <span v-if="errors.systemResource" class="error-message">{{ errors.systemResource }}</span>
      </div>

      <div class="form-group">
        <label for="accessLevel">Access Level</label>
        <select id="accessLevel" v-model="form.accessLevel">
          <option value="READ">Read</option>
          <option value="WRITE">Write</option>
          <option value="ADMIN">Admin</option>
        </select>
      </div>

      <div class="form-group">
        <label for="justification">Justification</label>
        <textarea
          id="justification"
          v-model="form.justification"
          rows="4"
          placeholder="Explain why you need access"
          :class="{ 'error': errors.justification }"
        ></textarea>
        <span v-if="errors.justification" class="error-message">{{ errors.justification }}</span>
      </div>

      <button type="submit" class="submit-btn" :disabled="isSubmitting">
        {{ isSubmitting ? 'Submitting...' : 'Submit Request' }}
      </button>
    </form>

    <Transition name="fade">
      <div v-if="showToast" :class="['toast', toastType]">
        {{ toastMessage }}
      </div>
    </Transition>
  </div>
</template>

<style scoped>
.form-container {
  max-width: 600px;
  margin: 0 auto;
  padding: 2rem;
}

h2 {
  margin-bottom: 1.5rem;
  color: #333;
}

.access-request-form {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

label {
  font-weight: 600;
  color: #444;
}

input,
select,
textarea {
  padding: 0.75rem;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 1rem;
  font-family: inherit;
  transition: border-color 0.2s;
}

input:focus,
select:focus,
textarea:focus {
  outline: none;
  border-color: #42b983;
}

input.error,
textarea.error {
  border-color: #e74c3c;
}

.error-message {
  color: #e74c3c;
  font-size: 0.875rem;
}

.submit-btn {
  padding: 0.75rem 1.5rem;
  background-color: #42b983;
  color: white;
  border: none;
  border-radius: 4px;
  font-size: 1rem;
  font-weight: 600;
  cursor: pointer;
  transition: background-color 0.2s;
}

.submit-btn:hover:not(:disabled) {
  background-color: #3aa876;
}

.submit-btn:disabled {
  background-color: #a0dcc0;
  cursor: not-allowed;
}

.toast {
  position: fixed;
  bottom: 2rem;
  left: 50%;
  transform: translateX(-50%);
  padding: 1rem 2rem;
  border-radius: 4px;
  color: white;
  font-weight: 500;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.toast.success {
  background-color: #42b983;
}

.toast.error {
  background-color: #e74c3c;
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
