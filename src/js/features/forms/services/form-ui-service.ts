import { Logger } from "@/core/logger";

/**
 * Handles DOM manipulation and UI operations for forms
 */
export class FormUIService {
  private static instance: FormUIService;

  private constructor() {}

  public static getInstance(): FormUIService {
    if (!FormUIService.instance) {
      FormUIService.instance = new FormUIService();
    }
    return FormUIService.instance;
  }

  /**
   * Initialize form deletion handlers
   */
  initializeFormDeletionHandlers(
    deleteFormCallback: (formId: string) => Promise<void>,
  ): void {
    document.addEventListener("DOMContentLoaded", () => {
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
            await deleteFormCallback(formId);
            this.removeFormCard(button);
            this.checkEmptyState();
          } catch (error) {
            Logger.error("Failed to delete form:", error);
            this.showError("Failed to delete form. Please try again.");
          }
        });
      });
    });
  }

  /**
   * Remove form card from DOM
   */
  private removeFormCard(button: Element): void {
    const formCard = button.closest(".form-card");
    if (formCard) {
      formCard.remove();
    }
  }

  /**
   * Check if forms grid is empty and show empty state
   */
  private checkEmptyState(): void {
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
  }

  /**
   * Show error message to user
   */
  private showError(message: string): void {
    // You can implement a more sophisticated error display system here
    Logger.error(message);
    // For now, just log to console. In a real app, you'd show a toast or modal
  }

  /**
   * Show success message to user
   */
  showSuccess(message: string): void {
    Logger.debug(message);
    // Implement success message display
  }

  /**
   * Update form card in the UI
   */
  updateFormCard(
    formId: string,
    updates: { title?: string; description?: string },
  ): void {
    const formCard = document
      .querySelector(`[data-form-id="${formId}"]`)
      ?.closest(".form-card");
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
