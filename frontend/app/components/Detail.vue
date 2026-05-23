<script setup lang="ts">
defineProps<{ label: string }>();
</script>

<template>
  <div>
    <dt class="text-xs font-medium text-muted uppercase tracking-wider mb-1">
      {{ label }}
    </dt>
    <dd class="text-sm text-fg font-mono break-all">
      <slot />
    </dd>
  </div>
</template>
