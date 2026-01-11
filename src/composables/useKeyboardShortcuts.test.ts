import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import {
  formatShortcut,
  useKeyboardShortcuts,
  type ShortcutConfig,
} from "./useKeyboardShortcuts";

// Helper to set platform
function setPlatform(platform: string) {
  Object.defineProperty(navigator, "platform", {
    value: platform,
    writable: true,
    configurable: true,
  });
}

describe("formatShortcut", () => {
  describe("on Mac", () => {
    beforeEach(() => {
      setPlatform("MacIntel");
    });

    it("formats meta+key shortcut with Mac symbols", () => {
      const shortcut: ShortcutConfig = {
        key: "s",
        meta: true,
        handler: () => {},
        description: "Save",
      };

      expect(formatShortcut(shortcut)).toBe("⌘S");
    });

    it("formats ctrl+key shortcut", () => {
      const shortcut: ShortcutConfig = {
        key: "c",
        ctrl: true,
        handler: () => {},
        description: "Copy",
      };

      expect(formatShortcut(shortcut)).toBe("CtrlC");
    });

    it("formats shift+key shortcut with Mac symbol", () => {
      const shortcut: ShortcutConfig = {
        key: "z",
        shift: true,
        handler: () => {},
        description: "Redo",
      };

      expect(formatShortcut(shortcut)).toBe("⇧Z");
    });

    it("formats alt+key shortcut with Mac symbol", () => {
      const shortcut: ShortcutConfig = {
        key: "t",
        alt: true,
        handler: () => {},
        description: "Toggle",
      };

      expect(formatShortcut(shortcut)).toBe("⌥T");
    });

    it("formats combined modifiers", () => {
      const shortcut: ShortcutConfig = {
        key: "z",
        meta: true,
        shift: true,
        handler: () => {},
        description: "Redo",
      };

      expect(formatShortcut(shortcut)).toBe("⇧⌘Z");
    });

    it("formats all modifiers together", () => {
      const shortcut: ShortcutConfig = {
        key: "k",
        ctrl: true,
        alt: true,
        shift: true,
        meta: true,
        handler: () => {},
        description: "Complex",
      };

      expect(formatShortcut(shortcut)).toBe("Ctrl⌥⇧⌘K");
    });
  });

  describe("on Windows/Linux", () => {
    beforeEach(() => {
      setPlatform("Win32");
    });

    it("formats meta+key shortcut as Ctrl", () => {
      const shortcut: ShortcutConfig = {
        key: "s",
        meta: true,
        handler: () => {},
        description: "Save",
      };

      expect(formatShortcut(shortcut)).toBe("Ctrl+S");
    });

    it("formats ctrl+key shortcut", () => {
      const shortcut: ShortcutConfig = {
        key: "c",
        ctrl: true,
        handler: () => {},
        description: "Copy",
      };

      expect(formatShortcut(shortcut)).toBe("Ctrl+C");
    });

    it("formats shift+key shortcut", () => {
      const shortcut: ShortcutConfig = {
        key: "z",
        shift: true,
        handler: () => {},
        description: "Redo",
      };

      expect(formatShortcut(shortcut)).toBe("Shift+Z");
    });

    it("formats alt+key shortcut", () => {
      const shortcut: ShortcutConfig = {
        key: "t",
        alt: true,
        handler: () => {},
        description: "Toggle",
      };

      expect(formatShortcut(shortcut)).toBe("Alt+T");
    });

    it("formats combined modifiers with plus signs", () => {
      const shortcut: ShortcutConfig = {
        key: "z",
        meta: true,
        shift: true,
        handler: () => {},
        description: "Redo",
      };

      expect(formatShortcut(shortcut)).toBe("Shift+Ctrl+Z");
    });
  });

  describe("key formatting", () => {
    beforeEach(() => {
      setPlatform("Win32");
    });

    it("uppercases single letter keys", () => {
      const shortcut: ShortcutConfig = {
        key: "a",
        meta: true,
        handler: () => {},
        description: "Select All",
      };

      expect(formatShortcut(shortcut)).toContain("A");
    });

    it("preserves special key names", () => {
      const shortcut: ShortcutConfig = {
        key: "Enter",
        meta: true,
        handler: () => {},
        description: "Submit",
      };

      expect(formatShortcut(shortcut)).toContain("ENTER");
    });
  });
});

describe("useKeyboardShortcuts", () => {
  let addEventListenerSpy: ReturnType<typeof vi.spyOn>;
  let removeEventListenerSpy: ReturnType<typeof vi.spyOn>;

  beforeEach(() => {
    addEventListenerSpy = vi.spyOn(window, "addEventListener");
    removeEventListenerSpy = vi.spyOn(window, "removeEventListener");
  });

  afterEach(() => {
    addEventListenerSpy.mockRestore();
    removeEventListenerSpy.mockRestore();
  });

  it("returns shortcuts ref", () => {
    const shortcuts: ShortcutConfig[] = [
      { key: "s", meta: true, handler: () => {}, description: "Save" },
    ];

    const result = useKeyboardShortcuts(shortcuts);

    expect(result.shortcuts.value).toHaveLength(1);
    expect(result.shortcuts.value[0].key).toBe("s");
  });

  it("returns isEnabled ref defaulting to true", () => {
    const result = useKeyboardShortcuts([]);

    expect(result.isEnabled.value).toBe(true);
  });

  it("enable() sets isEnabled to true", () => {
    const result = useKeyboardShortcuts([]);

    result.disable();
    expect(result.isEnabled.value).toBe(false);

    result.enable();
    expect(result.isEnabled.value).toBe(true);
  });

  it("disable() sets isEnabled to false", () => {
    const result = useKeyboardShortcuts([]);

    result.disable();
    expect(result.isEnabled.value).toBe(false);
  });

  it("toggle() toggles isEnabled", () => {
    const result = useKeyboardShortcuts([]);

    expect(result.isEnabled.value).toBe(true);
    result.toggle();
    expect(result.isEnabled.value).toBe(false);
    result.toggle();
    expect(result.isEnabled.value).toBe(true);
  });
});
