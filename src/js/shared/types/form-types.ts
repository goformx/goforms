// ===== src/js/forms/types/form-types.ts =====

/**
 * Core form configuration interface
 */
export interface FormConfig {
  formId: string;
  validationType: ValidationType;
  validationDelay?: number;
  options?: FormOptions;
}

/**
 * Validation type enumeration
 */
export type ValidationType = "realtime" | "onSubmit" | "hybrid";

/**
 * Form options for enhanced configuration
 */
export interface FormOptions {
  autoSave?: boolean;
  debounceMs?: number;
  showProgress?: boolean;
  allowDraft?: boolean;
  maxFileSize?: number;
}

/**
 * Comprehensive form schema interface
 */
export interface FormSchema {
  display: string;
  components: FormComponent[];
  metadata?: FormMetadata;
  settings?: FormSettings;
}

/**
 * Form component interface with proper typing
 */
export interface FormComponent {
  type: ComponentType;
  key: string;
  label: string;
  input: boolean;
  validate?: ValidationRule[];
  conditional?: ConditionalLogic;
  properties?: Record<string, unknown>;
  defaultValue?: unknown;
  placeholder?: string;
  required?: boolean;
  disabled?: boolean;
  hidden?: boolean;
}

/**
 * Component type enumeration
 */
export type ComponentType =
  | "textfield"
  | "textarea"
  | "select"
  | "checkbox"
  | "radio"
  | "button"
  | "email"
  | "password"
  | "number"
  | "date"
  | "file"
  | "url"
  | "phone"
  | "address"
  | "signature";

/**
 * Validation rule interface
 */
export interface ValidationRule {
  type:
    | "required"
    | "minLength"
    | "maxLength"
    | "pattern"
    | "custom"
    | "email"
    | "url";
  value?: string | number;
  message: string;
  enabled?: boolean;
}

/**
 * Conditional logic interface
 */
export interface ConditionalLogic {
  when: string;
  eq: unknown;
  show?: boolean;
  required?: boolean;
  disabled?: boolean;
}

/**
 * Form metadata interface
 */
export interface FormMetadata {
  title?: string;
  description?: string;
  version?: string;
  created?: Date;
  modified?: Date;
  author?: string;
  tags?: string[];
}

/**
 * Form settings interface
 */
export interface FormSettings {
  submitButtonText?: string;
  cancelButtonText?: string;
  showCancelButton?: boolean;
  allowMultipleSubmissions?: boolean;
  requireAuthentication?: boolean;
  redirectUrl?: string;
  successMessage?: string;
  errorMessage?: string;
}

/**
 * Server response interface
 */
export interface ServerResponse {
  message?: string;
  redirect?: string;
  success?: boolean;
  data?: unknown;
  errors?: Record<string, string[]>;
}

/**
 * Request options interface
 */
export interface RequestOptions {
  body: FormData | string;
  headers: Record<string, string>;
  timeout?: number;
  retries?: number;
}

/**
 * Form submission data interface
 */
export interface FormSubmissionData {
  formId: string;
  data: Record<string, unknown>;
  metadata?: {
    submittedAt: Date;
    userAgent?: string;
    ipAddress?: string;
    sessionId?: string;
  };
}

/**
 * Form validation result interface
 */
export interface FormValidationResult {
  isValid: boolean;
  errors: ValidationError[];
  warnings: ValidationWarning[];
}

/**
 * Validation error interface
 */
export interface ValidationError {
  field: string;
  message: string;
  code: string;
  value?: unknown;
}

/**
 * Validation warning interface
 */
export interface ValidationWarning {
  field: string;
  message: string;
  code: string;
  value?: unknown;
}

/**
 * Form builder state interface
 */
export interface FormBuilderState {
  isDirty: boolean;
  isSaving: boolean;
  hasErrors: boolean;
  lastSaved?: Date;
  autoSaveEnabled: boolean;
}

/**
 * Form field interface
 */
export interface FormField {
  name: string;
  type: ComponentType;
  value: unknown;
  isValid: boolean;
  error?: string;
  isRequired: boolean;
  isDisabled: boolean;
  isHidden: boolean;
}
