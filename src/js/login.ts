import { validation } from "./validation";

let isInitialized = false;

// Initialize validation when the page loads
document.addEventListener("DOMContentLoaded", () => {
  if (!isInitialized) {
    setupLoginForm();
    isInitialized = true;
  }
});

export function setupLoginForm() {
  const form = document.getElementById("login-form") as HTMLFormElement;

  if (form) {
    // Setup real-time validation
    validation.setupRealTimeValidation("login-form", "login");

    // Add input event listeners for real-time validation
    const inputs = form.querySelectorAll("input[id]");
    inputs.forEach((input) => {
      if (!input.id) return;
      const inputElement = input as HTMLInputElement;
      inputElement.addEventListener("input", async () => {
        validation.clearError(inputElement.id);
        inputElement.setAttribute("aria-invalid", "false");
        const result = await validation.validateForm(form, "login");
        if (!result.success && result.error) {
          result.error.errors.forEach((err) => {
            if (err.path[0] === inputElement.id) {
              validation.showError(inputElement.id, err.message);
              inputElement.setAttribute("aria-invalid", "true");
            }
          });
        }
      });
    });

    // Add form submit validation
    form.addEventListener("submit", async (e) => {
      e.preventDefault(); // Prevent default to handle validation first

      // Clear any existing errors
      validation.clearAllErrors();
      inputs.forEach((input) => {
        if (!input.id) return;
        const inputElement = input as HTMLInputElement;
        inputElement.setAttribute("aria-invalid", "false");
      });

      const result = await validation.validateForm(form, "login");

      if (!result.success) {
        if (result.error) {
          result.error.errors.forEach((err) => {
            validation.showError(err.path[0], err.message);
            const input = document.getElementById(
              err.path[0],
            ) as HTMLInputElement;
            if (input) {
              input.setAttribute("aria-invalid", "true");
            }
          });
        }
        return;
      }

      // If validation passes, submit the form
      form.submit();
    });
  }
}
