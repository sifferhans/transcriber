<script setup lang="ts">
// Mirrors Phoenix's CoreComponents.stat_card/1.
// Caption above a large value (slot).
defineProps<{ label: string }>();
</script>

<template>
  <div class="rounded-xl border border-border-1 bg-surface-default p-4">
    <div class="text-caption-1 text-text-muted">{{ label }}</div>
    <div class="mt-1 text-title-1 font-semibold text-text-default">
      <slot />
    </div>
  </div>
</template>
