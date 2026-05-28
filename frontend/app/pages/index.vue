<script setup lang="ts">
import type { TranscribeJob } from "~/types/job";

const { jobs, loading, error, fetchOnce } = useJobsList();
const showForm = ref(false);

const activeJobs = computed(() =>
    jobs.value.filter(
        (j) => j.status === "QUEUED" || j.status === "RUNNING",
    ),
);
const recentJobs = computed(() =>
    jobs.value.filter(
        (j) =>
            j.status === "COMPLETED" ||
            j.status === "FAILED" ||
            j.status === "CANCELED",
    ),
);

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
                {{ activeJobs.length }} active · {{ recentJobs.length }} recent
                <span v-if="loading" class="ml-2 text-text-hint"
                    >refreshing…</span
                >
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

        <div v-else class="space-y-8">
            <section>
                <h2
                    class="text-caption-1 font-medium text-text-muted uppercase tracking-wider mb-3"
                >
                    Active
                    <span class="text-text-hint ml-1"
                        >({{ activeJobs.length }})</span
                    >
                </h2>
                <div v-if="activeJobs.length" class="space-y-2">
                    <JobRow
                        v-for="j in activeJobs"
                        :key="j.id"
                        :job="j"
                        @cancel="fetchOnce"
                    />
                </div>
                <p v-else class="text-body-3 text-text-hint italic">
                    No active jobs.
                </p>
            </section>

            <section v-if="recentJobs.length">
                <h2
                    class="text-caption-1 font-medium text-text-muted uppercase tracking-wider mb-3"
                >
                    Recent
                    <span class="text-text-hint ml-1"
                        >({{ recentJobs.length }})</span
                    >
                </h2>
                <div class="space-y-2">
                    <JobRow
                        v-for="j in recentJobs"
                        :key="j.id"
                        :job="j"
                        @cancel="fetchOnce"
                    />
                </div>
            </section>
        </div>
    </div>
</template>
