<script setup lang="ts">
const props = defineProps<{ value: number }>();
const pct = computed(() => Math.max(0, Math.min(100, props.value)));
</script>

<template>
  <div class="h-2 bg-surface-indent rounded-full overflow-hidden ring-1 ring-border-1">
    <div
      class="h-full bg-primary-default ease-out-expo transition-all duration-300"
      :style="{ width: pct + '%' }"
    />
  </div>
</template>
