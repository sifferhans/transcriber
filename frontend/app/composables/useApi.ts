import type {
  ModelInfo,
  TranscribeInput,
  TranscribeJob,
} from "~/types/job";

// All calls go through the Nuxt server proxy (see nuxt.config.ts routeRules)
// so the browser never sees the Go API directly.
const BASE = "/api/transcriber";

export function useApi() {
  return {
    listJobs: () => $fetch<TranscribeJob[]>(`${BASE}/transcription/jobs`),
    getJob: (id: string) =>
      $fetch<TranscribeJob>(`${BASE}/transcription/job/${id}`),
    createJob: (input: TranscribeInput) =>
      $fetch<TranscribeJob>(`${BASE}/transcription/job`, {
        method: "POST",
        body: input,
      }),
    cancelJob: (id: string) =>
      $fetch<TranscribeJob>(`${BASE}/transcription/job/${id}`, {
        method: "DELETE",
      }),
    listModels: () => $fetch<ModelInfo[]>(`${BASE}/models`),
  };
}
