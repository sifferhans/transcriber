<script setup lang="ts">
// Mirrors Phoenix's CoreComponents.header/1.
// `title` and `subtitle` slots; `actions` slot pinned to the right.
</script>

<template>
  <header
    :class="[$slots.actions && 'flex items-center justify-between gap-6', 'pb-4']"
  >
    <div>
      <h1 class="text-heading-3 text-text-default"><slot /></h1>
      <p v-if="$slots.subtitle" class="text-body-3 text-text-muted">
        <slot name="subtitle" />
      </p>
    </div>
    <div v-if="$slots.actions" class="flex-none">
      <slot name="actions" />
    </div>
  </header>
</template>
