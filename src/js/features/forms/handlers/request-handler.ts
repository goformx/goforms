// ===== src/js/forms/handlers/request-handler.ts =====
import { validation } from "../validation/validation";
import type { RequestOptions } from "@/shared/types/form-types";
import { isAuthenticationEndpoint } from "@/shared/utils/endpoint-utils";

export class RequestHandler {
  /**
   * Sends form data to the server via AJAX
   */
  static async sendFormData(form: HTMLFormElement): Promise<Response> {
    console.group("Form Submission");

    try {
      const csrfToken = validation.getCSRFToken();
      const formData = new FormData(form);
      const isAuthEndpoint = isAuthenticationEndpoint(form.action);

      console.log("CSRF Token:", csrfToken ? "Present" : "Missing");
      console.log("Sending request to:", form.action);
      console.log("Cookies that will be sent:", document.cookie);

      const { body, headers } = this.prepareRequestData(
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
   * Prepares request data and headers based on endpoint type
   */
  private static prepareRequestData(
    formData: FormData,
    isAuthEndpoint: boolean,
    csrfToken: string | null,
  ): RequestOptions {
    const headers: Record<string, string> = {
      Accept: "application/json",
      "X-Requested-With": "XMLHttpRequest",
    };

    // Always add CSRF token to headers if available
    if (csrfToken) {
      headers["X-Csrf-Token"] = csrfToken;
    }

    if (isAuthEndpoint) {
      const data = Object.fromEntries(formData.entries());
      // Remove CSRF token from payload since it's in the header
      delete data.csrf_token;

      headers["Content-Type"] = "application/json";
      console.log("Cleaned Form Data:", data);
      return { body: JSON.stringify(data), headers };
    } else {
      // For non-auth endpoints, keep CSRF token in form data AND headers
      console.log("Form Data:", Object.fromEntries(formData.entries()));
      return { body: formData, headers };
    }
  }
}
