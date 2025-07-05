// ===== src/js/forms/handlers/request-handler.ts =====
import { Logger } from "@/core/logger";
import { HttpClient } from "@/core/http-client";
import { FormBuilderError } from "@/core/errors/form-builder-error";
import { isAuthenticationEndpoint } from "@/shared/utils/endpoint-utils";

export class RequestHandler {
  /**
   * Sends form data to the server via AJAX
   */
  static async sendFormData(form: HTMLFormElement): Promise<Response> {
    Logger.group("Form Submission");

    try {
      const formData = new FormData(form);
      const isAuthEndpoint = isAuthenticationEndpoint(form.action);

      Logger.debug("Sending request to:", form.action);
      Logger.debug("Cookies that will be sent:", document.cookie);

      if (isAuthEndpoint) {
        // For auth endpoints, convert to JSON and use HttpClient
        const data = Object.fromEntries(formData.entries());
        delete data["csrf_token"]; // Remove from payload since HttpClient handles it

        Logger.debug("Cleaned Form Data:", data);

        // For auth endpoints, we need to handle redirects properly
        // Use fetch directly instead of HttpClient to get the actual response
        const csrfToken = document.querySelector<HTMLMetaElement>('meta[name="csrf-token"]')?.content;

        const response = await fetch(form.action, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            'Accept': 'application/json',
            'X-Csrf-Token': csrfToken || '',
            'X-Requested-With': 'XMLHttpRequest',
          },
          body: JSON.stringify(data),
          credentials: 'include',
        });

        Logger.debug("Auth response status:", response.status);
        Logger.debug("Auth response headers:", Object.fromEntries(response.headers.entries()));

        // If we get a redirect, follow it
        if (response.redirected) {
          Logger.debug("Redirecting to:", response.url);
          window.location.href = response.url;
          return new Response(null, { status: 200 }); // Return success to prevent error handling
        }

        // For non-redirect responses, return the actual response
        return response;
      } else {
        // For non-auth endpoints, use HttpClient with FormData
        const response = await HttpClient.post(form.action, formData);

        // Create a proper Response object for the ResponseHandler
        return new Response(JSON.stringify(response), {
          status: 200,
          headers: {
            "Content-Type": "application/json",
            "access-control-allow-credentials": "true",
            "access-control-allow-headers":
              "Content-Type,Authorization,X-Csrf-Token,X-Requested-With",
            "access-control-allow-methods": "GET,POST,PUT,DELETE,OPTIONS",
            "access-control-allow-origin": "http://localhost:8090",
          },
        });
      }
    } catch (error) {
      Logger.error("Request failed:", error);

      // Convert to FormBuilderError for consistent error handling
      if (error instanceof FormBuilderError) {
        throw error;
      }

      // Handle different types of errors
      if (error instanceof TypeError) {
        throw FormBuilderError.networkError(
          "Network connection failed",
          form.action,
        );
      }

      // Handle HTTP errors from HttpClient
      if (error && typeof error === "object" && "status" in error) {
        const status = (error as any).status;
        const message = (error as any).message ?? "Request failed";

        switch (status) {
          case 400:
            throw FormBuilderError.validationError(
              "Invalid form data",
              undefined,
              new FormData(form),
            );
          case 403:
            throw FormBuilderError.csrfError("CSRF token validation failed");
          case 404:
            throw FormBuilderError.loadFailed(
              "Form endpoint not found",
              form.action,
            );
          case 429:
            throw FormBuilderError.networkError(
              "Rate limit exceeded",
              form.action,
              status,
            );
          case 500:
            throw FormBuilderError.saveFailed(
              "Server error occurred",
              form.action,
              error as unknown as Error,
            );
          default:
            throw FormBuilderError.networkError(
              `Server error: ${message}`,
              form.action,
              status,
            );
        }
      }

      // Generic error handling
      throw FormBuilderError.networkError("Unknown network error", form.action);
    } finally {
      Logger.groupEnd();
    }
  }
}
