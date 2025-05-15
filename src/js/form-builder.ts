import { Formio } from "@formio/js";
import goforms from "goforms";
import { FormService } from "./services/form-service";
import { builderOptions } from "./constants/builder-config";
import { setupBuilderEvents } from "./handlers/builder-events";

// Import Form.io styles
import "@formio/js/dist/formio.full.min.css";

// Register templates
Formio.use(goforms);

// Initialize form builder when the module is loaded
const formSchemaBuilder = document.getElementById("form-schema-builder");

if (!formSchemaBuilder) {
  console.error("Form schema builder element not found");
  throw new Error("Form schema builder element not found");
}

const formIdAttr = formSchemaBuilder.getAttribute("data-form-id");
if (!formIdAttr) {
  console.error("Form ID not found in data-form-id attribute");
  throw new Error("Form ID not found");
}

const formId = Number(formIdAttr);
if (isNaN(formId) || formId <= 0) {
  console.error("Invalid form ID:", formIdAttr);
  throw new Error("Invalid form ID");
}

initializeFormBuilder(formId).catch((error) => {
  console.error("Failed to initialize form builder:", error);
  // Show error to user
  const errorDiv = document.createElement("div");
  errorDiv.className = "gf-error-message";
  errorDiv.textContent =
    "Failed to load form builder. Please refresh the page.";
  formSchemaBuilder.appendChild(errorDiv);
});

async function initializeFormBuilder(formId: number): Promise<void> {
  const formService = FormService.getInstance();

  try {
    const schema = await formService.getSchema(formId);
    const builder = await Formio.builder(
      formSchemaBuilder,
      schema,
      builderOptions,
    );
    setupBuilderEvents(builder, formId, formService);
  } catch (error) {
    console.error("Error initializing form builder:", error);
    throw error; // Re-throw to be handled by the caller
  }
}
