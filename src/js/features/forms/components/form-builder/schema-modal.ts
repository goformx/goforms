import { dom } from "@/shared/utils/dom-utils";
import { Logger } from "@/core/logger";

// Configuration constants
const SCHEMA_MODAL_CONFIG = {
  MODAL_ID: "schema-modal",
  Z_INDEX: 10000,
  ANIMATION_DURATION: 300,
  MAX_SCHEMA_SIZE: 1024 * 1024, // 1MB limit
  DEBOUNCE_DELAY: 300,
} as const;

// Types for better structure
interface SchemaModalOptions {
  title?: string;
  metadata?: Record<string, any>;
  copyable?: boolean;
  downloadable?: boolean;
  searchable?: boolean;
  readonly?: boolean;
  onClose?: () => void;
  onCopy?: (schema: string) => void;
  onDownload?: (schema: string) => void;
}

interface ModalElements {
  modal: HTMLDivElement;
  content: HTMLDivElement;
  header: HTMLDivElement;
  body: HTMLDivElement;
  schemaContainer: HTMLPreElement;
  toolbar?: HTMLDivElement;
  searchInput?: HTMLInputElement;
}

/**
 * Enhanced Schema Modal with better features and accessibility
 */
export class SchemaModal {
  private elements: ModalElements | null = null;
  private originalFocus: HTMLElement | null = null;
  private debounceTimer: NodeJS.Timeout | null = null;
  private mutationObserver: MutationObserver | null = null;

  constructor(
    private readonly schemaString: string,
    private readonly options: SchemaModalOptions = {},
  ) {
    this.validateSchema();
  }

  /**
   * Validate schema before showing
   */
  private validateSchema(): void {
    if (!this.schemaString || typeof this.schemaString !== "string") {
      throw new Error("Schema string is required and must be a string");
    }

    if (this.schemaString.length > SCHEMA_MODAL_CONFIG.MAX_SCHEMA_SIZE) {
      Logger.warn("Schema is very large, performance may be affected");
    }

    // Validate JSON format
    try {
      JSON.parse(this.schemaString);
    } catch (error) {
      Logger.warn("Schema is not valid JSON:", error);
    }
  }

  /**
   * Show the modal
   */
  public show(): void {
    try {
      this.removeExistingModal();
      this.saveCurrentFocus();
      this.createModal();
      this.setupEventListeners();
      this.setupAccessibility();
      this.setupCleanupObserver();
      this.animateIn();

      Logger.debug("Schema modal opened");
    } catch (error) {
      Logger.error("Failed to show schema modal:", error);
      this.cleanup();
      throw error;
    }
  }

  /**
   * Remove any existing modal
   */
  private removeExistingModal(): void {
    const existing = document.getElementById(SCHEMA_MODAL_CONFIG.MODAL_ID);
    if (existing) {
      existing.remove();
    }
  }

  /**
   * Save current focus for restoration
   */
  private saveCurrentFocus(): void {
    this.originalFocus = document.activeElement as HTMLElement;
  }

  /**
   * Create the complete modal structure
   */
  private createModal(): void {
    const modal = this.createModalContainer();
    const content = this.createModalContent();
    const header = this.createModalHeader();
    const body = this.createModalBody();
    const schemaContainer = this.createSchemaContent();
    const toolbar = this.createToolbar();

    // Assemble structure
    body.appendChild(schemaContainer);
    content.appendChild(header);
    content.appendChild(body);
    if (toolbar) {
      content.insertBefore(toolbar, body);
    }
    modal.appendChild(content);

    // Store elements
    this.elements = {
      modal,
      content,
      header,
      body,
      schemaContainer,
      ...(toolbar && { toolbar }),
      ...(toolbar?.querySelector("input") && {
        searchInput: toolbar.querySelector("input") as HTMLInputElement,
      }),
    };

    // Add to document
    document.body.appendChild(modal);
  }

  /**
   * Create modal container with backdrop
   */
  private createModalContainer(): HTMLDivElement {
    const modal = dom.createElement<HTMLDivElement>("div", "schema-modal");
    modal.id = SCHEMA_MODAL_CONFIG.MODAL_ID;
    modal.setAttribute("role", "dialog");
    modal.setAttribute("aria-modal", "true");
    modal.setAttribute("aria-labelledby", "schema-modal-title");

    modal.style.cssText = `
      position: fixed;
      top: 0;
      left: 0;
      width: 100%;
      height: 100%;
      background: rgba(0, 0, 0, 0.6);
      display: flex;
      align-items: center;
      justify-content: center;
      z-index: ${SCHEMA_MODAL_CONFIG.Z_INDEX};
      padding: var(--spacing-4);
      backdrop-filter: blur(4px);
      opacity: 0;
      transition: opacity ${SCHEMA_MODAL_CONFIG.ANIMATION_DURATION}ms ease-out;
    `;

    return modal;
  }

