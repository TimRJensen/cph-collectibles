import { defineConfig } from "vite";
import vue from "@vitejs/plugin-vue";
import { resolve } from "node:path";

// https://vite.dev/config/
export default defineConfig({
  plugins: [vue()],
  root: resolve(__dirname, "src/pages"),
  publicDir: resolve(__dirname, "public"),
  base: "/",
  server: {
    host: "0.0.0.0",
    port: 3000,
    allowedHosts: ["localhost", "dashboard.localhost"],
  },
  build: {
    outDir: resolve(__dirname, "dist"),
    emptyOutDir: true,
    rollupOptions: {
      input: {
        landing: resolve(__dirname, "src/pages/index.html"),
        about: resolve(__dirname, "src/pages/about/index.html"),
        contact: resolve(__dirname, "src/pages/contact/index.html"),
        shop: resolve(__dirname, "src/pages/shop/index.html"),
        checkout: resolve(__dirname, "src/pages/checkout/index.html"),
        complete: resolve(__dirname, "src/pages/checkout/complete/index.html"),
      },
    },
  },
});