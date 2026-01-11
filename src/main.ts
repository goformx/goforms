import { createApp, h } from "vue";
import { createInertiaApp } from "@inertiajs/vue3";
import { Toaster } from "@/components/ui/sonner";
import "./assets/css/main.css";

// Type for page modules
type PageModule = { default: object };
type PageResolver = Record<string, () => Promise<PageModule>>;

// Import all page components
const pages = import.meta.glob<PageModule>("./pages/**/*.vue");

void createInertiaApp({
  resolve: (name: string) => {
    const resolver = pages as PageResolver;
    const match = resolver[`./pages/${name}.vue`];
    if (!match) {
      throw new Error(`Page not found: ${name}`);
    }
    return match();
  },
  setup({ el, App, props, plugin }) {
    const app = createApp({
      render: () =>
        h("div", [
          h(App, props),
          h(Toaster, {
            position: "top-right",
            richColors: true,
          }),
        ]),
    });

    app.use(plugin).mount(el);
  },
});
