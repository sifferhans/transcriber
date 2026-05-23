<script setup lang="ts">
import type { TranscribeJob } from "~/types/job";

const props = defineProps<{ job: TranscribeJob }>();
const emit = defineEmits<{ cancel: [] }>();

const api = useApi();
const canceling = ref(false);

const canCancel = computed(
  () => props.job.status === "PENDING" || props.job.status === "RUNNING",
);

async function onCancel(e: Event) {
  e.preventDefault();
  e.stopPropagation();
  if (!canCancel.value) return;
  canceling.value = true;
  try {
    await api.cancelJob(props.job.id);
    emit("cancel");
  } finally {
    canceling.value = false;
  }
}
</script>

<template>
  <NuxtLink
    :to="`/jobs/${job.id}`"
    class="block bg-white border border-slate-200 rounded-lg hover:border-slate-300 hover:shadow-sm transition px-4 py-3"
  >
    <div class="flex items-center gap-4">
      <div class="flex-1 min-w-0">
        <div class="flex items-center gap-2 mb-1">
          <JobStatusBadge :status="job.status" />
          <span class="text-xs text-slate-500 font-mono">{{
            job.id.slice(0, 8)
          }}</span>
          <span class="text-xs text-slate-400">·</span>
          <span class="text-xs text-slate-600">{{ job.model || "—" }}</span>
          <template v-if="job.language">
            <span class="text-xs text-slate-400">·</span>
            <span class="text-xs text-slate-600">{{ job.language }}</span>
          </template>
        </div>
        <div class="text-sm text-slate-700 truncate font-mono">
          {{ job.path }}
        </div>
      </div>

      <div class="flex items-center gap-3 shrink-0">
        <ProgressBar
          v-if="job.status === 'RUNNING'"
          :value="job.progress"
          class="w-28"
        />
        <span
          v-else-if="job.status === 'COMPLETED' && job.duration"
          class="text-xs text-slate-500"
        >
          {{ job.duration }}
        </span>
        <button
          v-if="canCancel"
          type="button"
          :disabled="canceling"
          class="text-slate-400 hover:text-red-600 disabled:opacity-50"
          :title="'Cancel job'"
          @click="onCancel"
        >
          <Icon name="mdi:close-circle-outline" size="20" />
        </button>
      </div>
    </div>
  </NuxtLink>
</template>
