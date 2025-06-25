// ===== src/js/forms/handlers/request-handler.ts =====
import { Logger } from "@/core/logger";
import { HttpClient } from "@/core/http-client";
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
        delete data.csrf_token; // Remove from payload since HttpClient handles it

        Logger.debug("Cleaned Form Data:", data);

        const response = await HttpClient.post(form.action, data);

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

      // Create an error response
      return new Response(
        JSON.stringify({
          success: false,
          message: error instanceof Error ? error.message : "Request failed",
        }),
        {
          status: 500,
          headers: {
            "Content-Type": "application/json",
            "access-control-allow-credentials": "true",
            "access-control-allow-headers":
              "Content-Type,Authorization,X-Csrf-Token,X-Requested-With",
            "access-control-allow-methods": "GET,POST,PUT,DELETE,OPTIONS",
            "access-control-allow-origin": "http://localhost:8090",
          },
        },
      );
    } finally {
      Logger.groupEnd();
    }
  }
}
