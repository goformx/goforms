import { onMounted, onUnmounted, type Ref, ref } from "vue";

export interface ShortcutConfig {
  key: string;
  ctrl?: boolean;
  meta?: boolean;
  shift?: boolean;
  alt?: boolean;
  handler: () => void;
  description: string;
  preventDefault?: boolean;
}

export interface UseKeyboardShortcutsOptions {
  enabled?: Ref<boolean>;
}

export interface UseKeyboardShortcutsReturn {
  shortcuts: Ref<ShortcutConfig[]>;
  isEnabled: Ref<boolean>;
  enable: () => void;
  disable: () => void;
  toggle: () => void;
}

/**
 * Composable for managing keyboard shortcuts with platform detection
 *
 * @param shortcuts Array of shortcut configurations
 * @param options Optional configuration options
 * @returns Keyboard shortcuts management interface
 *
 * @example
 * ```ts
 * const { shortcuts, enable, disable } = useKeyboardShortcuts([
 *   {
 *     key: 's',
 *     ctrl: true,
 *     handler: () => saveForm(),
 *     description: 'Save form'
 *   },
 *   {
 *     key: 'k',
 *     meta: true,
 *     handler: () => openCommandPalette(),
 *     description: 'Open command palette'
 *   }
 * ]);
 * ```
 */
export function useKeyboardShortcuts(
  shortcuts: ShortcutConfig[],
  options?: UseKeyboardShortcutsOptions,
): UseKeyboardShortcutsReturn {
  const isEnabled = options?.enabled ?? ref(true);
  const isMac = ref(false);

  // Detect platform
  onMounted(() => {
    isMac.value = navigator.platform.toLowerCase().includes("mac");
  });

  /**
   * Check if keyboard event matches a shortcut configuration
   */
  const matchesShortcut = (
    event: KeyboardEvent,
    shortcut: ShortcutConfig,
  ): boolean => {
    // Check key match (case-insensitive)
    const keyMatches = event.key.toLowerCase() === shortcut.key.toLowerCase();

    // Check modifiers
    const ctrlMatches = shortcut.ctrl ? event.ctrlKey : !event.ctrlKey;
    const metaMatches = shortcut.meta ? event.metaKey : !event.metaKey;
    const shiftMatches = shortcut.shift ? event.shiftKey : !event.shiftKey;
    const altMatches = shortcut.alt ? event.altKey : !event.altKey;

    return (
      keyMatches && ctrlMatches && metaMatches && shiftMatches && altMatches
    );
  };

  /**
   * Handle keyboard events and trigger matching shortcuts
   */
  const handleKeyDown = (event: KeyboardEvent): void => {
    if (!isEnabled.value) return;

    // Find matching shortcut
    const matchedShortcut = shortcuts.find((shortcut) =>
      matchesShortcut(event, shortcut),
    );

    if (matchedShortcut) {
      // Prevent default browser behavior if configured
      if (matchedShortcut.preventDefault !== false) {
        event.preventDefault();
      }

      // Execute handler
      matchedShortcut.handler();
    }
  };

  // Register event listener
  onMounted(() => {
    window.addEventListener("keydown", handleKeyDown);
  });

  // Cleanup on unmount
  onUnmounted(() => {
    window.removeEventListener("keydown", handleKeyDown);
  });

  /**
   * Enable keyboard shortcuts
   */
  const enable = (): void => {
    if (typeof isEnabled.value === "boolean") {
      isEnabled.value = true;
    }
  };

  /**
   * Disable keyboard shortcuts
   */
  const disable = (): void => {
    if (typeof isEnabled.value === "boolean") {
      isEnabled.value = false;
    }
  };

  /**
   * Toggle keyboard shortcuts on/off
   */
  const toggle = (): void => {
    if (typeof isEnabled.value === "boolean") {
      isEnabled.value = !isEnabled.value;
    }
  };

  return {
    shortcuts: ref(shortcuts),
    isEnabled,
    enable,
    disable,
    toggle,
  };
}

/**
 * Format shortcut display text based on platform
 *
 * @param shortcut Shortcut configuration
 * @returns Formatted shortcut text (e.g., "Cmd+S" on Mac, "Ctrl+S" on Windows)
 *
 * @example
 * ```ts
 * const shortcut = { key: 's', meta: true };
 * console.log(formatShortcut(shortcut)); // "Cmd+S" on Mac, "Ctrl+S" on Windows
 * ```
 */
export function formatShortcut(shortcut: ShortcutConfig): string {
  const isMac = navigator.platform.toLowerCase().includes("mac");
  const parts: string[] = [];

  // Add modifiers
  if (shortcut.ctrl) parts.push(isMac ? "Ctrl" : "Ctrl");
  if (shortcut.alt) parts.push(isMac ? "⌥" : "Alt");
  if (shortcut.shift) parts.push(isMac ? "⇧" : "Shift");
  if (shortcut.meta) parts.push(isMac ? "⌘" : "Ctrl");

  // Add key
  parts.push(shortcut.key.toUpperCase());

  return parts.join(isMac ? "" : "+");
}
