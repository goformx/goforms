/**
 * Form Handler Module
 *
 * Provides shared functionality for form handling including:
 * - Real-time validation with debouncing
 * - Form submission with CSRF protection
 * - Centralized error handling
 * - Server communication
 */

import { validation } from "../validation/validation";

export interface FormConfig {
  formId: string;
  validationType: string;
  validationDelay?: number;
}

interface ServerResponse {
  message?: string;
  redirect?: string;
}

/**
 * Debounces a function to prevent excessive calls
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
 */
export function setupForm(config: FormConfig): void {
  const form = document.querySelector<HTMLFormElement>(`#${config.formId}`);
  if (!form) {
    console.error(`Form with ID "${config.formId}" not found`);
    return;
  }

  validation.setupRealTimeValidation(form.id, config.validationType);
  setupRealTimeValidation(form, config.validationType, config.validationDelay);

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
): void {
  const inputs = form.querySelectorAll<HTMLInputElement>("input[id]");

  inputs.forEach((input) => {
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
): Promise<void> {
  try {
    validation.clearError(input.id);
    input.setAttribute("aria-invalid", "false");

    const result = await validation.validateForm(form, validationType);

    if (!result.success && result.error?.errors) {
      result.error.errors.forEach((err) => {
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
): Promise<void> {
  event.preventDefault();
  validation.clearAllErrors();

  // Reset aria-invalid attributes
  const inputs = form.querySelectorAll<HTMLInputElement>("input[id]");
  inputs.forEach((input) => input.setAttribute("aria-invalid", "false"));

  try {
    const result = await validation.validateForm(form, validationType);
    if (!result.success) {
      throw result.error;
    }

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
async function sendFormData(form: HTMLFormElement): Promise<Response> {
  console.group("Form Submission");

  try {
    const csrfToken = validation.getCSRFToken();
    const formData = new FormData(form);
    const isAuthEndpoint = isAuthenticationEndpoint(form.action);

    console.log("CSRF Token:", csrfToken ? "Present" : "Missing");
    console.log("Sending request to:", form.action);
    console.log("Cookies that will be sent:", document.cookie);

    const { body, headers } = prepareRequestData(
      formData,
      isAuthEndpoint,
      csrfToken,
    );

    const response = await fetch(form.action, {
      method: "POST",
      body,
      credentials: "include",
      headers,
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
 * Checks if the endpoint is an authentication endpoint
 */
function isAuthenticationEndpoint(action: string): boolean {
  return action.includes("/login") || action.includes("/signup");
}

/**
 * Prepares request data and headers based on endpoint type
 */
function prepareRequestData(
  formData: FormData,
  isAuthEndpoint: boolean,
  csrfToken: string | null,
): { body: FormData | string; headers: Record<string, string> } {
  const headers: Record<string, string> = {
    Accept: "application/json",
    "X-Requested-With": "XMLHttpRequest",
  };

  if (isAuthEndpoint) {
    const data = Object.fromEntries(formData.entries());
    delete data.csrf_token; // Remove CSRF token from payload since it's in the header

    if (csrfToken) {
      headers["X-Csrf-Token"] = csrfToken;
    }
    headers["Content-Type"] = "application/json";

    console.log("Cleaned Form Data:", data);
    return { body: JSON.stringify(data), headers };
  } else {
    console.log("Form Data:", Object.fromEntries(formData.entries()));
    return { body: formData, headers };
  }
}

/**
 * Handles the server's response to the form submission
 */
async function handleServerResponse(
  response: Response,
  form: HTMLFormElement,
): Promise<void> {
  console.group("Response Handler");

  try {
    const data: ServerResponse = await response.json();
    console.log("Response data:", data);

    if (response.redirected || data.redirect) {
      const redirectUrl = response.redirected ? response.url : data.redirect!;
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
function displayFormError(form: HTMLFormElement, message: string): void {
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
function displayFormSuccess(form: HTMLFormElement, message: string): void {
  console.debug("Displaying success message:", message);
  const formSuccess = form.querySelector(".form-success");

  if (formSuccess) {
    formSuccess.textContent = message;
    formSuccess.classList.remove("hidden");
  } else {
    console.warn("Form success container not found:", form.id);
  }
}

/**
 * Class-based form handler for more complex use cases
 */
export class FormHandler {
  private form: HTMLFormElement;
  private validationType: string;

  constructor(config: FormConfig) {
    const formElement = document.querySelector<HTMLFormElement>(
      `#${config.formId}`,
    );
    if (!formElement) {
      throw new Error(`Form with ID "${config.formId}" not found`);
    }

    this.form = formElement;
    this.validationType = config.validationType;

    this.initialize(config);
  }

  private initialize(config: FormConfig): void {
    validation.setupRealTimeValidation(this.form.id, config.validationType);
    setupRealTimeValidation(
      this.form,
      config.validationType,
      config.validationDelay,
    );

    this.form.addEventListener("submit", (event) =>
      this.handleFormSubmission(event),
    );
  }

  private async sendFormData(formData: FormData): Promise<Response> {
    console.group("Form Submission");

    try {
      const csrfToken = validation.getCSRFToken();
      console.log("CSRF Token from meta tag:", csrfToken);

      // Remove CSRF token from form data since we're using headers
      const cleanFormData = new FormData();
      for (const [key, value] of formData.entries()) {
        if (key !== "csrf_token") {
          cleanFormData.append(key, value);
        }
      }

      console.log("Sending request to:", this.form.action);
      console.log("Cookies that will be sent:", document.cookie);

      return validation.fetchWithAuth(this.form.action, {
        method: this.form.method,
        body: cleanFormData,
      });
    } finally {
      console.groupEnd();
    }
  }

  private async handleFormSubmission(event: Event): Promise<void> {
    event.preventDefault();
    validation.clearAllErrors();

    const inputs = this.form.querySelectorAll<HTMLInputElement>("input[id]");
    inputs.forEach((input) => input.setAttribute("aria-invalid", "false"));

    try {
      const result = await validation.validateForm(
        this.form,
        this.validationType,
      );
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
