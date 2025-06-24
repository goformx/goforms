/**
 * Error codes for form builder operations
 */
export enum ErrorCode {
  VALIDATION_FAILED = "VALIDATION_FAILED",
  NETWORK_ERROR = "NETWORK_ERROR",
  SCHEMA_ERROR = "SCHEMA_ERROR",
  PERMISSION_DENIED = "PERMISSION_DENIED",
  SANITIZATION_ERROR = "SANITIZATION_ERROR",
  CSRF_ERROR = "CSRF_ERROR",
  FORM_NOT_FOUND = "FORM_NOT_FOUND",
  INVALID_INPUT = "INVALID_INPUT",
  SAVE_FAILED = "SAVE_FAILED",
  LOAD_FAILED = "LOAD_FAILED",
}

/**
 * Enhanced form builder error handling with proper error codes and context
 */
export class FormBuilderError extends Error {
  constructor(
    message: string,
    public readonly code: ErrorCode,
    public readonly userMessage: string,
    public readonly context?: Record<string, unknown>,
    public readonly originalError?: Error,
  ) {
    super(message);
    this.name = "FormBuilderError";

    // Maintain proper stack trace
    if (Error.captureStackTrace) {
      Error.captureStackTrace(this, FormBuilderError);
    }
  }

  /**
   * Create a validation error
   */
  static validationError(
    message: string,
    field?: string,
    value?: unknown,
  ): FormBuilderError {
    return new FormBuilderError(
      `Validation failed: ${message}`,
      ErrorCode.VALIDATION_FAILED,
      message,
      { field, value },
    );
  }

  /**
   * Create a network error
   */
  static networkError(
    message: string,
    url?: string,
    status?: number,
  ): FormBuilderError {
    return new FormBuilderError(
      `Network error: ${message}`,
      ErrorCode.NETWORK_ERROR,
      "Network connection failed. Please check your internet connection and try again.",
      { url, status },
    );
  }

  /**
   * Create a schema error
   */
  static schemaError(message: string, schemaId?: string): FormBuilderError {
    return new FormBuilderError(
      `Schema error: ${message}`,
      ErrorCode.SCHEMA_ERROR,
      "Form configuration error. Please contact support.",
      { schemaId },
    );
  }

  /**
   * Create a CSRF error
   */
  static csrfError(message: string): FormBuilderError {
    return new FormBuilderError(
      `CSRF error: ${message}`,
      ErrorCode.CSRF_ERROR,
      "Security token expired. Please refresh the page and try again.",
    );
  }

  /**
   * Create a sanitization error
   */
  static sanitizationError(
    message: string,
    field?: string,
    value?: unknown,
  ): FormBuilderError {
    return new FormBuilderError(
      `Sanitization error: ${message}`,
      ErrorCode.SANITIZATION_ERROR,
      "Invalid input detected. Please check your input and try again.",
      { field, value },
    );
  }

  /**
   * Create a load failed error
   */
  static loadFailed(
    message: string,
    resourceId?: string,
    originalError?: Error,
  ): FormBuilderError {
    return new FormBuilderError(
      `Load failed: ${message}`,
      ErrorCode.LOAD_FAILED,
      "Failed to load resource. Please try again.",
      { resourceId },
      originalError,
    );
  }

  /**
   * Create a save failed error
   */
  static saveFailed(
    message: string,
    resourceId?: string,
    originalError?: Error,
  ): FormBuilderError {
    return new FormBuilderError(
      `Save failed: ${message}`,
      ErrorCode.SAVE_FAILED,
      "Failed to save changes. Please try again.",
      { resourceId },
      originalError,
    );
  }

  /**
   * Convert error to JSON for logging
   */
  toJSON(): Record<string, unknown> {
    return {
      name: this.name,
      message: this.message,
      code: this.code,
      userMessage: this.userMessage,
      context: this.context,
      stack: this.stack,
      originalError: this.originalError?.message,
    };
  }

  /**
   * Check if error is of specific type
   */
  isCode(code: ErrorCode): boolean {
    return this.code === code;
  }

  /**
   * Get context value safely
   */
  getContextValue<T>(key: string): T | undefined {
    return this.context?.[key] as T | undefined;
  }
}
