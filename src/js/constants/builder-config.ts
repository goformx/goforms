import { Formio } from "@formio/js";

type FormBuilderOptions = Parameters<typeof Formio.builder>[2];

export const createUserFieldSchema = (
  key: string,
  label: string,
  type: string,
  required = false,
) => ({
  title: label,
  key,
  icon:
    type === "email"
      ? "at"
      : type === "phoneNumber"
        ? "phone-square"
        : "terminal",
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

export const builderOptions: FormBuilderOptions = {
  display: "form" as "form" | "wizard" | "pdf",
  builder: {
    premium: false,
    basic: false,
    advanced: false,
    data: false,
    custom: {
      title: "User Fields",
      weight: 10,
      components: {
        firstName: createUserFieldSchema(
          "firstName",
          "First Name",
          "textfield",
          true,
        ),
        lastName: createUserFieldSchema(
          "lastName",
          "Last Name",
          "textfield",
          true,
        ),
        email: createUserFieldSchema("email", "Email", "email", true),
        phoneNumber: createUserFieldSchema(
          "mobilePhone",
          "Mobile Phone",
          "phoneNumber",
        ),
      },
    },
    customBasic: {
      title: "Basic Components",
      default: true,
      weight: 0,
      components: {
        textfield: true,
        textarea: true,
        email: true,
        phoneNumber: true,
      },
    },
    layout: {
      components: {
        table: false,
      },
    },
  },
  editForm: {
    textfield: [
      {
        key: "api",
        ignore: true,
      },
    ],
  },
  noDefaultSubmitButton: true,
  language: "en",
  template: "goforms",
};
