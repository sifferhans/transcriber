<script setup lang="ts">
type Variant = "neutral" | "success" | "warning" | "info" | "error";

withDefaults(
    defineProps<{
        variant?: Variant;
        icon?: string;
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
    <div
        :class="[
            'flex items-center gap-3 rounded-xl px-4 py-3 text-body-3',
            variants[variant],
        ]"
    >
        <Icon v-if="icon" :name="icon" size="20" class="shrink-0" />
        <div class="flex-1">
            <slot />
        </div>
    </div>
</template>
