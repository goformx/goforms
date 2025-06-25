/**
 * Login Form Handler
 *
 * Initializes and configures the login form using the enhanced form handler.
 */

import { Logger } from "@/core/logger";
import { EnhancedFormHandler } from "@/features/forms/handlers/enhanced-form-handler";
import type { FormConfig } from "@/shared/types/form-types";

Logger.debug("login.ts: Script loaded and executing");
Logger.debug("login.ts: Imports completed");

// Initialize the login form handler
document.addEventListener("DOMContentLoaded", () => {
  Logger.debug("login.ts: DOMContentLoaded event fired");

  try {
    const config: FormConfig = {
      formId: "user-login",
      validationType: "onSubmit",
    };

    Logger.debug("login.ts: Creating EnhancedFormHandler with config:", config);
    new EnhancedFormHandler(config);
    Logger.debug("login.ts: EnhancedFormHandler created successfully");
  } catch (error) {
    Logger.error("login.ts: Error creating EnhancedFormHandler:", error);
  }
});

Logger.debug("login.ts: Script execution completed");
