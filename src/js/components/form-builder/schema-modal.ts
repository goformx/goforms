import { dom } from "../../utils/dom-utils";

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
    background: rgba(0, 0, 0, 0.5);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 10000;
  `;

  // Create modal content
  const modalContent = dom.createElement<HTMLDivElement>(
    "div",
    "schema-modal-content",
  );
  modalContent.style.cssText = `
    background: var(--background-color, #ffffff);
    color: var(--text-color, #000000);
    border-radius: 8px;
    padding: 20px;
    max-width: 90%;
    max-height: 90%;
    overflow: auto;
    box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
    border: 1px solid var(--border-color, #e5e7eb);
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
    margin-bottom: 20px;
    padding-bottom: 10px;
    border-bottom: 1px solid var(--border-color, #e5e7eb);
  `;

  const modalTitle = dom.createElement<HTMLHeadingElement>("h3");
  modalTitle.textContent = "Form Schema";
  modalTitle.style.cssText = `
    margin: 0;
    font-size: 18px;
    font-weight: 600;
    color: var(--text-color, #000000);
  `;

  const closeBtn = createCloseButton(onClose);

  modalHeader.appendChild(modalTitle);
  modalHeader.appendChild(closeBtn);

  return modalHeader;
}

function createCloseButton(onClose: () => void): HTMLButtonElement {
  const closeBtn = dom.createElement<HTMLButtonElement>("button");
  closeBtn.textContent = "×";
  closeBtn.style.cssText = `
    background: none;
    border: none;
    font-size: 24px;
    cursor: pointer;
    padding: 0;
    width: 30px;
    height: 30px;
    display: flex;
    align-items: center;
    justify-content: center;
    border-radius: 4px;
    color: var(--text-color, #000000);
    transition: color 0.2s ease;
  `;

  closeBtn.addEventListener("click", onClose);
  closeBtn.addEventListener("mouseenter", () => {
    closeBtn.style.color = "var(--primary-color, #3b82f6)";
  });
  closeBtn.addEventListener("mouseleave", () => {
    closeBtn.style.color = "var(--text-color, #000000)";
  });

  return closeBtn;
}

function createSchemaContent(schemaString: string): HTMLPreElement {
  const schemaContent = dom.createElement<HTMLPreElement>("pre");
  schemaContent.textContent = schemaString;
  schemaContent.style.cssText = `
    background: var(--background-alt, #f8f9fa);
    color: var(--text-color, #000000);
    border: 1px solid var(--border-color, #e5e7eb);
    border-radius: 4px;
    padding: 16px;
    margin: 0;
    font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
    font-size: 12px;
    line-height: 1.5;
    overflow-x: auto;
    white-space: pre-wrap;
    word-wrap: break-word;
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
}
