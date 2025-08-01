import { Logger } from "@/core/logger";
import { FormBuilderError } from "@/core/errors/form-builder-error";

/**
 * Standardized API response structure
 */
export interface ApiResponse<T = unknown> {
  success: boolean;
  message?: string;
  data?: T;
}

/**
 * HTTP request options with AbortController support
 */
export interface HttpRequestOptions extends RequestInit {
  timeout?: number;
  signal?: AbortSignal;
}

/**
 * HTTP response wrapper with type safety
 */
export interface HttpResponse<T = unknown> {
  data: T;
  status: number;
  statusText: string;
  headers: Headers;
  url: string;
}

/**
 * Unified HTTP client for making authenticated requests
 */
export class HttpClient {
  private static readonly DEFAULT_TIMEOUT = 30000; // 30 seconds

  private static getCSRFToken(): string | null {
    const meta = document.querySelector<HTMLMetaElement>(
      "meta[name='csrf-token']",
    );
    if (!meta?.content) {
      // Return null instead of throwing error when CSRF token is not found
      // This allows the application to work when CSRF is disabled
      return null;
    }
    return meta.content;
  }

  private static createAbortController(timeout?: number): AbortController {
    const controller = new AbortController();

    if (timeout) {
      setTimeout(() => controller.abort(), timeout);
    }

    return controller;
  }

  private static async handleResponse<T>(
    response: Response,
  ): Promise<HttpResponse<T>> {
    Logger.log("Response status:", response.status);
    Logger.log(
      "Response headers:",
      Object.fromEntries(response.headers.entries()),
    );

    if (!response.ok) {
      throw FormBuilderError.networkError(
        `HTTP ${response.status}: ${response.statusText}`,
        response.url,
        response.status,
      );
    }

    // Try to parse JSON, fallback to text for empty responses
    const text = await response.text();
    let data: T;

    if (!text) {
      data = null as T;
    } else {
      try {
        const parsed = JSON.parse(text) as ApiResponse<T>;

        // Handle standardized API response format
        if (parsed && typeof parsed === "object" && "success" in parsed) {
          if (!parsed.success) {
            throw FormBuilderError.networkError(
              parsed.message ?? "API request failed",
              response.url,
              response.status,
            );
          }
          // Return the data field from the standardized response
          data = parsed.data as T;
        } else {
          // Handle legacy response format (direct data)
          data = parsed as T;
        }
      } catch (parseError) {
        // If JSON parsing fails, return the text
        Logger.warn("Failed to parse JSON response, returning text:", text);
        Logger.debug("Parse error details:", parseError);
        data = text as T;
      }
    }

    return {
      data,
      status: response.status,
      statusText: response.statusText,
      headers: response.headers,
      url: response.url,
    };
  }

  private static isPublicEndpoint(url: string): boolean {
    // Check if this is a public API endpoint that doesn't require CSRF
    const publicEndpoints = [
      "/api/v1/forms/", // Form submission endpoints
      "/health",
      "/healthz",
    ];

    return publicEndpoints.some((endpoint) => url.includes(endpoint));
  }

