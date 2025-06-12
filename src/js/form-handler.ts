/**
 * Form Handler Module
 *
 * Provides shared functionality for form handling including:
 * - Real-time validation
 * - Form submission
 * - Error handling
 * - Server communication
 */

import { validation } from "./validation";

export interface FormConfig {
  formId: string;
  validationType: string;
}

/**
 * Sets up a form with validation and submission handling
 *
 * @param config - Form configuration including form ID and validation type
 */
export function setupForm(config: FormConfig) {
  const form = document.getElementById(config.formId) as HTMLFormElement | null;
  if (!form) return;

  // Initialize validation
  validation.setupRealTimeValidation(form.id, config.validationType);

  // Setup real-time validation
  setupRealTimeValidation(form, config.validationType);

  // Setup form submission
  form.addEventListener("submit", (event) =>
    handleFormSubmission(event, form, config.validationType),
  );
}

/**
 * Sets up real-time validation for form inputs
 */
function setupRealTimeValidation(
  form: HTMLFormElement,
  validationType: string,
) {
  form.querySelectorAll<HTMLInputElement>("input[id]").forEach((input) => {
    input.addEventListener("input", () =>
      handleInputValidation(input, form, validationType),
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

  // Validate form
  const result = await validation.validateForm(form, validationType);
  if (!result.success) {
    result.error?.errors?.forEach((err) => {
      validation.showError(err.path[0], err.message);
      document
        .getElementById(err.path[0])
        ?.setAttribute("aria-invalid", "true");
    });
    return;
  }

  try {
    const response = await sendFormData(form);
    await handleServerResponse(response);
  } catch (error) {
    console.error("Error submitting form:", error);
    displayFormError("An error occurred. Please try again.");
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
async function handleServerResponse(response: Response) {
  const data = await response.json();

  if (response.redirected || data.redirect) {
    window.location.href = response.redirected ? response.url : data.redirect;
  } else if (!response.ok && data.message) {
    displayFormError(data.message);
  }
}

/**
 * Displays an error message in the form's error container
 */
function displayFormError(message: string) {
  const formError = document.getElementById("form_error");
  if (formError) formError.textContent = message;
}
