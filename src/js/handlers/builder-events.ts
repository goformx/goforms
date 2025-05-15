import { FormBuilder } from "@formio/js";
import { FormService, FormSchema } from "../services/form-service";

interface FormBuilderWithSchema extends FormBuilder {
  form: FormSchema;
  saveSchema: () => Promise<boolean>;
}

export const setupBuilderEvents = (
  builder: FormBuilder,
  formId: number,
  formService: FormService,
): void => {
  const typedBuilder = builder as FormBuilderWithSchema;

  // Add saveSchema method to the builder instance
  typedBuilder.saveSchema = async () => {
    try {
      await formService.saveSchema(formId, typedBuilder.form);
      return true;
    } catch (error) {
      console.error("Error saving schema:", error);
      return false;
    }
  };

  // Add component save event listener
  builder.on("saveComponent", () => {
    console.log("Component saved:", typedBuilder.form);
  });

  // Store builder instance globally
  (
    window as unknown as { formBuilderInstance?: FormBuilderWithSchema }
  ).formBuilderInstance = typedBuilder;
};
