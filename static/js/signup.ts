import { validation } from './validation';

export function setupSignupForm() {
  const signupForm = document.getElementById('signup-form') as HTMLFormElement;
  if (!signupForm) return;

  validation.setupRealTimeValidation('signup-form', 'signup');
  
  signupForm.addEventListener('submit', async (e) => {
    e.preventDefault();
    validation.clearAllErrors();
    
    const formData = new FormData(signupForm);
    const data = Object.fromEntries(formData.entries()) as Record<string, string>;
    
    const result = await validation.validateForm(signupForm, 'signup');
    if (result.success) {
      try {
        const response = await validation.fetchWithCSRF('/auth/signup', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify(data)
        });
        
        if (response.ok) {
          window.location.href = '/auth/login';
        } else {
          const error = await response.json();
          validation.showError('form', error.message || 'An error occurred during signup');
        }
      } catch (error) {
        validation.showError('form', 'An error occurred during signup');
      }
    } else if (result.error) {
      result.error.errors.forEach(err => {
        validation.showError(err.path[0], err.message);
      });
    }
  });
} 