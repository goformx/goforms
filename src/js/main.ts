import { setupSignupForm } from './signup';
import { setupLoginForm } from './login';
import { validation } from './validation';

// Main entry point for global initialization
document.addEventListener('DOMContentLoaded', () => {
  // Setup forms if they exist on the page
  setupSignupForm();
  setupLoginForm();
  
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

// Add a simple function to demonstrate TypeScript
function init() {
  const app = document.getElementById('app')
  if (app) {
    const timestamp = document.createElement('p')
    timestamp.textContent = `Page loaded at: ${new Date().toLocaleTimeString()}`
    app.appendChild(timestamp)
  }
}

init() 