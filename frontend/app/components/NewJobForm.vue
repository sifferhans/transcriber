<script setup lang="ts">
import type { ModelInfo, TranscribeJob } from "~/types/job";

const emit = defineEmits<{ created: [job: TranscribeJob]; cancel: [] }>();

const api = useApi();

const path = ref("");
const outputPath = ref("");
const language = ref("auto");
const format = ref("all");
const model = ref("");
const priority = ref(0);

const models = ref<ModelInfo[]>([]);
const submitting = ref(false);
const error = ref<string | null>(null);

const LANGS = [
  { value: "auto", label: "auto-detect" },
  { value: "en", label: "English" },
  { value: "no", label: "Norwegian" },
  { value: "de", label: "German" },
  { value: "es", label: "Spanish" },
  { value: "fr", label: "French" },
  { value: "it", label: "Italian" },
  { value: "pt", label: "Portuguese" },
  { value: "nl", label: "Dutch" },
  { value: "sv", label: "Swedish" },
];

const FORMATS = [
  { value: "all", label: "all (json+srt+vtt+txt)" },
  { value: "json", label: "json" },
  { value: "srt", label: "srt" },
  { value: "vtt", label: "vtt" },
  { value: "txt", label: "txt" },
];

const modelOptions = computed(() =>
  models.value.map((m) => ({
    value: m.id,
    label: m.name + (m.default ? " · default" : ""),
  })),
);

onMounted(async () => {
  try {
    models.value = await api.listModels();
    const def = models.value.find((m) => m.default);
    model.value = def?.id ?? models.value[0]?.id ?? "";
  } catch (e: unknown) {
    error.value = `Could not load models: ${e instanceof Error ? e.message : String(e)}`;
  }
});

async function submit() {
  if (!path.value || !outputPath.value) {
    error.value = "path and output_path are required";
    return;
  }
  submitting.value = true;
  error.value = null;
  try {
    const job = await api.createJob({
      path: path.value,
      output_path: outputPath.value,
      language: language.value,
      format: format.value,
      model: model.value || undefined,
      priority: priority.value || undefined,
    });
    emit("created", job);
  } catch (e: unknown) {
    const err = e as { data?: { error?: string }; message?: string };
    error.value = err?.data?.error ?? err?.message ?? String(e);
  } finally {
    submitting.value = false;
  }
}
</script>

<template>
  <form
    class="bg-surface-raise border border-border-1 rounded-lg p-6 space-y-4 shadow-resting"
    @submit.prevent="submit"
  >
    <h2 class="text-title-1 text-text-default">New transcription job</h2>

    <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
      <DesignInput
        v-model="path"
        label="Audio file path"
        placeholder="/mnt/storage/audio/foo.wav"
        required
        monospace
      />
      <DesignInput
        v-model="outputPath"
        label="Output directory"
        placeholder="/mnt/storage/out/foo/"
        required
        monospace
      />
      <DesignInput
        v-model="model"
        type="select"
        label="Model"
        prompt="(server default)"
        :options="modelOptions"
      />
      <DesignInput
        v-model="language"
        type="select"
        label="Language"
        :options="LANGS"
      />
      <DesignInput
        v-model="format"
        type="select"
        label="Format"
        :options="FORMATS"
      />
      <DesignInput
        v-model.number="priority"
        type="number"
        label="Priority"
        :min="0"
        :max="10"
      />
    </div>

    <DesignBanner v-if="error" variant="error" icon="tabler:alert-circle">
      {{ error }}
    </DesignBanner>

    <div class="flex items-center justify-end gap-2">
      <DesignButton
        variant="tertiary"
        size="small"
        label="Cancel"
        @click="emit('cancel')"
      />
      <DesignButton
        type="submit"
        size="small"
        :loading="submitting"
        :label="submitting ? 'Submitting…' : 'Submit'"
      />
    </div>
  </form>
</template>
