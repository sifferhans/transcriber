<script setup lang="ts">
// Mirrors Phoenix's CoreComponents.avatar/1.
// Shows the image when `src` is given; otherwise the first letters of `name`.

type Size = "small" | "medium" | "large";

const props = withDefaults(
  defineProps<{
    src?: string;
    name?: string;
    size?: Size;
  }>(),
  { size: "medium" },
);

const sizes: Record<Size, string> = {
  small: "size-6 text-caption-2",
  medium: "size-8 text-caption-1",
  large: "size-10 text-title-3",
};

function initials(name?: string): string {
  if (!name) return "?";
  const letters = name
    .split(/\s+/)
    .filter(Boolean)
    .slice(0, 2)
    .map((w) => w[0])
    .join("")
    .toUpperCase();
  return letters || "?";
}

const computed_initials = computed(() => initials(props.name));
</script>

<template>
  <span
    :class="[
      'relative inline-flex shrink-0 items-center justify-center overflow-hidden rounded-full',
      sizes[size],
    ]"
  >
    <img
      v-if="src"
      :src="src"
      :alt="name ?? ''"
      class="size-full object-cover"
    />
    <span
      v-else
      class="bg-primary-default text-on-primary flex size-full items-center justify-center font-semibold"
    >
      {{ computed_initials }}
    </span>
  </span>
</template>
