import { validation } from './validation';

export function setupLoginForm() {
  const loginForm = document.getElementById('login-form') as HTMLFormElement;
  if (!loginForm) return;

  validation.setupRealTimeValidation('login-form', 'login');
  
  loginForm.addEventListener('submit', async (e) => {
    e.preventDefault();
    validation.clearAllErrors();
    
    const formData = new FormData(loginForm);
    const data = Object.fromEntries(formData.entries()) as Record<string, string>;
    
    const result = await validation.validateForm(loginForm, 'login');
    if (result.success) {
      try {
        const response = await validation.fetchWithCSRF('/auth/login', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify(data)
        });
        
        if (response.ok) {
          window.location.href = '/';
        } else {
          const error = await response.json();
          validation.showError('form', error.message || 'Invalid email or password');
        }
      } catch (error) {
        validation.showError('form', 'An error occurred during login');
      }
    } else if (result.error) {
      result.error.errors.forEach(err => {
        validation.showError(err.path[0], err.message);
      });
    }
  });
} 