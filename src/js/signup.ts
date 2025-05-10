import { validation } from "./validation";

let isInitialized = false;
let initializationCount = 0;

// Initialize validation when the page loads
document.addEventListener("DOMContentLoaded", () => {
  console.log("DOMContentLoaded event fired", new Error().stack);
  if (!isInitialized) {
    initializationCount++;
    console.log("Initialization attempt #", initializationCount);
    setupSignupForm();
    isInitialized = true;
  } else {
    console.log("Preventing duplicate initialization");
  }
});

export function setupSignupForm() {
  console.log("Setting up signup form", new Error().stack);
  const form = document.getElementById("signup-form") as HTMLFormElement;
  const formError = document.getElementById("form_error") as HTMLDivElement;

  if (form) {
    console.log("Signup form found, form ID:", form.id);
    // Setup real-time validation
    console.log("Setting up real-time validation");
    validation.setupRealTimeValidation("signup-form", "signup");

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
        const result = await validation.validateForm(form, "signup");
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

      const result = await validation.validateForm(form, "signup");
      if (result.success) {
        console.log("Form validation successful, submitting...");
        try {
          const response = await validation.fetchWithCSRF(
            "/api/v1/auth/signup",
            {
              method: "POST",
              headers: {
                "Content-Type": "application/json",
              },
              body: JSON.stringify(result.data),
            },
          );

          if (response.ok) {
            console.log("Signup successful, redirecting to dashboard");
            window.location.href = "/dashboard";
          } else {
            console.log("Signup failed with status:", response.status);
            const error = await response.json();
            if (error.errors) {
              // Display field-specific errors
              Object.entries(error.errors).forEach(([field, message]) => {
                validation.showError(field, message as string);
              });
            } else {
              formError.textContent =
                error.message || "An error occurred during signup";
            }
          }
        } catch (error) {
          console.error("Unexpected error during signup:", error);
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
    console.error("Signup form not found");
  }
}
