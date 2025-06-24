import { Logger } from "@/core/logger";
import { showSchemaModal } from "@/features/forms/components/form-builder/schema-modal";

/**
 * Set up the View Schema button event handler
 */
export function setupViewSchemaButton(builder: any): void {
  const viewSchemaBtn = document.getElementById("view-schema-btn");
  if (!viewSchemaBtn) {
    Logger.warn("View Schema button not found");
    return;
  }

  viewSchemaBtn.addEventListener("click", () => {
    try {
      // Get current schema from builder
      const currentSchema = builder.form;
      const schemaString = JSON.stringify(currentSchema, null, 2);
      showSchemaModal(schemaString);
      Logger.debug("Schema modal opened");
    } catch (error) {
      Logger.error("Error showing schema modal:", error);
    }
  });

  Logger.debug("View Schema button event handler set up");
}
