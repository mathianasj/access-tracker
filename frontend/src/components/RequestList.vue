<script setup lang="ts">
import { ref, watch, computed } from 'vue';
import { useRouter, useRoute } from 'vue-router';
import { useQuery, useMutation } from '@vue/apollo-composable';
import { gql } from '@apollo/client/core';
import StatusBadge from './StatusBadge.vue';

const GET_ACCESS_REQUESTS = gql`
  query GetAccessRequests {
    accessRequests {
      id
      requestId
      requester
      systemResource
      accessLevel
      justification
      status
      createdAt
      updatedAt
    }
  }
`;

const UPDATE_ACCESS_REQUEST_STATUS = gql`
  mutation UpdateAccessRequestStatus($input: UpdateAccessRequestStatusInput!) {
    updateAccessRequestStatus(input: $input) {
      id
      status
      updatedAt
    }
  }
`;

const router = useRouter();
const route = useRoute();

const { result, loading, error, refetch } = useQuery(GET_ACCESS_REQUESTS);
const { mutate: updateStatus } = useMutation(UPDATE_ACCESS_REQUEST_STATUS);

const accessRequests = ref<any[]>([]);
const statusFilter = ref<string>('');
const requesterSearch = ref<string>('');

const pendingUpdate = ref<string | null>(null);
const showConfirmModal = ref(false);
const confirmTarget = ref<{ id: string; currentStatus: string; newStatus: string } | null>(null);

watch(result, (data) => {
  if (data?.accessRequests) {
    accessRequests.value = data.accessRequests;
  }
});

watch([statusFilter, requesterSearch], () => {
  updateQueryParams();
}, { immediate: true });

function updateQueryParams() {
  const query: Record<string, string> = {};
  if (statusFilter.value) {
    query.status = statusFilter.value;
  }
  if (requesterSearch.value) {
    query.requester = requesterSearch.value;
  }
  router.replace({ query });
}

function initFiltersFromQuery() {
  const queryStatus = route.query.status as string;
  const queryRequester = route.query.requester as string;

  if (queryStatus) {
    statusFilter.value = queryStatus;
  }
  if (queryRequester) {
    requesterSearch.value = queryRequester;
  }
}

initFiltersFromQuery();

const filteredRequests = computed(() => {
  let filtered = accessRequests.value;

  if (statusFilter.value) {
    filtered = filtered.filter(r => r.status === statusFilter.value);
  }

  if (requesterSearch.value) {
    const search = requesterSearch.value.toLowerCase();
    filtered = filtered.filter(r =>
      r.requester.toLowerCase().includes(search)
    );
  }

  return filtered;
});

function clearFilters() {
  statusFilter.value = '';
  requesterSearch.value = '';
}

function hasActiveFilters(): boolean {
  return statusFilter.value !== '' || requesterSearch.value !== '';
}

function requestStatusChange(request: any, newStatus: string) {
  if (newStatus === request.status) {
    return;
  }
  confirmTarget.value = {
    id: request.id,
    currentStatus: request.status,
    newStatus
  };
  showConfirmModal.value = true;
}

async function confirmStatusChange() {
  if (!confirmTarget.value) return;

  const { id, newStatus } = confirmTarget.value;
  pendingUpdate.value = id;

  const originalRequests = [...accessRequests.value];

  accessRequests.value = accessRequests.value.map(r =>
    r.id === id ? { ...r, status: newStatus } : r
  );

  showConfirmModal.value = false;
  confirmTarget.value = null;

  try {
    await updateStatus({
      input: {
        id,
        status: newStatus
      }
    });
    pendingUpdate.value = null;
    await refetch();
  } catch (err) {
    accessRequests.value = originalRequests;
    pendingUpdate.value = null;
    console.error('Failed to update status:', err);
    alert('Failed to update status. Please try again.');
  }
}

function cancelStatusChange() {
  showConfirmModal.value = false;
  confirmTarget.value = null;
}