  /**
   * Create modal content container
   */
  private createModalContent(): HTMLDivElement {
    const content = dom.createElement<HTMLDivElement>(
      "div",
      "schema-modal-content",
    );

    content.style.cssText = `
      background: var(--background);
      color: var(--text);
      border-radius: var(--border-radius-lg);
      padding: var(--spacing-6);
      max-width: 90vw;
      max-height: 90vh;
      width: 100%;
      overflow: hidden;
      box-shadow: var(--shadow-md);
      border: var(--card-border);
      display: flex;
      flex-direction: column;
      position: relative;
      transform: scale(0.9) translateY(20px);
      transition: all ${SCHEMA_MODAL_CONFIG.ANIMATION_DURATION}ms ease-out;
    `;

    return content;
  }

  /**
   * Create modal header with title and close button
   */
  private createModalHeader(): HTMLDivElement {
    const header = dom.createElement<HTMLDivElement>(
      "div",
      "schema-modal-header",
    );
    header.style.cssText = `
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: var(--spacing-4);
      padding-bottom: var(--spacing-4);
      border-bottom: 1px solid var(--border-color);
      flex-shrink: 0;
    `;

    // Title with metadata
    const titleContainer = this.createTitleContainer();
    const closeButton = this.createCloseButton();

    header.appendChild(titleContainer);
    header.appendChild(closeButton);

    return header;
  }

  /**
   * Create title container with metadata
   */
  private createTitleContainer(): HTMLDivElement {
    const container = dom.createElement<HTMLDivElement>("div");

    const title = dom.createElement<HTMLHeadingElement>("h3");
    title.id = "schema-modal-title";
    title.textContent = this.options.title ?? "Form Schema";
    title.style.cssText = `
      margin: 0 0 var(--spacing-2) 0;
      font-size: var(--font-size-xl);
      font-weight: var(--font-weight-semibold);
      color: var(--text);
      line-height: var(--line-height-tight);
    `;

    container.appendChild(title);

    // Add metadata if provided
    if (this.options.metadata) {
      const metadata = this.createMetadataDisplay();
      container.appendChild(metadata);
    }

    return container;
  }

  /**
   * Create metadata display
   */
  private createMetadataDisplay(): HTMLDivElement {
    const metadata = dom.createElement<HTMLDivElement>(
      "div",
      "schema-metadata",
    );
    metadata.style.cssText = `
      font-size: var(--font-size-sm);
      color: var(--text-light);
      display: flex;
      gap: var(--spacing-4);
      flex-wrap: wrap;
    `;

    const meta = this.options.metadata!;

    if (meta["componentCount"] !== undefined) {
      const count = dom.createElement<HTMLSpanElement>("span");
      count.textContent = `${meta["componentCount"]} components`;
      metadata.appendChild(count);
    }

    if (meta["componentTypes"]?.length) {
      const types = dom.createElement<HTMLSpanElement>("span");
      types.textContent = `Types: ${meta["componentTypes"].join(", ")}`;
      metadata.appendChild(types);
    }

    if (meta["schemaSize"] !== undefined) {
      const size = dom.createElement<HTMLSpanElement>("span");
      const sizeKB = Math.round((meta["schemaSize"] / 1024) * 100) / 100;
      size.textContent = `Size: ${sizeKB}KB`;
      metadata.appendChild(size);
    }

    return metadata;
  }

  /**
   * Create toolbar with actions
   */
  private createToolbar(): HTMLDivElement | null {
    if (
      !this.options.copyable &&
      !this.options.downloadable &&
      !this.options.searchable
    ) {
      return null;
    }

    const toolbar = dom.createElement<HTMLDivElement>("div", "schema-toolbar");
    toolbar.style.cssText = `
      display: flex;
      gap: var(--spacing-2);
      margin-bottom: var(--spacing-4);
      padding: var(--spacing-3);
      background: var(--background-alt);
      border-radius: var(--border-radius);
      border: 1px solid var(--border-color);
    `;

    // Search input
    if (this.options.searchable) {
      const searchInput = this.createSearchInput();
      toolbar.appendChild(searchInput);
    }

    // Action buttons container
    const actions = dom.createElement<HTMLDivElement>("div");
    actions.style.cssText = `
      display: flex;
      gap: var(--spacing-2);
      margin-left: auto;
    `;

    if (this.options.copyable) {
      const copyBtn = this.createActionButton("Copy", "📋", () =>
        this.copySchema(),
      );
      actions.appendChild(copyBtn);
    }

    if (this.options.downloadable) {
      const downloadBtn = this.createActionButton("Download", "💾", () =>
        this.downloadSchema(),
      );
      actions.appendChild(downloadBtn);
    }

    toolbar.appendChild(actions);
    return toolbar;
  }

