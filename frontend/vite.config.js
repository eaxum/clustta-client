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
  // Using array format to guarantee order - more specific aliases must come first
  const aliases = [];

  // In web mode, redirect services to HTTP adapter and mock @wailsio/runtime
  if (isWebMode) {
    // Use regex with $ to match exact import (not @/services/utils etc)
    aliases.push({ find: /^@\/services$/, replacement: join(PACKAGE_ROOT, "src/services/adapters/http.js") });
    aliases.push({ find: '@wailsio/runtime', replacement: join(PACKAGE_ROOT, "src/services/runtime/index.js") });
  }

  // The @ alias should come after more specific aliases
  aliases.push({ find: '@', replacement: join(PACKAGE_ROOT, "src") });

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
      // Proxy API requests in web mode to bypass CORS
      proxy: isWebMode ? {
        '/api': {
          target: 'https://api.clustta.com',
          changeOrigin: true,
          rewrite: (path) => path.replace(/^\/api/, ''),
          secure: true,
          // Cookie handling - rewrite domain so browser accepts cookies from proxy
          cookieDomainRewrite: {
            'api.clustta.com': 'localhost',
            '.clustta.com': 'localhost',
          },
          // Configure the proxy to pass through Set-Cookie headers
          configure: (proxy, options) => {
            proxy.on('proxyRes', (proxyRes, req, res) => {
              // Log the response for debugging
              console.log(`[Proxy] ${req.method} ${req.url} -> ${proxyRes.statusCode}`);
              
              // Log incoming cookies from request
              if (req.headers.cookie) {
                console.log(`[Proxy] Request Cookies: ${req.headers.cookie}`);
              }
              
              // Rewrite Set-Cookie headers to work with localhost
              const cookies = proxyRes.headers['set-cookie'];
              if (cookies) {
                console.log(`[Proxy] Original Set-Cookie: ${JSON.stringify(cookies)}`);
                proxyRes.headers['set-cookie'] = cookies.map(cookie => {
                  // Remove Domain attribute or rewrite it
                  let newCookie = cookie
                    .replace(/Domain=[^;]+;?/gi, '')
                    .replace(/Secure;?/gi, '')  // Remove Secure for localhost (http)
                    .replace(/SameSite=\w+;?/gi, 'SameSite=Lax;');  // Use Lax for localhost
                  console.log(`[Proxy] Rewritten Set-Cookie: ${newCookie}`);
                  return newCookie;
                });
              }
            });
          },
        },
      } : undefined,
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
