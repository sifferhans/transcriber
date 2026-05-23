import type { MaybeRefOrGetter } from "vue";
import type { TranscribeJob } from "~/types/job";

const ACTIVE_INTERVAL_MS = 1500;
const IDLE_INTERVAL_MS = 10_000;

// useJob polls a single job; intervals shorten while the job is active so
// progress feels live, then back off once it terminates.
export function useJob(id: MaybeRefOrGetter<string>) {
  const api = useApi();
  const job = ref<TranscribeJob | null>(null);
  const loading = ref(false);
  const error = ref<string | null>(null);

  let timer: ReturnType<typeof setTimeout> | null = null;
  let alive = true;

  const isActive = computed(() => {
    const s = job.value?.status;
    return s === "PENDING" || s === "RUNNING";
  });

  async function fetchOnce() {
    loading.value = true;
    try {
      job.value = await api.getJob(toValue(id));
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
