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
    <DesignHeader>
      Transcription jobs
      <template #subtitle>
        {{ jobs.length }} total
        <span v-if="loading" class="ml-2 text-text-hint">refreshing…</span>
      </template>
      <template #actions>
        <DesignButton
          :variant="showForm ? 'secondary' : 'primary'"
          :icon="showForm ? 'tabler:x' : 'tabler:plus'"
          :label="showForm ? 'Close' : 'New job'"
          @click="showForm = !showForm"
        />
      </template>
    </DesignHeader>

    <NewJobForm
      v-if="showForm"
      class="mb-6"
      @created="onCreated"
      @cancel="showForm = false"
    />

    <DesignBanner
      v-if="error"
      variant="error"
      icon="tabler:alert-circle"
      class="mb-4"
    >
      {{ error }}
    </DesignBanner>

    <DesignEmptyState
      v-if="!jobs.length && !loading"
      icon="tabler:wave-saw-tool"
      title="No jobs yet"
      description="Click “New job” to start a transcription."
      class="py-16"
    />

    <div v-else class="space-y-2">
      <JobRow v-for="j in jobs" :key="j.id" :job="j" @cancel="fetchOnce" />
    </div>
  </div>
</template>
