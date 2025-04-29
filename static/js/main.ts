import { setupSignupForm } from './signup';
import { setupLoginForm } from './login';

// Main entry point for global initialization
document.addEventListener('DOMContentLoaded', () => {
  // Setup forms if they exist on the page
  setupSignupForm();
  setupLoginForm();

  // Any other global initialization code can go here
  console.log('Application initialized');
}); 