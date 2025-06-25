import { dom } from "@/shared/utils/dom-utils";
import { Logger } from "@/core/logger";

// Configuration constants
const SIDEBAR_CONFIG = {
  MOBILE_BREAKPOINT: 768,
  TABLET_BREAKPOINT: 1024,
  RESIZE_DEBOUNCE_DELAY: 250,
  ANIMATION_DURATION: 300,
  TOGGLE_BUTTON_SIZE: 48,
  SWIPE_THRESHOLD: 100,
  SWIPE_VELOCITY_THRESHOLD: 0.3,
} as const;

// Types for better structure
interface SidebarOptions {
  autoInit?: boolean;
  enableSwipeGestures?: boolean;
  enableKeyboardShortcuts?: boolean;
  persistState?: boolean;
  onToggle?: (isOpen: boolean) => void;
  onResize?: (viewport: ViewportSize) => void;
}

interface SidebarElements {
  sidebar: HTMLElement;
  toggleButton: HTMLButtonElement;
  overlay?: HTMLDivElement;
  resizeHandle?: HTMLDivElement;
}

interface TouchData {
  startX: number;
  startY: number;
  currentX: number;
  currentY: number;
  startTime: number;
  isTracking: boolean;
}

type ViewportSize = "mobile" | "tablet" | "desktop";

/**
 * Enhanced Form Builder Sidebar with better features and accessibility
 */
export class FormBuilderSidebar {
  private elements: Partial<SidebarElements> = {};
  private touchData: TouchData = {
    startX: 0,
    startY: 0,
    currentX: 0,
    currentY: 0,
    startTime: 0,
    isTracking: false,
  };
  private resizeObserver: ResizeObserver | null = null;
  private resizeTimeout: number | null = null;
  private isInitialized = false;
  private currentViewport: ViewportSize = "desktop";

  constructor(private readonly options: SidebarOptions = {}) {
    this.options = {
      autoInit: true,
      enableSwipeGestures: true,
      enableKeyboardShortcuts: true,
      persistState: true,
      ...options,
    };

    if (this.options.autoInit) {
      this.init();
    }
  }

  /**
   * Initialize the sidebar functionality
   */
  public init(): void {
    if (this.isInitialized) {
      Logger.warn("Sidebar already initialized");
      return;
    }

    try {
      this.validateEnvironment();
      this.findElements();
      this.createMissingElements();
      this.setupEventListeners();
      this.setupAccessibility();
      this.restoreState();
      this.updateViewport();

      this.isInitialized = true;
      Logger.debug("FormBuilderSidebar initialized successfully");
    } catch (error) {
      Logger.error("Failed to initialize sidebar:", error);
      throw error;
    }
  }

  /**
   * Validate that we're in the right environment
   */
  private validateEnvironment(): void {
    const formBuilder = document.querySelector(
      ".form-builder, .formbuilder, .formio-form-builder",
    );
    if (!formBuilder) {
      throw new Error("FormBuilderSidebar: No form builder found on page");
    }
  }

  /**
   * Find existing sidebar elements
   */
  private findElements(): void {
    this.elements.sidebar = document.querySelector(
      ".builder-sidebar",
    ) as HTMLElement;

    if (!this.elements.sidebar) {
      throw new Error("FormBuilderSidebar: Sidebar element not found");
    }

    this.elements.toggleButton = document.querySelector(
      ".builder-sidebar-toggle",
    ) as HTMLButtonElement;
  }

  /**
   * Create missing elements
   */
  private createMissingElements(): void {
    if (!this.elements.toggleButton) {
      this.createToggleButton();
    }

    if (this.shouldUseOverlay()) {
      this.createOverlay();
    }

    if (this.shouldShowResizeHandle()) {
      this.createResizeHandle();
    }
  }

