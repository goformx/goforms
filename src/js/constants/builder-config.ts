import { Formio } from "@formio/js";

type FormBuilderOptions = Parameters<typeof Formio.builder>[2];

export enum FieldType {
  Text = "textfield",
  TextArea = "textarea",
  Email = "email",
  Phone = "phoneNumber",
}

export enum IconType {
  Text = "terminal",
  Email = "at",
  Phone = "phone-square",
}

// Icon mapping using Record type for type safety
const IconMap: Record<FieldType, IconType> = {
  [FieldType.Text]: IconType.Text,
  [FieldType.TextArea]: IconType.Text,
  [FieldType.Email]: IconType.Email,
  [FieldType.Phone]: IconType.Phone,
};

const getIconForType = (type: FieldType): IconType => IconMap[type];

export const createUserFieldSchema = (
  key: string,
  label: string,
  type: FieldType = FieldType.Text,
  required = false,
) => ({
  title: label,
  key,
  icon: getIconForType(type),
  schema: {
    label,
    type,
    key,
    input: true,
    ...(required && {
      validate: {
        required: true,
      },
    }),
  },
});

// Basic components configuration
const basicComponents = {
  textfield: true,
  textarea: true,
  email: true,
  phoneNumber: true,
};

// User fields configuration
const userFields = {
  firstName: createUserFieldSchema(
    "firstName",
    "First Name",
    FieldType.Text,
    true,
  ),
  lastName: createUserFieldSchema(
    "lastName",
    "Last Name",
    FieldType.Text,
    true,
  ),
  email: createUserFieldSchema("email", "Email", FieldType.Email, true),
  phoneNumber: createUserFieldSchema(
    "mobilePhone",
    "Mobile Phone",
    FieldType.Phone,
  ),
};

// Builder sections configuration
const builderSections = {
  premium: false,
  basic: false,
  advanced: false,
  data: false,
  custom: {
    title: "User Fields",
    weight: 0,
    components: userFields,
  },
  customBasic: {
    title: "Basic Components",
    default: true,
    weight: 1,
    components: basicComponents,
  },
  layout: {
    components: {
      table: false,
    },
  },
};

export const builderOptions: FormBuilderOptions = {
  display: "form" as "form" | "wizard" | "pdf",
  builder: builderSections,
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
};
