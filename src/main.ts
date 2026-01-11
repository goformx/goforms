import { createApp, h, type DefineComponent } from "vue";
import { createInertiaApp } from "@inertiajs/vue3";
import { Toaster } from "@/components/ui/sonner";
import "vue-sonner/style.css";
import "./assets/css/main.css";

// Import all page components with lazy loading
const pages = import.meta.glob<{ default: DefineComponent }>(
  "./pages/**/*.vue",
);

void createInertiaApp({
  resolve: async (name: string) => {
    const match = pages[`./pages/${name}.vue`];
    if (!match) {
      throw new Error(`Page not found: ${name}`);
    }
    // Await the dynamic import and return the default export
    const module = await match();
    return module.default;
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
