import { z } from "zod";
import { getValidationSchema } from "./validation/generator";

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

// Schema cache
const schemaCache: Record<string, z.ZodType<Record<string, string>>> = {};

export const validation = {
  clearError(fieldId: string): void {
    const errorElement = document.getElementById(`${fieldId}_error`);
    if (errorElement) {
      errorElement.textContent = "";
    }
    const input = document.getElementById(fieldId) as HTMLInputElement;
    if (input) {
      input.classList.remove("error");
    }
  },

  showError(fieldId: string, message: string): void {
    const errorElement = document.getElementById(`${fieldId}_error`);
    if (errorElement) {
      errorElement.textContent = message;
    }
    const input = document.getElementById(fieldId) as HTMLInputElement;
    if (input) {
      input.classList.add("error");
    }
  },

  clearAllErrors(): void {
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
    const form = document.getElementById(formId);
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
    let schema = schemaCache[schemaName];
    if (!schema) {
      schema = await getValidationSchema(schemaName);
      schemaCache[schemaName] = schema;
    }
    if (!schema) {
      return {
        success: false,
        error: {
          errors: [{ path: [], message: "Invalid schema name" }],
        },
      };
    }
    const formData = new FormData(form);
    const data = Object.fromEntries(formData.entries());
    try {
      const result = schema.parse(data);
      return { success: true, data: result };
    } catch (error) {
      if (error instanceof z.ZodError) {
        return {
          success: false,
          error: {
            errors: error.errors.map((err) => ({
              path: err.path.map((p) => String(p)),
              message: err.message,
            })),
          },
        };
      }
      throw error;
    }
  },

  showErrors: (form: HTMLFormElement, errors: Record<string, string>) => {
    Object.entries(errors).forEach(([field, message]) => {
      const input = form.querySelector(`[name="${field}"]`) as HTMLInputElement;
      if (input) {
        const errorElement = document.createElement("div");
        errorElement.className = "error-message";
        errorElement.textContent = message;
        input.parentElement?.appendChild(errorElement);
        input.classList.add("error");
      }
    });
  },

  clearErrors: (form: HTMLFormElement) => {
    form.querySelectorAll(".error-message").forEach((el) => el.remove());
    form
      .querySelectorAll(".error")
      .forEach((el) => el.classList.remove("error"));
  },

  // CSRF token handling
  getCSRFToken(): string | null {
    const meta = document.querySelector('meta[name="csrf-token"]');
    if (!meta) {
      console.error("CSRF token meta tag not found");
      return null;
    }
    const token = meta.getAttribute("content");
    if (!token) {
      console.error("CSRF token content is empty");
      return null;
    }
    console.debug("CSRF token found:", token);
    return token;
  },

  // Common fetch with CSRF
  async fetchWithCSRF(
    url: string,
    options: RequestInit = {},
  ): Promise<Response> {
    // Get CSRF token from meta tag
    const csrfToken = document
      .querySelector('meta[name="csrf-token"]')
      ?.getAttribute("content");
    // Prepare headers
    const headers = new Headers(options.headers || {});
    if (csrfToken) {
      headers.set("X-CSRF-Token", csrfToken);
    }
    headers.set("Content-Type", "application/json");
    // Make request with CSRF token and credentials
    return fetch(url, {
      ...options,
      headers,
      credentials: "include",
    });
  },

  // JWT token management
  getJWTToken(): string | null {
    return localStorage.getItem("jwt_token");
  },

  setJWTToken(token: string): void {
    localStorage.setItem("jwt_token", token);
  },

  clearJWTToken(): void {
    localStorage.removeItem("jwt_token");
  },
};
