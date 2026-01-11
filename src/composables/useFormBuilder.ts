import { ref, onMounted, onUnmounted, type Ref } from "vue";
import { Formio } from "@formio/js";
import goforms from "@goformx/formio";
import { Logger } from "@/lib/core/logger";

// Register GoFormX templates
Formio.use(goforms);

export interface FormSchema {
  display?: string;
  components: unknown[];
}

export interface FormBuilderOptions {
  containerId: string;
  formId: string;
  schema?: FormSchema;
  onSchemaChange?: (schema: FormSchema) => void;
  onSave?: (schema: FormSchema) => Promise<void>;
}

export interface UseFormBuilderReturn {
  builder: Ref<unknown | null>;
  schema: Ref<FormSchema>;
  isLoading: Ref<boolean>;
  error: Ref<string | null>;
  isSaving: Ref<boolean>;
  saveSchema: () => Promise<void>;
  getSchema: () => FormSchema;
  setSchema: (newSchema: FormSchema) => void;
}

const defaultSchema: FormSchema = {
  display: "form",
  components: [],
};

/**
 * Composable for integrating Form.io builder into Vue components
 */
export function useFormBuilder(
  options: FormBuilderOptions,
): UseFormBuilderReturn {
  const builder = ref<unknown | null>(null);
  const schema = ref<FormSchema>(options.schema ?? { ...defaultSchema });
  const isLoading = ref(true);
  const error = ref<string | null>(null);
  const isSaving = ref(false);

  let builderInstance: { schema: FormSchema; destroy?: () => void } | null =
    null;

  async function initializeBuilder() {
    const container = document.getElementById(options.containerId);
    if (!container) {
      error.value = `Container element #${options.containerId} not found`;
      isLoading.value = false;
      return;
    }

    try {
      Logger.debug("Initializing Form.io builder...");

      // Fetch existing schema if editing
      if (options.formId && !options.schema) {
        const response = await fetch(`/api/v1/forms/${options.formId}/schema`);
        if (response.ok) {
          const data = await response.json();
          if (data.success && data.data) {
            schema.value = data.data;
          }
        }
      }

      // Create the builder
      builderInstance = await Formio.builder(container, schema.value, {
        builder: {
          basic: {
            default: true,
            components: {
              textfield: true,
              textarea: true,
              number: true,
              password: true,
              checkbox: true,
              selectboxes: true,
              select: true,
              radio: true,
              button: true,
              email: true,
              url: true,
              phoneNumber: true,
              datetime: true,
            },
          },
          advanced: {
            default: false,
          },
          layout: {
            default: false,
          },
          data: {
            default: false,
          },
          premium: false,
        },
        noDefaultSubmitButton: false,
        editForm: {
          textfield: [
            { key: "display", components: [] },
            { key: "data", components: [] },
            { key: "validation", components: [] },
            { key: "api", components: [] },
            { key: "conditional", components: [] },
            { key: "logic", components: [] },
          ],
        },
      });

      builder.value = builderInstance;

      // Listen for schema changes
      if (
        builderInstance &&
        typeof (
          builderInstance as {
            on?: (event: string, callback: (s: FormSchema) => void) => void;
          }
        ).on === "function"
      ) {
        (
          builderInstance as {
            on: (event: string, callback: (s: FormSchema) => void) => void;
          }
        ).on("change", (newSchema: FormSchema) => {
          schema.value = newSchema;
          options.onSchemaChange?.(newSchema);
        });
      }

      Logger.debug("Form.io builder initialized successfully");
    } catch (err) {
      Logger.error("Failed to initialize Form.io builder:", err);
      error.value = "Failed to initialize form builder";
    } finally {
      isLoading.value = false;
    }
  }

  function getSchema(): FormSchema {
    if (builderInstance) {
      return builderInstance.schema;
    }
    return schema.value;
  }

  function setSchema(newSchema: FormSchema) {
    schema.value = newSchema;
    // The builder will pick up changes from schema ref
  }

  async function saveSchema() {
    if (!options.formId) {
      error.value = "No form ID provided";
      return;
    }

    isSaving.value = true;
    error.value = null;

    try {
      const currentSchema = getSchema();

      const response = await fetch(`/api/v1/forms/${options.formId}/schema`, {
        method: "PUT",
        headers: {
          "Content-Type": "application/json",
          "X-Requested-With": "XMLHttpRequest",
        },
        body: JSON.stringify(currentSchema),
      });

      if (!response.ok) {
        throw new Error("Failed to save schema");
      }

      await options.onSave?.(currentSchema);
      Logger.debug("Schema saved successfully");
    } catch (err) {
      Logger.error("Failed to save schema:", err);
      error.value = "Failed to save form schema";
      throw err;
    } finally {
      isSaving.value = false;
    }
  }

  onMounted(() => {
    initializeBuilder();
  });

  onUnmounted(() => {
    if (builderInstance && typeof builderInstance.destroy === "function") {
      builderInstance.destroy();
    }
  });

  return {
    builder,
    schema,
    isLoading,
    error,
    isSaving,
    saveSchema,
    getSchema,
    setSchema,
  };
}

/**
 * Composable for rendering Form.io forms
 */
export function useFormRenderer(options: {
  containerId: string;
  schema: FormSchema;
  onSubmit?: (submission: unknown) => Promise<void>;
}) {
  const form = ref<unknown | null>(null);
  const isLoading = ref(true);
  const error = ref<string | null>(null);

  async function initializeForm() {
    const container = document.getElementById(options.containerId);
    if (!container) {
      error.value = `Container element #${options.containerId} not found`;
      isLoading.value = false;
      return;
    }

    try {
      const formInstance = await Formio.createForm(container, options.schema);
      form.value = formInstance;

      if (options.onSubmit) {
        formInstance.on("submit", options.onSubmit);
      }
    } catch (err) {
      Logger.error("Failed to initialize form:", err);
      error.value = "Failed to load form";
    } finally {
      isLoading.value = false;
    }
  }

  onMounted(() => {
    initializeForm();
  });

  return {
    form,
    isLoading,
    error,
  };
}
