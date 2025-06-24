import { FormService } from "../forms/services/form-service";
import { dom } from "../utils/dom-utils";
import { showSchemaModal } from "../components/form-builder/schema-modal";

/**
 * Event handlers setup
 */
export function setupEventHandlers(builder: any, formId: string): void {
  setupViewSchemaHandler(builder);
  setupSaveHandler(builder, formId);
  setupNewFormHandler(builder, formId);
}

function setupViewSchemaHandler(builder: any): void {
  const viewSchemaBtn = dom.getElement<HTMLButtonElement>("view-schema-btn");
  if (!viewSchemaBtn) return;

  viewSchemaBtn.addEventListener("click", async () => {
    try {
      // Get the current schema
      const schema = await builder.saveSchema();
      if (!schema) {
        throw new Error("Failed to get form schema");
      }

      // Show the schema in a formatted way
      const schemaString = JSON.stringify(schema, null, 2);
      showSchemaModal(schemaString);
    } catch (error) {
      console.error("Failed to get schema:", error);
      dom.showError("Failed to get form schema. Please try again.");
    }
  });
}

function setupSaveHandler(builder: any, formId: string): void {
  const saveBtn = dom.getElement<HTMLButtonElement>("save-fields-btn");
  const feedback = dom.getElement<HTMLElement>("schema-save-feedback");

  if (!saveBtn || !feedback) return;

  saveBtn.addEventListener("click", async () => {
    const spinner = saveBtn.querySelector(".spinner") as HTMLElement;

    try {
      // Update UI state
      updateSaveUIState(feedback, saveBtn, spinner, "saving");

      // Get the current schema using saveSchema
      const schema = await builder.saveSchema();
      if (!schema) {
        throw new Error("Failed to get form schema");
      }

      // Save using form service
      const formService = FormService.getInstance();
      await formService.saveSchema(formId, schema);

      // Success state
      updateSaveUIState(feedback, saveBtn, spinner, "success");
    } catch (error) {
      console.error("Failed to save form fields:", error);
      const errorMessage =
        error instanceof Error ? error.message : "Error saving schema.";
      updateSaveUIState(feedback, saveBtn, spinner, "error", errorMessage);
    } finally {
      // Reset UI after 3 seconds
      setTimeout(() => {
        updateSaveUIState(feedback, saveBtn, spinner, "idle");
      }, 3000);
    }
  });
}

function setupNewFormHandler(builder: any, formId: string): void {
  // For new form creation, update the hidden schema field before form submission
  if (formId !== "new") return;

  const form = dom.getElement<HTMLFormElement>("new-form");
  if (!form) return;

  // Only add the submit handler if we're on the form builder page
  const isFormBuilderPage = dom.getElement<HTMLElement>("form-schema-builder");
  if (!isFormBuilderPage) return;

  form.addEventListener("submit", async (e) => {
    e.preventDefault();
    try {
      const schema = await builder.saveSchema();
      const schemaInput = dom.getElement<HTMLInputElement>("schema");
      if (schemaInput) {
        schemaInput.value = JSON.stringify(schema);
      }
      form.submit();
    } catch (_error) {
      dom.showError("Failed to save form schema. Please try again.");
    }
  });
}

type SaveUIState = "idle" | "saving" | "success" | "error";

function updateSaveUIState(
  feedback: HTMLElement,
  saveBtn: HTMLButtonElement,
  spinner: HTMLElement | null,
  state: SaveUIState,
  message?: string,
): void {
  switch (state) {
    case "saving":
      feedback.textContent = "Saving...";
      feedback.className = "schema-save-feedback";
      saveBtn.disabled = true;
      if (spinner) spinner.style.display = "inline-block";
      break;

    case "success":
      feedback.textContent = "Schema saved successfully.";
      feedback.className = "schema-save-feedback success";
      saveBtn.disabled = false;
      if (spinner) spinner.style.display = "none";
      break;

    case "error":
      feedback.textContent = message || "Error saving schema.";
      feedback.className = "schema-save-feedback error";
      saveBtn.disabled = false;
      if (spinner) spinner.style.display = "none";
      break;

    case "idle":
      feedback.textContent = "";
      feedback.className = "schema-save-feedback";
      saveBtn.disabled = false;
      if (spinner) spinner.style.display = "none";
      break;
  }
}
