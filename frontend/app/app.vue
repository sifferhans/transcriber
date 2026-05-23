<template>
  <div class="min-h-screen bg-slate-50 text-slate-900">
    <header class="bg-white border-b border-slate-200">
      <div
        class="mx-auto max-w-6xl px-6 py-4 flex items-center gap-3"
      >
        <Icon name="mdi:waveform" class="text-indigo-600" size="24" />
        <NuxtLink to="/" class="text-lg font-semibold tracking-tight">
          Transcriber
        </NuxtLink>
        <nav class="ml-auto flex items-center gap-4 text-sm">
          <NuxtLink
            to="/"
            class="text-slate-600 hover:text-slate-900"
            active-class="text-slate-900 font-medium"
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