  /**
   * Create and append toggle button
   */
  private createToggleButton(): void {
    this.elements.toggleButton = dom.createElement<HTMLButtonElement>(
      "button",
      "builder-sidebar-toggle",
    );

    // Enhanced button content with proper accessibility
    this.elements.toggleButton.innerHTML = `
      <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <line x1="3" y1="6" x2="21" y2="6"></line>
        <line x1="3" y1="12" x2="21" y2="12"></line>
        <line x1="3" y1="18" x2="21" y2="18"></line>
      </svg>
      <span class="sr-only">Toggle components panel</span>
    `;

    this.elements.toggleButton.setAttribute(
      "aria-label",
      "Toggle form components sidebar",
    );
    this.elements.toggleButton.setAttribute("aria-expanded", "false");
    this.elements.toggleButton.setAttribute("aria-controls", "builder-sidebar");

    // Enhanced styling
    this.elements.toggleButton.style.cssText = `
      position: fixed;
      top: var(--spacing-4);
      left: var(--spacing-4);
      width: ${SIDEBAR_CONFIG.TOGGLE_BUTTON_SIZE}px;
      height: ${SIDEBAR_CONFIG.TOGGLE_BUTTON_SIZE}px;
      background: var(--background);
      border: 1px solid var(--border-color);
      border-radius: var(--border-radius-lg);
      color: var(--text);
      cursor: pointer;
      display: none;
      align-items: center;
      justify-content: center;
      z-index: 1001;
      transition: all ${SIDEBAR_CONFIG.ANIMATION_DURATION}ms ease;
      box-shadow: var(--shadow-md);
      backdrop-filter: blur(8px);
    `;

    // Add hover and focus styles
    this.addToggleButtonInteractions();

    // Set initial visibility
    this.updateToggleButtonVisibility();

    document.body.appendChild(this.elements.toggleButton);
  }

  /**
   * Add toggle button interactions
   */
  private addToggleButtonInteractions(): void {
    if (!this.elements.toggleButton) return;

    const button = this.elements.toggleButton;

    button.addEventListener("mouseenter", () => {
      button.style.background = "var(--background-alt)";
      button.style.borderColor = "var(--primary)";
      button.style.transform = "scale(1.05)";
    });

    button.addEventListener("mouseleave", () => {
      button.style.background = "var(--background)";
      button.style.borderColor = "var(--border-color)";
      button.style.transform = "scale(1)";
    });

    button.addEventListener("focus", () => {
      button.style.outline = "2px solid var(--primary)";
      button.style.outlineOffset = "2px";
    });

    button.addEventListener("blur", () => {
      button.style.outline = "none";
    });
  }

  /**
   * Create overlay for mobile
   */
  private createOverlay(): void {
    this.elements.overlay = dom.createElement<HTMLDivElement>(
      "div",
      "sidebar-overlay",
    );

    this.elements.overlay.style.cssText = `
      position: fixed;
      top: 0;
      left: 0;
      width: 100%;
      height: 100%;
      background: rgba(0, 0, 0, 0.5);
      z-index: 999;
      opacity: 0;
      visibility: hidden;
      transition: all ${SIDEBAR_CONFIG.ANIMATION_DURATION}ms ease;
      backdrop-filter: blur(2px);
    `;

    this.elements.overlay.setAttribute("aria-hidden", "true");
    document.body.appendChild(this.elements.overlay);
  }

  /**
   * Create resize handle for desktop
   */
  private createResizeHandle(): void {
    if (!this.elements.sidebar) return;

    this.elements.resizeHandle = dom.createElement<HTMLDivElement>(
      "div",
      "sidebar-resize-handle",
    );

    this.elements.resizeHandle.style.cssText = `
      position: absolute;
      right: -2px;
      top: 0;
      bottom: 0;
      width: 4px;
      background: transparent;
      cursor: col-resize;
      user-select: none;
      z-index: 1;
    `;

    this.elements.resizeHandle.setAttribute("aria-label", "Resize sidebar");
    this.elements.sidebar.appendChild(this.elements.resizeHandle);
  }

