import "@fortawesome/fontawesome-free/css/all.min.css";
import { validation } from "./validation";
import "./form-builder-sidebar";

// Main entry point for global initialization
document.addEventListener("DOMContentLoaded", () => {
  // Setup real-time validation for forms
  const forms = document.querySelectorAll("form[data-validate]");
  forms.forEach((form) => {
    const schemaName = form.getAttribute("data-validate");
    if (schemaName && (schemaName === "signup" || schemaName === "login")) {
      validation.setupRealTimeValidation(form.id, schemaName);
    }
  });
});

// Enable HMR
if (import.meta.hot) {
  import.meta.hot.accept();
}
