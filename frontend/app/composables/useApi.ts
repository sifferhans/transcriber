import type { ModelInfo, TranscribeInput, TranscribeJob } from "~/types/job";

// Same-origin paths in both modes: in dev Nuxt proxies them to the Go API
// (see nuxt.config.ts routeRules), in the embedded build the Go binary
// serves them itself.
export function useApi() {
  return {
    listJobs: () => $fetch<TranscribeJob[]>(`/transcription/jobs`),
    getJob: (id: string) => $fetch<TranscribeJob>(`/transcription/job/${id}`),
    createJob: (input: TranscribeInput) =>
      $fetch<TranscribeJob>(`/transcription/job`, {
        method: "POST",
        body: input,
      }),
    cancelJob: (id: string) =>
      $fetch<TranscribeJob>(`/transcription/job/${id}`, {
        method: "DELETE",
      }),
    listModels: () => $fetch<ModelInfo[]>(`/models`),
  };
}
