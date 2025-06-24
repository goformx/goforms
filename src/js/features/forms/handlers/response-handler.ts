// ===== src/js/forms/handlers/response-handler.ts =====
import type { ServerResponse } from "@/shared/types/form-types";
import { UIManager } from "./ui-manager";

export class ResponseHandler {
  /**
   * Handles the server's response to the form submission
   */
  static async handleServerResponse(
    response: Response,
    form: HTMLFormElement,
  ): Promise<void> {
    console.group("Response Handler");

    try {
      const contentType = response.headers.get("content-type");
      let data: ServerResponse = {};

      // Only try to parse JSON if the response has JSON content
      if (contentType && contentType.includes("application/json")) {
        data = await response.json();
        console.log("Response data:", data);
      } else {
        console.log("Response has no JSON content, status:", response.status);
      }

      if (response.redirected || data.redirect) {
        const redirectUrl = response.redirected ? response.url : data.redirect!;
        console.log("Redirecting to:", redirectUrl);
        window.location.href = redirectUrl;
        return;
      }

      if (!response.ok) {
        const message = data.message || "An error occurred. Please try again.";
        console.warn("Request failed:", message);
        UIManager.displayFormError(form, message);
        return;
      }

      if (data.message) {
        console.log("Success message:", data.message);
        UIManager.displayFormSuccess(form, data.message);
      }
    } catch (error) {
      console.error("Error handling server response:", error);
      UIManager.displayFormError(form, "Error processing server response");
    } finally {
      console.groupEnd();
    }
  }
}
