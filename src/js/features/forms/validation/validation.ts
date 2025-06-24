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
    const form = document.getElementById(formId) as HTMLFormElement;
    if (!form) return;

    let schema = schemaCache[schemaName];
    if (!schema) {
      schema = await getValidationSchema(schemaName);
      schemaCache[schemaName] = schema;
    }
    if (!schema || !(schema instanceof z.ZodObject)) return;

    const schemaFields = schema.shape as Record<string, z.ZodType>;
    Object.keys(schemaFields).forEach((fieldId) => {
      const input = document.getElementById(fieldId);
      if (!input) return;

      input.addEventListener("input", () => {
        validation.clearError(fieldId);
        const value = (input as HTMLInputElement).value;
        const fieldSchema = schemaFields[fieldId];

        // Skip validation for empty fields during real-time validation
        if (!value && fieldId !== "password") {
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
            validation.showError(fieldId, result.error.errors[0].message);
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

  // CSRF token handling - now delegated to HttpClient
  getCSRFToken(): string {
    const meta = document.querySelector("meta[name='csrf-token']");
    if (meta) {
      const token = meta.getAttribute("content");
      if (token) {
        return token;
      }
    }

    throw new Error(
      "CSRF token not found. Please refresh the page and try again.",
    );
  },

  // Common fetch with authentication - now delegated to HttpClient
  async fetchWithAuth(
    url: string,
    options: RequestInit = {},
  ): Promise<Response> {
    // This method is kept for backward compatibility
    // New code should use HttpClient directly
    const { HttpClient } = await import("../../../core/http-client");
    return HttpClient.request(url, options);
  },
};
