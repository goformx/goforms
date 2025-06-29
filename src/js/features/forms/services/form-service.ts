import { Logger } from "@/core/logger";
import { FormApiService } from "./form-api-service";
import type { FormSchema } from "@/shared/types/form-types";

/**
 * Main form service that orchestrates API and UI operations
 * This is now a facade that delegates to focused services
 */
export class FormService {
  private static instance: FormService;
  private readonly apiService: FormApiService;

  private constructor() {
    this.apiService = FormApiService.getInstance();
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
    this.updateFormCard(formId, details);
    this.showSuccess("Form updated successfully");
  }

  public async deleteForm(formId: string): Promise<void> {
    return this.apiService.deleteForm(formId);
  }

  async submitForm(formId: string, data: FormData): Promise<Response> {
    return this.apiService.submitForm(formId, data);
  }

  // UI service methods
  showSuccess(message: string): void {
    // Find success container and show message
    const successContainer = document.querySelector(".success-message");
    if (successContainer) {
      successContainer.textContent = message;
      successContainer.classList.remove("hidden");

      // Auto-hide after 3 seconds
      setTimeout(() => {
        successContainer.classList.add("hidden");
      }, 3000);
    }
  }

  updateFormCard(
    formId: string,
    updates: { title?: string; description?: string },
  ): void {
    // Find the form card and update its content
    const formCard = document.querySelector(`[data-form-id="${formId}"]`);
    if (formCard) {
      if (updates.title) {
        const titleElement = formCard.querySelector(".form-title");
        if (titleElement) {
          titleElement.textContent = updates.title;
        }
      }

      if (updates.description) {
        const descElement = formCard.querySelector(".form-description");
        if (descElement) {
          descElement.textContent = updates.description;
        }
      }
    }
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
        const formCard = button.closest(".form-panel");
        if (formCard) {
          formCard.remove();
        }

        // If no forms left, show empty state
        const formsGrid = document.querySelector(".forms-grid");
        if (formsGrid && !formsGrid.querySelector(".form-panel")) {
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
