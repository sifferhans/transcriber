<script setup lang="ts">
const props = defineProps<{ value: number }>();
const pct = computed(() => Math.max(0, Math.min(100, props.value)));
</script>

<template>
  <div class="h-2 bg-slate-100 rounded-full overflow-hidden">
    <div
      class="h-full bg-indigo-500 transition-all duration-300 ease-out"
      :style="{ width: pct + '%' }"
    />
  </div>
</template>
