<script setup lang="ts">
// Mirrors Phoenix's CoreComponents.loading_state/1.
// Indeterminate spinner using the spinner-rotate / spinner-segment utilities.

withDefaults(defineProps<{ size?: number }>(), { size: 40 });
</script>

<template>
    <div class="flex size-full items-center justify-center px-4">
        <svg
            class="text-primary-contrast spinner-rotate"
            :width="size"
            :height="size"
            viewBox="0 0 24 24"
            fill="none"
            aria-hidden="true"
        >
            <circle
                cx="12"
                cy="12"
                r="10"
                stroke="currentColor"
                stroke-width="2.5"
                class="opacity-0"
            />
            <circle
                class="spinner-segment"
                cx="12"
                cy="12"
                r="10"
                stroke="currentColor"
                stroke-width="2.5"
                stroke-linecap="round"
            />
        </svg>
    </div>
</template>
