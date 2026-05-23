import tailwindcss from "@tailwindcss/vite";

const TRANSCRIBER_API =
  process.env.NUXT_TRANSCRIBER_API ?? "http://localhost:8888";

export default defineNuxtConfig({
  compatibilityDate: "2025-07-15",
  devtools: { enabled: true },
  ssr: false,
  css: ["~/assets/css/main.css"],
  modules: ["@nuxt/icon", "@nuxt/fonts"],
  // Disable the directory-name component prefix so `components/design/Foo.vue`
  // auto-imports as `<Foo>`, not `<DesignFoo>`.
  components: [{ path: "~/components", pathPrefix: false }],
  vite: {
    plugins: [tailwindcss()],
  },
  app: {
    head: {
      title: "Transcriber",
      htmlAttrs: {
        "data-theme": "dark",
      },
    },
  },
  // In dev (`pnpm dev`) the frontend runs on :3000 and the Go API on :8888,
  // so we proxy each API path through Nuxt to keep the browser same-origin.
  // In the embedded production build (`make build`) the Go binary serves
  // both the SPA and the API on the same port, so these paths resolve
  // natively — the frontend code uses identical URLs in both modes.
  routeRules: {
    "/transcription/**": { proxy: `${TRANSCRIBER_API}/transcription/**` },
    "/models": { proxy: `${TRANSCRIBER_API}/models` },
    "/healthz": { proxy: `${TRANSCRIBER_API}/healthz` },
    "/readyz": { proxy: `${TRANSCRIBER_API}/readyz` },
  },
});