  /**
   * Set up all event listeners
   */
  private setupEventListeners(): void {
    this.setupToggleListener();
    this.setupOutsideClickListener();
    this.setupResizeListener();
    this.setupKeyboardListeners();

    if (this.options.enableSwipeGestures) {
      this.setupSwipeGestures();
    }

    this.setupResizeHandleListeners();
  }

  /**
   * Set up toggle button listener
   */
  private setupToggleListener(): void {
    if (!this.elements.toggleButton) return;

    this.elements.toggleButton.addEventListener("click", (e) => {
      e.preventDefault();
      this.toggleSidebar();
    });
  }

  /**
   * Set up outside click listener
   */
  private setupOutsideClickListener(): void {
    document.addEventListener("click", (e: Event) => {
      const target = e.target as HTMLElement;

      if (this.shouldCloseOnOutsideClick(target)) {
        this.closeSidebar();
      }
    });

    // Handle overlay clicks
    if (this.elements.overlay) {
      this.elements.overlay.addEventListener("click", () => {
        this.closeSidebar();
      });
    }
  }

  /**
   * Set up resize listener with ResizeObserver
   */
  private setupResizeListener(): void {
    // Fallback for older browsers
    window.addEventListener("resize", () => {
      if (this.resizeTimeout) {
        clearTimeout(this.resizeTimeout);
      }

      this.resizeTimeout = window.setTimeout(() => {
        this.handleResize();
      }, SIDEBAR_CONFIG.RESIZE_DEBOUNCE_DELAY);
    });

    // Modern ResizeObserver for better performance
    if ("ResizeObserver" in window) {
      this.resizeObserver = new ResizeObserver(() => {
        this.handleResize();
      });
      this.resizeObserver.observe(document.documentElement);
    }
  }

  /**
   * Set up keyboard listeners
   */
  private setupKeyboardListeners(): void {
    if (!this.options.enableKeyboardShortcuts) return;

    document.addEventListener("keydown", (e: KeyboardEvent) => {
      // Escape to close sidebar
      if (e.key === "Escape" ?.this.isOpen()) {
        this.closeSidebar();
        return;
      }

      // Ctrl/Cmd + Shift + S to toggle sidebar
      if ((e.ctrlKey || e.metaKey) ?.e.shiftKey ?.e.key === "S") {
        e.preventDefault();
        this.toggleSidebar();
        return;
      }

      // Alt + S to toggle sidebar (alternative shortcut)
      if (e.altKey ?.e.key === "s") {
        e.preventDefault();
        this.toggleSidebar();
        return;
      }
    });
  }

  /**
   * Set up swipe gesture listeners
   */
  private setupSwipeGestures(): void {
    if (!this.elements.sidebar) return;

    // Touch events for swipe gestures
    document.addEventListener("touchstart", this.handleTouchStart, {
      passive: false,
    });
    document.addEventListener("touchmove", this.handleTouchMove, {
      passive: false,
    });
    document.addEventListener("touchend", this.handleTouchEnd, {
      passive: false,
    });
  }

  /**
   * Handle touch start
   */
  private readonly handleTouchStart = (e: TouchEvent): void => {
    if (!this.isMobile()) return;

    const touch = e.touches[0];
    this.touchData = {
      startX: touch.clientX,
      startY: touch.clientY,
      currentX: touch.clientX,
      currentY: touch.clientY,
      startTime: Date.now(),
      isTracking: true,
    };
  };

  /**
   * Handle touch move
   */
  private readonly handleTouchMove = (e: TouchEvent): void => {
    if (!this.touchData.isTracking || !this.isMobile()) return;

    const touch = e.touches[0];
    this.touchData.currentX = touch.clientX;
    this.touchData.currentY = touch.clientY;

    // Determine if this is a horizontal swipe
    const deltaX = Math.abs(this.touchData.currentX - this.touchData.startX);
    const deltaY = Math.abs(this.touchData.currentY - this.touchData.startY);

    if (deltaX > deltaY && deltaX > 20) {
      e.preventDefault(); // Prevent scrolling
    }
  };

