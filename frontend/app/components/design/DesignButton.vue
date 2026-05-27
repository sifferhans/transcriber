<script setup lang="ts">
type Variant = "primary" | "secondary" | "tertiary" | "danger";
type Size = "small" | "medium" | "large";

withDefaults(
    defineProps<{
        variant?: Variant;
        size?: Size;
        label?: string;
        icon?: string;
        loading?: boolean;
        disabled?: boolean;
        block?: boolean;
        to?: string;
        type?: "button" | "submit" | "reset";
    }>(),
    {
        variant: "primary",
        size: "medium",
        type: "button",
        loading: false,
        disabled: false,
        block: false,
    },
);

const variants: Record<Variant, string> = {
    primary: "bg-primary-default text-on-primary gradient-border-dark",
    secondary: "bg-surface-indent text-text-default",
    tertiary: "bg-transparent text-text-default hover:bg-surface-indent",
    danger: "bg-semantic-error text-text-light-default gradient-border-dark",
};

const sizes: Record<Size, string> = {
    small: "rounded-2xl px-3 py-1.5 text-title-3 gap-1",
    medium: "rounded-3xl px-4 py-2.5 text-title-2 gap-2",
    large: "rounded-4xl px-5 py-3.5 text-title-2 gap-2",
};

const base = [
    "inline-flex items-center justify-center select-none cursor-pointer",
    "transition-transform duration-200 ease-out-expo active:scale-95",
    "disabled:opacity-50 disabled:pointer-events-none disabled:cursor-not-allowed",
];

const nuxtLink = resolveComponent("NuxtLink");
</script>

<template>
    <component
        :is="to ? nuxtLink : 'button'"
        :to="to"
        :type="to ? undefined : type"
        :disabled="to ? undefined : disabled || loading"
        :class="[base, variants[variant], sizes[size], block && 'w-full']"
    >
        <svg
            v-if="loading"
            class="spinner-rotate"
            :width="size === 'small' ? 14 : 16"
            :height="size === 'small' ? 14 : 16"
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
        <Icon
            v-else-if="icon"
            :name="icon"
            :size="size === 'small' ? '14' : '16'"
        />
        <span v-if="$slots.default || label">
            <slot>{{ label }}</slot>
        </span>
    </component>
</template>
