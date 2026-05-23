<script setup lang="ts">
// Error-state wrapper around DesignViewState. Icon tinted with semantic-error.
withDefaults(
  defineProps<{
    icon?: string;
    title?: string;
    description?: string;
  }>(),
  { icon: "tabler:alert-circle" },
);
</script>

<template>
  <DesignViewState
    :icon="icon"
    icon-class="text-semantic-error!"
    :title="title ?? 'Something went wrong'"
    :description="description"
  >
    <template v-if="$slots.action" #action><slot name="action" /></template>
  </DesignViewState>
</template>
