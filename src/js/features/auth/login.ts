/**
 * Login Form Handler
 *
 * Initializes and configures the login form using the enhanced form handler.
 */

console.log("login.ts: Script loaded and executing");

import { EnhancedFormHandler } from "@/features/forms/handlers/enhanced-form-handler";
import type { FormConfig } from "@/shared/types/form-types";

console.log("login.ts: Imports completed");

// Initialize the login form handler
document.addEventListener("DOMContentLoaded", () => {
  console.log("login.ts: DOMContentLoaded event fired");

  try {
    const config: FormConfig = {
      formId: "login-form",
      validationType: "login",
    };

    console.log("login.ts: Creating EnhancedFormHandler with config:", config);
    new EnhancedFormHandler(config);
    console.log("login.ts: EnhancedFormHandler created successfully");
  } catch (error) {
    console.error("login.ts: Error creating EnhancedFormHandler:", error);
  }
});

console.log("login.ts: Script execution completed");
