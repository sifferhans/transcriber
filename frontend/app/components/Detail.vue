<script setup lang="ts">
defineProps<{ label: string }>();
</script>

<template>
  <div>
    <dt class="text-caption-2 text-text-hint uppercase tracking-wider mb-1">
      {{ label }}
    </dt>
    <dd class="text-body-3 text-text-default font-mono break-all">
      <slot />
    </dd>
  </div>
</template>
