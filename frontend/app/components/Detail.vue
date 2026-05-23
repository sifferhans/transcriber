<script setup lang="ts">
defineProps<{ label: string }>();
</script>

<template>
  <div>
    <dt
      class="text-xs font-medium text-slate-500 uppercase tracking-wider mb-1"
    >
      {{ label }}
    </dt>
    <dd class="text-sm text-slate-900 font-mono break-all">
      <slot />
    </dd>
  </div>
</template>
