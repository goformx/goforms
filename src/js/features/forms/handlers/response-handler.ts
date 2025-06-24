// ===== src/js/forms/handlers/response-handler.ts =====
import { Logger } from "@/core/logger";
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
    Logger.group("Response Handler");

    try {
      const contentType = response.headers.get("content-type");
      let data: ServerResponse = {};

      // Only try to parse JSON if the response has JSON content
      if (contentType && contentType.includes("application/json")) {
        data = await response.json();
        Logger.debug("Response data:", data);
      } else {
        Logger.debug("Response has no JSON content, status:", response.status);
      }

      if (response.redirected || data.redirect) {
        const redirectUrl = response.redirected ? response.url : data.redirect!;
        Logger.debug("Redirecting to:", redirectUrl);
        window.location.href = redirectUrl;
        return;
      }

      if (!response.ok) {
        const message = data.message || "An error occurred. Please try again.";
        Logger.warn("Request failed:", message);
        UIManager.displayFormError(form, message);
        return;
      }

      if (data.message) {
        Logger.debug("Success message:", data.message);
        UIManager.displayFormSuccess(form, data.message);
      }
    } catch (error) {
      Logger.error("Error handling server response:", error);
      UIManager.displayFormError(form, "Error processing server response");
    } finally {
      Logger.groupEnd();
    }
  }
}
