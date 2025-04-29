const { z } = window.Zod;

// Common validation schemas
const validationSchemas = {
    signup: z.object({
        first_name: z.string()
            .min(2, 'First name must be at least 2 characters')
            .max(50, 'First name must be less than 50 characters')
            .regex(/^[a-zA-Z\s-']+$/, 'First name can only contain letters, spaces, hyphens, and apostrophes'),
        last_name: z.string()
            .min(2, 'Last name must be at least 2 characters')
            .max(50, 'Last name must be less than 50 characters')
            .regex(/^[a-zA-Z\s-']+$/, 'Last name can only contain letters, spaces, hyphens, and apostrophes'),
        email: z.string()
            .email('Invalid email address')
            .min(5, 'Email must be at least 5 characters')
            .max(254, 'Email must be less than 254 characters'),
        password: z.string()
            .min(8, 'Password must be at least 8 characters')
            .max(72, 'Password must be less than 72 characters')
            .regex(/[A-Z]/, 'Password must contain at least one uppercase letter')
            .regex(/[a-z]/, 'Password must contain at least one lowercase letter')
            .regex(/[0-9]/, 'Password must contain at least one number')
            .regex(/[^A-Za-z0-9]/, 'Password must contain at least one special character')
    }),
    login: z.object({
        email: z.string()
            .email('Invalid email address'),
        password: z.string()
            .min(1, 'Password is required')
    })
};

// Validation utilities
const validation = {
    clearError(fieldId) {
        document.getElementById(`${fieldId}_error`).textContent = '';
    },

    showError(fieldId, message) {
        document.getElementById(`${fieldId}_error`).textContent = message;
    },

    clearAllErrors() {
        document.querySelectorAll('.error-message').forEach(el => el.textContent = '');
    },

    setupRealTimeValidation(formId, schema) {
        const form = document.getElementById(formId);
        if (!form) return;

        Object.keys(schema.shape).forEach(fieldId => {
            const input = document.getElementById(fieldId);
            if (!input) return;

            input.addEventListener('input', () => {
                validation.clearError(fieldId);
                const value = input.value;
                const result = schema.shape[fieldId].safeParse(value);
                if (!result.success) {
                    validation.showError(fieldId, result.error.errors[0].message);
                }
            });
        });
    },

    validateForm(formData, schema) {
        const result = schema.safeParse(formData);
        if (!result.success) {
            result.error.errors.forEach(error => {
                const field = error.path[0];
                validation.showError(field, error.message);
            });
            return false;
        }
        return true;
    },

    // CSRF token handling
    getCSRFToken() {
        return document.querySelector('meta[name="csrf-token"]')?.content;
    },

    // Common fetch with CSRF
    async fetchWithCSRF(url, options = {}) {
        const csrfToken = validation.getCSRFToken();
        return fetch(url, {
            ...options,
            headers: {
                ...options.headers,
                'X-CSRF-Token': csrfToken,
                'Content-Type': 'application/json',
            },
        });
    }
}; 