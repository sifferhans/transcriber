// Apply stored theme before hydration to avoid a flash from Nuxt's default `data-theme`.
export default defineNuxtPlugin(() => {
  const stored = window.localStorage.getItem("theme");
  if (stored === "light" || stored === "dark") {
    document.documentElement.dataset.theme = stored;
  }
});
