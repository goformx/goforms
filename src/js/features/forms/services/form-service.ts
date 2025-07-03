import { Logger } from "@/core/logger";
import { FormApiService } from "./form-api-service";
import type { FormSchema } from "@/shared/types/form-types";

/**
 * Main form service that orchestrates all form-related operations
 * Provides complete CRUD functionality with UI integration
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

  // =============================================
  // CREATE OPERATIONS
  // =============================================

  /**
   * Create a new form with basic details
   */
  public async createForm(
    title: string,
    description?: string,
  ): Promise<{ formId: string }> {
    try {
      Logger.group("Form Creation");
      Logger.debug("Creating new form:", { title, description });

      const formData: { title: string; description?: string } = { title };
      if (description) {
        formData.description = description;
      }
      const response = await this.apiService.createForm(formData);

      Logger.debug("Form created successfully:", response);
      Logger.groupEnd();

      this.showSuccess("Form created successfully");
      return response;
    } catch (error: unknown) {
      Logger.group("Form Creation Error");
      Logger.error("Failed to create form:", error);
      Logger.groupEnd();

      throw error;
    }
  }

  /**
   * Create form and redirect to edit page
   */
  public async createFormWithRedirect(
    title: string,
    description?: string,
  ): Promise<void> {
    try {
      const { formId } = await this.createForm(title, description);
      window.location.href = `/forms/${formId}/edit`;
    } catch (error: unknown) {
      alert(
        error instanceof Error
          ? error.message
          : "Failed to create form. Please try again.",
      );
    }
  }

  // =============================================
  // READ OPERATIONS
  // =============================================

  /**
   * Get form schema
   */
  public async getSchema(formId: string): Promise<FormSchema> {
    return this.apiService.getSchema(formId);
  }

  /**
   * Get complete form data
   */
  public async getForm(formId: string): Promise<any> {
    return this.apiService.getForm(formId);
  }

  /**
   * Get user's forms list
   */
  public async listForms(): Promise<any[]> {
    return this.apiService.listForms();
  }

  /**
   * Get form details (metadata only)
   */
  public async getFormDetails(formId: string): Promise<any> {
    return this.apiService.getFormDetails(formId);
  }

  // =============================================
  // UPDATE OPERATIONS
  // =============================================

  /**
   * Save form schema
   */
  public async saveSchema(formId: string, schema: any): Promise<any> {
    return this.apiService.saveSchema(formId, schema);
  }

  /**
   * Update form details (title, description)
   */
  public async updateFormDetails(
    formId: string,
    details: { title: string; description: string },
  ): Promise<void> {
    await this.apiService.updateFormDetails(formId, details);
    this.updateFormCard(formId, details);
    this.showSuccess("Form updated successfully");
  }

  /**
   * Update form status (draft, published, archived)
   */
  public async updateFormStatus(formId: string, status: string): Promise<void> {
    await this.apiService.updateFormStatus(formId, status);
    this.updateFormCard(formId, { status });
    this.showSuccess(`Form ${status} successfully`);
  }

  /**
   * Update form CORS settings
   */
  public async updateFormCors(
    formId: string,
    corsOrigins: string,
  ): Promise<void> {
    await this.apiService.updateFormCors(formId, corsOrigins);
    this.showSuccess("CORS settings updated successfully");
  }

  // =============================================
  // DELETE OPERATIONS
  // =============================================

  /**
   * Basic form deletion (API only)
   */
  public async deleteForm(formId: string): Promise<void> {
    return this.apiService.deleteForm(formId);
  }

  /**
   * Delete form with UI confirmation and error handling
   * This is the main method that should be called from UI components
   */
  public async deleteFormWithConfirmation(formId: string): Promise<void> {
    // Show confirmation dialog
    if (
      !confirm(
        "Are you sure you want to delete this form? This action cannot be undone.",
      )
    ) {
      return;
    }

    try {
      Logger.group("Form Deletion");
      Logger.debug("Starting form deletion for ID:", formId);

      // Call the API to delete the form
      await this.deleteForm(formId);

      Logger.debug("Form deleted successfully, removing from UI");
      Logger.groupEnd();

      // Remove the form card from the UI
      this.removeFormCard(formId);

      // Show success message
      this.showSuccess("Form deleted successfully");

      // If no forms left, show empty state
      this.checkAndShowEmptyState();
    } catch (error: unknown) {
      Logger.group("Form Deletion Error");
      Logger.error("Failed to delete form:", error);
      Logger.groupEnd();

      // Show error message to user
      alert(
        error instanceof Error
          ? error.message
          : "Failed to delete form. Please try again.",
      );
    }
  }

  // =============================================
  // SUBMISSION OPERATIONS
  // =============================================

  /**
   * Submit form data
   */
  public async submitForm(formId: string, data: FormData): Promise<Response> {
    return this.apiService.submitForm(formId, data);
  }

  // =============================================
  // UI SERVICE METHODS
  // =============================================

  /**
   * Show success message
   */
  public showSuccess(message: string): void {
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

  /**
   * Update form card in the UI
   */
  public updateFormCard(
    formId: string,
    updates: { title?: string; description?: string; status?: string },
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

      if (updates.status) {
        const statusElement = formCard.querySelector(".form-status");
        if (statusElement) {
          statusElement.textContent = updates.status;
          statusElement.className = `form-status form-status--${updates.status}`;
        }
      }
    }
  }

  /**
   * Remove a form card from the UI
   */
  private removeFormCard(formId: string): void {
    const formCard = document.querySelector(`[data-form-id="${formId}"]`);
    if (formCard) {
      formCard.remove();
    }
  }

  /**
   * Check if there are any forms left and show empty state if needed
   */
  private checkAndShowEmptyState(): void {
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
  }
}

// Re-export the FormSchema type for backward compatibility
export type { FormSchema } from "@/shared/types/form-types";
