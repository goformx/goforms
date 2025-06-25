import { z } from "zod";

// Simple schemas for your auth forms
export const loginSchema = z.object({
  email: z.string().email("Please enter a valid email"),
  password: z.string().min(1, "Password is required"),
});

export const signupSchema = z
  .object({
    email: z.string().email("Please enter a valid email"),
    password: z.string().min(8, "Password must be at least 8 characters"),
    confirm_password: z.string(),
  })
  .refine((data) => data.password === data.confirm_password, {
    message: "Passwords don't match",
    path: ["confirm_password"],
  });

// Add your other form schemas here as needed
export const contactSchema = z.object({
  name: z.string().min(1, "Name is required"),
  email: z.string().email("Please enter a valid email"),
  message: z.string().min(10, "Message must be at least 10 characters"),
});

// Simple validation function
export function validateForm<T>(
  schema: z.ZodSchema<T>,
  formData: FormData | Record<string, unknown>,
):
  | { success: true; data: T }
  | { success: false; errors: Record<string, string> } {
  const data =
    formData instanceof FormData
      ? Object.fromEntries(formData.entries())
      : formData;

  const result = schema.safeParse(data);

  if (result.success) {
    return { success: true, data: result.data };
  }

  const errors: Record<string, string> = {};
  result.error.errors.forEach((err) => {
    if (err.path.length > 0) {
      errors[err.path[0] as string] = err.message;
    }
  });

  return { success: false, errors };
}

// Real-time validation for auth forms only
export function setupRealTimeValidation(
  formId: string,
  schema: z.ZodSchema<any>,
): () => void {
  const form = document.getElementById(formId) as HTMLFormElement;
  if (!form) return () => {};

  const showError = (fieldName: string, message: string) => {
    const field = form.querySelector(`[name="${fieldName}"]`) as HTMLElement;
    if (!field) return;

    // Remove existing error
    const existingError = field.parentElement?.querySelector(".error-message");
    existingError?.remove();

    // Add new error
    const errorEl = document.createElement("div");
    errorEl.className = "error-message";
    errorEl.textContent = message;
    errorEl.style.color = "#ef4444";
    errorEl.style.fontSize = "0.875rem";
    errorEl.style.marginTop = "0.25rem";

    field.parentElement?.appendChild(errorEl);
    field.style.borderColor = "#ef4444";
  };

  const clearError = (fieldName: string) => {
    const field = form.querySelector(`[name="${fieldName}"]`) as HTMLElement;
    if (!field) return;

    const errorEl = field.parentElement?.querySelector(".error-message");
    errorEl?.remove();
    field.style.borderColor = "";
  };

  let debounceTimer: NodeJS.Timeout;

  const handleInput = (e: Event) => {
    const target = e.target as HTMLInputElement;
    const fieldName = target.name;
    if (!fieldName) return;

    clearTimeout(debounceTimer);
    debounceTimer = setTimeout(() => {
      const fieldSchema = (schema as any).shape[fieldName];
      if (!fieldSchema) return;

      const result = fieldSchema.safeParse(target.value);

      if (result.success) {
        clearError(fieldName);
      } else {
        showError(fieldName, result.error.errors[0]?.message ?? "Invalid");
      }

      // Special case for password confirmation
      if (fieldName === "password") {
        const confirmField = form.querySelector(
          '[name="confirm_password"]',
        ) as HTMLInputElement;
        if (confirmField?.value) {
          if (confirmField.value !== target.value) {
            showError("confirm_password", "Passwords don't match");
          } else {
            clearError("confirm_password");
          }
        }
      }
    }, 300);
  };

  form.addEventListener("input", handleInput);

  return () => {
    form.removeEventListener("input", handleInput);
    clearTimeout(debounceTimer);
  };
}

// Usage examples:
// For user forms (submit-time only):
// const result = validateForm(contactSchema, formData);

// For auth forms (real-time):
// const cleanup = setupRealTimeValidation('signup-form', signupSchema);
// cleanup(); // call when component unmounts
