<script setup lang="ts">
type Variant = "neutral" | "success" | "warning" | "info" | "error";

withDefaults(
    defineProps<{
        variant?: Variant;
        label?: string;
        class?: string;
    }>(),
    { variant: "neutral" },
);

const variants: Record<Variant, string> = {
    neutral: "bg-surface-indent text-text-muted",
    success: "bg-semantic-success/15 text-semantic-success",
    warning: "bg-semantic-warning/15 text-semantic-warning",
    info: "bg-semantic-info/15 text-semantic-info",
    error: "bg-semantic-error/15 text-semantic-error",
};
</script>

<template>
    <span
        :class="[
            'inline-flex items-center rounded-lg px-2 py-0.5 text-caption-1',
            variants[variant],
        ]"
    >
        <slot>{{ label }}</slot>
    </span>
</template>
