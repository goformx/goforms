import { z } from "zod";
import { getValidationSchema } from "@/features/forms/validation/generator";
import { Logger } from "@/core/logger";

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

/**
 * Improved validation manager with better schema caching and error handling
 */
export class ValidationManager {
  private static schemaCache = new Map<string, z.ZodType>();

  static async getSchema(schemaName: string): Promise<z.ZodType> {
    if (this.schemaCache.has(schemaName)) {
      return this.schemaCache.get(schemaName)!;
    }

    try {
      const schema = await getValidationSchema(schemaName);
      this.schemaCache.set(schemaName, schema);
      return schema;
    } catch (error) {
      Logger.error(`Failed to load validation schema: ${schemaName}`, error);
      throw new Error(`Validation schema '${schemaName}' not found`);
    }
  }

  static async validateForm(
    form: HTMLFormElement,
    schemaName: string,
  ): Promise<ValidationResult> {
    try {
      const schema = await this.getSchema(schemaName);
      const formData = new FormData(form);
      const data = Object.fromEntries(formData.entries()) as Record<
        string,
        string
      >;

      // Special handling for confirm_password field
      const passwordInput = form.querySelector<HTMLInputElement>(
        "input[name='password']",
      );
      const confirmPasswordInput = form.querySelector<HTMLInputElement>(
        "input[name='confirm_password']",
      );

      if (passwordInput && confirmPasswordInput && data.confirm_password) {
        if (data.password !== data.confirm_password) {
          return {
            success: false,
            error: {
              errors: [
                {
                  path: ["confirm_password"],
                  message: "Passwords don't match",
                },
              ],
            },
          };
        }
      }

      const result = schema.parse(data);
      return { success: true, data: result };
    } catch (error) {
      if (error instanceof z.ZodError) {
        return {
          success: false,
          error: {
            errors: error.errors.map((err) => ({
              path: err.path.map(String),
              message: err.message,
            })),
          },
        };
      }
      Logger.error("Validation error:", error);
      throw error;
    }
  }

  static async validateField(
    fieldName: string,
    value: string,
    schemaName: string,
    form?: HTMLFormElement,
  ): Promise<{ valid: boolean; error?: string }> {
    try {
      // Skip validation for empty fields during real-time validation (except password)
      if (!value && fieldName !== "password") {
        return { valid: true };
      }

      // Special handling for confirm_password
      if (fieldName === "confirm_password" && form) {
        const passwordInput = form.querySelector<HTMLInputElement>(
          "input[name='password']",
        );
        if (passwordInput && value !== passwordInput.value) {
          return { valid: false, error: "Passwords don't match" };
        }
        return { valid: true };
      }

      const schema = await this.getSchema(schemaName);
      if (!(schema instanceof z.ZodObject)) return { valid: true };

      const fieldSchema = schema.shape[fieldName];
      if (!fieldSchema) return { valid: true };

      fieldSchema.parse(value);
      return { valid: true };
    } catch (error) {
      if (error instanceof z.ZodError) {
        return { valid: false, error: error.errors[0]?.message };
      }
      Logger.error("Field validation error:", error);
      return { valid: false, error: "Validation error" };
    }
  }

  static clearSchemaCache(): void {
    this.schemaCache.clear();
    Logger.debug("Schema cache cleared");
  }

  static removeSchemaFromCache(schemaName: string): void {
    this.schemaCache.delete(schemaName);
    Logger.debug(`Schema '${schemaName}' removed from cache`);
  }

  static getCacheSize(): number {
    return this.schemaCache.size;
  }
}
