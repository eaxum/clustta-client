import { defineConfig } from "vite";
import vue from "@vitejs/plugin-vue";
import { join } from "path";
import http from "http";
import https from "https";

const host = process.env.TAURI_DEV_HOST;

// https://vitejs.dev/config/
const PACKAGE_ROOT = __dirname;

// Check if we're building for web mode
const isWebMode = process.env.VITE_PLATFORM === 'web';

// In desktop mode, web-only views still exist in the source tree but their
// adapter imports (e.g. `@clustta/web-adapters`) are unreachable at runtime.
// This plugin resolves those imports to an empty module so the build succeeds.
function webAdaptersStub() {
  const VIRTUAL_ID = '\0web-adapters-stub';
  return {
    name: 'web-adapters-stub',
    resolveId(id) {
      if (!isWebMode && (id === '@clustta/web-adapters' || id.startsWith('@clustta/web-adapters/'))) {
        return { id: VIRTUAL_ID, syntheticNamedExports: true };
      }
    },
    load(id) {
      if (id === VIRTUAL_ID) {
        return 'export default {}';
      }
    }
  };
}

export default defineConfig(async () => {
  // Build aliases based on platform
  // Using array format to guarantee order - more specific aliases must come first
  const aliases = [];

  // Resolve `@/services/adapters/*` to @clustta/web-adapters in both modes.
  // In web mode the real package is used; in desktop mode the stub plugin above
  // returns an empty module so the build doesn't need the package installed.
  aliases.push({ find: /^@\/services\/adapters\/(.+?)(\.js)?$/, replacement: "@clustta/web-adapters/$1" });

  // In web mode, also redirect the bare `@/services` import and mock
  // @wailsio/runtime so we can run without the desktop bindings.
  if (isWebMode) {
    aliases.push({ find: /^@\/services$/, replacement: "@clustta/web-adapters" });
    aliases.push({ find: '@wailsio/runtime', replacement: "@clustta/web-adapters/runtime/index.js" });
  }

  // The @ alias should come after more specific aliases
  aliases.push({ find: '@', replacement: join(PACKAGE_ROOT, "src") });

  return {
    plugins: [webAdaptersStub(), vue()],

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
      // Custom middleware to handle special proxy cases (like GET-with-body)
      // This runs before the proxy middleware
      ...(isWebMode ? {
        // Use configureServer hook via plugins instead - but we can use middleware option
      } : {}),
      // Proxy API requests in web mode to bypass CORS
      proxy: isWebMode ? {
        // Special proxy for /data endpoint that requires GET with body
        // Browser can't send body with GET, so we receive POST and forward as GET with body
        '/studio-data': {
          target: 'https://placeholder.invalid',
          changeOrigin: true,
          secure: false,
          configure: (proxy, options) => {
            // We'll completely bypass the proxy and use custom Node.js http(s) request
            proxy.on('proxyReq', (proxyReq, req, res) => {
              // Abort the proxy request - we'll handle it ourselves
              proxyReq.destroy();
            });
            
            // Override to handle it manually
            const originalWeb = proxy.web.bind(proxy);
            proxy.web = async function(req, res, opts) {
              const studioUrl = req.headers['x-studio-url'];
              if (!studioUrl) {
                res.writeHead(400, { 'Content-Type': 'application/json' });
                res.end(JSON.stringify({ error: 'Missing X-Studio-URL header' }));
                return;
              }
              
              try {
                const targetUrl = new URL(studioUrl);
                const endpoint = req.url.replace(/^\/studio-data/, '');
                const fullPath = targetUrl.pathname.replace(/\/$/, '') + endpoint;
                
                console.log(`[Studio Data Proxy] Converting POST → GET with body`);
                console.log(`[Studio Data Proxy] Target: ${targetUrl.origin}${fullPath}`);
                
                // Collect request body
                const chunks = [];
                req.on('data', chunk => chunks.push(chunk));
                req.on('end', () => {
                  const body = Buffer.concat(chunks);
                  console.log(`[Studio Data Proxy] Body: ${body.toString()}`);
                  
                  // Use correct protocol module
                  const httpModule = targetUrl.protocol === 'https:' ? https : http;
                  
                  const proxyOptions = {
                    hostname: targetUrl.hostname,
                    port: targetUrl.port || (targetUrl.protocol === 'https:' ? 443 : 80),
                    path: fullPath,
                    method: 'GET', // Force GET even though browser sent POST
                    headers: {
                      'Content-Type': 'application/json',
                      'Content-Length': body.length,
                      'Clustta-Agent': req.headers['clustta-agent'] || 'Clustta/Web',
                    },
                  };
                  
                  const proxyRequest = httpModule.request(proxyOptions, (proxyResponse) => {
                    console.log(`[Studio Data Proxy] Response: ${proxyResponse.statusCode}`);
                    
                    // Forward status and headers
                    res.writeHead(proxyResponse.statusCode, proxyResponse.headers);
                    
                    // Pipe response body
                    proxyResponse.pipe(res);
                  });
                  
                  proxyRequest.on('error', (err) => {
                    console.error(`[Studio Data Proxy] Error: ${err.message}`);
                    if (!res.headersSent) {
                      res.writeHead(502, { 'Content-Type': 'application/json' });
                      res.end(JSON.stringify({ error: 'Proxy error', message: err.message }));
                    }
                  });
                  
                  // Write the body to the GET request (Node.js allows this)
                  proxyRequest.write(body);
                  proxyRequest.end();
                });
              } catch (e) {
                console.error(`[Studio Data Proxy] Error: ${e.message}`);
                res.writeHead(500, { 'Content-Type': 'application/json' });
                res.end(JSON.stringify({ error: 'Proxy error', message: e.message }));
              }
            };
          },
        },
        // Dynamic proxy for studio servers - reads target URL from X-Studio-URL header
        // Using bypass function to handle dynamic target since router doesn't work as expected
        '/studio-proxy': {
          target: 'https://placeholder.invalid', // Placeholder - will be overridden
          changeOrigin: true,
          secure: false,
          configure: (proxy, options) => {
            // Store original createProxyReq to intercept target
            const originalWeb = proxy.web.bind(proxy);
            
            // Override the web method to change target per-request
            proxy.web = function(req, res, opts) {
              const studioUrl = req.headers['x-studio-url'];
              if (studioUrl) {
                try {
                  const url = new URL(studioUrl);
                  opts = { ...opts, target: url.origin };
                  console.log(`[Studio Proxy] Routing to: ${url.origin}`);
                } catch (e) {
                  console.error(`[Studio Proxy] URL parse error: ${e.message}`);
                }
              }
              return originalWeb(req, res, opts);
            };
            
            proxy.on('proxyReq', (proxyReq, req, res) => {
              const studioUrl = req.headers['x-studio-url'];
              if (studioUrl) {
                try {
                  const url = new URL(studioUrl);
                  const basePath = url.pathname.replace(/\/$/, ''); // Remove trailing slash
                  const endpoint = req.url.replace(/^\/studio-proxy/, '');
                  const newPath = basePath + endpoint;
                  proxyReq.path = newPath;
                  
                  // Set host header to match target
                  proxyReq.setHeader('Host', url.host);
                  
                  // Ensure custom headers are forwarded
                  if (req.headers['userid']) {
                    proxyReq.setHeader('UserId', req.headers['userid']);
                  }
                  if (req.headers['userdata']) {
                    proxyReq.setHeader('UserData', req.headers['userdata']);
                  }
                  if (req.headers['clustta-agent']) {
                    proxyReq.setHeader('Clustta-Agent', req.headers['clustta-agent']);
                  }
                  
                  console.log(`[Studio Proxy] ${req.method} -> ${url.origin}${newPath}`);
                } catch (e) {
                  console.error(`[Studio Proxy] URL parse error: ${e.message}`);
                }
              } else {
                console.error('[Studio Proxy] Missing X-Studio-URL header');
              }
            });
            
            proxy.on('proxyRes', (proxyRes, req, res) => {
              console.log(`[Studio Proxy] Response: ${proxyRes.statusCode}`);
            });
            
            proxy.on('error', (err, req, res) => {
              console.error('[Studio Proxy] Error:', err.message);
              // Send error response to client
              if (!res.headersSent) {
                res.writeHead(502, { 'Content-Type': 'application/json' });
                res.end(JSON.stringify({ error: 'Proxy error', message: err.message }));
              }
            });
          },
        },
        // Global API proxy
        '/api': {
          target: process.env.VITE_API_URL || 'https://api.clustta.com',
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
              // Rewrite Set-Cookie headers to work with localhost
              const cookies = proxyRes.headers['set-cookie'];
              if (cookies) {
                proxyRes.headers['set-cookie'] = cookies.map(cookie => {
                  // Remove Domain attribute or rewrite it
                  let newCookie = cookie
                    .replace(/Domain=[^;]+;?/gi, '')
                    .replace(/Secure;?/gi, '')  // Remove Secure for localhost (http)
                    .replace(/SameSite=\w+;?/gi, 'SameSite=Lax;');  // Use Lax for localhost
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
