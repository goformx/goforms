import { Formio } from "@formio/js";

type FormBuilderOptions = Parameters<typeof Formio.builder>[2];

export enum FieldType {
  Text = "textfield",
  TextArea = "textarea",
  Email = "email",
  Phone = "phoneNumber",
  Number = "number",
  Date = "datetime",
  Select = "select",
  Checkbox = "checkbox",
  Radio = "radio",
  File = "file",
}

export enum IconType {
  Text = "terminal",
  Email = "at",
  Phone = "phone-square",
  Number = "hash",
  Date = "calendar",
  Select = "list",
  Checkbox = "check-square",
  Radio = "dot-circle",
  File = "file",
}

// Enhanced validation configuration
export interface ValidationConfig {
  required?: boolean;
  minLength?: number;
  maxLength?: number;
  pattern?: string;
  custom?: string;
  min?: number;
  max?: number;
}

// Enhanced field schema interface
export interface UserFieldSchema {
  title: string;
  key: string;
  icon: IconType;
  schema: {
    label: string;
    type: FieldType;
    key: string;
    input: boolean;
    placeholder?: string;
    description?: string;
    validate?: ValidationConfig;
    conditional?: {
      show: boolean;
      when: string;
      eq: string;
    };
    data?: {
      values?: Array<{ label: string; value: string }>;
      url?: string;
    };
  };
}

// Type-safe icon mapping
const IconMap: Record<FieldType, IconType> = {
  [FieldType.Text]: IconType.Text,
  [FieldType.TextArea]: IconType.Text,
  [FieldType.Email]: IconType.Email,
  [FieldType.Phone]: IconType.Phone,
  [FieldType.Number]: IconType.Number,
  [FieldType.Date]: IconType.Date,
  [FieldType.Select]: IconType.Select,
  [FieldType.Checkbox]: IconType.Checkbox,
  [FieldType.Radio]: IconType.Radio,
  [FieldType.File]: IconType.File,
} as const;

const getIconForType = (type: FieldType): IconType => IconMap[type];

// Enhanced field schema factory with overloads for different field types
export function createUserFieldSchema(
  key: string,
  label: string,
  type: FieldType.Select,
  options: { values: Array<{ label: string; value: string }> },
  validation?: ValidationConfig,
): UserFieldSchema;

export function createUserFieldSchema(
  key: string,
  label: string,
  type: FieldType.Number,
  options?: { min?: number; max?: number; step?: number },
  validation?: ValidationConfig,
): UserFieldSchema;

export function createUserFieldSchema(
  key: string,
  label: string,
  type?: FieldType,
  validation?: ValidationConfig,
): UserFieldSchema;

export function createUserFieldSchema(
  key: string,
  label: string,
  type: FieldType = FieldType.Text,
  optionsOrValidation?:
    | ValidationConfig
    | {
        values?: Array<{ label: string; value: string }>;
        min?: number;
        max?: number;
        step?: number;
      },
  validation?: ValidationConfig,
): UserFieldSchema {
  const fieldType = type;
  const isValidationConfig = (obj: any): obj is ValidationConfig =>
    obj && typeof obj === "object" && !Array.isArray(obj.values);

  let finalValidation: ValidationConfig | undefined;
  let fieldOptions: any = {};

  if (validation) {
    finalValidation = validation;
    fieldOptions = optionsOrValidation || {};
  } else if (isValidationConfig(optionsOrValidation)) {
    finalValidation = optionsOrValidation;
  } else {
    fieldOptions = optionsOrValidation || {};
  }

  const baseSchema: UserFieldSchema["schema"] = {
    label,
    type: fieldType,
    key,
    input: true,
    ...(finalValidation && { validate: finalValidation }),
  };

  // Type-specific enhancements
  switch (fieldType) {
    case FieldType.Select:
      if (fieldOptions.values) {
        baseSchema.data = { values: fieldOptions.values };
      }
      break;
    case FieldType.Number:
      if (fieldOptions.min !== undefined) {
        baseSchema.validate = { ...baseSchema.validate, min: fieldOptions.min };
      }
      if (fieldOptions.max !== undefined) {
        baseSchema.validate = { ...baseSchema.validate, max: fieldOptions.max };
      }
      break;
    case FieldType.Email:
      baseSchema.validate = {
        ...baseSchema.validate,
        pattern: "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$",
      };
      break;
  }

  return {
    title: label,
    key,
    icon: getIconForType(fieldType),
    schema: baseSchema,
  };
}

