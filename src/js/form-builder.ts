import { Formio, FormBuilder } from "@formio/js";
import goforms from "goforms";

// Import Form.io styles
import "@formio/js/dist/formio.full.min.css";

// Register templates
Formio.use(goforms);

// Define builder options
const builderOptions = {
  display: "form" as "form" | "wizard" | "pdf",
  builder: {
    basic: {
      components: {
        textfield: true,
        textarea: true,
        email: true,
        phoneNumber: true,
        number: true,
        password: true,
        checkbox: true,
        selectboxes: true,
        select: true,
        radio: true,
        button: true,
      },
    },
    advanced: false,
    layout: false, // Disable layout components
    data: false, // Disable data components
    premium: false, // Disable premium components
    wizard: false, // Explicitly disable wizard
  },
  noDefaultSubmitButton: true,
  language: "en",
  template: "goforms",
};

// Initialize form builder when the module is loaded
const formSchemaBuilder = document.getElementById("form-schema-builder");

if (formSchemaBuilder) {
  const formIdAttr = formSchemaBuilder.getAttribute("data-form-id");
  const formId = formIdAttr ? Number(formIdAttr) : 0;

  if (formId > 0) {
    // Use Formio.builder instead of new FormBuilder
    Formio.builder(formSchemaBuilder, {}, builderOptions).then(
      (builder: FormBuilder) => {
        (
          window as unknown as { formBuilderInstance?: FormBuilder }
        ).formBuilderInstance = builder;
      },
    );
  }
}
