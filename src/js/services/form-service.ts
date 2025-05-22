export interface FormSchema {
  display: string;
  components: any[];
  [key: string]: any;
}

export class FormService {
  private static instance: FormService;
  private baseUrl = "/dashboard/forms";

  private constructor() {}

  static getInstance(): FormService {
    if (!FormService.instance) {
      FormService.instance = new FormService();
    }
    return FormService.instance;
  }

  private getCSRFToken(): string {
    return (
      document
        .querySelector('meta[name="csrf-token"]')
        ?.getAttribute("content") || ""
    );
  }

  async getSchema(formId: string): Promise<FormSchema> {
    console.log("Getting schema for formId:", formId);
    const response = await fetch(`${this.baseUrl}/${formId}/schema`);
    if (!response.ok) {
      throw new Error("Failed to load form schema");
    }
    return response.json();
  }

  async saveSchema(formId: string, schema: FormSchema): Promise<FormSchema> {
    const response = await fetch(`${this.baseUrl}/${formId}/schema`, {
      method: "PUT",
      headers: {
        "Content-Type": "application/json",
        "X-CSRF-Token": this.getCSRFToken(),
      },
      body: JSON.stringify(schema),
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.message || "Failed to save form schema");
    }

    return response.json();
  }

  async updateFormDetails(
    formId: string,
    details: { title: string; description: string },
  ): Promise<void> {
    const response = await fetch(`${this.baseUrl}/${formId}`, {
      method: "PUT",
      headers: {
        "Content-Type": "application/json",
        "X-CSRF-Token": this.getCSRFToken(),
      },
      body: JSON.stringify(details),
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.message || "Failed to update form details");
    }
  }
}
