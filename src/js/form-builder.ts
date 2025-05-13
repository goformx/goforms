console.log("form-builder.ts");

console.log("before import: goforms-template.");

import { Formio } from "@formio/js";
import { validation } from "./validation";
import goforms from "goforms-template";
// import "./semantic-test"; // Import the semantic test

// Import Form.io styles
import "@formio/js/dist/formio.full.min.css";

console.log("goforms before use", goforms);

// Register the goforms template
Formio.use(goforms);

console.log("goforms after use", goforms);

// Define builder options
const builderOptions = {
  display: "form",
  builder: {
    basic: false,
    advanced: false,
    data: false,
    customBasic: {
      title: "Basic Components",
      default: true,
      weight: 0,
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
  language: "en",
  i18n: {
    en: {
      "Basic Components": "Basic Components",
    },
  },
  noDefaultSubmitButton: true,
  templates: goforms.templates,
  framework: "goforms",
};

export interface FormBuilderOptions {
  disabled?: string[];
  noNewEdit?: boolean;
  noDefaultSubmitButton?: boolean;
  alwaysConfirmComponentRemoval?: boolean;
  formConfig?: Record<string, unknown>;
  resourceTag?: string;
  editForm?: Record<string, unknown>;
  language?: string;
  builder?: {
    basic?: Record<string, unknown>;
    advanced?: Record<string, unknown>;
    layout?: Record<string, unknown>;
    data?: boolean;
    premium?: boolean;
    resource?: Record<string, unknown>;
  };
  display?: "form" | "wizard" | "pdf";
  resourceFilter?: string;
  noSource?: boolean;
  showFullJsonSchema?: boolean;
  framework: string;
  templates?: Record<string, unknown>;
}

interface FormioBuilder {
  schema: Record<string, unknown>;
  setForm: (schema: Record<string, unknown>) => void;
  form?: {
    options?: {
      framework?: string;
    };
    templates?: Record<string, unknown>;
  };
}

interface FormioComponent {
  builderInfo?: {
    group: string;
    [key: string]: unknown;
  };
}

export class FormBuilder {
  private container: HTMLElement;
  private builder!: FormioBuilder;
  private formId: number;
  private currentSchema: Record<string, unknown> = {
    display: "form",
    components: [],
  };

  constructor(containerId: string, formId: number) {
    console.log("FormBuilder: constructor called with formId:", formId);
    const container = document.getElementById(containerId);
    if (!container) throw new Error(`Container ${containerId} not found`);
    this.container = container;
    // Add a custom class to verify template usage
    this.container.classList.add("goforms-template-active");
    this.formId = formId;
    this.init();
  }

  private init() {
    // Debug: Log builder options
    console.log("Builder options:", builderOptions);
    console.log("Registered templates:", goforms.templates);

    // Add a test component to verify template usage
    const testSchema = {
      display: "form",
      components: [
        {
          type: "textfield",
          label: "Test Field",
          key: "testField",
          inputType: "text",
          placeholder: "Enter text to test template",
        },
        {
          type: "button",
          label: "Test Button",
          key: "testButton",
          theme: "primary",
          leftIcon: "check",
          tooltip: "This is a test button",
        },
      ],
    };

    // Initialize the builder with our framework and templates
    Formio.builder(this.container, testSchema, builderOptions).then(
      (builder: FormioBuilder) => {
        this.builder = builder;
        // Debug: Log builder instance
        console.log("Builder instance:", {
          framework: (builder as any).form?.options?.framework,
          templates: Object.keys((builder as any).form?.templates || {}),
          components: (builder as any).form?.components,
        });
        this.loadExistingSchema();
      },
    );
  }

  private async loadExistingSchema() {
    try {
      if (this.formId === 0) {
        return;
      }
      console.log("Loading form schema for form ID:", this.formId);
      const response = await validation.fetchWithCSRF(
        `/dashboard/forms/${this.formId}/schema`,
        {
          method: "GET",
          headers: {
            "Content-Type": "application/json",
          },
        },
      );
      if (response.ok) {
        const schema = await response.json();
        console.log("Loaded form schema:", schema);
        this.builder.setForm(schema);
        this.currentSchema = schema;
      } else {
        if (response.status === 401) {
          console.error("Not authenticated, redirecting to login");
          window.location.href = "/login";
        } else {
          console.error(
            "Failed to load form schema:",
            response.status,
            response.statusText,
          );
        }
      }
    } catch (error) {
      console.error("Failed to load form schema:", error);
    }
  }

  public async saveSchema(): Promise<boolean> {
    try {
      const formioSchema = this.builder.schema;
      const response = await validation.fetchWithCSRF(
        `/dashboard/forms/${this.formId}/schema`,
        {
          method: "PUT",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify(formioSchema),
        },
      );
      if (response.ok) {
        console.log("Schema saved successfully");
        this.currentSchema = formioSchema;
        return true;
      } else {
        throw new Error("Failed to save schema");
      }
    } catch (error) {
      console.error("Failed to save form schema:", error);
      return false;
    }
  }
}

// Initialize form builder when the module is loaded
const formSchemaBuilder = document.getElementById("form-schema-builder");
if (formSchemaBuilder) {
  const formIdAttr = formSchemaBuilder.getAttribute("data-form-id");
  if (formIdAttr) {
    const formId = parseInt(formIdAttr, 10);
    if (!isNaN(formId)) {
      (window as { formBuilderInstance?: FormBuilder }).formBuilderInstance =
        new FormBuilder("form-schema-builder", formId);
    } else {
      console.error("FormBuilder: Invalid form ID:", formIdAttr);
    }
  }
}

// Type assertion for Formio.Components.components
const components = Object.values(
  Formio.Components.components,
) as FormioComponent[];

console.log(
  "Basic components in Form.io:",
  components
    .filter((c) => c.builderInfo?.group === "basic")
    .map((c) => c.builderInfo),
);
