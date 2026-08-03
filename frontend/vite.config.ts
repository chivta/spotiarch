import { defineConfig, loadEnv } from "vite";
import react from "@vitejs/plugin-react";

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, ".", "");
  return {
    base: "/",
    plugins: [react()],
    server: {
      host: true,
      proxy: {
        "/api": env.VITE_API_URL || "http://localhost:3000",
      },
    },
  };
});
