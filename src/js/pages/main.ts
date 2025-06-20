import "@fortawesome/fontawesome-free/css/all.min.css";
import "../components/form-builder-sidebar";
// @ts-expect-error Vite module preload polyfill
import "vite/modulepreload-polyfill";

// Enable HMR
if (import.meta.hot) {
  import.meta.hot.accept();
}
