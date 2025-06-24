import { FormService } from "../forms/services/form-service";
import { dom } from "../shared/utils/dom-utils";

/**
 * Event handlers setup
 */
export function setupEventHandlers(builder: any, formId: string): void {
  const viewSchemaBtn = dom.getElement<HTMLButtonElement>("view-schema-btn");
  if (viewSchemaBtn) {
    viewSchemaBtn.addEventListener("click", async () => {
      try {
        // Get the current schema
        const schema = await builder.saveSchema();
        if (!schema) {
          throw new Error("Failed to get form schema");
        }

        // Show the schema in a formatted way
        const schemaString = JSON.stringify(schema, null, 2);
        // Create a modal to display the schema
        showSchemaModal(schemaString);
      } catch (error) {
        console.error("Failed to get schema:", error);
        dom.showError("Failed to get form schema. Please try again.");
      }
    });
  }

  const saveBtn = dom.getElement<HTMLButtonElement>("save-fields-btn");
  const feedback = dom.getElement<HTMLElement>("schema-save-feedback");
  if (saveBtn && feedback) {
    saveBtn.addEventListener("click", async () => {
      const spinner = saveBtn.querySelector(".spinner") as HTMLElement;
      try {
        feedback.textContent = "Saving...";
        feedback.className = "schema-save-feedback";
        saveBtn.disabled = true;
        if (spinner) spinner.style.display = "inline-block";

        // Get the current schema using saveSchema
        const schema = await builder.saveSchema();
        if (!schema) {
          throw new Error("Failed to get form schema");
        }

        // Save using form service
        const formService = FormService.getInstance();
        await formService.saveSchema(formId, schema);

        feedback.textContent = "Schema saved successfully.";
        feedback.className = "schema-save-feedback success";
      } catch (error) {
        console.error("Failed to save form fields:", error);
        feedback.textContent =
          error instanceof Error ? error.message : "Error saving schema.";
        feedback.className = "schema-save-feedback error";
      } finally {
        saveBtn.disabled = false;
        if (spinner) spinner.style.display = "none";
        setTimeout(() => {
          feedback.textContent = "";
          feedback.className = "schema-save-feedback";
        }, 3000);
      }
    });
  }

  // For new form creation, update the hidden schema field before form submission
  if (formId === "new") {
    const form = dom.getElement<HTMLFormElement>("new-form");
    if (form) {
      // Only add the submit handler if we're on the form builder page (not the new form page)
      const isFormBuilderPage = dom.getElement<HTMLElement>(
        "form-schema-builder",
      );
      if (isFormBuilderPage) {
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
    }
  }
}

/**
 * Show schema in a modal dialog
 */
function showSchemaModal(schemaString: string): void {
  // Remove existing modal if present
  const existingModal = document.getElementById("schema-modal");
  if (existingModal) {
    existingModal.remove();
  }

  // Create modal container
  const modal = dom.createElement<HTMLDivElement>("div", "schema-modal");
  modal.id = "schema-modal";
  modal.style.cssText = `
    position: fixed;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    background: rgba(0, 0, 0, 0.5);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 10000;
  `;

  // Create modal content
  const modalContent = dom.createElement<HTMLDivElement>(
    "div",
    "schema-modal-content",
  );
  modalContent.style.cssText = `
    background: var(--background-color, #ffffff);
    color: var(--text-color, #000000);
    border-radius: 8px;
    padding: 20px;
    max-width: 90%;
    max-height: 90%;
    overflow: auto;
    box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
    border: 1px solid var(--border-color, #e5e7eb);
  `;

  // Create modal header
  const modalHeader = dom.createElement<HTMLDivElement>(
    "div",
    "schema-modal-header",
  );
  modalHeader.style.cssText = `
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 20px;
    padding-bottom: 10px;
    border-bottom: 1px solid var(--border-color, #e5e7eb);
  `;

  const modalTitle = dom.createElement<HTMLHeadingElement>("h3");
  modalTitle.textContent = "Form Schema";
  modalTitle.style.cssText = `
    margin: 0;
    font-size: 18px;
    font-weight: 600;
    color: var(--text-color, #000000);
  `;

  const closeBtn = dom.createElement<HTMLButtonElement>("button");
  closeBtn.textContent = "×";
  closeBtn.style.cssText = `
    background: none;
    border: none;
    font-size: 24px;
    cursor: pointer;
    padding: 0;
    width: 30px;
    height: 30px;
    display: flex;
    align-items: center;
    justify-content: center;
    border-radius: 4px;
    color: var(--text-color, #000000);
    transition: color 0.2s ease;
  `;
  closeBtn.addEventListener("click", () => modal.remove());
  closeBtn.addEventListener("mouseenter", () => {
    closeBtn.style.color = "var(--primary-color, #3b82f6)";
  });
  closeBtn.addEventListener("mouseleave", () => {
    closeBtn.style.color = "var(--text-color, #000000)";
  });

  modalHeader.appendChild(modalTitle);
  modalHeader.appendChild(closeBtn);

  // Create schema content
  const schemaContent = dom.createElement<HTMLPreElement>("pre");
  schemaContent.textContent = schemaString;
  schemaContent.style.cssText = `
    background: var(--background-alt, #f8f9fa);
    color: var(--text-color, #000000);
    border: 1px solid var(--border-color, #e5e7eb);
    border-radius: 4px;
    padding: 16px;
    margin: 0;
    font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
    font-size: 12px;
    line-height: 1.5;
    overflow-x: auto;
    white-space: pre-wrap;
    word-wrap: break-word;
  `;

  // Assemble modal
  modalContent.appendChild(modalHeader);
  modalContent.appendChild(schemaContent);
  modal.appendChild(modalContent);

  // Add to document
  document.body.appendChild(modal);

  // Close modal when clicking outside
  modal.addEventListener("click", (e) => {
    if (e.target === modal) {
      modal.remove();
    }
  });

  // Close modal with Escape key
  const handleEscape = (e: KeyboardEvent) => {
    if (e.key === "Escape") {
      modal.remove();
      document.removeEventListener("keydown", handleEscape);
    }
  };
  document.addEventListener("keydown", handleEscape);
}
