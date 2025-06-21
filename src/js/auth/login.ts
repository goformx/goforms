/**
 * Login Form Handler
 *
 * Initializes and configures the login form using the enhanced form handler.
 */

import { EnhancedFormHandler } from "./handlers/enhanced-form-handler";

// Initialize form when DOM is ready
document.addEventListener("DOMContentLoaded", () => {
  new EnhancedFormHandler({
    formId: "login-form",
    validationType: "login",
  });
});
