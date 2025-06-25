import { z } from "zod";
import { getValidationSchema } from "./generator";
import { ValidationManager } from "@/shared/utils/validation-manager";
import { ErrorManager } from "@/shared/utils/error-manager";
import { Logger } from "@/core/logger";

// Types
export type FormData = Record<string, string>;
export type ValidationResult = {
  success: boolean;
  data?: FormData;
  error?: {
    errors: Array<{
      path: string[];
      message: string;
    }>;
  };
};

// Legacy schema cache for backward compatibility
const schemaCache: Record<string, z.ZodType<Record<string, string>>> = {};

export const validation = {
  clearError(fieldId: string): void {
    ErrorManager.clearFieldError(fieldId);
  },

  showError(fieldId: string, message: string): void {
    ErrorManager.showFieldError(fieldId, message);
  },

  clearAllErrors(): void {
    // This method is kept for backward compatibility
    // It clears all errors globally, which is less precise than the new ErrorManager
    document.querySelectorAll(".error-message").forEach((el) => {
      el.textContent = "";
    });
    document.querySelectorAll(".error").forEach((el) => {
      el.classList.remove("error");
    });
  },

  async setupRealTimeValidation(
    formId: string,
    schemaName: string,
  ): Promise<void> {
    Logger.debug("Setting up real-time validation", { formId, schemaName });

    const form = document.getElementById(formId) as HTMLFormElement;
    if (!form) {
      Logger.error("Form not found", { formId });
      return;
    }

    let schema = schemaCache[schemaName];
    if (!schema) {
      Logger.debug("Fetching validation schema", { schemaName });
      schema = await getValidationSchema(schemaName);
      schemaCache[schemaName] = schema;
    }
    if (!schema || !(schema instanceof z.ZodObject)) {
      Logger.error("Invalid schema", { schemaName, schemaType: typeof schema });
      return;
    }

    const schemaFields = schema.shape as Record<string, z.ZodType>;
    Logger.debug("Setting up validation for fields", {
      fields: Object.keys(schemaFields),
    });

    Object.keys(schemaFields).forEach((fieldId) => {
      const input = document.getElementById(fieldId);
      if (!input) {
        Logger.warn("Input field not found", { fieldId });
        return;
      }

      Logger.debug("Adding input listener", { fieldId });
      input.addEventListener("input", () => {
        Logger.debug("Input event triggered", { fieldId });
        validation.clearError(fieldId);
        const value = (input as HTMLInputElement).value;
        const fieldSchema = schemaFields[fieldId];

        // Skip validation for empty fields during real-time validation
        if (!value && fieldId !== "password") {
          Logger.debug("Skipping validation for empty field", { fieldId });
          return;
        }

        // Special handling for confirm_password
        if (fieldId === "confirm_password") {
          const passwordInput = document.getElementById(
            "password",
          ) as HTMLInputElement;
          if (passwordInput && value !== passwordInput.value) {
            validation.showError(fieldId, "Passwords don't match");
            return;
          }
        }
        if (fieldSchema instanceof z.ZodType) {
          const result = fieldSchema.safeParse(value);
          if (!result.success) {
            Logger.debug("Validation failed", {
              fieldId,
              error: result.error.errors[0].message,
            });
            validation.showError(fieldId, result.error.errors[0].message);
          } else {
            Logger.debug("Validation passed", { fieldId });
          }
        }
      });
      // For password field, also validate confirm_password when password changes
      if (fieldId === "password") {
        input.addEventListener("input", () => {
          const confirmInput = document.getElementById(
            "confirm_password",
          ) as HTMLInputElement;
          if (confirmInput && confirmInput.value) {
            if (confirmInput.value !== (input as HTMLInputElement).value) {
              validation.showError("confirm_password", "Passwords don't match");
            } else {
              validation.clearError("confirm_password");
            }
          }
        });
      }
    });
  },

  async validateForm(
    form: HTMLFormElement,
    schemaName: string,
  ): Promise<ValidationResult> {
    // Use the new ValidationManager for better error handling
    try {
      return await ValidationManager.validateForm(form, schemaName);
    } catch (error) {
      Logger.error("Form validation failed:", error);
      return {
        success: false,
        error: {
          errors: [{ path: [], message: "Validation failed" }],
        },
      };
    }
  },

  showErrors: (form: HTMLFormElement, errors: Record<string, string>) => {
    ErrorManager.showErrors(form, errors);
  },

  clearErrors: (form: HTMLFormElement) => {
    ErrorManager.clearErrors(form);
  },

  // Note: CSRF token handling and HTTP requests are now handled by HttpClient
  // Use HttpClient.get(), HttpClient.post(), etc. directly instead of fetchWithAuth
};
