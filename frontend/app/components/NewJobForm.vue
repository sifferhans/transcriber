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
  { v: "auto", l: "auto-detect" },
  { v: "en", l: "English" },
  { v: "no", l: "Norwegian" },
  { v: "de", l: "German" },
  { v: "es", l: "Spanish" },
  { v: "fr", l: "French" },
  { v: "it", l: "Italian" },
  { v: "pt", l: "Portuguese" },
  { v: "nl", l: "Dutch" },
  { v: "sv", l: "Swedish" },
];

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

const inputClass =
  "w-full px-3 py-2 border border-line rounded-md text-sm font-mono bg-surface text-fg placeholder:text-muted focus:outline-none focus:ring-2 focus:ring-brand-500/40 focus:border-brand-500";
const selectClass =
  "w-full px-3 py-2 border border-line rounded-md text-sm bg-surface text-fg focus:outline-none focus:ring-2 focus:ring-brand-500/40 focus:border-brand-500";
const labelClass = "block text-xs font-medium text-fg mb-1";
</script>

<template>
  <form
    class="bg-surface border border-line rounded-lg p-6 space-y-4"
    @submit.prevent="submit"
  >
    <h2 class="text-base font-semibold tracking-tight text-fg">
      New transcription job
    </h2>

    <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
      <div>
        <label :class="labelClass">Audio file path</label>
        <input
          v-model="path"
          required
          placeholder="/mnt/storage/audio/foo.wav"
          :class="inputClass"
        />
      </div>
      <div>
        <label :class="labelClass">Output directory</label>
        <input
          v-model="outputPath"
          required
          placeholder="/mnt/storage/out/foo/"
          :class="inputClass"
        />
      </div>
      <div>
        <label :class="labelClass">Model</label>
        <select v-model="model" :class="selectClass">
          <option value="">(server default)</option>
          <option v-for="m in models" :key="m.id" :value="m.id">
            {{ m.name }}{{ m.default ? " · default" : "" }}
          </option>
        </select>
      </div>
      <div>
        <label :class="labelClass">Language</label>
        <select v-model="language" :class="selectClass">
          <option v-for="l in LANGS" :key="l.v" :value="l.v">{{ l.l }}</option>
        </select>
      </div>
      <div>
        <label :class="labelClass">Format</label>
        <select v-model="format" :class="selectClass">
          <option value="all">all (json+srt+vtt+txt)</option>
          <option value="json">json</option>
          <option value="srt">srt</option>
          <option value="vtt">vtt</option>
          <option value="txt">txt</option>
        </select>
      </div>
      <div>
        <label :class="labelClass">Priority</label>
        <input
          v-model.number="priority"
          type="number"
          min="0"
          max="10"
          :class="inputClass"
        />
      </div>
    </div>

    <div v-if="error" class="text-sm text-danger-600">{{ error }}</div>

    <div class="flex items-center justify-end gap-2">
      <DesignButton variant="ghost" @click="emit('cancel')"> Cancel </DesignButton>
      <DesignButton type="submit" :loading="submitting">
        {{ submitting ? "Submitting…" : "Submit" }}
      </DesignButton>
    </div>
  </form>
</template>
