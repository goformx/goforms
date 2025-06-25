import { Logger } from "@/core/logger";

// Types and Interfaces
interface DOMIds {
  DEMO_FORM: string;
  API_RESPONSE: string;
  MESSAGES_LIST: string;
}

interface Templates {
  API_RESPONSE: (type: string, data: unknown) => string;
  DEFAULT_RESPONSE: string;
}

interface FormData {
  name: string;
  email: string;
}

interface Submission extends FormData {
  timestamp?: string;
}

interface APIResponse {
  type: string;
  data: unknown;
  timestamp: Date;
}

// Constants and Configuration
const DOM_IDS: DOMIds = {
  DEMO_FORM: "demo-form",
  API_RESPONSE: "api-response",
  MESSAGES_LIST: "messages-list",
};

const TEMPLATES: Templates = {
  API_RESPONSE: (type: string, data: unknown) => `
    <div class="api-response api-response--${type}">
      <div class="api-response__header">
        <span class="api-response__type">${type.toUpperCase()}</span>
        <span class="api-response__timestamp">${new Date().toLocaleTimeString()}</span>
      </div>
      <div class="api-response__content">
        <pre>${JSON.stringify(data, null, 2)}</pre>
      </div>
    </div>
  `,
  DEFAULT_RESPONSE: `
    <div class="api-response api-response--default">
      <div class="api-response__content">
        <p>No API responses yet. Submit the form to see responses here.</p>
      </div>
    </div>
  `,
};

// Utility Functions
const formatDate = (date: Date): string => {
  return new Intl.DateTimeFormat("default", {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(date);
};

const logDebug = (message: string, data?: unknown): void => {
  const timestamp = new Date().toISOString();
  Logger.debug(`[${timestamp}] ${message}`, data ?? "");
};

const logError = (message: string, error: Error): void => {
  const timestamp = new Date().toISOString();
  Logger.error(`[${timestamp}] ${message}`, error);
  if (error?.stack) Logger.error(`[${timestamp}] Error stack:`, error.stack);
};

// DOM Helpers
const getElement = (id: string): HTMLElement | null => {
  const element = document.getElementById(id);
  if (!element) {
    logDebug(`Element not found: ${id}`);
  }
  return element;
};

// API Response Display Component
class APIResponseDisplay {
  private readonly container: HTMLElement | null;
  private responses: APIResponse[];
  private readonly maxResponses: number;

  constructor(containerId: string) {
    this.container = getElement(containerId);
    this.responses = [];
    this.maxResponses = 5;
    this.showDefault();
  }

  showDefault(): void {
    if (this.container) {
      this.container.innerHTML = TEMPLATES.DEFAULT_RESPONSE;
    }
  }

  addResponse(type: string, data: unknown): void {
    this.responses.unshift({ type, data, timestamp: new Date() });
    if (this.responses.length > this.maxResponses) {
      this.responses.pop();
    }
    this.render();
  }

  render(): void {
    if (!this.container) return;

    this.container.innerHTML = this.responses
      .map(({ type, data }) => TEMPLATES.API_RESPONSE(type, data))
      .join("");
  }

  clear(): void {
    this.responses = [];
    this.showDefault();
  }
}

// Form Handler Component
class DemoForm {
  private readonly form: HTMLFormElement | null;
  private readonly apiResponse: APIResponseDisplay;
  private readonly messagesList: HTMLElement | null;

  constructor(formId: string) {
    this.form = getElement(formId) as HTMLFormElement;
    this.apiResponse = new APIResponseDisplay(DOM_IDS.API_RESPONSE);
    this.messagesList = getElement(DOM_IDS.MESSAGES_LIST);
    this.setupListeners();
  }

  private setupListeners(): void {
    if (this.form) {
      this.form.addEventListener("submit", this.handleSubmit.bind(this));
    }
  }

  private getFormData(): FormData {
    const formData = new FormData(this.form!);
    return {
      name: formData.get("name") as string,
      email: formData.get("email") as string,
    };
  }

  private async handleSubmit(event: Event): Promise<void> {
    event.preventDefault();
    logDebug("Form submission started...");

    const formData = this.getFormData();
    logDebug("Form data:", formData);

    try {
      // Simulate API call
      await new Promise((resolve) => setTimeout(resolve, 500));

      // Add new submission to the list
      this.addSubmissionToList(formData);

      // Show success message
      this.showMessage(
        "success",
        "Form submitted successfully! Check the submissions list.",
      );
      this.form?.reset();
    } catch (error) {
      logError("Submit error:", error as Error);
      this.apiResponse.addResponse("error", {
        error: (error as Error).message,
      });
      this.showMessage("error", "Failed to submit form. Please try again.");
    }
  }

  private addSubmissionToList(submission: Submission): void {
    if (!this.messagesList) return;

    const item = document.createElement("div");
    item.className = "message-item";
    item.innerHTML = `
      <div class="message-header">
        <strong>${submission.name}</strong>
        <span class="timestamp">${submission.timestamp || formatDate(new Date())}</span>
      </div>
      <div class="message-content">
        <span class="email">${submission.email}</span>
      </div>
    `;
    this.messagesList.insertBefore(item, this.messagesList.firstChild);
  }

  private showMessage(type: string, text: string): void {
    if (!this.form?.parentNode) return;

    const message = document.createElement("div");
    message.className = `alert alert-${type}`;
    message.textContent = text;
    this.form.parentNode.insertBefore(message, this.form);

    setTimeout(() => {
      message.remove();
    }, 5000);
  }
}

// Initialize form handler when DOM is loaded
document.addEventListener("DOMContentLoaded", () => {
  const form = document.getElementById("demo-form");
  const messagesList = document.getElementById("messages-list");

  if (!form || !messagesList) return;

  // Demo submissions to show in the list
  const demoSubmissions: Submission[] = [
    {
      name: "John Smith",
      email: "john@example.com",
      timestamp: "2 minutes ago",
    },
    {
      name: "Sarah Wilson",
      email: "s.wilson@example.com",
      timestamp: "5 minutes ago",
    },
    {
      name: "Mike Johnson",
      email: "mike.j@example.com",
      timestamp: "10 minutes ago",
    },
  ];

  // Display demo submissions
  demoSubmissions.forEach((submission) => {
    const item = document.createElement("div");
    item.className = "message-item";
    item.innerHTML = `
      <div class="message-header">
        <strong>${submission.name}</strong>
        <span class="timestamp">${submission.timestamp}</span>
      </div>
      <div class="message-content">
        <span class="email">${submission.email}</span>
      </div>
    `;
    messagesList.insertBefore(item, messagesList.firstChild);
  });

  // Initialize the form handler
  new DemoForm(DOM_IDS.DEMO_FORM);
});
