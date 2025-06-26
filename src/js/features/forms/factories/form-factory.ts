import { FormController } from "../controllers/form-controller";
import type { FormConfig } from "@/shared/types/form-types";

/**
 * Factory function to create a FormController
 * Replaces the old setupForm function with better error handling
 */
export function createFormController(config: FormConfig): FormController {
  return new FormController(config);
}

/**
 * Advanced factory with additional configuration options
 */
export function createAdvancedFormController(
  config: FormConfig,
  _options?: {
    autoReset?: boolean;
    customValidation?: (form: HTMLFormElement) => Promise<boolean>;
    onSuccess?: (response: any) => void;
    onError?: (error: unknown) => void;
  },
): FormController {
  // For now, just return the basic controller
  // This can be enhanced later with the advanced options
  return new FormController(config);
}

/**
 * Legacy function for backward compatibility
 * @deprecated Use createFormController instead
 */
export function setupForm(config: FormConfig): void {
  new FormController(config);
}
