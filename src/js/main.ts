import { validation } from './validation';

// Main entry point for global initialization
document.addEventListener('DOMContentLoaded', () => {
  // Setup real-time validation for forms
  const forms = document.querySelectorAll('form[data-validate]');
  forms.forEach(form => {
    const schemaName = form.getAttribute('data-validate');
    if (schemaName && (schemaName === 'signup' || schemaName === 'login')) {
      validation.setupRealTimeValidation(form.id, schemaName);
    }
  });

  // Any other global initialization code can go here
  console.log('Application initialized');
});

// Enable HMR
if (import.meta.hot) {
  import.meta.hot.accept()
}

console.log('Development server is running!')