function formatDate(dateString: string): string {
  const date = new Date(dateString);
  return date.toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric'
  });
}

function getAccessLevelBadgeClass(level: string): string {
  switch (level) {
    case 'ADMIN':
      return 'level-admin';
    case 'WRITE':
      return 'level-write';
    default:
      return 'level-read';
  }
}
</script>

<template>
  <div class="request-list-container">
    <div class="header">
      <h2>Access Requests</h2>
      <router-link to="/request/new" class="new-request-btn">
        + New Request
      </router-link>
    </div>

    <div class="filters">
      <div class="filter-group">
        <label for="statusFilter">Status</label>
        <select id="statusFilter" v-model="statusFilter">
          <option value="">All Statuses</option>
          <option value="PENDING">Pending</option>
          <option value="APPROVED">Approved</option>
          <option value="DENIED">Denied</option>
        </select>
      </div>

      <div class="filter-group">
        <label for="requesterSearch">Requester</label>
        <input
          id="requesterSearch"
          v-model="requesterSearch"
          type="text"
          placeholder="Search by requester..."
        />
      </div>

      <button
        v-if="hasActiveFilters()"
        @click="clearFilters"
        class="clear-filters-btn"
      >
        Clear Filters
      </button>
    </div>

    <div v-if="loading" class="loading">Loading...</div>
    <div v-else-if="error" class="error">Error loading requests: {{ error.message }}</div>
    <template v-else>
      <div class="results-count" v-if="hasActiveFilters()">
        Showing {{ filteredRequests.length }} of {{ accessRequests.length }} requests
      </div>

      <table class="request-table" v-if="filteredRequests.length > 0">
        <thead>
          <tr>
            <th>Requester</th>
            <th>System/Resource</th>
            <th>Access Level</th>
            <th>Status</th>
            <th>Date</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="request in filteredRequests" :key="request.id" :class="{ 'updating': pendingUpdate === request.id }">
            <td>{{ request.requester }}</td>
            <td>{{ request.systemResource }}</td>
            <td>
              <span :class="['level-badge', getAccessLevelBadgeClass(request.accessLevel)]">
                {{ request.accessLevel }}
              </span>
            </td>
            <td>
              <div class="status-cell">
                <StatusBadge :status="request.status" />
                <select
                  :value="request.status"
                  @change="(e) => requestStatusChange(request, (e.target as HTMLSelectElement).value)"
                  :disabled="pendingUpdate === request.id"
                  class="status-select"
                >
                  <option value="PENDING">Pending</option>
                  <option value="APPROVED">Approved</option>
                  <option value="DENIED">Denied</option>
                </select>
              </div>
            </td>
            <td>{{ formatDate(request.createdAt) }}</td>
          </tr>
        </tbody>
      </table>
      <div v-else class="empty-state">
        <p v-if="hasActiveFilters()">No requests match your filters.</p>
        <p v-else>No access requests found.</p>
        <router-link v-if="!hasActiveFilters()" to="/request/new" class="empty-state-link">
          Create your first request
        </router-link>
        <button v-else @click="clearFilters" class="empty-state-link">
          Clear filters
        </button>
      </div>
    </template>

    <Teleport to="body">
      <div v-if="showConfirmModal" class="modal-overlay" @click.self="cancelStatusChange">
        <div class="modal">
          <h3>Confirm Status Change</h3>
          <p v-if="confirmTarget">
            Are you sure you want to change the status from
            <strong>{{ confirmTarget.currentStatus }}</strong> to
            <strong>{{ confirmTarget.newStatus }}</strong>?
          </p>
          <div class="modal-actions">
            <button @click="cancelStatusChange" class="btn-cancel">Cancel</button>
            <button @click="confirmStatusChange" class="btn-confirm">Confirm</button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<style scoped>
.request-list-container {
  padding: 2rem;
  max-width: 1200px;
  margin: 0 auto;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1.5rem;
}

