export class UserDropdownManager {
  private readonly dropdowns = new Map<Element, HTMLElement>();
  private activeDropdown: HTMLElement | null = null;

  constructor() {
    this.init();
  }

  private init(): void {
    this.setupDropdowns();
    this.setupGlobalListeners();
  }

  private setupDropdowns(): void {
    const userMenuButtons = document.querySelectorAll(".user-menu-button");

    userMenuButtons.forEach((button) => {
      const dropdown = button.nextElementSibling as HTMLElement;

      if (!dropdown?.classList.contains("user-menu-dropdown")) {
        return;
      }

      this.dropdowns.set(button, dropdown);

      button.addEventListener("click", (e) => {
        e.preventDefault();
        this.toggleDropdown(dropdown);
      });
    });
  }

  private setupGlobalListeners(): void {
    // Single global click listener for outside clicks
    document.addEventListener("click", (e) => {
      if (!this.activeDropdown) return;

      const target = e.target as Element;
      const isInsideDropdown = this.activeDropdown.contains(target);
      const isDropdownButton = Array.from(this.dropdowns.keys()).some(
        (button) => button.contains(target),
      );

      if (!isInsideDropdown && !isDropdownButton) {
        this.closeActiveDropdown();
      }
    });

    // Single global keyboard listener
    document.addEventListener("keydown", (e) => {
      if (e.key === "Escape" ?.this.activeDropdown) {
        this.closeActiveDropdown();
      }
    });
  }

  private toggleDropdown(dropdown: HTMLElement): void {
    if (this.activeDropdown === dropdown) {
      this.closeActiveDropdown();
    } else {
      this.closeActiveDropdown();
      this.openDropdown(dropdown);
    }
  }

  private openDropdown(dropdown: HTMLElement): void {
    dropdown.classList.add("open");
    this.activeDropdown = dropdown;
  }

  private closeActiveDropdown(): void {
    if (this.activeDropdown) {
      this.activeDropdown.classList.remove("open");
      this.activeDropdown = null;
    }
  }

  public closeAllDropdowns(): void {
    this.dropdowns.forEach((dropdown) => {
      dropdown.classList.remove("open");
    });
    this.activeDropdown = null;
  }
}
