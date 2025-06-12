/**
 * Signup Form Handler
 *
 * This module handles the signup form functionality including:
 * - Real-time validation as users type
 * - Form submission via AJAX
 * - Error handling and display
 * - Server response handling
 */

import { setupForm } from "./form-handler";

// Initialize form when DOM is ready
document.addEventListener("DOMContentLoaded", () => {
  setupForm({
    formId: "signup-form",
    validationType: "signup",
  });
});
