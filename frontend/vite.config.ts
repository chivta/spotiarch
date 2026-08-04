import { defineConfig, loadEnv } from "vite";
import react from "@vitejs/plugin-react";

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, ".", "");
  return {
    base: "/",
    plugins: [react()],
    server: {
      host: true,
      // The SPA calls a relative /api so the browser stays same-origin, matching
      // production where traefik serves both from one host. VITE_API_URL is the
      // server-side proxy target, not something the browser ever sees.
      proxy: {
        "/api": env.VITE_API_URL || "http://localhost:8080",
      },
    },
  };
});
