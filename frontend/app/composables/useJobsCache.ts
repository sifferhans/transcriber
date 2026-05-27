import type { TranscribeJob } from "~/types/job";

// useJobsCache lets detail pages render from cache so the view-transition target exists at navigation.
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
