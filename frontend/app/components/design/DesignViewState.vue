<script setup lang="ts">
withDefaults(
    defineProps<{
        icon?: string;
        iconClass?: string;
        title?: string;
        description?: string;
    }>(),
    {},
);
</script>

<template>
    <div
        class="flex size-full flex-col items-center justify-center px-4 text-center"
    >
        <div v-if="icon" class="mb-6">
            <Icon
                :name="icon"
                :class="['text-text-hint', iconClass]"
                size="32"
            />
        </div>
        <p v-if="title" class="text-heading-3 text-text-default">{{ title }}</p>
        <p v-if="description" class="text-body-2 text-text-hint mt-1">
            {{ description }}
        </p>
        <div
            v-if="$slots.action"
            class="mt-6 flex flex-wrap items-center justify-center gap-2"
        >
            <slot name="action" />
        </div>
    </div>
</template>
