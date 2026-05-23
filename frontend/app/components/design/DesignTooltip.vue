<script setup lang="ts">
// Mirrors Phoenix's CoreComponents.tooltip/1.
// CSS-only — shows on hover or focus-within, no JS state. Use `text` for
// plain content or the `content` slot for richer markup.

type Placement = "top" | "bottom" | "left" | "right";

const props = withDefaults(
  defineProps<{
    text?: string;
    placement?: Placement;
    gap?: number;
  }>(),
  { placement: "top", gap: 8 },
);

const placements: Record<Placement, string> = {
  top: "bottom-full left-1/2 -translate-x-1/2 mb-[var(--tooltip-gap)] origin-bottom",
  bottom: "top-full left-1/2 -translate-x-1/2 mt-[var(--tooltip-gap)] origin-top",
  left: "right-full top-1/2 -translate-y-1/2 mr-[var(--tooltip-gap)] origin-right",
  right: "left-full top-1/2 -translate-y-1/2 ml-[var(--tooltip-gap)] origin-left",
};
</script>

<template>
  <span class="relative inline-flex group">
    <slot />
    <span
      role="tooltip"
      :style="{ '--tooltip-gap': `${props.gap}px` }"
      :class="[
        'pointer-events-none absolute z-50 max-w-xs whitespace-normal rounded-lg gradient-border bg-surface-raise px-2.5 py-1.5',
        'text-caption-1 text-text-default shadow-floating',
        'opacity-0 scale-95 transition-[opacity,transform] duration-200 ease-out-expo',
        'group-hover:opacity-100 group-hover:scale-100',
        'group-focus-within:opacity-100 group-focus-within:scale-100',
        placements[placement],
      ]"
    >
      <slot name="content">{{ text }}</slot>
    </span>
  </span>
</template>
