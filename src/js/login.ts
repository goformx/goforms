/**
 * Login Form Handler
 *
 * Initializes and configures the login form using the shared form handler.
 */

import { setupForm } from "./form-handler";

// Initialize form when DOM is ready
document.addEventListener("DOMContentLoaded", () => {
  setupForm({
    formId: "login-form",
    validationType: "login",
  });
});