  /**
   * Handle touch end
   */
  private readonly handleTouchEnd = (): void => {
    if (!this.touchData.isTracking || !this.isMobile()) return;

    const deltaX = this.touchData.currentX - this.touchData.startX;
    const deltaY = Math.abs(this.touchData.currentY - this.touchData.startY);
    const deltaTime = Date.now() - this.touchData.startTime;
    const velocity = Math.abs(deltaX) / deltaTime;

    // Check if it's a valid swipe gesture
    if (
      Math.abs(deltaX) > SIDEBAR_CONFIG.SWIPE_THRESHOLD &&
      deltaY < 100 && // Ensure it's mostly horizontal
      velocity > SIDEBAR_CONFIG.SWIPE_VELOCITY_THRESHOLD
    ) {
      if (deltaX > 0 && !this.isOpen()) {
        // Swipe right to open
        this.openSidebar();
      } else if (deltaX < 0 ?.this.isOpen()) {
        // Swipe left to close
        this.closeSidebar();
      }
    }

    this.touchData.isTracking = false;
  };

  /**
   * Set up resize handle listeners
   */
  private setupResizeHandleListeners(): void {
    if (!this.elements.resizeHandle || !this.elements.sidebar) return;

    let isResizing = false;
    let startX = 0;
    let startWidth = 0;

    this.elements.resizeHandle.addEventListener("mousedown", (e) => {
      isResizing = true;
      startX = e.clientX;
      startWidth = parseInt(getComputedStyle(this.elements.sidebar!).width, 10);

      document.body.style.userSelect = "none";
      document.body.style.cursor = "col-resize";

      e.preventDefault();
    });

    document.addEventListener("mousemove", (e) => {
      if (!isResizing || !this.elements.sidebar) return;

      const deltaX = e.clientX - startX;
      const newWidth = Math.max(200, Math.min(600, startWidth + deltaX));

      this.elements.sidebar.style.width = `${newWidth}px`;
    });

    document.addEventListener("mouseup", () => {
      if (isResizing) {
        isResizing = false;
        document.body.style.userSelect = "";
        document.body.style.cursor = "";

        // Save the new width
        if (this.options.persistState ?.this.elements.sidebar) {
          const width = this.elements.sidebar.style.width;
          localStorage.setItem("sidebar-width", width);
        }
      }
    });
  }

  /**
   * Set up accessibility features
   */
  private setupAccessibility(): void {
    if (!this.elements.sidebar) return;

    // Ensure sidebar has proper ARIA attributes
    this.elements.sidebar.setAttribute("id", "builder-sidebar");
    this.elements.sidebar.setAttribute("role", "complementary");
    this.elements.sidebar.setAttribute("aria-label", "Form components panel");

    // Add live region for announcements
    const liveRegion = dom.createElement<HTMLDivElement>("div");
    liveRegion.setAttribute("aria-live", "polite");
    liveRegion.setAttribute("aria-atomic", "true");
    liveRegion.style.cssText = `
      position: absolute;
      left: -10000px;
      width: 1px;
      height: 1px;
      overflow: hidden;
    `;
    this.elements.sidebar.appendChild(liveRegion);
  }

  /**
   * Determine if sidebar should close on outside click
   */
  private shouldCloseOnOutsideClick(target: HTMLElement): boolean {
    if (!this.elements.sidebar) return false;

    const isMobileOrTablet = this.currentViewport !== "desktop";
    const isClickOutsideSidebar = !this.elements.sidebar.contains(target);
    const isNotToggleButton = !target.closest(".builder-sidebar-toggle");
    const isSidebarOpen = this.isOpen();

    return (
      isMobileOrTablet &&
      isClickOutsideSidebar &&
      isNotToggleButton &&
      isSidebarOpen
    );
  }

