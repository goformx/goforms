import { z } from 'zod';
import { getValidationSchema } from './validation/generator';

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

// Common validation schemas
export const validationSchemas = {
  signup: z.object({
    username: z.string()
      .min(3, 'Username must be at least 3 characters')
      .max(50, 'Username must be less than 50 characters')
      .regex(/^[a-zA-Z0-9_]+$/, 'Username can only contain letters, numbers, and underscores'),
    email: z.string()
      .email('Invalid email address'),
    password: z.string()
      .min(8, 'Password must be at least 8 characters')
      .regex(/[A-Z]/, 'Password must contain at least one uppercase letter')
      .regex(/[a-z]/, 'Password must contain at least one lowercase letter')
      .regex(/[0-9]/, 'Password must contain at least one number'),
    confirmPassword: z.string()
  }).refine((data) => data.password === data.confirmPassword, {
    message: "Passwords don't match",
    path: ["confirmPassword"]
  }),

  login: z.object({
    email: z.string()
      .email('Invalid email address'),
    password: z.string()
      .min(1, 'Password is required')
  })
} as const;

// Validation utilities
export const validation = {
  clearError(fieldId: string): void {
    const errorElement = document.getElementById(`${fieldId}_error`);
    if (errorElement) {
      errorElement.textContent = '';
    }
  },

  showError(fieldId: string, message: string): void {
    const errorElement = document.getElementById(`${fieldId}_error`);
    if (errorElement) {
      errorElement.textContent = message;
    }
  },

  clearAllErrors(): void {
    document.querySelectorAll('.error-message').forEach(el => {
      if (el instanceof HTMLElement) {
        el.textContent = '';
      }
    });
  },

  async setupRealTimeValidation(formId: string, schemaName: string): Promise<void> {
    const form = document.getElementById(formId);
    if (!form) return;

    const schema = await getValidationSchema(schemaName);
    if (!schema) return;

    const schemaFields = schema instanceof z.ZodObject ? schema.shape : {};
    Object.keys(schemaFields).forEach(fieldId => {
      const input = document.getElementById(fieldId);
      if (!input) return;

      input.addEventListener('input', () => {
        validation.clearError(fieldId);
        const value = (input as HTMLInputElement).value;
        const fieldSchema = schemaFields[fieldId];
        if (fieldSchema instanceof z.ZodType) {
          const result = fieldSchema.safeParse(value);
          if (!result.success) {
            validation.showError(fieldId, result.error.errors[0].message);
          }
        }
      });
    });
  },

  async validateForm(form: HTMLFormElement, schemaName: string): Promise<ValidationResult> {
    const schema = await getValidationSchema(schemaName);
    if (!schema) {
      return { 
        success: false, 
        error: { 
          errors: [{ path: [], message: 'Failed to load validation schema' }] 
        } 
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
            errors: error.errors.map(err => ({
              path: err.path.map(p => String(p)),
              message: err.message
            }))
          }
        };
      }
      throw error;
    }
  },

  showErrors: (form: HTMLFormElement, errors: Record<string, string>) => {
    Object.entries(errors).forEach(([field, message]) => {
      const input = form.querySelector(`[name="${field}"]`) as HTMLInputElement;
      if (input) {
        const errorElement = document.createElement('div');
        errorElement.className = 'error-message';
        errorElement.textContent = message;
        input.parentElement?.appendChild(errorElement);
        input.classList.add('error');
      }
    });
  },

  clearErrors: (form: HTMLFormElement) => {
    form.querySelectorAll('.error-message').forEach(el => el.remove());
    form.querySelectorAll('.error').forEach(el => el.classList.remove('error'));
  },

  // CSRF token handling
  getCSRFToken(): string | null {
    return document.querySelector('meta[name="csrf-token"]')?.getAttribute('content') ?? null;
  },

  // Common fetch with CSRF
  async fetchWithCSRF(url: string, options: RequestInit = {}): Promise<Response> {
    const csrfToken = validation.getCSRFToken();
    return fetch(url, {
      ...options,
      headers: {
        ...options.headers,
        'X-CSRF-Token': csrfToken ?? '',
        'Content-Type': 'application/json',
      },
    });
  }
}; 