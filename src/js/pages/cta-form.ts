import { Logger } from "@/core/logger";
import { Formio } from "@formio/js";
import { FormService } from "@/features/forms/services/form-service";

const CTA_FORM_ID = "61af2a0f-5b54-476f-9bf6-c2ee6ce5b822";
const BASE_URL = window.location.origin;

class CTAForm {
  private readonly formService: FormService;
  private formInstance: any;

  constructor() {
    this.formService = FormService.getInstance();
  }

  async initialize(containerId: string): Promise<void> {
    try {
      const container = document.getElementById(containerId);
      if (!container) {
        throw new Error(`Container element #${containerId} not found`);
      }

      // Fetch the form schema
      const schema = await this.formService.getSchema(CTA_FORM_ID);

      // Create the form instance
      this.formInstance = await Formio.createForm(container, schema, {
        readOnly: false,
        noAlerts: true,
        i18n: {
          en: {
            submit: "Get Started",
          },
        },
      });

      // Handle form submission
      this.formInstance.on("submit", async (submission: any) => {
        try {
          // Submit the form data
          const response = await fetch(
            `${BASE_URL}/api/v1/forms/${CTA_FORM_ID}/submissions`,
            {
              method: "POST",
              headers: {
                "Content-Type": "application/json",
              },
              body: JSON.stringify(submission.data),
            },
          );

          if (!response.ok) {
            throw new Error("Failed to submit form");
          }

          // Show success message
          this.showMessage("Thank you! We'll be in touch soon.", "success");
          this.formInstance.resetForm();
        } catch (error) {
          Logger.error("Form submission error:", error);
          this.showMessage("Something went wrong. Please try again.", "error");
        }
      });
    } catch (error) {
      Logger.error("Failed to initialize CTA form:", error);
      this.showMessage(
        "Failed to load form. Please refresh the page.",
        "error",
      );
    }
  }

  private showMessage(message: string, type: "success" | "error"): void {
    const messageContainer = document.createElement("div");
    messageContainer.className = `gf-message gf-message--${type}`;
    messageContainer.textContent = message;

    const formContainer = this.formInstance.element;
    formContainer.parentNode.insertBefore(
      messageContainer,
      formContainer.nextSibling,
    );

    // Remove message after 5 seconds
    setTimeout(() => {
      messageContainer.remove();
    }, 5000);
  }
}

// Initialize the CTA form when the DOM is ready
document.addEventListener("DOMContentLoaded", () => {
  const ctaForm = new CTAForm();
  ctaForm.initialize("cta-form-container");
});
