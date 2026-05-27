<script setup lang="ts">
const { theme, toggle } = useTheme();
</script>

<template>
    <div class="min-h-screen bg-surface-indent text-text-default">
        <header class="bg-surface-default border-b border-border-1">
            <div class="mx-auto max-w-6xl px-6 py-4 flex items-center gap-3">
                <NuxtLink to="/" class="text-title-2 items-center flex gap-2">
                    <Icon
                        name="tabler:wave-saw-tool"
                        class="text-primary-default"
                        size="20"
                    />
                    Transcriber
                </NuxtLink>
                <nav class="ml-auto flex items-center gap-4 text-body-3">
                    <DesignButton
                        :aria-label="`Switch to ${theme === 'dark' ? 'light' : 'dark'} theme`"
                        :title="`Switch to ${theme === 'dark' ? 'light' : 'dark'} theme`"
                        :icon="theme === 'dark' ? 'tabler:sun' : 'tabler:moon'"
                        variant="tertiary"
                        @click="toggle"
                    />
                </nav>
            </div>
        </header>

        <main class="mx-auto max-w-6xl px-6 py-8">
            <NuxtRouteAnnouncer />
            <NuxtPage />
        </main>
    </div>
</template>
