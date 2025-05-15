import { Formio, FormBuilder } from "@formio/js";
import goforms from "goforms-template";
import bootstrap from "@formio/bootstrap";

// Import Form.io styles
import "@formio/js/dist/formio.full.min.css";

// Register templates
Formio.use(goforms);
Formio.use(bootstrap);

// Define builder options
/** @type {import('@formio/js').FormBuilder['options']} */
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
    advanced: {
      components: {}, // Empty to disable all advanced components
    },
    layout: false, // Disable layout components
    data: false, // Disable data components
    premium: false, // Disable premium components
    wizard: false, // Explicitly disable wizard
  },
  noDefaultSubmitButton: true,
  language: "en",
  template: "bootstrap5",
};

// Initialize form builder when the module is loaded
const formSchemaBuilder = document.getElementById("form-schema-builder");

if (formSchemaBuilder) {
  const formIdAttr = formSchemaBuilder.getAttribute("data-form-id");
  const formId = formIdAttr ? Number(formIdAttr) : 0;

  if (formId > 0) {
    (
      window as unknown as { formBuilderInstance?: FormBuilder }
    ).formBuilderInstance = new FormBuilder(
      formSchemaBuilder,
      {},
      builderOptions,
    );
  }
}
