// Apply the stored theme as early as possible so first paint matches the
// user's preference. Without this the initial render uses whatever
// `data-theme` Nuxt put on <html>, then flips after hydration.

export default defineNuxtPlugin(() => {
  const stored = window.localStorage.getItem("theme");
  if (stored === "light" || stored === "dark") {
    document.documentElement.dataset.theme = stored;
  }
});
