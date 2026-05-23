<template>
  <div class="min-h-screen bg-surface-indent text-text-default">
    <header class="bg-surface-default border-b border-border-1">
      <div class="mx-auto max-w-6xl px-6 py-4 flex items-center gap-3">
        <Icon
          name="tabler:wave-saw-tool"
          class="text-primary-default"
          size="24"
        />
        <NuxtLink to="/" class="text-title-1 tracking-tight">
          Transcriber
        </NuxtLink>
        <nav class="ml-auto flex items-center gap-4 text-body-3">
          <NuxtLink
            to="/"
            class="text-text-muted hover:text-text-default"
            active-class="text-text-default font-medium"
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
