export interface FormSchema {
  display: string;
  components: any[];
  [key: string]: any;
}

export class FormService {
  private static instance: FormService;
  private baseUrl: string;

  private constructor() {
    // Use environment-based base URL
    this.baseUrl =
      process.env.NODE_ENV === "production"
        ? "https://goformx.com" // Production URL
        : "http://localhost:8090"; // Development URL
  }

  public static getInstance(): FormService {
    if (!FormService.instance) {
      FormService.instance = new FormService();
    }
    return FormService.instance;
  }

  // Set base URL (useful for testing or custom deployments)
  public setBaseUrl(url: string): void {
    this.baseUrl = url;
  }

  private getCSRFToken(): string {
    return (
      document
        .querySelector('meta[name="csrf-token"]')
        ?.getAttribute("content") || ""
    );
  }

  async getSchema(formId: string): Promise<FormSchema> {
    const response = await fetch(
      `${this.baseUrl}/api/v1/forms/${formId}/schema`,
    );
    if (!response.ok) {
      throw new Error("Failed to load form schema");
    }
    return response.json().then((data) => data as FormSchema);
  }

  async saveSchema(formId: string, schema: FormSchema): Promise<FormSchema> {
    const response = await fetch(
      `${this.baseUrl}/api/v1/forms/${formId}/schema`,
      {
        method: "PUT",
        headers: {
          "Content-Type": "application/json",
          "X-CSRF-Token": this.getCSRFToken(),
        },
        body: JSON.stringify(schema),
      },
    );

    const responseText = await response.text();

    if (!response.ok) {
      throw new Error(`Server error: ${response.status} - ${responseText}`);
    }

    // Parse the response as JSON
    let data: FormSchema;
    try {
      data = JSON.parse(responseText);
    } catch (_error) {
      throw new Error("Invalid response from server");
    }

    if (!data) {
      throw new Error("Empty response from server");
    }

    return data;
  }

  async updateFormDetails(
    formId: string,
    details: { title: string; description: string },
  ): Promise<void> {
    const response = await fetch(`${this.baseUrl}/dashboard/forms/${formId}`, {
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

  async deleteForm(formId: string): Promise<void> {
    const response = await fetch(`${this.baseUrl}/dashboard/forms/${formId}`, {
      method: "DELETE",
      headers: {
        "X-CSRF-Token": this.getCSRFToken(),
      },
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.message || "Failed to delete form");
    }
  }

  async submitForm(formId: string, data: FormData): Promise<Response> {
    const response = await fetch(
      `${this.baseUrl}/api/v1/forms/${formId}/submit`,
      {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Accept: "application/json",
        },
        body: JSON.stringify(data),
      },
    );

    if (!response.ok) {
      const error = await response
        .json()
        .catch(() => ({ message: "Failed to submit form" }));
      throw new Error(error.message || "Failed to submit form");
    }

    return response;
  }
}