  private static async makeRequest<T>(
    url: string,
    options: HttpRequestOptions = {},
  ): Promise<HttpResponse<T>> {
    Logger.group("HTTP Request");
    Logger.log("URL:", url);
    Logger.log("Method:", options.method ?? "GET");

    const timeout = options.timeout ?? this.DEFAULT_TIMEOUT;
    const abortController = this.createAbortController(timeout);

    // Merge signals if both are provided
    const signal = options.signal
      ? this.mergeSignals([options.signal, abortController.signal])
      : abortController.signal;

    try {
      const headers = new Headers(options.headers);

      // Add CSRF token for non-GET requests only if CSRF is enabled and not a public endpoint
      if (options.method !== "GET" && !this.isPublicEndpoint(url)) {
        const csrfToken = this.getCSRFToken();
        if (csrfToken) {
          headers.set("X-Csrf-Token", csrfToken);
          Logger.log("CSRF Token included in request");
        } else {
          throw new Error("CSRF token not found");
        }
      } else if (this.isPublicEndpoint(url)) {
        Logger.log("Skipping CSRF token for public endpoint");
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
        signal,
      });

      return await this.handleResponse<T>(response);
    } catch (error) {
      Logger.error("Request failed:", error);

      // Handle abort errors
      if (error instanceof Error && error.name === "AbortError") {
        throw FormBuilderError.networkError("Request timeout", url);
      }

      // If it's already a FormBuilderError, re-throw it
      if (error instanceof FormBuilderError) {
        throw error;
      }

      // Convert other errors to FormBuilderError
      throw FormBuilderError.networkError(
        `Network error: ${error instanceof Error ? error.message : String(error)}`,
        url,
      );
    } finally {
      Logger.groupEnd();
    }
  }

  /**
   * Merge multiple AbortSignals into one
   */
  private static mergeSignals(signals: AbortSignal[]): AbortSignal {
    const controller = new AbortController();

    signals.forEach((signal) => {
      if (signal.aborted) {
        controller.abort();
      } else {
        signal.addEventListener("abort", () => controller.abort(), {
          once: true,
        });
      }
    });

    return controller.signal;
  }

  static async get<T = unknown>(
    url: string,
    options: HttpRequestOptions = {},
  ): Promise<HttpResponse<T>> {
    return this.makeRequest<T>(url, { ...options, method: "GET" });
  }

  static async post<T = unknown>(
    url: string,
    body?: FormData | string | object,
    options: HttpRequestOptions = {},
  ): Promise<HttpResponse<T>> {
    let requestBody: FormData | string | null = null;

    // Convert objects to JSON string
    if (body && typeof body === "object" && !(body instanceof FormData)) {
      requestBody = JSON.stringify(body);
    } else if (body) {
      requestBody = body as FormData | string;
    }

    return this.makeRequest<T>(url, {
      ...options,
      method: "POST",
      body: requestBody,
    });
  }

  static async put<T = unknown>(
    url: string,
    body?: FormData | string | object,
    options: HttpRequestOptions = {},
  ): Promise<HttpResponse<T>> {
    let requestBody: FormData | string | null = null;

    // Convert objects to JSON string
    if (body && typeof body === "object" && !(body instanceof FormData)) {
      requestBody = JSON.stringify(body);
    } else if (body) {
      requestBody = body as FormData | string;
    }

    return this.makeRequest<T>(url, {
      ...options,
      method: "PUT",
      body: requestBody,
    });
  }

  static async delete<T = unknown>(
    url: string,
    options: HttpRequestOptions = {},
  ): Promise<HttpResponse<T>> {
    return this.makeRequest<T>(url, { ...options, method: "DELETE" });
  }

  /**
   * Create a new HTTP client instance with default options
   */
  static create(defaultOptions: HttpRequestOptions = {}): typeof HttpClient {
    return class extends HttpClient {
      static override async get<T = unknown>(
        url: string,
        options: HttpRequestOptions = {},
      ): Promise<HttpResponse<T>> {
        return super.get<T>(url, { ...defaultOptions, ...options });
      }

      static override async post<T = unknown>(
        url: string,
        body?: FormData | string | object,
        options: HttpRequestOptions = {},
      ): Promise<HttpResponse<T>> {
        return super.post<T>(url, body, { ...defaultOptions, ...options });
      }

      static override async put<T = unknown>(
        url: string,
        body?: FormData | string | object,
        options: HttpRequestOptions = {},
      ): Promise<HttpResponse<T>> {
        return super.put<T>(url, body, { ...defaultOptions, ...options });
      }

      static override async delete<T = unknown>(
        url: string,
        options: HttpRequestOptions = {},
      ): Promise<HttpResponse<T>> {
        return super.delete<T>(url, { ...defaultOptions, ...options });
      }
    };
  }
}
