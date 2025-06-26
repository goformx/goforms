/**
 * Signup Form Handler
 *
 * Initializes and configures the signup form using the enhanced form handler.
 */

import { Logger } from "@/core/logger";
import { EnhancedFormHandler } from "@/features/forms/handlers/enhanced-form-handler";
import type { FormConfig } from "@/shared/types/form-types";
import { createFormId } from "@/shared/types/form-types";

// Initialize the signup form handler
document.addEventListener("DOMContentLoaded", () => {
  try {
    const formConfig: FormConfig = {
      formId: createFormId("user-signup"),
      validationType: "realtime",
      options: {
        autoSave: false,
        showProgress: true,
      },
    };

    new EnhancedFormHandler(formConfig);
  } catch (error) {
    Logger.error("Error creating EnhancedFormHandler:", error);
  }
});
