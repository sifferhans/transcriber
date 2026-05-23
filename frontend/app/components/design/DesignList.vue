<script setup lang="ts">
// Mirrors Phoenix's CoreComponents.list/1.
// Renders `items` as title/value rows separated by border-1. Items are
// declared in TS rather than via slot iteration since v-for inside a slot
// is more awkward than a typed array.

interface Item {
  title: string;
  value: string | number;
}

defineProps<{ items: Item[] }>();
</script>

<template>
  <ul
    class="divide-y divide-border-1 rounded-md border border-border-1 bg-surface-default"
  >
    <li
      v-for="item in items"
      :key="item.title"
      class="flex flex-col gap-1 px-4 py-3"
    >
      <div class="text-title-3 font-semibold text-text-default">
        {{ item.title }}
      </div>
      <div class="text-body-3 text-text-muted">{{ item.value }}</div>
    </li>
  </ul>
</template>
