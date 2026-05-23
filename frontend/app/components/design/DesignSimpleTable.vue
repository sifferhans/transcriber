<script setup lang="ts">
// Mirrors Phoenix's CoreComponents.simple_table/1.
// Card-bordered table; consumer provides <tr>/<td> rows via the default slot.
// `empty` renders when the slot is empty.

defineProps<{
  columns: string[];
  empty?: string;
}>();
</script>

<template>
  <div class="overflow-hidden rounded-xl border border-border-1">
    <table class="w-full">
      <thead>
        <tr class="bg-surface-indent text-left">
          <th
            v-for="col in columns"
            :key="col"
            class="px-4 py-2.5 text-caption-1 font-medium text-text-muted"
          >
            {{ col }}
          </th>
        </tr>
      </thead>
      <tbody>
        <slot />
      </tbody>
    </table>
    <div
      v-if="empty && !$slots.default"
      class="px-4 py-12 text-center text-body-2 text-text-hint"
    >
      {{ empty }}
    </div>
  </div>
</template>
