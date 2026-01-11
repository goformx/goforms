import { createApp, h, type DefineComponent } from "vue";
import { createInertiaApp } from "@inertiajs/vue3";
import { Toaster } from "@/components/ui/sonner";
import "./assets/css/main.css";

// Type for page modules with proper Vue component type
type PageModule = { default: DefineComponent };

// Import all page components
const pages = import.meta.glob<PageModule>("./pages/**/*.vue");

void createInertiaApp({
  resolve: async (name: string) => {
    const match = pages[`./pages/${name}.vue`];
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
