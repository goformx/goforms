import { Formio, FormBuilder } from "@formio/js";
import goforms from "goforms";
import { FormService, FormSchema } from "./services/form-service";

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
const formService = FormService.getInstance();

if (formSchemaBuilder) {
  const formIdAttr = formSchemaBuilder.getAttribute("data-form-id");
  const formId = formIdAttr ? Number(formIdAttr) : 0;

  if (formId > 0) {
    // Load schema and initialize builder
    formService
      .getSchema(formId)
      .then((schema) => {
        Formio.builder(formSchemaBuilder, schema, builderOptions).then(
          (builder: FormBuilder) => {
            // Add saveSchema method to the builder instance
            (builder as any).saveSchema = async () => {
              try {
                const schema = builder.form as FormSchema;
                await formService.saveSchema(formId, schema);
                return true;
              } catch (error) {
                console.error("Error saving schema:", error);
                return false;
              }
            };

            (
              window as unknown as { formBuilderInstance?: FormBuilder }
            ).formBuilderInstance = builder;
          },
        );
      })
      .catch((error) => {
        console.error("Error loading schema:", error);
      });
  }
}
