<template>
  <div class="min-h-screen bg-canvas text-fg font-display">
    <header class="bg-surface border-b border-line">
      <div class="mx-auto max-w-6xl px-6 py-4 flex items-center gap-3">
        <Icon name="tabler:wave-saw-tool" class="text-brand-600" size="24" />
        <NuxtLink to="/" class="text-lg font-semibold tracking-tight">
          Transcriber
        </NuxtLink>
        <nav class="ml-auto flex items-center gap-4 text-sm">
          <NuxtLink
            to="/"
            class="text-fg-muted hover:text-fg"
            active-class="text-fg font-medium"
          >
            Jobs
          </NuxtLink>
        </nav>
      </div>
    </header>

    <main class="mx-auto max-w-6xl px-6 py-8">
      <NuxtRouteAnnouncer />
      <NuxtPage />
    </main>
  </div>
</template>
