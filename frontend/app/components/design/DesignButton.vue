<script setup lang="ts">
// DesignButton is the primary button primitive of the design system.
// Tokens (primary-default / on-primary / surface-* / text-*) ensure variants
// stay coherent across light and dark themes.
//
// Focus styling is handled globally by main.css (the :where() rule); we do
// not set per-component focus rings.

type Variant = "primary" | "secondary" | "ghost" | "danger" | "link";
type Size = "sm" | "md" | "lg" | "icon";

withDefaults(
  defineProps<{
    variant?: Variant;
    size?: Size;
    type?: "button" | "submit" | "reset";
    disabled?: boolean;
    loading?: boolean;
    block?: boolean;
  }>(),
  {
    variant: "primary",
    size: "md",
    type: "button",
    disabled: false,
    loading: false,
    block: false,
  },
);

const variants: Record<Variant, string> = {
  primary:
    "bg-primary-default hover:bg-primary-contrast text-on-primary shadow-resting",
  secondary:
    "bg-surface-raise hover:bg-surface-indent text-text-default border border-border-1",
  ghost:
    "bg-transparent hover:bg-surface-indent text-text-muted hover:text-text-default",
  danger:
    "bg-semantic-error hover:bg-semantic-error/90 text-white shadow-resting",
  link:
    "bg-transparent text-primary-default hover:text-primary-contrast hover:underline px-0!",
};

const sizes: Record<Size, string> = {
  sm: "px-2.5 py-1 text-caption-1 gap-1 rounded-sm",
  md: "px-4 py-2 text-title-3 gap-2 rounded-md",
  lg: "px-5 py-2.5 text-title-2 gap-2 rounded-md",
  icon: "p-1.5 text-title-3 rounded-md",
};
</script>

<template>
  <button
    :type="type"
    :disabled="disabled || loading"
    :class="[
      'inline-flex items-center justify-center transition-colors',
      'disabled:opacity-50 disabled:cursor-not-allowed',
      variants[variant],
      sizes[size],
      block && 'w-full',
    ]"
  >
    <!-- SVG spinner uses the spinner-rotate / spinner-segment utilities. -->
    <svg
      v-if="loading"
      class="spinner-rotate"
      :width="size === 'sm' ? 12 : 14"
      :height="size === 'sm' ? 12 : 14"
      viewBox="0 0 24 24"
      aria-hidden="true"
    >
      <circle
        class="spinner-segment"
        cx="12"
        cy="12"
        r="10"
        fill="none"
        stroke="currentColor"
        stroke-width="2.5"
        stroke-linecap="round"
      />
    </svg>
    <slot />
  </button>
</template>
