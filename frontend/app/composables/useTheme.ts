// Theme toggle backed by localStorage, reflected onto `<html data-theme>` (consumed by CSS variables).
export type Theme = "light" | "dark";

const STORAGE_KEY = "theme";
const DEFAULT_THEME: Theme = "dark";

// Module-scoped so all callers share the same reactive state.
const theme = ref<Theme>(DEFAULT_THEME);
let initialized = false;

function readInitial(): Theme {
  if (typeof window === "undefined") return DEFAULT_THEME;
  const stored = window.localStorage.getItem(STORAGE_KEY);
  if (stored === "light" || stored === "dark") return stored;
  return DEFAULT_THEME;
}

function apply(next: Theme) {
  if (typeof document === "undefined") return;
  document.documentElement.dataset.theme = next;
}

export function useTheme() {
  if (!initialized && typeof window !== "undefined") {
    theme.value = readInitial();
    apply(theme.value);
    initialized = true;
  }

  function setTheme(next: Theme) {
    theme.value = next;
    apply(next);
    if (typeof window !== "undefined") {
      window.localStorage.setItem(STORAGE_KEY, next);
    }
  }

  function toggle() {
    setTheme(theme.value === "dark" ? "light" : "dark");
  }

  return { theme: readonly(theme), toggle, setTheme };
}
