/**
 * Form builder error handling
 */
export class FormBuilderError extends Error {
  constructor(
    message: string,
    public readonly userMessage: string,
  ) {
    super(message);
    this.name = "FormBuilderError";
  }
}
