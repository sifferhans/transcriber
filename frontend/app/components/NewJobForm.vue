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
</script>

<template>
  <form
    class="bg-white border border-slate-200 rounded-lg p-6 space-y-4"
    @submit.prevent="submit"
  >
    <h2 class="text-base font-semibold tracking-tight">New transcription job</h2>

    <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
      <div>
        <label class="block text-xs font-medium text-slate-700 mb-1">
          Audio file path
        </label>
        <input
          v-model="path"
          required
          placeholder="/mnt/storage/audio/foo.wav"
          class="w-full px-3 py-2 border border-slate-300 rounded text-sm font-mono focus:outline-none focus:ring-2 focus:ring-indigo-500/40"
        />
      </div>
      <div>
        <label class="block text-xs font-medium text-slate-700 mb-1">
          Output directory
        </label>
        <input
          v-model="outputPath"
          required
          placeholder="/mnt/storage/out/foo/"
          class="w-full px-3 py-2 border border-slate-300 rounded text-sm font-mono focus:outline-none focus:ring-2 focus:ring-indigo-500/40"
        />
      </div>
      <div>
        <label class="block text-xs font-medium text-slate-700 mb-1">
          Model
        </label>
        <select
          v-model="model"
          class="w-full px-3 py-2 border border-slate-300 rounded text-sm bg-white"
        >
          <option value="">(server default)</option>
          <option v-for="m in models" :key="m.id" :value="m.id">
            {{ m.name }}{{ m.default ? " · default" : "" }}
          </option>
        </select>
      </div>
      <div>
        <label class="block text-xs font-medium text-slate-700 mb-1">
          Language
        </label>
        <select
          v-model="language"
          class="w-full px-3 py-2 border border-slate-300 rounded text-sm bg-white"
        >
          <option v-for="l in LANGS" :key="l.v" :value="l.v">{{ l.l }}</option>
        </select>
      </div>
      <div>
        <label class="block text-xs font-medium text-slate-700 mb-1">
          Format
        </label>
        <select
          v-model="format"
          class="w-full px-3 py-2 border border-slate-300 rounded text-sm bg-white"
        >
          <option value="all">all (json+srt+vtt+txt)</option>
          <option value="json">json</option>
          <option value="srt">srt</option>
          <option value="vtt">vtt</option>
          <option value="txt">txt</option>
        </select>
      </div>
      <div>
        <label class="block text-xs font-medium text-slate-700 mb-1">
          Priority
        </label>
        <input
          v-model.number="priority"
          type="number"
          min="0"
          max="10"
          class="w-full px-3 py-2 border border-slate-300 rounded text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500/40"
        />
      </div>
    </div>

    <div v-if="error" class="text-sm text-red-600">{{ error }}</div>

    <div class="flex items-center justify-end gap-2">
      <button
        type="button"
        class="px-4 py-2 text-sm text-slate-600 hover:text-slate-900"
        @click="emit('cancel')"
      >
        Cancel
      </button>
      <button
        type="submit"
        :disabled="submitting"
        class="bg-indigo-600 hover:bg-indigo-700 disabled:opacity-50 text-white px-4 py-2 rounded text-sm font-medium inline-flex items-center gap-2"
      >
        <Icon
          v-if="submitting"
          name="mdi:loading"
          class="animate-spin"
        />
        {{ submitting ? "Submitting…" : "Submit" }}
      </button>
    </div>
  </form>
</template>
