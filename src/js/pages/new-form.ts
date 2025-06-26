/**
 * New Form Handler
 *
 * Initializes and configures the new form creation using the enhanced form handler.
 */

import { Logger } from "@/core/logger";
import { EnhancedFormHandler } from "@/features/forms/handlers/enhanced-form-handler";
import type { FormConfig } from "@/shared/types/form-types";
import { createFormId } from "@/shared/types/form-types";

Logger.debug("new-form.ts: Script loaded and executing");

// Initialize the new form handler
document.addEventListener("DOMContentLoaded", () => {
  Logger.debug("new-form.ts: DOMContentLoaded event fired");

  try {
    const formConfig: FormConfig = {
      formId: createFormId("new-form"),
      validationType: "realtime",
      options: {
        autoSave: true,
        showProgress: true,
      },
    };

    Logger.debug(
      "new-form.ts: Creating EnhancedFormHandler with config:",
      formConfig,
    );
    new EnhancedFormHandler(formConfig);
    Logger.debug("new-form.ts: EnhancedFormHandler created successfully");
  } catch (error) {
    Logger.error("new-form.ts: Error creating EnhancedFormHandler:", error);
  }
});

Logger.debug("new-form.ts: Script execution completed");