  /**
   * Create search input
   */
  private createSearchInput(): HTMLInputElement {
    const input = dom.createElement<HTMLInputElement>("input");
    input.type = "text";
    input.placeholder = "Search schema...";
    input.setAttribute("aria-label", "Search schema content");

    input.style.cssText = `
      flex: 1;
      padding: var(--spacing-2) var(--spacing-3);
      border: 1px solid var(--border-color);
      border-radius: var(--border-radius);
      background: var(--background);
      color: var(--text);
      font-size: var(--font-size-sm);
      outline: none;
      transition: border-color var(--form-transition-duration) var(--form-transition-timing);
    `;

    input.addEventListener("input", (e) => {
      const target = e.target as HTMLInputElement;
      this.debounceSearch(target.value);
    });

    input.addEventListener("focus", () => {
      input.style.borderColor = "var(--primary)";
    });

    input.addEventListener("blur", () => {
      input.style.borderColor = "var(--border-color)";
    });

    return input;
  }

  /**
   * Create action button
   */
  private createActionButton(
    text: string,
    icon: string,
    onClick: () => void,
  ): HTMLButtonElement {
    const button = dom.createElement<HTMLButtonElement>("button");
    button.innerHTML = `${icon} ${text}`;
    button.setAttribute("aria-label", text);

    button.style.cssText = `
      padding: var(--spacing-2) var(--spacing-3);
      border: 1px solid var(--border-color);
      border-radius: var(--border-radius);
      background: var(--background);
      color: var(--text);
      font-size: var(--font-size-sm);
      cursor: pointer;
      transition: all var(--form-transition-duration) var(--form-transition-timing);
      white-space: nowrap;
    `;

    button.addEventListener("click", onClick);
    button.addEventListener("mouseenter", () => {
      button.style.background = "var(--primary)";
      button.style.color = "var(--background)";
      button.style.borderColor = "var(--primary)";
    });
    button.addEventListener("mouseleave", () => {
      button.style.background = "var(--background)";
      button.style.color = "var(--text)";
      button.style.borderColor = "var(--border-color)";
    });

    return button;
  }

  /**
   * Create modal body container
   */
  private createModalBody(): HTMLDivElement {
    const body = dom.createElement<HTMLDivElement>("div", "schema-modal-body");
    body.style.cssText = `
      flex: 1;
      min-height: 0;
      overflow: hidden;
      display: flex;
      flex-direction: column;
    `;

    return body;
  }

  /**
   * Create close button
   */
  private createCloseButton(): HTMLButtonElement {
    const closeBtn = dom.createElement<HTMLButtonElement>("button");
    closeBtn.innerHTML = "×";
    closeBtn.setAttribute("aria-label", "Close modal");

    closeBtn.style.cssText = `
      background: none;
      border: none;
      font-size: var(--font-size-2xl);
      cursor: pointer;
      padding: var(--spacing-2);
      width: 40px;
      height: 40px;
      display: flex;
      align-items: center;
      justify-content: center;
      border-radius: var(--border-radius);
      color: var(--text-light);
      transition: all var(--form-transition-duration) var(--form-transition-timing);
      background: var(--background-alt);
      border: 1px solid var(--border-color);
    `;

    closeBtn.addEventListener("click", () => this.close());
    closeBtn.addEventListener("mouseenter", () => {
      closeBtn.style.color = "var(--primary)";
      closeBtn.style.background = "var(--background)";
      closeBtn.style.borderColor = "var(--primary)";
    });
    closeBtn.addEventListener("mouseleave", () => {
      closeBtn.style.color = "var(--text-light)";
      closeBtn.style.background = "var(--background-alt)";
      closeBtn.style.borderColor = "var(--border-color)";
    });

    return closeBtn;
  }

  /**
   * Create schema content display
   */
  private createSchemaContent(): HTMLPreElement {
    const pre = dom.createElement<HTMLPreElement>("pre");
    pre.textContent = this.schemaString;
    pre.setAttribute("aria-label", "Schema content");

    pre.style.cssText = `
      background: var(--background-alt);
      color: var(--text);
      border: 1px solid var(--border-color);
      border-radius: var(--border-radius);
      padding: var(--spacing-4);
      margin: 0;
      font-family: var(--font-mono);
      font-size: var(--font-size-sm);
      line-height: var(--line-height-normal);
      overflow: auto;
      white-space: pre-wrap;
      word-wrap: break-word;
      flex: 1;
      min-height: 0;
      tab-size: 2;
    `;

    return pre;
  }

