import { validation } from './validation';

export function setupLoginForm() {
  // Only run on login page
  if (!document.getElementById('login-form')) {
    return;
  }

  const loginForm = document.getElementById('login-form') as HTMLFormElement;
  if (!loginForm) {
    return;
  }

  // Setup real-time validation
  validation.setupRealTimeValidation('login-form', 'login')
    .catch(error => {
      console.error('Failed to setup real-time validation:', error);
    });

  // Handle form submission
  loginForm.addEventListener('submit', async (e) => {
    e.preventDefault();
    console.log('Login form submitted');
    validation.clearAllErrors();
    
    const formData = new FormData(loginForm);
    const data = Object.fromEntries(formData.entries()) as Record<string, string>;
    console.log('Form data:', { email: data.email, password: '***' });
    
    const result = await validation.validateForm(loginForm, 'login');
    if (result.success) {
      try {
        console.log('Sending login request...');
        const response = await validation.fetchWithCSRF('/api/v1/auth/login', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            'Accept': 'application/json',
          },
          body: JSON.stringify(data)
        });
        
        if (response.ok) {
          // Store tokens in localStorage for API calls
          const tokens = await response.json();
          validation.setJWTToken(tokens.access_token);
          
          // Redirect to dashboard
          window.location.href = '/dashboard';
        } else {
          const error = await response.json();
          if (response.status === 401) {
            validation.showError('form', 'Invalid email or password');
          } else if (error.errors) {
            // Handle field-specific errors
            Object.entries(error.errors).forEach(([field, message]) => {
              validation.showError(field, message as string);
            });
          } else {
            validation.showError('form', error.error || 'An error occurred during login');
          }
        }
      } catch (error) {
        console.error('Login error:', error);
        validation.showError('form', 'An error occurred during login. Please try again.');
      }
    } else if (result.error) {
      result.error.errors.forEach(err => {
        validation.showError(err.path[0], err.message);
      });
    }
  });
} 