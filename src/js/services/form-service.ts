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
    try {
      console.log("Form-service: Starting schema save for form:", formId);
      console.log("Form-service: Schema to save:", schema);

      const response = await fetch(`${this.baseUrl}/${formId}/schema`, {
        method: "PUT",
        headers: {
          "Content-Type": "application/json",
          "X-CSRF-Token": this.getCSRFToken(),
        },
        body: JSON.stringify(schema),
      });

      console.log("Form-service: Server response status:", response.status);
      const responseText = await response.text();
      console.log("Form-service: Server response body:", responseText);

      if (!response.ok) {
        console.error(
          "Form-service: Server error:",
          response.status,
          responseText,
        );
        throw new Error(`Server error: ${response.status} - ${responseText}`);
      }

      // Parse the response as JSON
      let data: FormSchema;
      try {
        data = JSON.parse(responseText);
        console.log("Form-service: Parsed response data:", data);
      } catch (e) {
        console.error("Form-service: Failed to parse response as JSON:", e);
        throw new Error("Invalid response from server");
      }

      if (!data) {
        console.error("Form-service: Empty response from server");
        throw new Error("Empty response from server");
      }

      console.log("Form-service: Save successful, returning schema");
      return data;
    } catch (error) {
      console.error("Form-service: Error in saveSchema:", error);
      throw error;
    }
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
