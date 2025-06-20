import { Logger } from "./logger";

/**
 * Unified HTTP client for making authenticated requests
 */
export class HttpClient {
  private static getCSRFToken(): string {
    // Try meta tag first
    const meta = document.querySelector<HTMLMetaElement>(
      "meta[name='csrf-token']",
    );
    if (meta?.content) {
      return meta.content;
    }

    // Try hidden input as fallback
    const input = document.querySelector<HTMLInputElement>(
      "input[name='csrf_token']",
    );
    if (input?.value) {
      return input.value;
    }

    throw new Error(
      "CSRF token not found. Please refresh the page and try again.",
    );
  }

  private static isAuthEndpoint(url: string): boolean {
    return url.includes("/login") || url.includes("/signup");
  }

  private static prepareFormDataForAuth(formData: FormData): FormData {
    const newFormData = new FormData();

    // Copy all form fields, renaming csrf_token to _csrf for middleware compatibility
    for (const [key, value] of formData.entries()) {
      if (key === "csrf_token") {
        newFormData.append("_csrf", value as string);
      } else {
        newFormData.append(key, value);
      }
    }

    Logger.debug(
      "Prepared auth FormData:",
      Object.fromEntries(newFormData.entries()),
    );
    return newFormData;
  }

  static async request(
    url: string,
    options: RequestInit = {},
  ): Promise<Response> {
    Logger.group("HTTP Request");
    Logger.log("URL:", url);
    Logger.log("Method:", options.method || "GET");

    try {
      const headers = new Headers(options.headers);

      // Add CSRF token for non-GET requests
      if (options.method !== "GET") {
        try {
          const csrfToken = this.getCSRFToken();
          if (this.isAuthEndpoint(url)) {
            headers.set("X-Csrf-Token", csrfToken);
          }
          Logger.log("CSRF Token:", csrfToken ? "Present" : "Missing");
        } catch (error) {
          Logger.error("CSRF token error:", error);
          throw error;
        }
      }

      // Set content type if not FormData
      if (!(options.body instanceof FormData)) {
        headers.set("Content-Type", "application/json");
      }

      // Add common headers
      headers.set("Accept", "application/json");
      headers.set("X-Requested-With", "XMLHttpRequest");

      Logger.log("Headers:", Object.fromEntries(headers.entries()));
      Logger.log("Cookies:", document.cookie);

      const response = await fetch(url, {
        ...options,
        headers,
        credentials: "include",
      });

      Logger.log("Response status:", response.status);
      Logger.log(
        "Response headers:",
        Object.fromEntries(response.headers.entries()),
      );

      return response;
    } catch (error) {
      Logger.error("Request failed:", error);
      throw error;
    } finally {
      Logger.groupEnd();
    }
  }

  static async get(url: string, options: RequestInit = {}): Promise<Response> {
    return this.request(url, { ...options, method: "GET" });
  }

  static async post(
    url: string,
    body?: FormData | string,
    options: RequestInit = {},
  ): Promise<Response> {
    let processedBody = body;

    // Special handling for auth endpoints with FormData
    if (body instanceof FormData && this.isAuthEndpoint(url)) {
      processedBody = this.prepareFormDataForAuth(body);
    }

    return this.request(url, {
      ...options,
      method: "POST",
      body: processedBody,
    });
  }

  static async put(
    url: string,
    body?: FormData | string,
    options: RequestInit = {},
  ): Promise<Response> {
    return this.request(url, {
      ...options,
      method: "PUT",
      body,
    });
  }

  static async delete(
    url: string,
    options: RequestInit = {},
  ): Promise<Response> {
    return this.request(url, { ...options, method: "DELETE" });
  }
}