  /**
   * Set up event listeners
   */
  private setupEventListeners(): void {
    if (!this.elements) return;

    // Close on backdrop click
    this.elements.modal.addEventListener("click", (e) => {
      if (e.target === this.elements!.modal) {
        this.close();
      }
    });

    // Keyboard events
    document.addEventListener("keydown", this.handleKeydown);

    // Prevent body scroll
    document.body.style.overflow = "hidden";

    // Clean up on removal
    this.setupCleanupObserver();
  }

  /**
   * Handle keyboard events
   */
  private readonly handleKeydown = (e: KeyboardEvent): void => {
    switch (e.key) {
      case "Escape":
        this.close();
        break;
      case "Tab":
        this.handleTabKey(e);
        break;
      default:
        // Handle other shortcuts if needed
        break;
    }
  };

  /**
   * Handle tab key for focus trap
   */
  private handleTabKey(e: KeyboardEvent): void {
    if (!this.elements) return;

    const focusableElements = this.elements.content.querySelectorAll(
      "button, [href], input, select, textarea, [tabindex]:not([tabindex='-1'])",
    );

    if (focusableElements.length === 0) return;

    const firstElement = focusableElements[0] as HTMLElement;
    const lastElement = focusableElements[
      focusableElements.length - 1
    ] as HTMLElement;

    if (e.shiftKey) {
      if (document.activeElement === firstElement) {
        e.preventDefault();
        lastElement.focus();
      }
    } else {
      if (document.activeElement === lastElement) {
        e.preventDefault();
        firstElement.focus();
      }
    }
  }

  /**
   * Set up accessibility features
   */
  private setupAccessibility(): void {
    if (!this.elements) return;

    // Focus the first focusable element
    const focusableElements = this.elements.content.querySelectorAll(
      "button, [href], input, select, textarea, [tabindex]:not([tabindex='-1'])",
    );

    if (focusableElements.length > 0) {
      (focusableElements[0] as HTMLElement).focus();
    }

    // Announce modal opening to screen readers
    const announcement = dom.createElement<HTMLDivElement>("div");
    announcement.setAttribute("aria-live", "polite");
    announcement.setAttribute("aria-atomic", "true");
    announcement.style.cssText = `
      position: absolute;
      left: -10000px;
      width: 1px;
      height: 1px;
      overflow: hidden;
    `;
    announcement.textContent = "Schema modal opened";
    this.elements.content.appendChild(announcement);
  }

  /**
   * Set up cleanup observer
   */
  private setupCleanupObserver(): void {
    this.mutationObserver = new MutationObserver(() => {
      if (!document.getElementById(SCHEMA_MODAL_CONFIG.MODAL_ID)) {
        this.cleanup();
      }
    });

    this.mutationObserver.observe(document.body, {
      childList: true,
      subtree: true,
    });
  }

  /**
   * Animate modal in
   */
  private animateIn(): void {
    if (!this.elements) return;

    // Animate in
    this.elements.modal.style.opacity = "1";
    this.elements.content.style.transform = "scale(1) translateY(0)";
  }

  /**
   * Debounced search function
   */
  private debounceSearch(term: string): void {
    if (this.debounceTimer) {
      clearTimeout(this.debounceTimer);
    }

    this.debounceTimer = setTimeout(() => {
      this.performSearch(term);
    }, SCHEMA_MODAL_CONFIG.DEBOUNCE_DELAY);
  }

  /**
   * Perform search highlighting
   */
  private performSearch(term: string): void {
    if (!this.elements) return;

    if (!term.trim()) {
      // Reset to original content
      this.elements.schemaContainer.innerHTML = "";
      this.elements.schemaContainer.textContent = this.schemaString;
      return;
    }

    // Simple highlighting (could be enhanced with regex for better matching)
    const highlighted = this.schemaString.replace(
      new RegExp(this.escapeRegExp(term), "gi"),
      (match) =>
        `<mark style="background: var(--primary); color: var(--background); padding: 1px 2px; border-radius: 2px;">${match}</mark>`,
    );

    this.elements.schemaContainer.innerHTML = highlighted;
  }

  /**
   * Escape string for regex
   */
  private escapeRegExp(string: string): string {
    return string.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
  }

