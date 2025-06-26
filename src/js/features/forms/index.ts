// Main entry point for form handling

// Export services
export * from "./services";

// Export new controller and factories
export { FormController } from "./controllers/form-controller";
export {
  createFormController,
  setupForm,
  createAdvancedFormController,
} from "./factories/form-factory";

// Export handlers (legacy - will be deprecated)
export { setupForm as setupFormLegacy } from "./handlers/form-handler";
export { EnhancedFormHandler } from "./handlers/enhanced-form-handler";

// Export types
export type { FormConfig, ServerResponse } from "@/shared/types";

// Re-export utilities if needed externally
export { debounce } from "@/shared/utils/debounce";
export { isAuthenticationEndpoint } from "@/shared/utils/endpoint-utils";
