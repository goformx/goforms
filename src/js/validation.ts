import { z } from 'zod';

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

// Schema names type
export type SchemaName = keyof typeof validationSchemas;

// Common validation schemas
export const validationSchemas = {
  signup: z.object({
    first_name: z.string()
      .min(2, 'First name must be at least 2 characters')
      .max(50, 'First name must be less than 50 characters'),
    last_name: z.string()
      .min(2, 'Last name must be at least 2 characters')
      .max(50, 'Last name must be less than 50 characters'),
    email: z.string()
      .email('Please enter a valid email address')
      .min(5, 'Email must be at least 5 characters')
      .max(100, 'Email must be less than 100 characters'),
    password: z.string()
      .min(8, 'Password must be at least 8 characters')
      .max(100, 'Password must be less than 100 characters')
      .regex(/[A-Z]/, 'Password must contain at least one uppercase letter')
      .regex(/[a-z]/, 'Password must contain at least one lowercase letter')
      .regex(/[0-9]/, 'Password must contain at least one number')
      .regex(/[^A-Za-z0-9]/, 'Password must contain at least one special character'),
    confirm_password: z.string()
  }).refine((data) => data.password === data.confirm_password, {
    message: "Passwords don't match",
    path: ["confirm_password"]
  }),

  login: z.object({
    email: z.string()
      .email('Invalid email address'),
    password: z.string()
      .min(1, 'Password is required')
  }),

  contact: z.object({
    name: z.string()
      .min(2, 'Name must be at least 2 characters')
      .max(100, 'Name must be less than 100 characters'),
    email: z.string()
      .email('Please enter a valid email address')
      .min(5, 'Email must be at least 5 characters')
      .max(100, 'Email must be less than 100 characters'),
    message: z.string()
      .min(10, 'Message must be at least 10 characters')
      .max(1000, 'Message must be less than 1000 characters')
  }),

  demo: z.object({
    name: z.string()
      .min(2, 'Name must be at least 2 characters')
      .max(100, 'Name must be less than 100 characters'),
    email: z.string()
      .email('Please enter a valid email address')
      .min(5, 'Email must be at least 5 characters')
      .max(100, 'Email must be less than 100 characters')
  }),

  editForm: z.object({
    title: z.string()
      .min(3, 'Title must be at least 3 characters')
      .max(255, 'Title must be less than 255 characters'),
    description: z.string()
      .max(1000, 'Description must be less than 1000 characters')
      .optional()
  }),

  newForm: z.object({
    title: z.string()
      .min(3, 'Title must be at least 3 characters')
      .max(255, 'Title must be less than 255 characters'),
    description: z.string()
      .max(1000, 'Description must be less than 1000 characters')
      .optional()
  })
} as const;

// Validation utilities
export const validation = {
  clearError(fieldId: string): void {
    const errorElement = document.getElementById(`${fieldId}_error`);
    if (errorElement) {
      errorElement.textContent = '';
    }
    const input = document.getElementById(fieldId) as HTMLInputElement;
    if (input) {
      input.classList.remove('error');
    }
  },

  showError(fieldId: string, message: string): void {
    const errorElement = document.getElementById(`${fieldId}_error`);
    if (errorElement) {
      errorElement.textContent = message;
    }
    const input = document.getElementById(fieldId) as HTMLInputElement;
    if (input) {
      input.classList.add('error');
    }
  },

  clearAllErrors(): void {
    document.querySelectorAll('.error-message').forEach(el => {
      el.textContent = '';
    });
    document.querySelectorAll('.error').forEach(el => {
      el.classList.remove('error');
    });
  },

  setupRealTimeValidation(formId: string, schemaName: SchemaName): void {
    const form = document.getElementById(formId);
    if (!form) return;

    const schema = validationSchemas[schemaName];
    if (!schema || !(schema instanceof z.ZodObject)) return;

    const schemaFields = schema.shape as Record<string, z.ZodType>;
    Object.keys(schemaFields).forEach(fieldId => {
      const input = document.getElementById(fieldId);
      if (!input) return;

      input.addEventListener('input', () => {
        validation.clearError(fieldId);
        const value = (input as HTMLInputElement).value;
        const fieldSchema = schemaFields[fieldId];
        
        // Special handling for confirm_password
        if (fieldId === 'confirm_password') {
          const passwordInput = document.getElementById('password') as HTMLInputElement;
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
      if (fieldId === 'password') {
        input.addEventListener('input', () => {
          const confirmInput = document.getElementById('confirm_password') as HTMLInputElement;
          if (confirmInput && confirmInput.value) {
            if (confirmInput.value !== (input as HTMLInputElement).value) {
              validation.showError('confirm_password', "Passwords don't match");
            } else {
              validation.clearError('confirm_password');
            }
          }
        });
      }
    });
  },

  async validateForm(form: HTMLFormElement, schemaName: SchemaName): Promise<ValidationResult> {
    const schema = validationSchemas[schemaName];
    if (!schema) {
      return { 
        success: false, 
        error: { 
          errors: [{ path: [], message: 'Invalid schema name' }] 
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
    const meta = document.querySelector('meta[name="csrf-token"]');
    if (!meta) {
      console.error('CSRF token meta tag not found');
      return null;
    }
    const token = meta.getAttribute('content');
    if (!token) {
      console.error('CSRF token content is empty');
      return null;
    }
    console.debug('CSRF token found:', token);
    return token;
  },

  // Common fetch with CSRF
  async fetchWithCSRF(url: string, options: RequestInit = {}): Promise<Response> {
    // Get CSRF token from form's hidden input
    const csrfInput = document.querySelector('input[name="csrf_token"]') as HTMLInputElement;
    if (!csrfInput) {
      console.error('CSRF token input not found');
      throw new Error('CSRF token not found');
    }

    const csrfToken = csrfInput.value;
    if (!csrfToken) {
      console.error('CSRF token value is empty');
      throw new Error('CSRF token is empty');
    }

    // Add CSRF token to headers
    const headers = new Headers(options.headers);
    headers.set('X-CSRF-Token', csrfToken);

    // Add JWT token to Authorization header if available
    const jwtToken = this.getJWTToken();
    if (jwtToken) {
      headers.set('Authorization', `Bearer ${jwtToken}`);
    }

    // Make request with CSRF token
    return fetch(url, {
      ...options,
      headers,
      credentials: 'same-origin'
    });
  },

  // JWT token management
  setJWTToken(token: string): void {
    localStorage.setItem('jwt_token', token);
  },

  getJWTToken(): string | null {
    return localStorage.getItem('jwt_token');
  },

  clearJWTToken(): void {
    localStorage.removeItem('jwt_token');
  }
};

// Export types for external use
export type SignupFormData = z.infer<typeof validationSchemas.signup>; 