h2 {
  margin: 0;
  color: #333;
}

.new-request-btn {
  background-color: #42b983;
  color: white;
  padding: 0.5rem 1rem;
  border-radius: 4px;
  text-decoration: none;
  font-weight: 600;
  transition: background-color 0.2s;
}

.new-request-btn:hover {
  background-color: #3aa876;
}

.filters {
  display: flex;
  gap: 1rem;
  margin-bottom: 1rem;
  align-items: flex-end;
  flex-wrap: wrap;
}

.filter-group {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.filter-group label {
  font-size: 0.75rem;
  font-weight: 600;
  color: #666;
  text-transform: uppercase;
}

.filter-group select,
.filter-group input {
  padding: 0.5rem;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 0.875rem;
  min-width: 150px;
}

.filter-group input {
  min-width: 200px;
}

.clear-filters-btn {
  padding: 0.5rem 1rem;
  background-color: #6b7280;
  color: white;
  border: none;
  border-radius: 4px;
  font-size: 0.875rem;
  cursor: pointer;
  transition: background-color 0.2s;
}

.clear-filters-btn:hover {
  background-color: #4b5563;
}

.results-count {
  font-size: 0.875rem;
  color: #666;
  margin-bottom: 0.5rem;
}

.loading,
.error {
  text-align: center;
  padding: 2rem;
  color: #666;
}

.error {
  color: #e74c3c;
}

.request-table {
  width: 100%;
  border-collapse: collapse;
  background: white;
  border-radius: 8px;
  overflow: hidden;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.request-table th,
.request-table td {
  padding: 1rem;
  text-align: left;
  border-bottom: 1px solid #eee;
}

.request-table th {
  background-color: #f8f9fa;
  font-weight: 600;
  color: #555;
  font-size: 0.875rem;
  text-transform: uppercase;
}

.request-table tr:last-child td {
  border-bottom: none;
}

.request-table tr:hover td {
  background-color: #f8f9fa;
}

.request-table tr.updating td {
  opacity: 0.6;
}

.level-badge {
  display: inline-block;
  padding: 0.25rem 0.5rem;
  border-radius: 4px;
  font-size: 0.75rem;
  font-weight: 600;
}

.level-read {
  background-color: #e0f2fe;
  color: #0369a1;
}

.level-write {
  background-color: #fef9c3;
  color: #854d0e;
}

.level-admin {
  background-color: #f3e8ff;
  color: #6b21a8;
}

.status-cell {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.status-select {
  padding: 0.25rem 0.5rem;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 0.75rem;
  cursor: pointer;
}

.status-select:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.empty-state {
  text-align: center;
  padding: 4rem 2rem;
  background: white;
  border-radius: 8px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.empty-state p {
  color: #666;
  margin-bottom: 1rem;
}

.empty-state-link {
  color: #42b983;
  font-weight: 600;
  background: none;
  border: none;
  cursor: pointer;
  font-size: 1rem;
  text-decoration: none;
}

.empty-state-link:hover {
  text-decoration: underline;
}

.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal {
  background: white;
  padding: 2rem;
  border-radius: 8px;
  max-width: 400px;
  width: 90%;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.15);
}

.modal h3 {
  margin: 0 0 1rem 0;
  color: #333;
}

.modal p {
  color: #666;
  margin-bottom: 1.5rem;
}

.modal strong {
  color: #333;
}

.modal-actions {
  display: flex;
  gap: 1rem;
  justify-content: flex-end;
}

.btn-cancel {
  padding: 0.5rem 1rem;
  background-color: #e5e7eb;
  color: #374151;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-weight: 600;
}

.btn-cancel:hover {
  background-color: #d1d5db;
}

.btn-confirm {
  padding: 0.5rem 1rem;
  background-color: #42b983;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-weight: 600;
}

.btn-confirm:hover {
  background-color: #3aa876;
}
</style>
