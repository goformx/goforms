import "@fortawesome/fontawesome-free/css/all.min.css";
import "./features/forms/components/form-builder/sidebar";
import { UserDropdownManager } from "@/features/user/components/user-dropdown";
// @ts-expect-error Vite module preload polyfill
import "vite/modulepreload-polyfill";

// Enable HMR
if (import.meta.hot) {
  import.meta.hot.accept();
}

// Initialize when DOM is ready
document.addEventListener("DOMContentLoaded", () => {
  new UserDropdownManager();
});
