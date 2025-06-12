/**
 * Form Handler Module
 *
 * Provides shared functionality for form handling including:
 * - Real-time validation with debouncing
 * - Form submission
 * - Centralized error handling
 * - Server communication
 */

import { validation } from "./validation";

export interface FormConfig {
  formId: string;
  validationType: string;
  validationDelay?: number; // Optional delay for debounced validation
}

/**
 * Debounces a function to prevent excessive calls
 *
 * @param fn - Function to debounce
 * @param delay - Delay in milliseconds
 * @returns Debounced function
 */
function debounce<T extends (...args: any[]) => any>(
  fn: T,
  delay = 300,
): (...args: Parameters<T>) => void {
  let timer: NodeJS.Timeout;
  return (...args: Parameters<T>) => {
    clearTimeout(timer);
    timer = setTimeout(() => fn(...args), delay);
  };
}

/**
 * Sets up a form with validation and submission handling
 *
 * @param config - Form configuration including form ID and validation type
 */
export function setupForm(config: FormConfig) {
  // Use querySelector for better type safety
  const form = document.querySelector<HTMLFormElement>(`#${config.formId}`);
  if (!form) {
    console.error(`Form with ID "${config.formId}" not found`);
    return;
  }

  // Initialize validation
  validation.setupRealTimeValidation(form.id, config.validationType);

  // Setup real-time validation with debouncing
  setupRealTimeValidation(form, config.validationType, config.validationDelay);

  // Setup form submission
  form.addEventListener("submit", (event) =>
    handleFormSubmission(event, form, config.validationType),
  );
}

/**
 * Sets up real-time validation for form inputs with debouncing
 */
function setupRealTimeValidation(
  form: HTMLFormElement,
  validationType: string,
  delay = 300,
) {
  form.querySelectorAll<HTMLInputElement>("input[id]").forEach((input) => {
    input.addEventListener(
      "input",
      debounce(() => handleInputValidation(input, form, validationType), delay),
    );
  });
}

/**
 * Handles real-time validation for individual input fields
 */
async function handleInputValidation(
  input: HTMLInputElement,
  form: HTMLFormElement,
  validationType: string,
) {
  try {
    validation.clearError(input.id);
    input.setAttribute("aria-invalid", "false");

    const result = await validation.validateForm(form, validationType);

    if (!result.success) {
      result.error?.errors?.forEach((err) => {
        if (err.path[0] === input.id) {
          validation.showError(input.id, err.message);
          input.setAttribute("aria-invalid", "true");
        }
      });
    }
  } catch (error) {
    console.error("Validation error:", error);
    displayFormError(form, "Validation error occurred");
  }
}

/**
 * Handles form submission including validation and server communication
 */
async function handleFormSubmission(
  event: Event,
  form: HTMLFormElement,
  validationType: string,
) {
  event.preventDefault();
  validation.clearAllErrors();

  // Reset aria-invalid attributes
  form
    .querySelectorAll<HTMLInputElement>("input[id]")
    .forEach((input) => input.setAttribute("aria-invalid", "false"));

  try {
    // Validate form
    const result = await validation.validateForm(form, validationType);
    if (!result.success) {
      throw result.error; // Throw for unified error handling
    }

    // Submit form data and handle response
    const response = await sendFormData(form);
    await handleServerResponse(response, form);
  } catch (error) {
    console.error("Form submission error:", error);
    displayFormError(form, "An error occurred. Please try again.");
  }
}

/**
 * Sends form data to the server via AJAX
 */
async function sendFormData(form: HTMLFormElement) {
  return fetch(form.action, {
    method: "POST",
    body: new FormData(form),
    credentials: "include",
    headers: { "X-Requested-With": "XMLHttpRequest" },
  });
}

/**
 * Handles the server's response to the form submission
 */
async function handleServerResponse(response: Response, form: HTMLFormElement) {
  try {
    const data = await response.json();

    if (response.redirected || data.redirect) {
      window.location.href = response.redirected ? response.url : data.redirect;
    } else if (!response.ok && data.message) {
      displayFormError(form, data.message);
    }
  } catch (error) {
    console.error("Error handling server response:", error);
    displayFormError(form, "Error processing server response");
  }
}

/**
 * Displays an error message in the form's error container
 * Uses a class selector for more flexible error container targeting
 */
function displayFormError(form: HTMLFormElement, message: string) {
  const formError = form.querySelector(".form-error");
  if (formError) {
    formError.textContent = message;
  } else {
    console.warn("Form error container not found:", form.id);
  }
}
