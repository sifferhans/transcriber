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
      class="text-body-3 text-primary-default hover:text-primary-contrast hover:underline inline-flex items-center gap-1 mb-4"
    >
      <Icon name="tabler:arrow-left" /> Back to jobs
    </NuxtLink>

    <div
      v-if="error"
      class="bg-semantic-error/10 text-semantic-error ring-1 ring-semantic-error/30 rounded-md px-4 py-3 text-body-3"
    >
      {{ error }}
    </div>

    <div v-else-if="!job && loading" class="text-text-hint">Loading…</div>

    <div
      v-else-if="job"
      class="bg-surface-raise border border-border-1 rounded-lg p-6 space-y-6 shadow-resting"
    >
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-3">
          <JobStatusBadge :status="job.status" />
          <span class="font-mono text-body-3 text-text-hint">{{ job.id }}</span>
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
          class="flex items-center justify-between mb-1 text-caption-1 text-text-muted"
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

      <div v-if="job.result" class="border-t border-border-1 pt-4">
        <Detail label="Result">{{ job.result }}</Detail>
      </div>

      <div v-if="job.error" class="border-t border-border-1 pt-4">
        <div
          class="text-caption-2 text-semantic-error uppercase tracking-wider mb-1"
        >
          Error
        </div>
        <pre
          class="text-body-3 text-semantic-error bg-semantic-error/10 p-3 rounded-md font-mono overflow-x-auto whitespace-pre-wrap"
          >{{ job.error }}</pre
        >
      </div>
    </div>
  </div>
</template>
