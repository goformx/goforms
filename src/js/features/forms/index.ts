// ===== src/js/forms/index.ts =====
// Main entry point for form handling
export { setupForm } from "./handlers/form-handler";
export { EnhancedFormHandler } from "./handlers/enhanced-form-handler";
export type { FormConfig, ServerResponse } from "../forms/types/form-types";

// Re-export utilities if needed externally
export { debounce } from "./utils/debounce";
export { isAuthenticationEndpoint } from "./utils/endpoint-utils";
