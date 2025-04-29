import { validation, validationSchemas } from './validation';

document.addEventListener('DOMContentLoaded', () => {
  // Setup signup form validation
  const signupForm = document.getElementById('signup-form');
  if (signupForm) {
    validation.setupRealTimeValidation('signup-form', validationSchemas.signup);
    
    signupForm.addEventListener('submit', async (e) => {
      e.preventDefault();
      validation.clearAllErrors();
      
      const formData = new FormData(signupForm as HTMLFormElement);
      const data = Object.fromEntries(formData.entries()) as Record<string, string>;
      
      if (validation.validateForm(data, validationSchemas.signup)) {
        try {
          const response = await validation.fetchWithCSRF('/auth/signup', {
            method: 'POST',
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
      }
    });
  }

  // Setup login form validation
  const loginForm = document.getElementById('login-form');
  if (loginForm) {
    validation.setupRealTimeValidation('login-form', validationSchemas.login);
    
    loginForm.addEventListener('submit', async (e) => {
      e.preventDefault();
      validation.clearAllErrors();
      
      const formData = new FormData(loginForm as HTMLFormElement);
      const data = Object.fromEntries(formData.entries()) as Record<string, string>;
      
      if (validation.validateForm(data, validationSchemas.login)) {
        try {
          const response = await validation.fetchWithCSRF('/auth/login', {
            method: 'POST',
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
      }
    });
  }
}); 