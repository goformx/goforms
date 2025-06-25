import { Logger } from "@/core/logger";
import { FormService } from "@/features/forms/services/form-service";
import { formState } from "@/features/forms/state/form-state";
import { FormBuilderError, ErrorCode } from "@/core/errors/form-builder-error";

// Configuration constants
const SAVE_FIELDS_CONFIG = {
  BUTTON_ID: "save-fields-btn",
  FEEDBACK_ID: "schema-save-feedback",
  SUCCESS_TIMEOUT: 3000,
  ERROR_TIMEOUT: 5000,
  BUTTON_STATES: {
    IDLE: "Save Fields",
    SAVING: "Saving...",
  },
  STYLES: {
    SUCCESS: "#28a745",
    ERROR: "#dc3545",
  },
} as const;

// Types for better type safety
interface SaveFieldsElements {
  saveButton: HTMLButtonElement;
  feedbackElement: HTMLSpanElement;
  spinner: HTMLElement | null;
  buttonText: HTMLElement | null;
}

interface SaveFieldsState {
  isLoading: boolean;
  currentTimeout: NodeJS.Timeout | null;
}

/**
 * Enhanced Save Fields button handler with better architecture
 */
export class SaveFieldsHandler {
  private elements: SaveFieldsElements;
  private state: SaveFieldsState = {
    isLoading: false,
    currentTimeout: null,
  };
  private formService: FormService;
  private abortController: AbortController | null = null;

  constructor(
    private readonly formId: string,
    formService?: FormService,
  ) {
    this.formService = formService ?? FormService.getInstance();
    this.elements = this.getRequiredElements();
    this.setupEventListeners();
    Logger.debug("SaveFieldsHandler initialized for form:", formId);
  }

  /**
   * Get and validate required DOM elements
   */
  private getRequiredElements(): SaveFieldsElements {
    const saveButton = document.getElementById(
      SAVE_FIELDS_CONFIG.BUTTON_ID,
    ) as HTMLButtonElement;

    const feedbackElement = document.getElementById(
      SAVE_FIELDS_CONFIG.FEEDBACK_ID,
    ) as HTMLSpanElement;

    if (!saveButton) {
      throw new FormBuilderError(
        "Save Fields button not found",
        ErrorCode.FORM_NOT_FOUND,
        "Required form elements are missing. Please refresh the page.",
      );
    }

    if (!feedbackElement) {
      throw new FormBuilderError(
        "Schema save feedback element not found",
        ErrorCode.FORM_NOT_FOUND,
        "Required form elements are missing. Please refresh the page.",
      );
    }

    return {
      saveButton,
      feedbackElement,
      spinner: saveButton.querySelector(".spinner") as HTMLElement,
      buttonText: saveButton.querySelector("span:not(.spinner)") as HTMLElement,
    };
  }

  /**
   * Set up event listeners with proper cleanup
   */
  private setupEventListeners(): void {
    this.elements.saveButton.addEventListener("click", this.handleSaveClick);

    // Handle page unload cleanup
    window.addEventListener("beforeunload", this.cleanup);
  }

  /**
   * Handle save button click with arrow function to preserve context
   */
  private handleSaveClick = async (event: Event): Promise<void> => {
    event.preventDefault();

    if (this.state.isLoading) {
      Logger.warn("Save operation already in progress");
      return;
    }

    try {
      await this.saveFormFields();
    } catch (error) {
      this.handleSaveError(error);
    }
  };

  /**
   * Main save operation with better error handling
   */
  private async saveFormFields(): Promise<void> {
    // Cancel any previous operation
    this.abortController?.abort();
    this.abortController = new AbortController();

    try {
      this.setLoadingState(true);
      this.clearFeedback();

      const builder = this.getFormBuilder();
      const currentSchema = this.extractSchema(builder);

      await this.formService.saveSchema(this.formId, currentSchema);

      this.showSuccessFeedback();
      Logger.debug("Form schema saved successfully:", this.formId);
    } catch (error) {
      if (error instanceof DOMException && error.name === "AbortError") {
        Logger.debug("Save operation was cancelled");
        return;
      }
      throw error;
    } finally {
      this.setLoadingState(false);
    }
  }

  /**
   * Get form builder with validation
   */
  private getFormBuilder(): any {
    const builder = formState.get("formBuilder");

    if (!builder) {
      throw new FormBuilderError(
        "Form builder not found",
        ErrorCode.FORM_NOT_FOUND,
        "Form builder not found. Please refresh the page.",
      );
    }

    return builder;
  }

  /**
   * Extract and validate schema from builder
   */
  private extractSchema(builder: any): any {
    const schema = builder.form;

    if (!schema || typeof schema !== "object") {
      throw new FormBuilderError(
        "Invalid form schema",
        ErrorCode.SCHEMA_ERROR,
        "Form schema is invalid. Please check your form configuration.",
      );
    }

    // Basic schema validation
    if (!schema.components || !Array.isArray(schema.components)) {
      throw new FormBuilderError(
        "Form schema missing components",
        ErrorCode.SCHEMA_ERROR,
        "Form must contain at least one component.",
      );
    }

    return schema;
  }

