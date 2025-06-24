import { Logger } from "@/core/logger";
import { FormApiService } from "./form-api-service";
import { FormUIService } from "./form-ui-service";
import type { FormSchema } from "@/shared/types/form-types";

/**
 * Main form service that orchestrates API and UI operations
 * This is now a facade that delegates to focused services
 */
export class FormService {
  private static instance: FormService;
  private apiService: FormApiService;
  private uiService: FormUIService;

  private constructor() {
    this.apiService = FormApiService.getInstance();
    this.uiService = FormUIService.getInstance();

    // Initialize UI handlers
    this.uiService.initializeFormDeletionHandlers(
      this.apiService.deleteForm.bind(this.apiService),
    );
  }

  public static getInstance(): FormService {
    if (!FormService.instance) {
      FormService.instance = new FormService();
    }
    return FormService.instance;
  }

  public setBaseUrl(url: string): void {
    this.apiService.setBaseUrl(url);
  }

  // Delegate API operations to FormApiService
  async getSchema(formId: string): Promise<FormSchema> {
    return this.apiService.getSchema(formId);
  }

  async saveSchema(formId: string, schema: any): Promise<any> {
    return this.apiService.saveSchema(formId, schema);
  }

  async updateFormDetails(
    formId: string,
    details: { title: string; description: string },
  ): Promise<void> {
    await this.apiService.updateFormDetails(formId, details);
    this.uiService.updateFormCard(formId, details);
    this.uiService.showSuccess("Form updated successfully");
  }

  public async deleteForm(formId: string): Promise<void> {
    return this.apiService.deleteForm(formId);
  }

  async submitForm(formId: string, data: FormData): Promise<Response> {
    return this.apiService.submitForm(formId, data);
  }

  // Expose UI service methods for external use
  showSuccess(message: string): void {
    this.uiService.showSuccess(message);
  }

  updateFormCard(
    formId: string,
    updates: { title?: string; description?: string },
  ): void {
    this.uiService.updateFormCard(formId, updates);
  }
}

// Re-export the FormSchema type for backward compatibility
export type { FormSchema } from "@/shared/types/form-types";

// Initialize form deletion handlers
document.addEventListener("DOMContentLoaded", () => {
  const formService = FormService.getInstance();

  document.querySelectorAll(".delete-form").forEach((button) => {
    button.addEventListener("click", async (e) => {
      e.preventDefault();
      const formId = button.getAttribute("data-form-id");
      if (!formId) return;

      if (
        !confirm(
          "Are you sure you want to delete this form? This action cannot be undone.",
        )
      ) {
        return;
      }

      try {
        await formService.deleteForm(formId);
        const formCard = button.closest(".form-card");
        if (formCard) {
          formCard.remove();
        }

        // If no forms left, show empty state
        const formsGrid = document.querySelector(".forms-grid");
        if (formsGrid && !formsGrid.querySelector(".form-card")) {
          formsGrid.innerHTML = `
            <div class="empty-state">
              <i class="bi bi-file-earmark-text"></i>
              <p>You haven't created any forms yet.</p>
              <a href="/forms/new" class="btn btn-primary">Create Your First Form</a>
            </div>
          `;
        }
      } catch (error) {
        Logger.error("Failed to delete form:", error);
        alert(error instanceof Error ? error.message : "Failed to delete form");
      }
    });
  });
});
