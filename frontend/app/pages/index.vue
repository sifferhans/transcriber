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
        <h1 class="text-2xl font-semibold tracking-tight text-fg">
          Transcription jobs
        </h1>
        <p class="text-sm text-fg-muted mt-1">
          {{ jobs.length }} total
          <span v-if="loading" class="ml-2 text-muted">refreshing…</span>
        </p>
      </div>
      <DesignButton
        :variant="showForm ? 'secondary' : 'primary'"
        @click="showForm = !showForm"
      >
        <Icon :name="showForm ? 'tabler:x' : 'tabler:plus'" />
        {{ showForm ? "Close" : "New job" }}
      </DesignButton>
    </div>

    <NewJobForm
      v-if="showForm"
      class="mb-6"
      @created="onCreated"
      @cancel="showForm = false"
    />

    <div
      v-if="error"
      class="bg-danger-50 text-danger-800 border border-danger-200 rounded-md px-4 py-3 mb-4 text-sm"
    >
      {{ error }}
    </div>

    <div
      v-if="!jobs.length && !loading"
      class="text-center py-16 text-fg-muted border border-dashed border-line rounded-lg"
    >
      <Icon
        name="tabler:wave-saw-tool"
        size="32"
        class="mx-auto mb-2 text-line-strong"
      />
      <p>No jobs yet. Click "New job" to start one.</p>
    </div>

    <div v-else class="space-y-2">
      <JobRow v-for="j in jobs" :key="j.id" :job="j" @cancel="fetchOnce" />
    </div>
  </div>
</template>