  /**
   * Set loading state with UI updates
   */
  private setLoadingState(loading: boolean): void {
    this.state.isLoading = loading;
    this.elements.saveButton.disabled = loading;

    if (this.elements.spinner) {
      this.elements.spinner.style.display = loading ? "inline-block" : "none";
    }

    if (this.elements.buttonText) {
      this.elements.buttonText.textContent = loading
        ? SAVE_FIELDS_CONFIG.BUTTON_STATES.SAVING
        : SAVE_FIELDS_CONFIG.BUTTON_STATES.IDLE;
    }

    // Add loading class for CSS styling
    this.elements.saveButton.classList.toggle("loading", loading);
  }

  /**
   * Show success feedback with auto-clear
   */
  private showSuccessFeedback(): void {
    this.showFeedback(
      "Fields saved successfully!",
      "success",
      SAVE_FIELDS_CONFIG.STYLES.SUCCESS,
      SAVE_FIELDS_CONFIG.SUCCESS_TIMEOUT,
    );
  }

  /**
   * Handle and display save errors
   */
  private handleSaveError(error: unknown): void {
    Logger.group("Save Fields Error");

    if (error instanceof FormBuilderError) {
      Logger.error("FormBuilderError:", {
        code: error.code,
        message: error.message,
        userMessage: error.userMessage,
        context: error.context,
      });

      this.showFeedback(
        error.userMessage,
        "error",
        SAVE_FIELDS_CONFIG.STYLES.ERROR,
        SAVE_FIELDS_CONFIG.ERROR_TIMEOUT,
      );
    } else if (error instanceof Error) {
      Logger.error("Standard Error:", {
        name: error.name,
        message: error.message,
        stack: error.stack,
      });

      this.showFeedback(
        "Failed to save fields. Please try again.",
        "error",
        SAVE_FIELDS_CONFIG.STYLES.ERROR,
        SAVE_FIELDS_CONFIG.ERROR_TIMEOUT,
      );
    } else {
      Logger.error("Unknown error:", { type: typeof error, error });

      this.showFeedback(
        "An unexpected error occurred. Please refresh and try again.",
        "error",
        SAVE_FIELDS_CONFIG.STYLES.ERROR,
        SAVE_FIELDS_CONFIG.ERROR_TIMEOUT,
      );
    }

    Logger.groupEnd();
  }

  /**
   * Generic feedback display with auto-clear
   */
  private showFeedback(
    message: string,
    className: string,
    color: string,
    timeout: number,
  ): void {
    this.clearFeedback();

    this.elements.feedbackElement.textContent = message;
    this.elements.feedbackElement.className = className;
    this.elements.feedbackElement.style.color = color;

    // Auto-clear with cleanup of previous timeouts
    this.state.currentTimeout = setTimeout(() => {
      this.clearFeedback();
    }, timeout);
  }

  /**
   * Clear feedback display and timeouts
   */
  private clearFeedback(): void {
    if (this.state.currentTimeout) {
      clearTimeout(this.state.currentTimeout);
      this.state.currentTimeout = null;
    }

    this.elements.feedbackElement.textContent = "";
    this.elements.feedbackElement.className = "";
    this.elements.feedbackElement.style.color = "";
  }

  /**
   * Manual save trigger for external use
   */
  public async save(): Promise<void> {
    return this.saveFormFields();
  }

  /**
   * Check if save operation is in progress
   */
  public get isLoading(): boolean {
    return this.state.isLoading;
  }

  /**
   * Cancel current save operation
   */
  public cancel(): void {
    this.abortController?.abort();
    this.setLoadingState(false);
    this.clearFeedback();
  }

  /**
   * Cleanup resources and event listeners
   */
  private cleanup = (): void => {
    this.abortController?.abort();
    this.clearFeedback();
    this.elements.saveButton.removeEventListener("click", this.handleSaveClick);
    window.removeEventListener("beforeunload", this.cleanup);
  };

  /**
   * Destroy the handler and cleanup resources
   */
  public destroy(): void {
    this.cleanup();
    Logger.debug("SaveFieldsHandler destroyed for form:", this.formId);
  }
}

/**
 * Factory function for backwards compatibility and easier testing
 */
export function setupSaveFieldsButton(
  formId: string,
  formService?: FormService,
): SaveFieldsHandler {
  try {
    return new SaveFieldsHandler(formId, formService);
  } catch (error) {
    Logger.error("Failed to setup Save Fields button:", error);
    throw error;
  }
}

/**
 * Utility function to setup multiple save handlers
 */
export function setupMultipleSaveHandlers(
  formConfigs: Array<{ formId: string; formService?: FormService }>,
): SaveFieldsHandler[] {
  return formConfigs.map((config) =>
    setupSaveFieldsButton(config.formId, config.formService),
  );
}
