import { setupSignupForm } from './signup';
import { setupLoginForm } from './login';
import { validation } from './validation';

// Main entry point for global initialization
document.addEventListener('DOMContentLoaded', () => {
  // Setup forms if they exist on the page
  setupSignupForm();
  setupLoginForm();
  
  // Setup validation for forms
  const forms = document.querySelectorAll('form[data-validate]');
  forms.forEach(form => {
    const schemaName = form.getAttribute('data-validate');
    if (schemaName) {
      validation.setupRealTimeValidation(form.id, schemaName);
    }
  });

  // Any other global initialization code can go here
  console.log('Application initialized');
}); 