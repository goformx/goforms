/**
 * Signup Form Handler
 *
 * Initializes and configures the signup form using the enhanced form handler.
 */

import { Logger } from "@/core/logger";
import { EnhancedFormHandler } from "@/features/forms/handlers/enhanced-form-handler";
import type { FormConfig } from "@/shared/types/form-types";

// Initialize the signup form handler
document.addEventListener("DOMContentLoaded", () => {
  try {
    const config: FormConfig = {
      formId: "user-signup",
      validationType: "onSubmit",
    };

    new EnhancedFormHandler(config);
  } catch (error) {
    Logger.error("Error creating EnhancedFormHandler:", error);
  }
});