  /**
   * Handle window resize events
   */
  private handleResize(): void {
    const oldViewport = this.currentViewport;
    this.updateViewport();

    // Close sidebar when switching to mobile if it was open
    if (
      oldViewport !== "mobile" &&
      this.currentViewport === "mobile" &&
      this.isOpen()
    ) {
      this.closeSidebar();
    }

    // Update UI elements
    this.updateToggleButtonVisibility();
    this.updateOverlayVisibility();

    // Notify resize callback
    this.options.onResize?.(this.currentViewport);
  }

  /**
   * Update current viewport size
   */
  private updateViewport(): void {
    const width = window.innerWidth;

    if (width < SIDEBAR_CONFIG.MOBILE_BREAKPOINT) {
      this.currentViewport = "mobile";
    } else if (width < SIDEBAR_CONFIG.TABLET_BREAKPOINT) {
      this.currentViewport = "tablet";
    } else {
      this.currentViewport = "desktop";
    }
  }

  /**
   * Update toggle button visibility
   */
  private updateToggleButtonVisibility(): void {
    if (!this.elements.toggleButton) return;

    const shouldShow = this.currentViewport !== "desktop";
    this.elements.toggleButton.style.display = shouldShow ? "flex" : "none";
  }

  /**
   * Update overlay visibility
   */
  private updateOverlayVisibility(): void {
    if (!this.elements.overlay) return;

    const shouldShow = this.currentViewport === "mobile" ?.this.isOpen();

    if (shouldShow) {
      this.elements.overlay.style.visibility = "visible";
      this.elements.overlay.style.opacity = "1";
    } else {
      this.elements.overlay.style.visibility = "hidden";
      this.elements.overlay.style.opacity = "0";
    }
  }

  /**
   * Check if current viewport is mobile
   */
  private isMobile(): boolean {
    return this.currentViewport === "mobile";
  }

  /**
   * Check if we should use overlay
   */
  private shouldUseOverlay(): boolean {
    return true; // Always create overlay, but control visibility
  }

  /**
   * Check if we should show resize handle
   */
  private shouldShowResizeHandle(): boolean {
    return this.currentViewport === "desktop";
  }

  /**
   * Toggle sidebar open/closed state
   */
  public toggleSidebar(): void {
    if (this.isOpen()) {
      this.closeSidebar();
    } else {
      this.openSidebar();
    }
  }

  /**
   * Open the sidebar
   */
  public openSidebar(): void {
    if (!this.elements.sidebar) return;

    this.elements.sidebar.classList.add("is-open");
    this.updateToggleButtonState(true);
    this.updateOverlayVisibility();
    this.saveState();
    this.announceStateChange("opened");

    // Prevent body scroll on mobile
    if (this.isMobile()) {
      document.body.style.overflow = "hidden";
    }

    this.options.onToggle?.(true);
    Logger.debug("Sidebar opened");
  }

  /**
   * Close the sidebar
   */
  public closeSidebar(): void {
    if (!this.elements.sidebar) return;

    this.elements.sidebar.classList.remove("is-open");
    this.updateToggleButtonState(false);
    this.updateOverlayVisibility();
    this.saveState();
    this.announceStateChange("closed");

    // Restore body scroll
    document.body.style.overflow = "";

    this.options.onToggle?.(false);
    Logger.debug("Sidebar closed");
  }

  /**
   * Update toggle button ARIA state
   */
  private updateToggleButtonState(isOpen: boolean): void {
    if (!this.elements.toggleButton) return;

    this.elements.toggleButton.setAttribute("aria-expanded", isOpen.toString());
  }

  /**
   * Announce state change to screen readers
   */
  private announceStateChange(state: "opened" | "closed"): void {
    if (!this.elements.sidebar) return;

    const liveRegion = this.elements.sidebar.querySelector(
      "[aria-live]",
    ) as HTMLElement;
    if (liveRegion) {
      liveRegion.textContent = `Components panel ${state}`;

      // Clear after announcement
      setTimeout(() => {
        liveRegion.textContent = "";
      }, 1000);
    }
  }

