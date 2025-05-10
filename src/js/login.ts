import { validation } from "./validation";

let isInitialized = false;

// Initialize validation when the page loads
document.addEventListener("DOMContentLoaded", () => {
  console.log("DOMContentLoaded event fired");
  if (!isInitialized) {
    setupLoginForm();
    isInitialized = true;
  }
});

export function setupLoginForm() {
  console.log("Setting up login form");
  const form = document.getElementById("login-form") as HTMLFormElement;
  const formError = document.getElementById("form_error") as HTMLDivElement;

  if (form) {
    console.log("Login form found, form ID:", form.id);
    // Setup real-time validation
    console.log("Setting up real-time validation");
    validation.setupRealTimeValidation("login-form", "login");

    // Add input event listeners for real-time validation
    const inputs = form.querySelectorAll("input[id]");
    console.log(
      "Found inputs:",
      Array.from(inputs).map((input) => input.id),
    );
    inputs.forEach((input) => {
      if (!input.id) return;
      console.log("Adding input listener for:", input.id);
      const inputElement = input as HTMLInputElement;
      inputElement.addEventListener("input", async () => {
        console.log(
          "Input event for:",
          inputElement.id,
          "Value:",
          inputElement.value,
        );
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
      console.log("Form submit event");
      e.preventDefault();

      // Clear previous errors
      validation.clearAllErrors();

      const result = await validation.validateForm(form, "login");
      if (result.success && result.data) {
        console.log("Form validation successful, submitting...");
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
            console.log("Login successful");
            const tokens = await response.json();
            console.log("Login response tokens:", tokens);
            validation.setJWTToken(tokens.AccessToken);
            console.log("JWT token stored:", validation.getJWTToken());
            console.log("JWT token stored, redirecting to dashboard");
            window.location.href = "/dashboard";
          } else {
            console.log("Login failed with status:", response.status);
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
        } catch (error) {
          console.error("Unexpected error during login:", error);
          formError.textContent = "An unexpected error occurred";
        }
      } else if (result.error) {
        console.log("Form validation failed:", result.error);
        // Display validation errors
        result.error.errors.forEach((err) => {
          validation.showError(err.path[0], err.message);
        });
      }
    });
  } else {
    console.error("Login form not found");
  }
}
