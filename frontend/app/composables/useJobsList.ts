import type { TranscribeJob } from "~/types/job";

const ACTIVE_INTERVAL_MS = 2000;
const IDLE_INTERVAL_MS = 10_000;

// useJobsList fetches the full job list on mount and re-polls at an interval
// that adapts to whether any job is currently active. Cleans up on unmount.
export function useJobsList() {
  const api = useApi();
  const cache = useJobsCache();
  // `jobs` is useState (not ref) so it survives the route navigation away
  // from "/" and back: when the user returns from a detail page, the list
  // is already in the DOM on the first render, which lets the view
  // transition find a target with `view-transition-name: job-<id>`.
  const jobs = useState<TranscribeJob[]>("jobs-list", () => []);
  const loading = ref(false);
  const error = ref<string | null>(null);

  let timer: ReturnType<typeof setTimeout> | null = null;
  let alive = true;

  const hasActive = computed(() =>
    jobs.value.some((j) => j.status === "PENDING" || j.status === "RUNNING"),
  );

  async function fetchOnce() {
    loading.value = true;
    try {
      const all = await api.listJobs();
      // newest first
      jobs.value = all.slice().reverse();
      cache.setMany(all);
      error.value = null;
    } catch (e: unknown) {
      error.value = errorMessage(e);
    } finally {
      loading.value = false;
    }
  }

  function scheduleNext() {
    if (!alive) return;
    const delay = hasActive.value ? ACTIVE_INTERVAL_MS : IDLE_INTERVAL_MS;
    timer = setTimeout(async () => {
      await fetchOnce();
      scheduleNext();
    }, delay);
  }

  onMounted(async () => {
    await fetchOnce();
    scheduleNext();
  });

  onBeforeUnmount(() => {
    alive = false;
    if (timer) clearTimeout(timer);
  });

  return { jobs, loading, error, fetchOnce };
}

function errorMessage(e: unknown): string {
  if (e instanceof Error) return e.message;
  return String(e);
}
