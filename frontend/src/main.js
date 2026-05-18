import * as WailsRuntime from "@wailsio/runtime";
window.wails = WailsRuntime;

import { createPinia } from "pinia";
import { createApp } from "vue";
import App from "./App.vue";
import { AppService } from "@/services";
import { applyTheme } from "@/theme/apply";

// Apply a default palette immediately so the very first paint already uses
// the design-token system. The real values (from user settings) are loaded
// later by `useThemeStore.initialize()`.
applyTheme({ mode: "system", tint: "neutral" });

// Tag <html> with the host OS so CSS can target platform-specific quirks
// (e.g. missing backdrop-filter rendering on WebKitGTK / Linux).
AppService.GetOS()
  .then((os) => { document.documentElement.dataset.os = os; })
  .catch(() => {});
import router from "./router";
import i18n from "./i18n";
import { stopPropagation } from "./directives.js";
import { rightClick } from "./directives.js";
import { escDirective } from "./directives.js";
import { returnDirective } from "./directives.js";
import { focusDirective } from "./directives.js";
import tooltip from "./directives/tooltip.js";
// import { devtools } from "@vue/devtools";

if (process.env.NODE_ENV === "development") {
  // devtools.connect("http://127.0.0.1", 8098);
}

import "./assets/global.css";

const app = createApp(App);
app.use(createPinia());
app.use(i18n);

// Always use router for both desktop and web mode
app.use(router);

// app.use(VueCropper);
app.directive("stop-propagation", stopPropagation);
app.directive("right-click", rightClick);
app.directive("tooltip", tooltip);
app.directive("esc", escDirective);
app.directive("return", returnDirective);
app.directive("focus", focusDirective);
app.mount("#app");
