<script setup lang="ts">
import type { JobStatus } from "~/types/job";

defineProps<{ status: JobStatus }>();

type Variant = "neutral" | "success" | "warning" | "info" | "error";

const variants: Record<JobStatus, Variant> = {
    QUEUED: "neutral",
    RUNNING: "info",
    COMPLETED: "success",
    FAILED: "error",
    CANCELED: "warning",
};
</script>

<template>
    <DesignBadge :variant="variants[status]" :label="status.toLowerCase()" />
</template>
