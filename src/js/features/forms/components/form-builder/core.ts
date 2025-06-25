import { Logger } from "@/core/logger";
import { Formio } from "@formio/js";
import { FormService } from "@/features/forms/services/form-service";
import type { FormSchema } from "@/features/forms/services/form-service";
import { builderOptions } from "@/core/config/builder-config";
import { FormBuilderError, ErrorCode } from "@/core/errors/form-builder-error";
import { dom } from "@/shared/utils/dom-utils";
import { formState } from "@/features/forms/state/form-state";

/**
 * Form builder validation
 */
export function validateFormBuilder(): {
  builder: HTMLElement;
  formId: string;
} {
  const builder = dom.getElement<HTMLElement>("form-schema-builder");
  if (!builder) {
    throw new FormBuilderError(
      "Form builder element not found",
      ErrorCode.FORM_NOT_FOUND,
      "Form builder element not found. Please refresh the page.",
    );
  }

  const formId = builder.getAttribute("data-form-id");
  if (!formId) {
    throw new FormBuilderError(
      "Form ID not found",
      ErrorCode.FORM_NOT_FOUND,
      "Form ID not found. Please refresh the page.",
    );
  }

  return { builder, formId };
}

/**
 * Schema management
 */
export async function getFormSchema(formId: string): Promise<FormSchema> {
  // For new form creation, return a default schema with submit button
  if (formId === "new") {
    return {
      display: "form",
      components: [
        {
          type: "button",
          key: "submit",
          label: "Submit",
          input: true,
          required: false,
        },
      ],
    };
  }

  const formService = FormService.getInstance();
  try {
    const schema = await formService.getSchema(formId);
    Logger.info("Schema fetched successfully:", schema);
    return schema;
  } catch (error) {
    Logger.error("Error in getFormSchema:", error);
    Logger.error("Error type:", typeof error);
    Logger.error(
      "Error message:",
      error instanceof Error ? error.message : String(error),
    );

    throw new FormBuilderError(
      "Failed to fetch schema",
      ErrorCode.LOAD_FAILED,
      "Failed to load form schema. Please try again later.",
    );
  }
}

/**
 * Builder initialization
 */
export async function createFormBuilder(
  container: HTMLElement,
  schema: FormSchema,
): Promise<any> {
  try {
    Logger.debug("createFormBuilder: Starting initialization");
    Logger.debug("createFormBuilder: Container element:", container);
    Logger.debug(
      "createFormBuilder: Container innerHTML:",
      container.innerHTML,
    );
    Logger.debug("createFormBuilder: Container dimensions:", {
      offsetWidth: container.offsetWidth,
      offsetHeight: container.offsetHeight,
      clientWidth: container.clientWidth,
      clientHeight: container.clientHeight,
    });

    // Ensure schema has required properties
    const formSchema = {
      ...schema,
      display: "form",
      components: schema.components || [],
    };

    Logger.debug("createFormBuilder: Form schema:", formSchema);
    Logger.debug("createFormBuilder: Builder options:", builderOptions);

    // Create builder with standalone configuration
    const builder = await Formio.builder(container, formSchema, {
      ...builderOptions,
      // Standalone mode - no server communication
      noDefaultSubmitButton: false,
      showSchema: true,
      showJSONEditor: true,
      showPreview: true,
      // Disable all project-related features
      projectUrl: null,
      appUrl: null,
      apiUrl: null,
      // Disable all server communication
      noAlerts: true,
      readOnly: false,
      // Prevent project settings requests
      project: null,
      settings: null,
    });

    Logger.debug("createFormBuilder: Form.io builder created:", builder);
    Logger.debug(
      "createFormBuilder: Container after builder:",
      container.innerHTML,
    );

    // Store builder instance in state management instead of global window
    formState.set("formBuilder", builder);
    return builder;
  } catch (error) {
    Logger.error("Form builder initialization error:", error);
    throw new FormBuilderError(
      "Failed to initialize builder",
      ErrorCode.SCHEMA_ERROR,
      "Failed to initialize form builder. Please refresh the page.",
    );
  }
}
