/**
 * Form state management to replace global window assignments
 */
export class FormState {
  private static instance: FormState;
  private state = new Map<string, any>();

  private constructor() {}

  public static getInstance(): FormState {
    if (!FormState.instance) {
      FormState.instance = new FormState();
    }
    return FormState.instance;
  }

  set(key: string, value: any): void {
    this.state.set(key, value);
  }

  get<T>(key: string): T | undefined {
    return this.state.get(key) as T | undefined;
  }

  has(key: string): boolean {
    return this.state.has(key);
  }

  delete(key: string): boolean {
    return this.state.delete(key);
  }

  clear(): void {
    this.state.clear();
  }
}

// Export a singleton instance for easy access
export const formState = FormState.getInstance();