// Predefined field configurations
export const PredefinedFields = {
  // Contact Information
  firstName: () =>
    createUserFieldSchema("firstName", "First Name", FieldType.Text, {
      required: true,
      minLength: 2,
      maxLength: 50,
    }),

  lastName: () =>
    createUserFieldSchema("lastName", "Last Name", FieldType.Text, {
      required: true,
      minLength: 2,
      maxLength: 50,
    }),

  email: () =>
    createUserFieldSchema("email", "Email Address", FieldType.Email, {
      required: true,
    }),

  phone: () =>
    createUserFieldSchema("phoneNumber", "Phone Number", FieldType.Phone, {
      pattern: "^[\\+]?[1-9][\\d\\s\\-\\(\\)]{7,15}$",
    }),

  // Demographics
  age: () =>
    createUserFieldSchema(
      "age",
      "Age",
      FieldType.Number,
      { min: 13, max: 120 },
      {
        required: true,
      },
    ),

  gender: () =>
    createUserFieldSchema("gender", "Gender", FieldType.Select, {
      values: [
        { label: "Male", value: "male" },
        { label: "Female", value: "female" },
        { label: "Other", value: "other" },
        { label: "Prefer not to say", value: "not_specified" },
      ],
    }),

  // Address
  address: () =>
    createUserFieldSchema("address", "Street Address", FieldType.TextArea, {
      required: true,
      maxLength: 200,
    }),

  city: () =>
    createUserFieldSchema("city", "City", FieldType.Text, {
      required: true,
      maxLength: 100,
    }),

  // Preferences
  newsletter: () =>
    createUserFieldSchema(
      "newsletter",
      "Subscribe to Newsletter",
      FieldType.Checkbox,
    ),
} as const;

// Enhanced basic components with more options
const basicComponents = {
  textfield: true,
  textarea: true,
  email: true,
  phoneNumber: true,
  number: true,
  select: true,
  checkbox: true,
  radio: true,
  datetime: true,
  file: true,
};

// Dynamic user fields - can be customized per form
const createUserFields = () => ({
  firstName: PredefinedFields.firstName(),
  lastName: PredefinedFields.lastName(),
  email: PredefinedFields.email(),
  phone: PredefinedFields.phone(),
  age: PredefinedFields.age(),
  gender: PredefinedFields.gender(),
  address: PredefinedFields.address(),
  city: PredefinedFields.city(),
  newsletter: PredefinedFields.newsletter(),
});

// Enhanced builder configuration with better organization
const createBuilderSections = (customFields = createUserFields()) => ({
  premium: false,
  basic: false,
  advanced: false,
  data: false,

  // Custom user fields section
  custom: {
    title: "User Fields",
    weight: 0,
    default: true,
    components: customFields,
  },

  // Standard form components
  customBasic: {
    title: "Basic Components",
    weight: 1,
    components: basicComponents,
  },

  // Layout controls (minimal)
  layout: {
    title: "Layout",
    weight: 2,
    components: {
      table: false,
      panel: true,
      fieldset: true,
      columns: true,
    },
  },
});

// Enhanced builder options with better defaults
export const createBuilderOptions = (
  customFields?: ReturnType<typeof createUserFields>,
  overrides: Partial<FormBuilderOptions> = {},
): FormBuilderOptions => ({
  display: "form" as const,
  builder: createBuilderSections(customFields),

  editForm: {
    textfield: [
      {
        key: "api",
        ignore: true,
      },
    ],
  },

  noDefaultSubmitButton: false,
  language: "en",
  template: "goforms",

  submitButton: {
    type: "button",
    label: "Submit",
    key: "submit",
    size: "lg",
    block: true,
    action: "submit",
    theme: "primary",
    disableOnInvalid: true,
    customClass: "gf-button gf-button--primary gf-button--lg",
    unique: true,
    position: "bottom",
  },

  // Development helpers
  showSchema: import.meta.env.DEV,
  showJSONEditor: import.meta.env.DEV,
  showPreview: true,

  // Apply any custom overrides
  ...overrides,
});

// Default export for backwards compatibility
export const builderOptions = createBuilderOptions();
