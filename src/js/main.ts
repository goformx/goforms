import { validation } from './validation';
import { FormBuilder } from './form-builder';

// Main entry point for global initialization
document.addEventListener('DOMContentLoaded', () => {
  console.log('Main: DOMContentLoaded fired');

  // Setup real-time validation for forms
  const forms = document.querySelectorAll('form[data-validate]');
  forms.forEach(form => {
    const schemaName = form.getAttribute('data-validate');
    if (schemaName && (schemaName === 'signup' || schemaName === 'login')) {
      validation.setupRealTimeValidation(form.id, schemaName);
    }
  });

  // Initialize form builder if we're on the edit page
  const formSchemaBuilder = document.getElementById('form-schema-builder');
  if (formSchemaBuilder) {
    console.log('Main: Found form-schema-builder element', formSchemaBuilder);
    const formIdAttr = formSchemaBuilder.getAttribute('data-form-id');
    console.log('Main: data-form-id attribute value:', formIdAttr);
    const formId = parseInt(formIdAttr || '', 10);
    console.log('Main: Parsed formId:', formId);
    if (!isNaN(formId)) {
      console.log('Main: Initializing FormBuilder with formId:', formId);
      new FormBuilder('form-schema-builder', formId);
    } else {
      console.error('Main: Invalid form ID:', formIdAttr);
    }
  } else {
    console.log('Main: form-schema-builder element not found');
  }

  console.log('Application initialized');
});

// Enable HMR
if (import.meta.hot) {
  import.meta.hot.accept()
}

console.log('Development server is running!')
console.log('Testing Vite HMR rebuild - ' + new Date().toISOString());
