import { validation } from './validation';

export function setupLoginForm() {
  // Only run on login page
  if (!window.location.pathname.includes('/login')) {
    return;
  }

  const loginForm = document.getElementById('login-form') as HTMLFormElement;
  if (!loginForm) {
    console.warn('Login form not found on login page');
    return;
  }

  // Setup real-time validation
  validation.setupRealTimeValidation('login-form', 'login')
    .catch(error => {
      console.error('Failed to setup real-time validation:', error);
    });
  
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
          console.log('Login successful, redirecting...');
          window.location.href = '/';
        } else {
          const error = await response.json();
          console.error('Login failed:', error);
          if (response.status === 401) {
            // Show the error under both email and password fields for invalid credentials
            validation.showError('email', 'Invalid email or password');
            validation.showError('password', 'Invalid email or password');
          } else {
            validation.showError('form', error.error || 'An error occurred during login');
          }
        }
      } catch (error) {
        console.error('Login error:', error);
        validation.showError('form', 'An error occurred during login');
      }
    } else if (result.error) {
      console.error('Validation errors:', result.error);
      result.error.errors.forEach(err => {
        validation.showError(err.path[0], err.message);
      });
    }
  });
} 