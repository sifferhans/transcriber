<script setup lang="ts">
const props = defineProps<{ value: number }>();
const pct = computed(() => Math.max(0, Math.min(100, props.value)));
</script>

<template>
  <div class="h-2 bg-canvas rounded-full overflow-hidden ring-1 ring-line">
    <div
      class="h-full bg-brand-500 transition-all duration-300 ease-out"
      :style="{ width: pct + '%' }"
    />
  </div>
</template>
