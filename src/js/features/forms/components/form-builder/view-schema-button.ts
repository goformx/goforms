import { Logger } from "@/core/logger";
import { FormBuilderError, ErrorCode } from "@/core/errors/form-builder-error";
import { showSchemaModal } from "@/features/forms/components/form-builder/schema-modal";

// Configuration constants
const VIEW_SCHEMA_CONFIG = {
  BUTTON_ID: "view-schema-btn",
  DEBOUNCE_DELAY: 300,
  DEFAULT_INDENT: 2,
} as const;

// Types for better type safety
interface SchemaViewerElements {
  viewButton: HTMLButtonElement;
}

interface SchemaViewerOptions {
  indent?: number;
  onSchemaView?: (schema: any) => void;
  onError?: (error: Error) => void;
  validateSchema?: boolean;
}

/**
 * Enhanced Schema Viewer handler with better architecture
 */
export class SchemaViewerHandler {
  private readonly elements: SchemaViewerElements;
  private debounceTimer: NodeJS.Timeout | null = null;
  private readonly options: Required<SchemaViewerOptions>;

  constructor(
    private builder: any,
    options: SchemaViewerOptions = {},
  ) {
    this.options = {
      indent: VIEW_SCHEMA_CONFIG.DEFAULT_INDENT,
      onSchemaView: () => {},
      onError: (error: Error) => {
        Logger.error("Schema viewer error:", error);
      },
      validateSchema: true,
      ...options,
    };

    this.validateBuilder();
    this.elements = this.getRequiredElements();
    this.setupEventListeners();
    Logger.debug("SchemaViewerHandler initialized");
  }

  /**
   * Validate builder instance
   */
  private validateBuilder(): void {
    if (!this.builder) {
      throw new FormBuilderError(
        "Builder instance is required",
        ErrorCode.FORM_NOT_FOUND,
        "Form builder is not available.",
      );
    }

    if (typeof this.builder.form === "undefined") {
      throw new FormBuilderError(
        "Builder does not have form property",
        ErrorCode.SCHEMA_ERROR,
        "Form builder is not properly initialized.",
      );
    }
  }

  /**
   * Get and validate required DOM elements
   */
  private getRequiredElements(): SchemaViewerElements {
    const viewButton = document.getElementById(
      VIEW_SCHEMA_CONFIG.BUTTON_ID,
    ) as HTMLButtonElement;

    if (!viewButton) {
      throw new FormBuilderError(
        "View Schema button not found",
        ErrorCode.FORM_NOT_FOUND,
        "Required form elements are missing. Please refresh the page.",
      );
    }

    return { viewButton };
  }

  /**
   * Set up event listeners with proper cleanup
   */
  private setupEventListeners(): void {
    this.elements.viewButton.addEventListener("click", this.handleViewClick);

    // Optional: Add keyboard shortcut (Ctrl/Cmd + Shift + S)
    document.addEventListener("keydown", this.handleKeyboardShortcut);

    // Cleanup on page unload
    window.addEventListener("beforeunload", this.cleanup);
  }

  /**
   * Handle view button click with debouncing
   */
  private readonly handleViewClick = (event: Event): void => {
    event.preventDefault();

    // Debounce rapid clicks
    if (this.debounceTimer) {
      clearTimeout(this.debounceTimer);
    }

    this.debounceTimer = setTimeout(() => {
      this.viewSchema();
    }, VIEW_SCHEMA_CONFIG.DEBOUNCE_DELAY);
  };

  /**
   * Handle keyboard shortcut (Ctrl/Cmd + Shift + S)
   */
  private readonly handleKeyboardShortcut = (event: KeyboardEvent): void => {
    if (
      (event.ctrlKey || event.metaKey) &&
      event.shiftKey &&
      event.key === "S"
    ) {
      event.preventDefault();
      this.viewSchema();
    }
  };

  /**
   * Main schema viewing operation
   */
  private viewSchema(): void {
    try {
      this.addLoadingState();

      const schema = this.extractSchema();
      const validatedSchema = this.options.validateSchema
        ? this.validateSchema(schema)
        : schema;

      const formattedSchema = this.formatSchema(validatedSchema);

      this.showSchema(formattedSchema);
      this.options.onSchemaView(validatedSchema);

      Logger.debug("Schema modal opened successfully");
    } catch (error) {
      this.handleViewError(error);
    } finally {
      this.removeLoadingState();
    }
  }

  /**
   * Extract schema from builder with validation
   */
  private extractSchema(): any {
    const schema = this.builder.form;

    if (!schema) {
      throw new FormBuilderError(
        "No schema available",
        ErrorCode.SCHEMA_ERROR,
        "Form schema is empty. Please add components to view the schema.",
      );
    }

    return schema;
  }

  /**
   * Validate schema structure and content
   */
  private validateSchema(schema: any): any {
    // Basic structure validation
    if (typeof schema !== "object") {
      throw new FormBuilderError(
        "Invalid schema format",
        ErrorCode.SCHEMA_ERROR,
        "Form schema has an invalid format.",
      );
    }

    // Check for required properties
    const requiredProps = ["components"];
    const missingProps = requiredProps.filter((prop) => !(prop in schema));

    if (missingProps.length > 0) {
      Logger.warn("Schema missing properties:", missingProps);
    }

    // Validate components array
    if (schema.components && !Array.isArray(schema.components)) {
      throw new FormBuilderError(
        "Invalid components format",
        ErrorCode.SCHEMA_ERROR,
        "Form components have an invalid format.",
      );
    }

    // Check for circular references
    try {
      JSON.stringify(schema);
    } catch (_error) {
      throw new FormBuilderError(
        "Schema contains circular references",
        ErrorCode.SCHEMA_ERROR,
        "Form schema contains invalid references and cannot be displayed.",
      );
    }

    return schema;
  }

