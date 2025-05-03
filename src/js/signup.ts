import { validation } from './validation';

export function setupSignupForm() {
  const form = document.getElementById('signup-form') as HTMLFormElement;
  const formError = document.getElementById('form_error') as HTMLDivElement;

  if (form) {
    form.addEventListener('submit', async (e) => {
      e.preventDefault();
      
      // Clear previous errors
      validation.clearAllErrors();
      
      const result = await validation.validateForm(form, 'signup');
      if (result.success) {
        try {
          const response = await validation.fetchWithCSRF('/api/v1/auth/signup', {
            method: 'POST',
            headers: {
              'Content-Type': 'application/json'
            },
            body: JSON.stringify(result.data)
          });

          if (response.ok) {
            window.location.href = '/dashboard';
          } else {
            const error = await response.json();
            if (error.errors) {
              // Display field-specific errors
              Object.entries(error.errors).forEach(([field, message]) => {
                validation.showError(field, message as string);
              });
            } else {
              formError.textContent = error.message || 'An error occurred during signup';
            }
          }
        } catch (error) {
          formError.textContent = 'An unexpected error occurred';
        }
      } else if (result.error) {
        // Display validation errors
        result.error.errors.forEach(err => {
          validation.showError(err.path[0], err.message);
        });
      }
    });
  }
} 