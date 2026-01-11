import { ref, onMounted, onUnmounted, watch, type Ref } from "vue";
import { Formio } from "@formio/js";
import goforms from "@goformx/formio";
import { Logger } from "@/lib/core/logger";
import { useFormBuilderState, type FormComponent } from "./useFormBuilderState";

// Register GoFormX templates
Formio.use(goforms);

export interface FormSchema {
  display?: string;
  components: FormComponent[];
}

export interface FormBuilderOptions {
  containerId: string;
  formId: string;
  schema?: FormSchema;
  onSchemaChange?: (schema: FormSchema) => void;
  onSave?: (schema: FormSchema) => Promise<void>;
  autoSave?: boolean;
  autoSaveDelay?: number;
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
  // New methods
  selectedField: Ref<string | null>;
  selectField: (fieldKey: string | null) => void;
  duplicateField: (fieldKey: string) => void;
  deleteField: (fieldKey: string) => void;
  undo: () => void;
  redo: () => void;
  canUndo: Ref<boolean>;
  canRedo: Ref<boolean>;
  exportSchema: () => string;
  importSchema: (json: string) => void;
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

  // Initialize builder state with undo/redo
  const {
    selectedField,
    selectField,
    pushHistory,
    undo: undoHistory,
    redo: redoHistory,
    canUndo,
    canRedo,
    markDirty,
  } = useFormBuilderState(options.formId);

  let builderInstance: { schema: FormSchema; destroy?: () => void } | null =
    null;
  let autoSaveTimeout: ReturnType<typeof setTimeout> | null = null;

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

      // Create the builder with Form.io's built-in sidebar
      builderInstance = await Formio.builder(container, schema.value, {
        builder: {
          basic: {
            default: true,
            weight: 0,
            title: "Basic",
            components: {
              textfield: true,
              textarea: true,
              number: true,
              checkbox: true,
              select: true,
              radio: true,
              email: true,
              phoneNumber: true,
              datetime: true,
              button: true,
            },
          },
          layout: {
            default: false,
            weight: 10,
            title: "Layout",
            components: {
              panel: true,
              columns: true,
              fieldset: true,
            },
          },
          advanced: false,
          data: false,
          premium: false,
        },
        noDefaultSubmitButton: false,
        i18n: {
          en: {
            searchFields: "Search fields...",
            dragAndDropComponent: "Drag and drop fields here",
            basic: "Basic",
            advanced: "Advanced",
            layout: "Layout",
            data: "Data",
            premium: "Premium",
          },
        },
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
      const builderWithEvents = builderInstance as unknown as {
        on?: (event: string, callback: (s: FormSchema) => void) => void;
      };

      if (builderInstance && typeof builderWithEvents.on === "function") {
        builderWithEvents.on("change", (newSchema: FormSchema) => {
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
    void initializeBuilder();
  });

  onUnmounted(() => {
    if (builderInstance && typeof builderInstance.destroy === "function") {
      builderInstance.destroy();
    }
    // Clear auto-save timeout
    if (autoSaveTimeout) {
      clearTimeout(autoSaveTimeout);
    }
  });

  // Auto-save functionality
  if (options.autoSave) {
    watch(
      schema,
      () => {
        // Clear existing timeout
        if (autoSaveTimeout) {
          clearTimeout(autoSaveTimeout);
        }

        // Set new timeout
        const delay = options.autoSaveDelay ?? 2000;
        autoSaveTimeout = setTimeout(() => {
          void saveSchema();
        }, delay);
      },
      { deep: true },
    );
  }

  // Schema change handler with history tracking
  watch(
    schema,
    (newSchema) => {
      pushHistory(newSchema);
      markDirty();
      options.onSchemaChange?.(newSchema);
    },
    { deep: true },
  );

  /**
   * Undo last change
   */
  function undo() {
    const previousSchema = undoHistory();
    if (previousSchema) {
      setSchema(previousSchema);
    }
  }

  /**
   * Redo last undone change
   */
  function redo() {
    const nextSchema = redoHistory();
    if (nextSchema) {
      setSchema(nextSchema);
    }
  }

  /**
   * Find a component by key in the schema
   */
  function findComponent(
    components: unknown[],
    key: string,
  ): FormComponent | null {
    for (const component of components) {
      const comp = component as FormComponent;
      if (comp.key === key) {
        return comp;
      }
      // Recursively search in nested components
      if (comp["components"]) {
        const found = findComponent(comp["components"] as unknown[], key);
        if (found) return found;
      }
    }
    return null;
  }

  /**
   * Duplicate a field by key
   */
  function duplicateField(fieldKey: string) {
    const currentSchema = getSchema();
    const component = findComponent(currentSchema.components, fieldKey);

    if (!component) {
      Logger.warn(`Component with key "${fieldKey}" not found`);
      return;
    }

    // Create a duplicate with a new key
    const duplicate = JSON.parse(JSON.stringify(component)) as FormComponent;
    duplicate.key = `${component.key}_copy`;
    duplicate.label = `${component.label ?? component.type} (Copy)`;

    // Add to schema
    currentSchema.components.push(duplicate);
    setSchema(currentSchema);

    Logger.debug(`Duplicated component: ${fieldKey}`);
  }

  /**
   * Delete a field by key
   */
  function deleteField(fieldKey: string) {
    const currentSchema = getSchema();

    // Remove component from array
    const filterComponents = (components: FormComponent[]): FormComponent[] => {
      return components.filter((comp) => {
        if (comp.key === fieldKey) return false;

        // Recursively filter nested components
        if (comp["components"]) {
          comp["components"] = filterComponents(
            comp["components"] as FormComponent[],
          );
        }
        return true;
      });
    };

    currentSchema.components = filterComponents(currentSchema.components);
    setSchema(currentSchema);

    Logger.debug(`Deleted component: ${fieldKey}`);
  }

  /**
   * Export schema as JSON string
   */
  function exportSchema(): string {
    const currentSchema = getSchema();
    return JSON.stringify(currentSchema, null, 2);
  }

  /**
   * Import schema from JSON string
   */
  function importSchema(json: string) {
    try {
      const imported = JSON.parse(json) as FormSchema;
      setSchema(imported);
      Logger.debug("Schema imported successfully");
    } catch (err) {
      Logger.error("Failed to import schema:", err);
      error.value = "Invalid schema JSON";
    }
  }

  return {
    builder,
    schema,
    isLoading,
    error,
    isSaving,
    saveSchema,
    getSchema,
    setSchema,
    // New methods
    selectedField,
    selectField,
    duplicateField,
    deleteField,
    undo,
    redo,
    canUndo,
    canRedo,
    exportSchema,
    importSchema,
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
    void initializeForm();
  });

  return {
    form,
    isLoading,
    error,
  };
}
