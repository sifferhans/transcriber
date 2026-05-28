import type { TranscribeJob } from "~/types/job";

const ACTIVE_INTERVAL_MS = 2000;
const IDLE_INTERVAL_MS = 10_000;

export function useJobsList() {
  const api = useApi();
  const cache = useJobsCache();
  // useState (not ref) so it survives navigation, keeping the view-transition target in the DOM.
  const jobs = useState<TranscribeJob[]>("jobs-list", () => []);
  const loading = ref(false);
  const error = ref<string | null>(null);

  let timer: ReturnType<typeof setTimeout> | null = null;
  let alive = true;

  const hasActive = computed(() =>
    jobs.value.some((j) => j.status === "QUEUED" || j.status === "RUNNING"),
  );

  async function fetchOnce() {
    loading.value = true;
    try {
      const all = await api.listJobs();
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
