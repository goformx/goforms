import { dom } from "../../../../shared/utils/dom-utils";

/**
 * Form Builder Sidebar Management
 */
export class FormBuilderSidebar {
  private sidebar: HTMLElement | null = null;
  private toggleButton: HTMLButtonElement | null = null;
  private resizeTimeout: number | null = null;
  private readonly MOBILE_BREAKPOINT = 768;
  private readonly RESIZE_DEBOUNCE_DELAY = 250;

  constructor() {
    this.init();
  }

  /**
   * Initialize the sidebar functionality
   */
  private init(): void {
    // Only initialize on form builder pages
    const formBuilder = document.querySelector(".formbuilder");
    if (!formBuilder) return;

    this.sidebar = document.querySelector(".builder-sidebar");
    if (!this.sidebar) return;

    this.createToggleButton();
    this.setupEventListeners();
  }

  /**
   * Create and append toggle button if it doesn't exist
   */
  private createToggleButton(): void {
    if (document.querySelector(".builder-sidebar-toggle")) return;

    this.toggleButton = dom.createElement<HTMLButtonElement>(
      "button",
      "builder-sidebar-toggle",
    );
    this.toggleButton.innerHTML = '<i class="bi bi-list"></i>';
    this.toggleButton.setAttribute(
      "aria-label",
      "Toggle form components sidebar",
    );

    document.body.appendChild(this.toggleButton);
  }

  /**
   * Set up all event listeners
   */
  private setupEventListeners(): void {
    this.setupToggleListener();
    this.setupOutsideClickListener();
    this.setupResizeListener();
  }

  /**
   * Toggle sidebar on button click
   */
  private setupToggleListener(): void {
    if (!this.toggleButton) return;

    this.toggleButton.addEventListener("click", () => {
      this.toggleSidebar();
    });
  }

  /**
   * Close sidebar when clicking outside on mobile
   */
  private setupOutsideClickListener(): void {
    document.addEventListener("click", (e: Event) => {
      const target = e.target as HTMLElement;

      if (!this.shouldCloseOnOutsideClick(target)) return;

      this.closeSidebar();
    });
  }

  /**
   * Handle window resize
   */
  private setupResizeListener(): void {
    window.addEventListener("resize", () => {
      if (this.resizeTimeout) {
        clearTimeout(this.resizeTimeout);
      }

      this.resizeTimeout = window.setTimeout(() => {
        this.handleResize();
      }, this.RESIZE_DEBOUNCE_DELAY);
    });
  }

  /**
   * Determine if sidebar should close on outside click
   */
  private shouldCloseOnOutsideClick(target: HTMLElement): boolean {
    if (!this.sidebar) return false;

    const isMobile = this.isMobile();
    const isClickOutsideSidebar = !this.sidebar.contains(target);
    const isNotToggleButton = !target.closest(".builder-sidebar-toggle");
    const isSidebarOpen = this.sidebar.classList.contains("is-open");

    return (
      isMobile && isClickOutsideSidebar && isNotToggleButton && isSidebarOpen
    );
  }

  /**
   * Handle window resize events
   */
  private handleResize(): void {
    if (!this.isMobile()) {
      this.closeSidebar();
    }
  }

  /**
   * Check if current viewport is mobile
   */
  private isMobile(): boolean {
    return window.innerWidth < this.MOBILE_BREAKPOINT;
  }

  /**
   * Toggle sidebar open/closed state
   */
  public toggleSidebar(): void {
    if (!this.sidebar) return;

    this.sidebar.classList.toggle("is-open");
  }

  /**
   * Open the sidebar
   */
  public openSidebar(): void {
    if (!this.sidebar) return;

    this.sidebar.classList.add("is-open");
  }

  /**
   * Close the sidebar
   */
  public closeSidebar(): void {
    if (!this.sidebar) return;

    this.sidebar.classList.remove("is-open");
  }

  /**
   * Check if sidebar is currently open
   */
  public isOpen(): boolean {
    return this.sidebar?.classList.contains("is-open") ?? false;
  }

  /**
   * Destroy the sidebar instance and clean up event listeners
   */
  public destroy(): void {
    if (this.resizeTimeout) {
      clearTimeout(this.resizeTimeout);
      this.resizeTimeout = null;
    }

    if (this.toggleButton) {
      this.toggleButton.remove();
      this.toggleButton = null;
    }

    this.sidebar = null;
  }
}

/**
 * Initialize sidebar when DOM is ready
 */
function initializeSidebar(): void {
  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", () => {
      new FormBuilderSidebar();
    });
  } else {
    new FormBuilderSidebar();
  }
}

// Auto-initialize
initializeSidebar();

// Export for manual initialization if needed
export { initializeSidebar };
