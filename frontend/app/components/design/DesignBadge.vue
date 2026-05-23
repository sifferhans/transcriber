<script setup lang="ts">
// DesignBadge is the canonical pill/tag of the design system. Tones reuse the
// semantic-* and primary tokens; tinted backgrounds and rings are derived via
// opacity modifiers so a single hue powers fg/bg/ring for each tone.

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

const tones: Record<Tone, { bg: string; text: string; ring: string; dot: string }> = {
  neutral: {
    bg: "bg-surface-indent",
    text: "text-text-muted",
    ring: "ring-border-1",
    dot: "bg-text-hint",
  },
  brand: {
    bg: "bg-primary-default/15",
    text: "text-primary-contrast",
    ring: "ring-primary-default/30",
    dot: "bg-primary-default",
  },
  success: {
    bg: "bg-semantic-success/15",
    text: "text-semantic-success",
    ring: "ring-semantic-success/30",
    dot: "bg-semantic-success",
  },
  warning: {
    bg: "bg-semantic-warning/15",
    text: "text-semantic-warning",
    ring: "ring-semantic-warning/30",
    dot: "bg-semantic-warning",
  },
  danger: {
    bg: "bg-semantic-error/15",
    text: "text-semantic-error",
    ring: "ring-semantic-error/30",
    dot: "bg-semantic-error",
  },
  info: {
    bg: "bg-semantic-info/15",
    text: "text-semantic-info",
    ring: "ring-semantic-info/30",
    dot: "bg-semantic-info",
  },
};

const sizes: Record<Size, string> = {
  sm: "text-caption-2 px-2 py-0.5 gap-1",
  md: "text-caption-1 px-2.5 py-1 gap-1.5",
};
</script>

<template>
  <span
    :class="[
      'inline-flex items-center rounded-full ring-1',
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
