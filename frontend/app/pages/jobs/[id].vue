<script setup lang="ts">
const route = useRoute();
const id = computed(() => route.params.id as string);

const { job, loading, error, fetchOnce } = useJob(id);
const api = useApi();
const canceling = ref(false);

const canCancel = computed(
  () => job.value?.status === "PENDING" || job.value?.status === "RUNNING",
);

async function onCancel() {
  canceling.value = true;
  try {
    await api.cancelJob(id.value);
    await fetchOnce();
  } finally {
    canceling.value = false;
  }
}
</script>

<template>
  <div>
    <NuxtLink
      to="/"
      class="text-sm text-brand-600 hover:text-brand-700 hover:underline inline-flex items-center gap-1 mb-4"
    >
      <Icon name="tabler:arrow-left" /> Back to jobs
    </NuxtLink>

    <div
      v-if="error"
      class="bg-danger-50 text-danger-800 border border-danger-200 rounded-md px-4 py-3 text-sm"
    >
      {{ error }}
    </div>

    <div v-else-if="!job && loading" class="text-muted">Loading…</div>

    <div
      v-else-if="job"
      class="bg-surface border border-line rounded-lg p-6 space-y-6"
    >
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-3">
          <JobStatusBadge :status="job.status" />
          <span class="font-mono text-sm text-muted">{{ job.id }}</span>
        </div>
        <DesignButton
          v-if="canCancel"
          variant="danger"
          size="sm"
          :loading="canceling"
          @click="onCancel"
        >
          <Icon v-if="!canceling" name="tabler:circle-x" />
          Cancel
        </DesignButton>
      </div>

      <div v-if="job.status === 'RUNNING'">
        <div
          class="flex items-center justify-between mb-1 text-xs text-fg-muted"
        >
          <span>Progress</span>
          <span>{{ job.progress }}%</span>
        </div>
        <ProgressBar :value="job.progress" />
      </div>

      <dl class="grid grid-cols-1 md:grid-cols-2 gap-x-8 gap-y-4">
        <Detail label="Input">{{ job.path }}</Detail>
        <Detail label="Output dir">{{ job.output_path }}</Detail>
        <Detail label="Model">{{ job.model || "—" }}</Detail>
        <Detail label="Language">{{ job.language || "—" }}</Detail>
        <Detail label="Format">{{ job.format || "—" }}</Detail>
        <Detail label="Priority">{{ job.priority }}</Detail>
        <Detail label="Duration">{{ job.duration || "—" }}</Detail>
        <Detail label="Callback">{{ job.callback || "—" }}</Detail>
      </dl>

      <div v-if="job.result" class="border-t border-line pt-4">
        <Detail label="Result">{{ job.result }}</Detail>
      </div>

      <div v-if="job.error" class="border-t border-line pt-4">
        <div
          class="text-xs font-medium text-danger-700 uppercase tracking-wider mb-1"
        >
          Error
        </div>
        <pre
          class="text-sm text-danger-800 bg-danger-50 p-3 rounded-md font-mono overflow-x-auto whitespace-pre-wrap"
          >{{ job.error }}</pre
        >
      </div>
    </div>
  </div>
</template>
