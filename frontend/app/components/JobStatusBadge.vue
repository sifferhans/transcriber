<script setup lang="ts">
import type { JobStatus } from "~/types/job";

defineProps<{ status: JobStatus }>();

type Tone = "neutral" | "brand" | "success" | "warning" | "danger" | "info";

const tones: Record<JobStatus, Tone> = {
  PENDING: "neutral",
  RUNNING: "brand",
  COMPLETED: "success",
  FAILED: "danger",
  CANCELED: "warning",
};
</script>

<template>
  <DesignBadge
    :tone="tones[status]"
    :dot="status === 'PENDING' || status === 'RUNNING'"
  >
    {{ status.toLowerCase() }}
  </DesignBadge>
</template>
