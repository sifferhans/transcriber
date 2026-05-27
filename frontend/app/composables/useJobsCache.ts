import type { TranscribeJob } from "~/types/job";

// useJobsCache is a tiny per-app store keyed by job id. The list fetch writes
// into it, and the detail page reads from it for its initial render so the
// view-transition target (the card with `view-transition-name: job-<id>`) is
// in the DOM the moment navigation commits — otherwise the transition fires
// while the detail page is still showing its loading state and there's
// nothing to morph into.
export function useJobsCache() {
  const map = useState<Record<string, TranscribeJob>>("jobs-cache", () => ({}));

  function set(job: TranscribeJob) {
    map.value = { ...map.value, [job.id]: job };
  }

  function setMany(jobs: TranscribeJob[]) {
    const next = { ...map.value };
    for (const j of jobs) next[j.id] = j;
    map.value = next;
  }

  function get(id: string): TranscribeJob | null {
    return map.value[id] ?? null;
  }

  return { set, setMany, get };
}
