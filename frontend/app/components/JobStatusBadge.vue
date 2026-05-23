<script setup lang="ts">
import type { JobStatus } from "~/types/job";

defineProps<{ status: JobStatus }>();

const styles: Record<JobStatus, string> = {
  PENDING: "bg-slate-100 text-slate-700 ring-slate-200",
  RUNNING: "bg-indigo-100 text-indigo-700 ring-indigo-200",
  COMPLETED: "bg-emerald-100 text-emerald-700 ring-emerald-200",
  FAILED: "bg-red-100 text-red-700 ring-red-200",
  CANCELED: "bg-amber-100 text-amber-700 ring-amber-200",
};
</script>

<template>
  <span
    :class="[
      'text-xs font-medium px-2 py-0.5 rounded-full ring-1',
      styles[status],
    ]"
  >
    {{ status.toLowerCase() }}
  </span>
</template>