  /**
   * Format schema as JSON string with pretty printing
   */
  private formatSchema(schema: any): string {
    try {
      return JSON.stringify(
        schema,
        this.createJSONReplacer(),
        this.options.indent,
      );
    } catch (error) {
      Logger.error("Error formatting schema:", error);
      throw new FormBuilderError(
        "Failed to format schema",
        ErrorCode.SCHEMA_ERROR,
        "Schema cannot be formatted for display.",
      );
    }
  }

  /**
   * Create JSON replacer function for better formatting
   */
  private createJSONReplacer(): (key: string, value: any) => any {
    return (key: string, value: any) => {
      // Remove internal Form.io properties that clutter the view
      if (key.startsWith("_") || (key === "id" && typeof value === "string")) {
        return undefined;
      }

      // Format dates nicely
      if (value instanceof Date) {
        return value.toISOString();
      }

      // Handle functions (shouldn't be in schema, but just in case)
      if (typeof value === "function") {
        return "[Function]";
      }

      return value;
    };
  }

  /**
   * Show schema in modal with enhanced information
   */
  private showSchema(formattedSchema: string): void {
    try {
      // Note: showSchemaModal only accepts one parameter, so we'll just show the schema
      showSchemaModal(formattedSchema);
    } catch (error) {
      Logger.error("Error showing schema modal:", error);
      // Fallback to basic modal
      showSchemaModal(formattedSchema);
    }
  }

  /**
   * Add loading state to button
   */
  private addLoadingState(): void {
    this.elements.viewButton.disabled = true;
    this.elements.viewButton.classList.add("loading");

    // Store original text and show loading
    const originalText = this.elements.viewButton.textContent;
    this.elements.viewButton.dataset.originalText = originalText || "";
    this.elements.viewButton.textContent = "Loading...";
  }

  /**
   * Remove loading state from button
   */
  private removeLoadingState(): void {
    this.elements.viewButton.disabled = false;
    this.elements.viewButton.classList.remove("loading");

    // Restore original text
    const originalText = this.elements.viewButton.dataset.originalText;
    if (originalText) {
      this.elements.viewButton.textContent = originalText;
      delete this.elements.viewButton.dataset.originalText;
    }
  }

  /**
   * Handle view operation errors
   */
  private handleViewError(error: unknown): void {
    Logger.group("View Schema Error");

    if (error instanceof FormBuilderError) {
      Logger.error("FormBuilderError:", {
        code: error.code,
        message: error.message,
        userMessage: error.userMessage,
        context: error.context,
      });
    } else if (error instanceof Error) {
      Logger.error("Standard Error:", {
        name: error.name,
        message: error.message,
        stack: error.stack,
      });
    } else {
      Logger.error("Unknown error:", { type: typeof error, error });
    }

    Logger.groupEnd();

    // Call error callback
    this.options.onError(
      error instanceof Error
        ? error
        : new Error("Unknown error occurred while viewing schema"),
    );
  }

  /**
   * Manual schema view trigger for external use
   */
  public view(): void {
    this.viewSchema();
  }

  /**
   * Get current schema without showing modal
   */
  public getSchema(): any {
    return this.extractSchema();
  }

  /**
   * Get formatted schema string
   */
  public getFormattedSchema(): string {
    const schema = this.extractSchema();
    return this.formatSchema(schema);
  }

  /**
   * Update builder instance
   */
  public updateBuilder(newBuilder: any): void {
    this.builder = newBuilder;
    this.validateBuilder();
    Logger.debug("Builder instance updated");
  }

  /**
   * Cleanup resources and event listeners
   */
  private readonly cleanup = (): void => {
    if (this.debounceTimer) {
      clearTimeout(this.debounceTimer);
    }

    this.elements.viewButton.removeEventListener("click", this.handleViewClick);
    document.removeEventListener("keydown", this.handleKeyboardShortcut);
    window.removeEventListener("beforeunload", this.cleanup);
  };

  /**
   * Destroy the handler and cleanup resources
   */
  public destroy(): void {
    this.cleanup();
    Logger.debug("SchemaViewerHandler destroyed");
  }
}

/**
 * Factory function for backwards compatibility
 */
export function setupViewSchemaButton(
  builder: any,
  options: SchemaViewerOptions = {},
): SchemaViewerHandler {
  try {
    return new SchemaViewerHandler(builder, options);
  } catch (error) {
    Logger.error("Failed to setup View Schema button:", error);
    throw error;
  }
}

/**
 * Utility function for multiple builders
 */
export function setupMultipleSchemaViewers(
  configs: Array<{ builder: any; options?: SchemaViewerOptions }>,
): SchemaViewerHandler[] {
  return configs.map((config) =>
    setupViewSchemaButton(config.builder, config.options),
  );
}
