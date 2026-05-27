<script setup lang="ts">
// Empty-state wrapper around DesignViewState with friendly defaults.
withDefaults(
    defineProps<{
        icon?: string;
        title?: string;
        description?: string;
    }>(),
    { icon: "tabler:inbox" },
);
</script>

<template>
    <DesignViewState
        :icon="icon"
        :title="title ?? 'Nothing here yet'"
        :description="description"
    >
        <template v-if="$slots.action" #action><slot name="action" /></template>
    </DesignViewState>
</template>
