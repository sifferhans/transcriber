import tailwindcss from "@tailwindcss/vite";

const TRANSCRIBER_API =
  process.env.NUXT_TRANSCRIBER_API ?? "http://localhost:8888";

export default defineNuxtConfig({
  compatibilityDate: "2025-07-15",
  devtools: { enabled: true },
  ssr: false,
  css: ["~/assets/css/main.css"],
  modules: ["@nuxt/icon", "@nuxt/fonts"],
  vite: {
    plugins: [tailwindcss()],
  },
  // Proxy /api/transcriber/** to the Go API. The browser talks only to the
  // Nuxt origin so there's no CORS to configure and the backend URL never
  // ends up in the client bundle.
  routeRules: {
    "/api/transcriber/**": {
      proxy: { to: `${TRANSCRIBER_API}/**` },
    },
  },
});
