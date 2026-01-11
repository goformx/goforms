import { Logger } from "@/core/logger";
import { FormBuilderError } from "@/core/errors/form-builder-error";
import {
  validateForm,
  loginSchema,
  signupSchema,
  contactSchema,
} from "@/shared/validation";
import { debounce } from "@/shared/utils/debounce";
import { UIManager } from "./ui-manager";
import type { z } from "zod";

// Schema mapping for form validation
const SCHEMA_MAP = {
  "login-form": loginSchema,
  "user-login": loginSchema,
  "signup-form": signupSchema,
  "user-signup": signupSchema,
  "contact-form": contactSchema,
} as const;

type SchemaName = keyof typeof SCHEMA_MAP;

export class ValidationHandler {
  /**
   * Sets up real-time validation for form inputs with debouncing
   */
  static setupRealTimeValidation(form: HTMLFormElement, delay = 300): void {
    try {
      const inputs = form.querySelectorAll<HTMLInputElement>(
        "input[id], textarea[id], select[id]",
      );

      inputs.forEach((input) => {
        input.addEventListener(
          "input",
          debounce(() => this.handleInputValidation(input, form), delay),
        );

        // Also validate on blur for better UX
        input.addEventListener("blur", () =>
          this.handleInputValidation(input, form),
        );
      });

      Logger.debug(
        `Real-time validation setup for ${inputs.length} inputs in form: ${form.id}`,
      );
    } catch (_error) {
      throw FormBuilderError.schemaError(
        "Failed to setup real-time validation",
        form.id,
      );
    }
  }

  /**
   * Handles real-time validation for individual input fields
   */
  private static async handleInputValidation(
    input: HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement,
    form: HTMLFormElement,
  ): Promise<void> {
    try {
      // Clear previous errors for this field
      this.clearFieldError(input);
      UIManager.setAriaInvalid(input.id, false);

      // Get the validation schema for this form
      const schema = this.getSchemaForForm(form.id);
      if (!schema) {
        // No schema validation needed for this form
        return;
      }

      // Validate just this field
      const fieldValue = input.value;
      const fieldName = input.name || input.id;

      // Try to validate the individual field if it exists in the schema
      try {
        const fieldSchema = (schema as any).shape[fieldName];
        if (fieldSchema) {
          fieldSchema.parse(fieldValue);
          // Field is valid
          this.showFieldSuccess(input);
        }
      } catch (validationError) {
        if (validationError instanceof Error) {
          const errorMessage = this.extractErrorMessage(validationError);
          this.showFieldError(input, errorMessage);
          UIManager.setAriaInvalid(input.id, true);
        }
      }
    } catch (error) {
      Logger.error("Input validation error:", error);
      // Don't show error to user for real-time validation failures
    }
  }

  /**
   * Validates the entire form before submission
   */
  static async validateFormSubmission(
    form: HTMLFormElement,
    schemaName: string,
  ): Promise<boolean> {
    try {
      // Clear all previous errors
      this.clearAllFormErrors(form);
      UIManager.resetAriaInvalid(form);

      // Get form data
      const formData = new FormData(form);

      // Get the validation schema
      const schema = this.getSchemaForForm(schemaName);
      if (!schema) {
        Logger.warn(`No validation schema found for form: ${schemaName}`);
        return true; // Allow submission if no schema is defined
      }

      // Validate using the shared validation function
      const result = validateForm(schema, formData);

      if (!result.success) {
        // Show all validation errors
        Object.entries(result.errors).forEach(([fieldName, errorMessage]) => {
          const field = form.querySelector(
            `[name="${fieldName}"], #${fieldName}`,
          ) as HTMLElement;
          if (field) {
            this.showFieldError(field, errorMessage);
            UIManager.setAriaInvalid(fieldName, true);
          }
        });

        throw FormBuilderError.validationError(
          "Form contains invalid fields",
          undefined,
          formData,
        );
      }

      return true;
    } catch (error) {
      if (error instanceof FormBuilderError) {
        throw error;
      }

      Logger.error("Form validation failed:", error);
      throw FormBuilderError.validationError(
        "Validation process failed",
        undefined,
        error,
      );
    }
  }

  /**
   * Get validation schema for a form
   */
  private static getSchemaForForm(formId: string): z.ZodSchema<any> | null {
    // Normalize form ID to match schema keys
    const normalizedId = formId.toLowerCase().replace(/[^a-z0-9-]/g, "-");

    if (normalizedId in SCHEMA_MAP) {
      return SCHEMA_MAP[normalizedId as SchemaName];
    }

    // For dynamic forms or forms without predefined schemas
    return null;
  }

  /**
   * Extract user-friendly error message from validation error
   */
  private static extractErrorMessage(error: Error): string {
    // Handle Zod validation errors
    if ("issues" in error && Array.isArray((error as any).issues)) {
      const issues = (error as any).issues;
      return issues[0]?.message ?? "Invalid input";
    }

    return error.message ?? "Invalid input";
  }

  /**
   * Show error for a specific field
   */
  private static showFieldError(field: HTMLElement, message: string): void {
    // Remove existing error
    this.clearFieldError(field);

    // Create error element
    const errorEl = document.createElement("div");
    errorEl.className = "form-error-message";
    errorEl.textContent = message;
    errorEl.style.color = "var(--error-color, #ef4444)";
    errorEl.style.fontSize = "0.875rem";
    errorEl.style.marginTop = "0.25rem";

    // Add error styling to field
    field.classList.add("error");
    field.style.borderColor = "var(--error-color, #ef4444)";

    // Insert error message
    const container = field.parentElement;
    if (container) {
      container.appendChild(errorEl);
    }
  }

  /**
   * Show success state for a field
   */
  private static showFieldSuccess(field: HTMLElement): void {
    this.clearFieldError(field);
    field.classList.add("success");
    field.style.borderColor = "var(--success-color, #10b981)";
  }

  /**
   * Clear error for a specific field
   */
  private static clearFieldError(field: HTMLElement): void {
    // Remove error styling
    field.classList.remove("error", "success");
    field.style.borderColor = "";

    // Remove error message
    const container = field.parentElement;
    if (container) {
      const errorEl = container.querySelector(".form-error-message");
      errorEl?.remove();
    }
  }

  /**
   * Clear all form errors
   */
  private static clearAllFormErrors(form: HTMLFormElement): void {
    // Clear all field errors
    const fields = form.querySelectorAll<HTMLElement>(
      "input, textarea, select",
    );
    fields.forEach((field) => this.clearFieldError(field));

    // Clear form-level error messages
    const formErrors = form.querySelectorAll(".form-error-message");
    formErrors.forEach((error) => error.remove());
  }

  /**
   * Add custom validation schema for dynamic forms
   */
  static addCustomSchema(formId: string, schema: z.ZodSchema<any>): void {
    (SCHEMA_MAP as any)[formId] = schema;
  }

  /**
   * Remove custom validation schema
   */
  static removeCustomSchema(formId: string): void {
    delete (SCHEMA_MAP as any)[formId];
  }
}