  /**
   * Copy schema to clipboard
   */
  private async copySchema(): Promise<void> {
    try {
      await navigator.clipboard.writeText(this.schemaString);
      Logger.debug("Schema copied to clipboard");
      this.options.onCopy?.(this.schemaString);

      // Show temporary feedback
      this.showCopyFeedback();
    } catch (error) {
      Logger.error("Failed to copy schema:", error);
      // Fallback for older browsers
      this.fallbackCopy();
    }
  }

  /**
   * Show copy feedback
   */
  private showCopyFeedback(): void {
    if (!this.elements) return;

    const feedback = dom.createElement<HTMLDivElement>("div");
    feedback.textContent = "Copied to clipboard!";
    feedback.style.cssText = `
      position: absolute;
      top: var(--spacing-4);
      right: var(--spacing-4);
      background: var(--primary);
      color: var(--background);
      padding: var(--spacing-2) var(--spacing-3);
      border-radius: var(--border-radius);
      font-size: var(--font-size-sm);
      z-index: 1;
      animation: fadeInOut 2s ease-in-out forwards;
    `;

    this.elements.content.appendChild(feedback);

    setTimeout(() => {
      if (feedback.parentNode) {
        feedback.parentNode.removeChild(feedback);
      }
    }, 2000);
  }

  /**
   * Fallback copy method
   */
  private fallbackCopy(): void {
    const textArea = document.createElement("textarea");
    textArea.value = this.schemaString;
    textArea.style.position = "fixed";
    textArea.style.left = "-999999px";
    textArea.style.top = "-999999px";
    document.body.appendChild(textArea);
    textArea.focus();
    textArea.select();

    try {
      document.execCommand("copy");
      this.showCopyFeedback();
    } catch (error) {
      Logger.error("Fallback copy failed:", error);
    }

    document.body.removeChild(textArea);
  }

  /**
   * Download schema as JSON file
   */
  private downloadSchema(): void {
    try {
      const blob = new Blob([this.schemaString], { type: "application/json" });
      const url = URL.createObjectURL(blob);
      const link = document.createElement("a");

      link.href = url;
      link.download = `form-schema-${new Date().toISOString().split("T")[0]}.json`;
      link.style.display = "none";

      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);

      URL.revokeObjectURL(url);

      Logger.debug("Schema downloaded");
      this.options.onDownload?.(this.schemaString);
    } catch (error) {
      Logger.error("Failed to download schema:", error);
    }
  }

  /**
   * Close modal with animation
   */
  public close(): void {
    if (!this.elements) return;

    // Animate out
    this.elements.modal.style.opacity = "0";
    this.elements.content.style.transform = "scale(0.9) translateY(20px)";

    // Remove after animation
    setTimeout(() => {
      if (this.elements?.modal.parentNode) {
        this.elements.modal.parentNode.removeChild(this.elements.modal);
      }
      this.cleanup();
    }, SCHEMA_MODAL_CONFIG.ANIMATION_DURATION);

    this.options.onClose?.();
  }

  /**
   * Clean up resources
   */
  private cleanup(): void {
    // Restore focus
    if (this.originalFocus) {
      this.originalFocus.focus();
    }

    // Restore body scroll
    document.body.style.overflow = "";

    // Remove event listeners
    document.removeEventListener("keydown", this.handleKeydown);

    // Clean up timers
    if (this.debounceTimer) {
      clearTimeout(this.debounceTimer);
    }

    // Disconnect observer
    if (this.mutationObserver) {
      this.mutationObserver.disconnect();
    }

    this.elements = null;
    Logger.debug("Schema modal cleanup completed");
  }
}

/**
 * Enhanced show schema modal function with options support
 */
export function showSchemaModal(
  schemaString: string,
  metadata?: Record<string, any>,
): void {
  const modal = new SchemaModal(schemaString, {
    title: "Form Schema",
    ...(metadata && { metadata }),
    copyable: true,
    downloadable: true,
    searchable: true,
  });
  modal.show();
}

/**
 * Show schema modal with custom options
 */
export function showSchemaModalWithOptions(
  schemaString: string,
  options: SchemaModalOptions,
): SchemaModal {
  const modal = new SchemaModal(schemaString, options);
  modal.show();
  return modal;
}

// Add CSS animation keyframes to document if not already present
if (
  typeof document !== "undefined" &&
  !document.getElementById("schema-modal-animations")
) {
  const style = document.createElement("style");
  style.id = "schema-modal-animations";
  style.textContent = `
    @keyframes fadeInOut {
      0% { opacity: 0; transform: translateY(-10px); }
      20% { opacity: 1; transform: translateY(0); }
      80% { opacity: 1; transform: translateY(0); }
      100% { opacity: 0; transform: translateY(-10px); }
    }
  `;
  document.head.appendChild(style);
}
