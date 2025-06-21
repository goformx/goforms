/**
 * Signup Form Handler
 *
 * This module handles the signup form functionality using the enhanced form handler
 */

import { EnhancedFormHandler } from "../forms/handlers/enhanced-form-handler";

// Initialize form when DOM is ready
document.addEventListener("DOMContentLoaded", () => {
  new EnhancedFormHandler({
    formId: "signup-form",
    validationType: "signup",
  });
});
