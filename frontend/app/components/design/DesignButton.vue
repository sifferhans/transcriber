<script setup lang="ts">
// DesignButton is the primary button primitive of the design system.
// Variants combine semantic color tokens (brand/danger/etc.) with size scales.
// Consumers should reach for this rather than hand-rolling Tailwind classes.

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
    "bg-brand-600 hover:bg-brand-700 text-white shadow-sm focus-visible:ring-brand-500/40",
  secondary:
    "bg-surface hover:bg-canvas text-fg border border-line hover:border-line-strong focus-visible:ring-brand-500/40",
  ghost:
    "bg-transparent hover:bg-canvas text-fg-muted hover:text-fg focus-visible:ring-brand-500/40",
  danger:
    "bg-danger-600 hover:bg-danger-700 text-white shadow-sm focus-visible:ring-danger-500/40",
  link: "bg-transparent text-brand-600 hover:text-brand-700 hover:underline focus-visible:ring-brand-500/40 px-0!",
};

const sizes: Record<Size, string> = {
  sm: "px-2.5 py-1 text-xs gap-1 rounded-sm",
  md: "px-4 py-2 text-sm gap-2 rounded-md",
  lg: "px-5 py-2.5 text-base gap-2 rounded-md",
  icon: "p-1.5 text-sm rounded-md",
};
</script>

<template>
  <button
    :type="type"
    :disabled="disabled || loading"
    :class="[
      'inline-flex items-center justify-center font-medium font-display transition-colors',
      'disabled:opacity-50 disabled:cursor-not-allowed',
      'focus:outline-none focus-visible:ring-2',
      variants[variant],
      sizes[size],
      block && 'w-full',
    ]"
  >
    <Icon
      v-if="loading"
      name="tabler:loader-2"
      class="animate-spin"
      :size="size === 'sm' ? '14' : '16'"
    />
    <slot />
  </button>
</template>
