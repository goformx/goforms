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
  const formError = document.getElementById("form_error") as HTMLDivElement;

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
        const result = await validation.validateForm(form, "login");
        if (!result.success && result.error) {
          result.error.errors.forEach((err) => {
            if (err.path[0] === inputElement.id) {
              validation.showError(inputElement.id, err.message);
            }
          });
        }
      });
    });

    form.addEventListener("submit", async (e) => {
      e.preventDefault();

      // Clear previous errors
      validation.clearAllErrors();

      const result = await validation.validateForm(form, "login");
      if (result.success && result.data) {
        try {
          const response = await validation.fetchWithCSRF(
            "/api/v1/auth/login",
            {
              method: "POST",
              headers: {
                "Content-Type": "application/json",
              },
              body: JSON.stringify({
                Email: result.data.email,
                Password: result.data.password,
              }),
            },
          );

          if (response.ok) {
            const tokens = await response.json();
            validation.setJWTToken(tokens.AccessToken);
            window.location.href = "/dashboard";
          } else {
            const error = await response.json();
            if (error.errors) {
              // Display field-specific errors
              Object.entries(error.errors).forEach(([field, message]) => {
                validation.showError(field, message as string);
              });
            } else {
              formError.textContent =
                error.message || "An error occurred during login";
            }
          }
        } catch (_error) {
          formError.textContent = "An unexpected error occurred";
        }
      } else if (result.error) {
        // Display validation errors
        result.error.errors.forEach((err) => {
          validation.showError(err.path[0], err.message);
        });
      }
    });
  }
}
