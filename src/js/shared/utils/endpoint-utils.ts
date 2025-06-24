// ===== src/js/forms/utils/endpoint-utils.ts =====
/**
 * Checks if the endpoint is an authentication endpoint
 */
export function isAuthenticationEndpoint(action: string): boolean {
  return action.includes("/login") || action.includes("/signup");
}
