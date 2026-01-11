import { ref, type Ref } from "vue";
import { z, type ZodSchema, type ZodError } from "zod";

export interface ValidationResult {
  valid: boolean;
  errors: Record<string, string>;
}

export interface UseFormValidationReturn<T> {
  errors: Ref<Record<string, string>>;
  validate: (data: unknown) => ValidationResult;
  validateField: (field: keyof T, value: unknown) => string | null;
  clearErrors: () => void;
  clearFieldError: (field: keyof T) => void;
  setFieldError: (field: keyof T, message: string) => void;
  hasErrors: Ref<boolean>;
}

/**
 * Composable for form validation using Zod schemas
 */
export function useFormValidation<T extends Record<string, unknown>>(
  schema: ZodSchema<T>,
): UseFormValidationReturn<T> {
  const errors = ref<Record<string, string>>({});
  const hasErrors = ref(false);

  function formatZodErrors(error: ZodError): Record<string, string> {
    const formatted: Record<string, string> = {};
    for (const issue of error.issues) {
      const path = issue.path.join(".");
      if (path && !formatted[path]) {
        formatted[path] = issue.message;
      }
    }
    return formatted;
  }

  function validate(data: unknown): ValidationResult {
    try {
      schema.parse(data);
      errors.value = {};
      hasErrors.value = false;
      return { valid: true, errors: {} };
    } catch (error) {
      if (error instanceof z.ZodError) {
        errors.value = formatZodErrors(error);
        hasErrors.value = true;
        return { valid: false, errors: errors.value };
      }
      throw error;
    }
  }

  function validateField(field: keyof T, value: unknown): string | null {
    try {
      // Create a partial schema for the single field
      const fieldSchema = (schema as z.ZodObject<z.ZodRawShape>).shape[
        field as string
      ];
      if (fieldSchema) {
        fieldSchema.parse(value);
        delete errors.value[field as string];
        hasErrors.value = Object.keys(errors.value).length > 0;
        return null;
      }
      return null;
    } catch (error) {
      if (error instanceof z.ZodError) {
        const message = error.issues[0]?.message ?? "Invalid value";
        errors.value[field as string] = message;
        hasErrors.value = true;
        return message;
      }
      throw error;
    }
  }

  function clearErrors(): void {
    errors.value = {};
    hasErrors.value = false;
  }

  function clearFieldError(field: keyof T): void {
    delete errors.value[field as string];
    hasErrors.value = Object.keys(errors.value).length > 0;
  }

  function setFieldError(field: keyof T, message: string): void {
    errors.value[field as string] = message;
    hasErrors.value = true;
  }

  return {
    errors,
    validate,
    validateField,
    clearErrors,
    clearFieldError,
    setFieldError,
    hasErrors,
  };
}

// Common validation schemas
export const loginSchema = z.object({
  email: z.string().email("Please enter a valid email address"),
  password: z.string().min(1, "Password is required"),
});

export const signupSchema = z
  .object({
    email: z.string().email("Please enter a valid email address"),
    password: z
      .string()
      .min(8, "Password must be at least 8 characters")
      .regex(/[A-Z]/, "Password must contain at least one uppercase letter")
      .regex(/[a-z]/, "Password must contain at least one lowercase letter")
      .regex(/[0-9]/, "Password must contain at least one number")
      .regex(
        /[^A-Za-z0-9]/,
        "Password must contain at least one special character",
      ),
    confirmPassword: z.string().min(1, "Please confirm your password"),
  })
  .refine((data) => data.password === data.confirmPassword, {
    message: "Passwords don't match",
    path: ["confirmPassword"],
  });

export const forgotPasswordSchema = z.object({
  email: z.string().email("Please enter a valid email address"),
});

export type LoginFormData = z.infer<typeof loginSchema>;
export type SignupFormData = z.infer<typeof signupSchema>;
export type ForgotPasswordFormData = z.infer<typeof forgotPasswordSchema>;
