import { defineConfig } from "vite";
import vue from "@vitejs/plugin-vue";
import { join } from "path";

const host = process.env.TAURI_DEV_HOST;

// https://vitejs.dev/config/
const PACKAGE_ROOT = __dirname;

// Check if we're building for web mode
const isWebMode = process.env.VITE_PLATFORM === 'web';

export default defineConfig(async () => {
  // Build aliases based on platform
  const aliases = {
    "@": join(PACKAGE_ROOT, "src") + "/",
  };

  // In web mode, redirect @/services to HTTP adapter and mock @wailsio/runtime
  if (isWebMode) {
    aliases["@/services"] = join(PACKAGE_ROOT, "src/services/adapters/http.js");
    aliases["@wailsio/runtime"] = join(PACKAGE_ROOT, "src/services/runtime/index.js");
  } else {
    // Desktop mode uses Wails adapter (default)
    aliases["@/services"] = join(PACKAGE_ROOT, "src/services/adapters/wails.js");
  }

  return {
    plugins: [vue()],

    // Vite options tailored for Tauri development and only applied in `tauri dev` or `tauri build`
    //
    // 1. prevent vite from obscuring rust errors
    clearScreen: false,
    // 2. tauri expects a fixed port, fail if that port is not available
    server: {
      port: 1420,
      strictPort: true,
      host: host || false,
      hmr: host
        ? {
            protocol: "ws",
            host,
            port: 1421,
          }
        : undefined,
      watch: {
        // 3. tell vite to ignore watching `src-tauri`
        ignored: ["**/src-tauri/**"],
      },
    },
    envPrefix: ["VITE_", "TAURI_"],
    build: {
      target: ["es2022", "chrome100", "safari15"],
      minify: !process.env.TAURI_DEBUG ? "esbuild" : false,
      sourcemap: !!process.env.TAURI_DEBUG,
    },
    root: PACKAGE_ROOT,
    resolve: {
      alias: aliases,
      extensions: [".mjs", ".js", ".ts", ".jsx", ".tsx", ".json", ".vue"],
    },
  };
});
