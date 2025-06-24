import { dom } from "@/shared/utils/dom-utils";

/**
 * Show schema in a modal dialog
 */
export function showSchemaModal(schemaString: string): void {
  // Remove existing modal if present
  const existingModal = document.getElementById("schema-modal");
  if (existingModal) {
    existingModal.remove();
  }

  // Create modal container
  const modal = dom.createElement<HTMLDivElement>("div", "schema-modal");
  modal.id = "schema-modal";
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
    z-index: 10000;
    padding: var(--spacing-4);
    backdrop-filter: blur(4px);
  `;

  // Create modal content
  const modalContent = dom.createElement<HTMLDivElement>(
    "div",
    "schema-modal-content",
  );
  modalContent.style.cssText = `
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
  `;

  // Create modal header
  const modalHeader = createModalHeader(() => modal.remove());

  // Create schema content
  const schemaContent = createSchemaContent(schemaString);

  // Assemble modal
  modalContent.appendChild(modalHeader);
  modalContent.appendChild(schemaContent);
  modal.appendChild(modalContent);

  // Add to document
  document.body.appendChild(modal);

  // Set up event listeners
  setupModalEventListeners(modal);

  // Focus trap for accessibility
  setupFocusTrap(modalContent);
}

function createModalHeader(onClose: () => void): HTMLDivElement {
  const modalHeader = dom.createElement<HTMLDivElement>(
    "div",
    "schema-modal-header",
  );
  modalHeader.style.cssText = `
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: var(--spacing-6);
    padding-bottom: var(--spacing-4);
    border-bottom: 1px solid var(--border-color);
    flex-shrink: 0;
  `;

  const modalTitle = dom.createElement<HTMLHeadingElement>("h3");
  modalTitle.textContent = "Form Schema";
  modalTitle.style.cssText = `
    margin: 0;
    font-size: var(--font-size-xl);
    font-weight: var(--font-weight-semibold);
    color: var(--text);
    line-height: var(--line-height-tight);
  `;

  const closeBtn = createCloseButton(onClose);

  modalHeader.appendChild(modalTitle);
  modalHeader.appendChild(closeBtn);

  return modalHeader;
}

function createCloseButton(onClose: () => void): HTMLButtonElement {
  const closeBtn = dom.createElement<HTMLButtonElement>("button");
  closeBtn.textContent = "×";
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

  closeBtn.addEventListener("click", onClose);
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

  // Keyboard support
  closeBtn.addEventListener("keydown", (e) => {
    if (e.key === "Enter" || e.key === " ") {
      e.preventDefault();
      onClose();
    }
  });

  return closeBtn;
}

function createSchemaContent(schemaString: string): HTMLPreElement {
  const schemaContent = dom.createElement<HTMLPreElement>("pre");
  schemaContent.textContent = schemaString;
  schemaContent.style.cssText = `
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
    max-height: 60vh;
  `;

  return schemaContent;
}

function setupModalEventListeners(modal: HTMLDivElement): void {
  // Close modal when clicking outside
  modal.addEventListener("click", (e) => {
    if (e.target === modal) {
      modal.remove();
    }
  });

  // Close modal with Escape key
  const handleEscape = (e: KeyboardEvent) => {
    if (e.key === "Escape") {
      modal.remove();
      document.removeEventListener("keydown", handleEscape);
    }
  };
  document.addEventListener("keydown", handleEscape);

  // Prevent body scroll when modal is open
  document.body.style.overflow = "hidden";

  // Clean up when modal is removed
  const observer = new MutationObserver(() => {
    if (!document.getElementById("schema-modal")) {
      document.body.style.overflow = "";
      observer.disconnect();
      document.removeEventListener("keydown", handleEscape);
    }
  });
  observer.observe(document.body, { childList: true, subtree: true });
}

function setupFocusTrap(modalContent: HTMLDivElement): void {
  const focusableElements = modalContent.querySelectorAll(
    "button, [href], input, select, textarea, [tabindex]:not([tabindex='-1'])",
  );

  if (focusableElements.length === 0) return;

  const firstElement = focusableElements[0] as HTMLElement;
  const lastElement = focusableElements[
    focusableElements.length - 1
  ] as HTMLElement;

  // Focus first element
  firstElement.focus();

  // Handle tab key for focus trap
  const handleTab = (e: KeyboardEvent) => {
    if (e.key === "Tab") {
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
  };

  modalContent.addEventListener("keydown", handleTab);
}
