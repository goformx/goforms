// ===== src/js/forms/types/form-types.ts =====

/**
 * Branded types for better type safety
 */
export type FormId = string & { readonly __brand: "FormId" };
export type ComponentKey = string & { readonly __brand: "ComponentKey" };
export type FieldName = string & { readonly __brand: "FieldName" };

/**
 * Core form configuration interface
 */
export interface FormConfig {
  readonly formId: FormId;
  readonly validationType: ValidationType;
  readonly validationDelay?: number;
  readonly options?: FormOptions;
}

/**
 * Validation type enumeration - using const assertion for better type inference
 */
export const VALIDATION_TYPES = ["realtime", "onSubmit", "hybrid"] as const;
export type ValidationType = (typeof VALIDATION_TYPES)[number];

/**
 * Form options for enhanced configuration
 */
export interface FormOptions {
  readonly autoSave?: boolean;
  readonly debounceMs?: number;
  readonly showProgress?: boolean;
  readonly allowDraft?: boolean;
  readonly maxFileSize?: number;
}

/**
 * Comprehensive form schema interface
 */
export interface FormSchema {
  readonly display: string;
  readonly components: readonly FormComponent[];
  readonly metadata?: FormMetadata;
  readonly settings?: FormSettings;
}

/**
 * Form component interface with proper typing
 */
export interface FormComponent {
  readonly type: ComponentType;
  readonly key: ComponentKey;
  readonly label: string;
  readonly input: boolean;
  readonly validate?: readonly ValidationRule[];
  readonly conditional?: ConditionalLogic;
  readonly properties?: Readonly<Record<string, unknown>>;
  readonly defaultValue?: unknown;
  readonly placeholder?: string;
  readonly required?: boolean;
  readonly disabled?: boolean;
  readonly hidden?: boolean;
}

/**
 * Component type enumeration - using const assertion
 */
export const COMPONENT_TYPES = [
  "textfield",
  "textarea",
  "select",
  "checkbox",
  "radio",
  "button",
  "email",
  "password",
  "number",
  "date",
  "file",
  "url",
  "phone",
  "address",
  "signature",
] as const;

export type ComponentType = (typeof COMPONENT_TYPES)[number];

/**
 * Validation rule interface
 */
export interface ValidationRule {
  readonly type: ValidationRuleType;
  readonly value?: string | number;
  readonly message: string;
  readonly enabled?: boolean;
}

/**
 * Validation rule types - using const assertion
 */
export const VALIDATION_RULE_TYPES = [
  "required",
  "minLength",
  "maxLength",
  "pattern",
  "custom",
  "email",
  "url",
] as const;

export type ValidationRuleType = (typeof VALIDATION_RULE_TYPES)[number];

/**
 * Conditional logic interface
 */
export interface ConditionalLogic {
  readonly when: string;
  readonly eq: unknown;
  readonly show?: boolean;
  readonly required?: boolean;
  readonly disabled?: boolean;
}

/**
 * Form metadata interface
 */
export interface FormMetadata {
  readonly title?: string;
  readonly description?: string;
  readonly version?: string;
  readonly created?: Date;
  readonly modified?: Date;
  readonly author?: string;
  readonly tags?: readonly string[];
}

/**
 * Form settings interface
 */
export interface FormSettings {
  readonly submitButtonText?: string;
  readonly cancelButtonText?: string;
  readonly showCancelButton?: boolean;
  readonly allowMultipleSubmissions?: boolean;
  readonly requireAuthentication?: boolean;
  readonly redirectUrl?: string;
  readonly successMessage?: string;
  readonly errorMessage?: string;
}

/**
 * Server response interface with generic type
 */
export interface ServerResponse<T = unknown> {
  readonly success: boolean;
  readonly message?: string;
  readonly data?: T;
  readonly errors?: Readonly<Record<string, readonly string[]>>;
}

/**
 * Request options interface
 */
export interface RequestOptions {
  readonly body: FormData | string;
  readonly headers: Readonly<Record<string, string>>;
  readonly timeout?: number;
  readonly retries?: number;
}

/**
 * Form submission data interface
 */
export interface FormSubmissionData {
  readonly formId: FormId;
  readonly data: Readonly<Record<string, unknown>>;
  readonly metadata?: {
    readonly submittedAt: Date;
    readonly userAgent?: string;
    readonly ipAddress?: string;
    readonly sessionId?: string;
  };
}

/**
 * Form validation result interface
 */
export interface FormValidationResult {
  readonly isValid: boolean;
  readonly errors: readonly ValidationError[];
  readonly warnings: readonly ValidationWarning[];
}

/**
 * Validation error interface
 */
export interface ValidationError {
  readonly field: FieldName;
  readonly message: string;
  readonly code: string;
  readonly value?: unknown;
}

/**
 * Validation warning interface
 */
export interface ValidationWarning {
  readonly field: FieldName;
  readonly message: string;
  readonly code: string;
  readonly value?: unknown;
}

/**
 * Form builder state interface
 */
export interface FormBuilderState {
  readonly isDirty: boolean;
  readonly isSaving: boolean;
  readonly hasErrors: boolean;
  readonly lastSaved?: Date;
  readonly autoSaveEnabled: boolean;
}

/**
 * Form field interface
 */
export interface FormField {
  readonly name: FieldName;
  readonly type: ComponentType;
  readonly value: unknown;
  readonly isValid: boolean;
  readonly error?: string;
  readonly isRequired: boolean;
  readonly isDisabled: boolean;
  readonly isHidden: boolean;
}

/**
 * Type guards for runtime type checking
 */
export const isValidationType = (value: unknown): value is ValidationType => {
  return VALIDATION_TYPES.includes(value as ValidationType);
};

export const isComponentType = (value: unknown): value is ComponentType => {
  return COMPONENT_TYPES.includes(value as ComponentType);
};

export const isValidationRuleType = (
  value: unknown,
): value is ValidationRuleType => {
  return VALIDATION_RULE_TYPES.includes(value as ValidationRuleType);
};

/**
 * Utility types for better type inference
 */
export type FormComponentMap = ReadonlyMap<ComponentKey, FormComponent>;
export type FormFieldMap = ReadonlyMap<FieldName, FormField>;
export type ValidationErrorMap = ReadonlyMap<
  FieldName,
  readonly ValidationError[]
>;

/**
 * Type-safe form builder functions
 */
export const createFormId = (id: string): FormId => id as FormId;
export const createComponentKey = (key: string): ComponentKey =>
  key as ComponentKey;
export const createFieldName = (name: string): FieldName => name as FieldName;
