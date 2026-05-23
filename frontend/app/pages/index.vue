<script setup lang="ts">
import type { TranscribeJob } from "~/types/job";

const { jobs, loading, error, fetchOnce } = useJobsList();
const showForm = ref(false);

async function onCreated(_job: TranscribeJob) {
  showForm.value = false;
  await fetchOnce();
}
</script>

<template>
  <div>
    <div class="flex items-center justify-between mb-6">
      <div>
        <h1 class="text-2xl font-semibold tracking-tight">
          Transcription jobs
        </h1>
        <p class="text-sm text-slate-500 mt-1">
          {{ jobs.length }} total
          <span v-if="loading" class="ml-2 text-slate-400">refreshing…</span>
        </p>
      </div>
      <button
        type="button"
        class="bg-indigo-600 hover:bg-indigo-700 text-white px-4 py-2 rounded text-sm font-medium inline-flex items-center gap-2"
        @click="showForm = !showForm"
      >
        <Icon :name="showForm ? 'mdi:close' : 'mdi:plus'" />
        {{ showForm ? "Close" : "New job" }}
      </button>
    </div>

    <NewJobForm
      v-if="showForm"
      class="mb-6"
      @created="onCreated"
      @cancel="showForm = false"
    />

    <div
      v-if="error"
      class="bg-red-50 text-red-800 border border-red-200 rounded px-4 py-3 mb-4 text-sm"
    >
      {{ error }}
    </div>

    <div
      v-if="!jobs.length && !loading"
      class="text-center py-16 text-slate-500 border border-dashed border-slate-300 rounded-lg"
    >
      <Icon
        name="mdi:waveform"
        size="32"
        class="mx-auto mb-2 text-slate-300"
      />
      <p>No jobs yet. Click "New job" to start one.</p>
    </div>

    <div v-else class="space-y-2">
      <JobRow
        v-for="j in jobs"
        :key="j.id"
        :job="j"
        @cancel="fetchOnce"
      />
    </div>
  </div>
</template>
