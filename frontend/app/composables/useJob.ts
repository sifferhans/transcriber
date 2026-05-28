import type { MaybeRefOrGetter } from "vue";
import type { TranscribeJob } from "~/types/job";

const ACTIVE_INTERVAL_MS = 1500;
const IDLE_INTERVAL_MS = 10_000;

export function useJob(id: MaybeRefOrGetter<string>) {
  const api = useApi();
  const cache = useJobsCache();
  // Seed from cache so the detail card is in the DOM at first render for the view transition.
  const job = ref<TranscribeJob | null>(cache.get(toValue(id)));
  const loading = ref(false);
  const error = ref<string | null>(null);

  let timer: ReturnType<typeof setTimeout> | null = null;
  let alive = true;

  const isActive = computed(() => {
    const s = job.value?.status;
    return s === "QUEUED" || s === "RUNNING";
  });

  async function fetchOnce() {
    loading.value = true;
    try {
      const fresh = await api.getJob(toValue(id));
      job.value = fresh;
      cache.set(fresh);
      error.value = null;
    } catch (e: unknown) {
      error.value = e instanceof Error ? e.message : String(e);
    } finally {
      loading.value = false;
    }
  }

  function scheduleNext() {
    if (!alive) return;
    const delay = isActive.value ? ACTIVE_INTERVAL_MS : IDLE_INTERVAL_MS;
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

  return { job, loading, error, fetchOnce };
}
