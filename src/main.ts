import { createApp, h } from "vue";
import { createInertiaApp } from "@inertiajs/vue3";
import "./assets/css/main.css";

// Type for page modules
type PageModule = { default: object };
type PageResolver = Record<string, () => Promise<PageModule>>;

// Import all page components
const pages = import.meta.glob<PageModule>("./pages/**/*.vue");

createInertiaApp({
  resolve: (name: string) => {
    const resolver = pages as PageResolver;
    const match = resolver[`./pages/${name}.vue`];
    if (!match) {
      throw new Error(`Page not found: ${name}`);
    }
    return match();
  },
  setup({ el, App, props, plugin }) {
    createApp({ render: () => h(App, props) })
      .use(plugin)
      .mount(el);
  },
});
