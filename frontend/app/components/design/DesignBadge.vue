<script setup lang="ts">
// DesignBadge is the canonical pill/tag of the design system. Tones map to
// the same semantic color tokens as DesignButton, so the visual language
// stays consistent across the app.

type Tone = "neutral" | "brand" | "success" | "warning" | "danger" | "info";
type Size = "sm" | "md";

withDefaults(
  defineProps<{
    tone?: Tone;
    size?: Size;
    dot?: boolean;
  }>(),
  {
    tone: "neutral",
    size: "sm",
    dot: false,
  },
);

const tones: Record<
  Tone,
  { bg: string; text: string; ring: string; dot: string }
> = {
  neutral: {
    bg: "bg-canvas",
    text: "text-fg-muted",
    ring: "ring-line",
    dot: "bg-muted",
  },
  brand: {
    bg: "bg-brand-100",
    text: "text-brand-700",
    ring: "ring-brand-200",
    dot: "bg-brand-500",
  },
  success: {
    bg: "bg-success-100",
    text: "text-success-700",
    ring: "ring-success-200",
    dot: "bg-success-500",
  },
  warning: {
    bg: "bg-warning-100",
    text: "text-warning-700",
    ring: "ring-warning-200",
    dot: "bg-warning-500",
  },
  danger: {
    bg: "bg-danger-100",
    text: "text-danger-700",
    ring: "ring-danger-200",
    dot: "bg-danger-500",
  },
  info: {
    bg: "bg-info-100",
    text: "text-info-700",
    ring: "ring-info-200",
    dot: "bg-info-500",
  },
};

const sizes: Record<Size, string> = {
  sm: "text-xs px-2 py-0.5 gap-1",
  md: "text-sm px-2.5 py-1 gap-1.5",
};
</script>

<template>
  <span
    :class="[
      'inline-flex items-center font-medium font-display rounded-full ring-1',
      tones[tone].bg,
      tones[tone].text,
      tones[tone].ring,
      sizes[size],
    ]"
  >
    <span
      v-if="dot"
      :class="['inline-block w-1.5 h-1.5 rounded-full', tones[tone].dot]"
      aria-hidden="true"
    />
    <slot />
  </span>
</template>