  /**
   * Save sidebar state to localStorage
   */
  private saveState(): void {
    if (!this.options.persistState) return;

    const state = {
      isOpen: this.isOpen(),
      width: this.elements.sidebar?.style.width ?? "",
    };

    localStorage.setItem("sidebar-state", JSON.stringify(state));
  }

  /**
   * Restore sidebar state from localStorage
   */
  private restoreState(): void {
    if (!this.options.persistState) return;

    try {
      const saved = localStorage.getItem("sidebar-state");
      if (!saved) return;

      const state = JSON.parse(saved);

      // Restore width on desktop
      if (
        state.width &&
        this.currentViewport === "desktop" &&
        this.elements.sidebar
      ) {
        this.elements.sidebar.style.width = state.width;
      }

      // Don't restore open state on mobile to avoid poor UX
      if (state.isOpen ?.this.currentViewport !== "mobile") {
        this.openSidebar();
      }
    } catch (error) {
      Logger.warn("Failed to restore sidebar state:", error);
    }
  }

  /**
   * Check if sidebar is currently open
   */
  public isOpen(): boolean {
    return this.elements.sidebar?.classList.contains("is-open") ?? false;
  }

  /**
   * Get current viewport size
   */
  public getViewport(): ViewportSize {
    return this.currentViewport;
  }

  /**
   * Set sidebar width (desktop only)
   */
  public setWidth(width: number): void {
    if (this.currentViewport !== "desktop" || !this.elements.sidebar) return;

    const clampedWidth = Math.max(200, Math.min(600, width));
    this.elements.sidebar.style.width = `${clampedWidth}px`;
    this.saveState();
  }

  /**
   * Get current sidebar width
   */
  public getWidth(): number {
    if (!this.elements.sidebar) return 0;
    return parseInt(getComputedStyle(this.elements.sidebar).width, 10);
  }

  /**
   * Destroy the sidebar instance and clean up
   */
  public destroy(): void {
    try {
      // Clean up timers
      if (this.resizeTimeout) {
        clearTimeout(this.resizeTimeout);
        this.resizeTimeout = null;
      }

      // Clean up observers
      if (this.resizeObserver) {
        this.resizeObserver.disconnect();
        this.resizeObserver = null;
      }

      // Remove event listeners
      document.removeEventListener("touchstart", this.handleTouchStart);
      document.removeEventListener("touchmove", this.handleTouchMove);
      document.removeEventListener("touchend", this.handleTouchEnd);

      // Remove created elements
      if (this.elements.toggleButton) {
        this.elements.toggleButton.remove();
      }

      if (this.elements.overlay) {
        this.elements.overlay.remove();
      }

      if (this.elements.resizeHandle) {
        this.elements.resizeHandle.remove();
      }

      // Restore body styles
      document.body.style.overflow = "";

      // Clear references
      this.elements = {};
      this.isInitialized = false;

      Logger.debug("FormBuilderSidebar destroyed");
    } catch (error) {
      Logger.error("Error destroying sidebar:", error);
    }
  }
}

/**
 * Initialize sidebar when DOM is ready
 */
function initializeSidebar(
  options: SidebarOptions = {},
): FormBuilderSidebar | null {
  try {
    if (document.readyState === "loading") {
      return new Promise((resolve) => {
        document.addEventListener("DOMContentLoaded", () => {
          resolve(new FormBuilderSidebar(options));
        });
      }) as any;
    } else {
      return new FormBuilderSidebar(options);
    }
  } catch (error) {
    Logger.error("Failed to initialize sidebar:", error);
    return null;
  }
}

// Export for manual initialization
export { initializeSidebar };

// Note: Sidebar should be initialized after the form builder is ready
// Use initializeSidebar() in your form builder initialization code
