import { createPinia } from "pinia";
import { createApp } from "vue";
import App from "./App.vue";
import router from "@/router/index.js";
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
app.use(router);
// app.use(VueCropper);
app.directive("stop-propagation", stopPropagation);
app.directive("right-click", rightClick);
app.directive("tooltip", tooltip);
app.directive("esc", escDirective);
app.directive("return", returnDirective);
app.directive("focus", focusDirective);
app.mount("#app");
