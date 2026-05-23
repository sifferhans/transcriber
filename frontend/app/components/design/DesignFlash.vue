<script setup lang="ts">
// Mirrors Phoenix's CoreComponents.flash/1 — a fixed-position toast.
// In Phoenix the flash is server-driven; here it's a controlled component:
// parent v-model:open="..." or use the `dismiss` event to react.

type Kind = "info" | "error" | "success" | "warning";

withDefaults(
  defineProps<{
    kind?: Kind;
    title?: string;
  }>(),
  { kind: "info" },
);

const open = defineModel<boolean>("open", { default: true });

const styles: Record<Kind, { bg: string; icon: string }> = {
  info: { bg: "bg-semantic-info", icon: "tabler:info-circle" },
  error: { bg: "bg-semantic-error", icon: "tabler:alert-circle" },
  success: { bg: "bg-semantic-success", icon: "tabler:circle-check" },
  warning: { bg: "bg-semantic-warning", icon: "tabler:alert-triangle" },
};
</script>

<template>
  <div
    v-if="open"
    role="alert"
    class="fixed top-4 right-4 z-50"
  >
    <div
      :class="[
        'flex items-start gap-3 p-4 rounded-md shadow-floating w-80 sm:w-96 max-w-80 sm:max-w-96 text-wrap text-body-3 text-text-light-default',
        styles[kind].bg,
      ]"
    >
      <Icon :name="styles[kind].icon" size="20" class="shrink-0" />
      <div class="flex-1">
        <p v-if="title" class="text-title-3 font-semibold">{{ title }}</p>
        <p><slot /></p>
      </div>
      <button
        type="button"
        class="group self-start cursor-pointer"
        :aria-label="'close'"
        @click="open = false"
      >
        <Icon
          name="tabler:x"
          size="20"
          class="opacity-60 group-hover:opacity-100"
        />
      </button>
    </div>
  </div>
</template>
