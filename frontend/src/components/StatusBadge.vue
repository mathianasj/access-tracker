<script setup lang="ts">
defineProps<{
  status: 'PENDING' | 'APPROVED' | 'DENIED';
}>();
</script>

<template>
  <span :class="['status-badge', status.toLowerCase()]">
    {{ status }}
  </span>
</template>

<style scoped>
.status-badge {
  display: inline-block;
  padding: 0.25rem 0.75rem;
  border-radius: 9999px;
  font-size: 0.75rem;
  font-weight: 600;
  text-transform: uppercase;
}

.status-badge.pending {
  background-color: #fef3c7;
  color: #92400e;
}

.status-badge.approved {
  background-color: #d1fae5;
  color: #065f46;
}

.status-badge.denied {
  background-color: #fee2e2;
  color: #991b1b;
}
</style>
