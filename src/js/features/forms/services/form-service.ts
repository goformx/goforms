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
