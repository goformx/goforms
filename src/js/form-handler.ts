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
  console.group("Form Submission");
  const csrfToken = validation.getCSRFToken();
  console.log("CSRF Token:", csrfToken ? "Present" : "Missing");
  // Convert FormData to JSON for auth endpoints
  const formData = new FormData(form);
  const isAuthEndpoint =
    form.action.includes("/login") || form.action.includes("/signup");
  // For auth endpoints, clean up the data before sending
  let body: FormData | string;
  if (isAuthEndpoint) {
    const data = Object.fromEntries(formData.entries());
    // Remove CSRF token from payload since it's in the header
    delete data.csrf_token;
    body = JSON.stringify(data);
    console.log("Cleaned Form Data:", data);
  } else {
    body = formData;
    console.log("Form Data:", Object.fromEntries(formData.entries()));
  }
  try {
    console.log("Sending request to:", form.action);
    const response = await fetch(form.action, {
      method: "POST",
      body,
      credentials: "include",
      headers: {
        "Accept": "application/json",
        "X-CSRF-Token": csrfToken,
        ...(isAuthEndpoint && { "Content-Type": "application/json" }),
      },
    });
    console.log("Response status:", response.status);
    console.log(
      "Response headers:",
      Object.fromEntries(response.headers.entries()),
    );
    return response;
  } catch (error) {
    console.error("Request failed:", error);
    throw error;
  } finally {
    console.groupEnd();
  }
}

/**
 * Handles the server's response to the form submission
 */
async function handleServerResponse(response: Response, form: HTMLFormElement) {
  console.group("Response Handler");
  try {
    const data = await response.json();
    console.log("Response data:", data);

    if (response.redirected || data.redirect) {
      const redirectUrl = response.redirected ? response.url : data.redirect;
      console.log("Redirecting to:", redirectUrl);
      window.location.href = redirectUrl;
      return;
    }

    if (!response.ok) {
      const message = data.message || "An error occurred. Please try again.";
      console.warn("Request failed:", message);
      displayFormError(form, message);
      return;
    }

    // Handle successful response without redirect
    if (data.message) {
      console.log("Success message:", data.message);
      displayFormSuccess(form, data.message);
    }
  } catch (error) {
    console.error("Error handling server response:", error);
    displayFormError(form, "Error processing server response");
  } finally {
    console.groupEnd();
  }
}

/**
 * Displays an error message in the form's error container
 */
function displayFormError(form: HTMLFormElement, message: string) {
  console.debug("Displaying error message:", message);
  const formError = form.querySelector(".form-error");
  if (formError) {
    formError.textContent = message;
    formError.classList.remove("hidden");
  } else {
    console.warn("Form error container not found:", form.id);
  }
}

/**
 * Displays a success message in the form's success container
 */
function displayFormSuccess(form: HTMLFormElement, message: string) {
  console.debug("Displaying success message:", message);
  const formSuccess = form.querySelector(".form-success");
  if (formSuccess) {
    formSuccess.textContent = message;
    formSuccess.classList.remove("hidden");
  } else {
    console.warn("Form success container not found:", form.id);
  }
}

export class FormHandler {
  private form: HTMLFormElement;

  constructor(config: FormConfig) {
    const formElement = document.querySelector<HTMLFormElement>(
      `#${config.formId}`,
    );
    if (!formElement) {
      throw new Error(`Form with ID "${config.formId}" not found`);
    }
    this.form = formElement;

    validation.setupRealTimeValidation(this.form.id, config.validationType);
    setupRealTimeValidation(
      this.form,
      config.validationType,
      config.validationDelay,
    );

    // Setup form submission
    this.form.addEventListener("submit", (event) =>
      this.handleFormSubmission(event, config.validationType),
    );
  }

  private async sendFormData(formData: FormData): Promise<Response> {
    console.group("Form Submission");
    const csrfToken = validation.getCSRFToken();
    console.log("CSRF Token:", csrfToken ? "Present" : "Missing");
    // Convert FormData to JSON for auth endpoints
    const isAuthEndpoint =
      this.form.action.includes("/login") ||
      this.form.action.includes("/signup");
    // For auth endpoints, clean up the data before sending
    let body: FormData | string;
    if (isAuthEndpoint) {
      const data = Object.fromEntries(formData.entries());
      // Remove CSRF token from payload since it's in the header
      delete data.csrf_token;
      body = JSON.stringify(data);
      console.log("Cleaned Form Data:", data);
    } else {
      body = formData;
      console.log("Form Data:", Object.fromEntries(formData.entries()));
    }
    try {
      console.log("Sending request to:", this.form.action);
      const response = await fetch(this.form.action, {
        method: "POST",
        body,
        credentials: "include",
        headers: {
          "Accept": "application/json",
          "X-CSRF-Token": csrfToken,
          ...(isAuthEndpoint && { "Content-Type": "application/json" }),
        },
      });
      console.log("Response status:", response.status);
      console.log(
        "Response headers:",
        Object.fromEntries(response.headers.entries()),
      );
      return response;
    } catch (error) {
      console.error("Request failed:", error);
      throw error;
    } finally {
      console.groupEnd();
    }
  }

  private async handleFormSubmission(
    event: Event,
    validationType: string,
  ): Promise<void> {
    event.preventDefault();
    validation.clearAllErrors();

    this.form
      .querySelectorAll<HTMLInputElement>("input[id]")
      .forEach((input) => input.setAttribute("aria-invalid", "false"));

    try {
      const result = await validation.validateForm(this.form, validationType);
      if (!result.success) {
        const errorMessage =
          result.error?.errors?.[0]?.message ||
          "Please check the form for errors.";
        this.showError(errorMessage);
        return;
      }

      const formData = new FormData(this.form);
      const response = await this.sendFormData(formData);
      await handleServerResponse(response, this.form);
    } catch (error) {
      console.error("Form submission error:", error);
      this.showError("An unexpected error occurred. Please try again.");
    }
  }

  private showError(message: string, field?: string): void {
    console.debug(
      "Displaying error message:",
      message,
      field ? `for field: ${field}` : "",
    );
    const errorContainer = this.form.querySelector(".form-error");
    if (errorContainer) {
      errorContainer.textContent = message;
      errorContainer.classList.remove("hidden");
    }

    if (field) {
      const fieldElement = this.form.querySelector(`[name="${field}"]`);
      if (fieldElement) {
        fieldElement.classList.add("error");
      }
    }
  }
}
