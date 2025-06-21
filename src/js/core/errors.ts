export class FormBuilderError extends Error {
  constructor(
    message: string,
    public originalError?: any,
  ) {
    super(message);
    this.name = "FormBuilderError";
  }
}
