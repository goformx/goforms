// ===== src/js/forms/types/form-types.ts =====
export interface FormConfig {
  formId: string;
  validationType: string;
  validationDelay?: number;
}

export interface ServerResponse {
  message?: string;
  redirect?: string;
}

export interface RequestOptions {
  body: FormData | string;
  headers: Record<string, string>;
}
