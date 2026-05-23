<script setup lang="ts">
// Mirrors Phoenix's CoreComponents.switch/1.
// Visual track + thumb react to the underlying checkbox via group-has-checked.

withDefaults(
  defineProps<{
    label?: string;
    disabled?: boolean;
  }>(),
  { disabled: false },
);

const model = defineModel<boolean>({ default: false });
</script>

<template>
  <label
    class="group inline-flex cursor-pointer items-center gap-2.5 has-disabled:cursor-not-allowed has-disabled:opacity-50"
  >
    <input
      v-model="model"
      type="checkbox"
      :disabled="disabled"
      class="sr-only"
    />
    <span
      class="inline-flex h-6 w-10 shrink-0 items-center rounded-full bg-text-default/15 p-0.5 transition-colors duration-200 ease-out-expo group-has-checked:bg-primary-contrast"
    >
      <span
        class="size-5 rounded-full bg-white shadow-resting transition-transform duration-200 ease-out-expo group-has-checked:translate-x-4"
      />
    </span>
    <span v-if="label" class="text-body-3 text-text-default select-none">
      {{ label }}
    </span>
  </label>
</template>
