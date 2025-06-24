import { Formio } from "@formio/js";
import { FormService } from "@/features/forms/services/form-service";
import type { FormSchema } from "@/features/forms/services/form-service";
import { builderOptions } from "@/core/config/builder-config";
import { FormBuilderError } from "@/core/errors/form-builder-error";
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
      "Form builder element not found. Please refresh the page.",
    );
  }

  const formId = builder.getAttribute("data-form-id");
  if (!formId) {
    throw new FormBuilderError(
      "Form ID not found",
      "Form ID not found. Please refresh the page.",
    );
  }

  return { builder, formId };
}

/**
 * Schema management
 */
export async function getFormSchema(formId: string): Promise<FormSchema> {
  // For new form creation, return a default schema
  if (formId === "new") {
    return {
      display: "form",
      components: [],
    };
  }

  const formService = FormService.getInstance();
  try {
    return await formService.getSchema(formId);
  } catch {
    throw new FormBuilderError(
      "Failed to fetch schema",
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
    // Configure Formio for standalone mode (no server required)
    Formio.setProjectUrl(null);
    Formio.setBaseUrl(null);

    // Ensure schema has required properties
    const formSchema = {
      ...schema,
      display: "form",
      components: schema.components || [],
    };

    // Create builder with options
    const builder = await Formio.builder(container, formSchema, {
      ...builderOptions,
      noDefaultSubmitButton: true,
      builder: {
        ...builderOptions.builder,
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
      },
    });

    // Store builder instance in state management instead of global window
    formState.set("formBuilder", builder);
    return builder;
  } catch (error) {
    console.error("Form builder initialization error:", error);
    throw new FormBuilderError(
      "Failed to initialize builder",
      "Failed to initialize form builder. Please refresh the page.",
    );
  }
}
