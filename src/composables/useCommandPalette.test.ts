import { describe, it, expect, beforeEach, vi } from "vitest";
import { useCommandPalette, type Command } from "./useCommandPalette";

// Mock localStorage
const localStorageMock = (() => {
  let store: Record<string, string> = {};
  return {
    getItem: vi.fn((key: string) => store[key] ?? null),
    setItem: vi.fn((key: string, value: string) => {
      store[key] = value;
    }),
    removeItem: vi.fn((key: string) => {
      delete store[key];
    }),
    clear: vi.fn(() => {
      store = {};
    }),
  };
})();

Object.defineProperty(window, "localStorage", { value: localStorageMock });

describe("useCommandPalette", () => {
  const mockHandler = vi.fn();

  const testCommands: Command[] = [
    {
      id: "create-form",
      label: "Create New Form",
      icon: "Plus",
      keywords: ["new", "add"],
      handler: mockHandler,
      section: "Actions",
    },
    {
      id: "dashboard",
      label: "Go to Dashboard",
      icon: "Home",
      keywords: ["home", "main"],
      handler: mockHandler,
      section: "Navigation",
    },
    {
      id: "settings",
      label: "Open Settings",
      icon: "Cog",
      keywords: ["preferences", "config"],
      handler: mockHandler,
      section: "Navigation",
    },
    {
      id: "logout",
      label: "Sign Out",
      icon: "LogOut",
      keywords: ["exit", "leave"],
      handler: mockHandler,
      section: "Account",
    },
  ];

  beforeEach(() => {
    vi.clearAllMocks();
    localStorageMock.clear();
  });

  describe("open/close/toggle", () => {
    it("opens the command palette", () => {
      const { isOpen, open } = useCommandPalette(testCommands);

      expect(isOpen.value).toBe(false);
      open();
      expect(isOpen.value).toBe(true);
    });

    it("closes the command palette", () => {
      const { isOpen, open, close } = useCommandPalette(testCommands);

      open();
      expect(isOpen.value).toBe(true);
      close();
      expect(isOpen.value).toBe(false);
    });

    it("toggles the command palette", () => {
      const { isOpen, toggle } = useCommandPalette(testCommands);

      expect(isOpen.value).toBe(false);
      toggle();
      expect(isOpen.value).toBe(true);
      toggle();
      expect(isOpen.value).toBe(false);
    });

    it("clears query when opening", () => {
      const { query, open } = useCommandPalette(testCommands);

      query.value = "test";
      open();
      expect(query.value).toBe("");
    });

    it("clears query when closing", () => {
      const { query, open, close } = useCommandPalette(testCommands);

      open();
      query.value = "test";
      close();
      expect(query.value).toBe("");
    });
  });

  describe("filteredCommands", () => {
    it("returns default commands when no query", () => {
      const { filteredCommands } = useCommandPalette(testCommands);

      expect(filteredCommands.value.length).toBeGreaterThan(0);
      expect(filteredCommands.value.length).toBeLessThanOrEqual(10);
    });

    it("filters by exact label match", () => {
      const { query, filteredCommands } = useCommandPalette(testCommands);

      query.value = "Create New Form";
      expect(filteredCommands.value[0].id).toBe("create-form");
    });

    it("filters by partial label match", () => {
      const { query, filteredCommands } = useCommandPalette(testCommands);

      query.value = "Dashboard";
      expect(filteredCommands.value[0].id).toBe("dashboard");
    });

    it("filters by keyword match", () => {
      const { query, filteredCommands } = useCommandPalette(testCommands);

      query.value = "home";
      expect(filteredCommands.value.some((cmd) => cmd.id === "dashboard")).toBe(
        true,
      );
    });

    it("returns empty array when no matches", () => {
      const { query, filteredCommands } = useCommandPalette(testCommands);

      query.value = "xyznonexistent";
      expect(filteredCommands.value.length).toBe(0);
    });

    it("limits results to 10", () => {
      const manyCommands: Command[] = Array.from({ length: 20 }, (_, i) => ({
        id: `cmd-${i}`,
        label: `Command ${i}`,
        handler: mockHandler,
      }));

      const { query, filteredCommands } = useCommandPalette(manyCommands);

      query.value = "Command";
      expect(filteredCommands.value.length).toBeLessThanOrEqual(10);
    });

    it("prioritizes exact matches over partial matches", () => {
      const { query, filteredCommands } = useCommandPalette(testCommands);

      query.value = "settings";
      // "Settings" should be ranked higher than commands that just contain "settings"
      expect(filteredCommands.value[0].label.toLowerCase()).toContain(
        "settings",
      );
    });

    it("is case insensitive", () => {
      const { query, filteredCommands } = useCommandPalette(testCommands);

      query.value = "DASHBOARD";
      expect(filteredCommands.value.some((cmd) => cmd.id === "dashboard")).toBe(
        true,
      );
    });
  });

  describe("executeCommand", () => {
    it("executes the command handler", () => {
      const handler = vi.fn();
      const commands: Command[] = [{ id: "test", label: "Test", handler }];
      const { executeCommand } = useCommandPalette(commands);

      executeCommand("test");

      expect(handler).toHaveBeenCalledOnce();
    });

    it("closes palette after execution", () => {
      const commands: Command[] = [
        { id: "test", label: "Test", handler: mockHandler },
      ];
      const { isOpen, open, executeCommand } = useCommandPalette(commands);

      open();
      executeCommand("test");

      expect(isOpen.value).toBe(false);
    });

    it("adds command to recent commands", () => {
      const commands: Command[] = [
        { id: "test", label: "Test", handler: mockHandler },
      ];
      const { executeCommand, recentCommands } = useCommandPalette(commands);

      executeCommand("test");

      expect(recentCommands.value[0].id).toBe("test");
    });

    it("moves executed command to top of recent", () => {
      const commands: Command[] = [
        { id: "cmd1", label: "Command 1", handler: mockHandler },
        { id: "cmd2", label: "Command 2", handler: mockHandler },
      ];
      const { executeCommand, recentCommands } = useCommandPalette(commands);

      executeCommand("cmd1");
      executeCommand("cmd2");
      executeCommand("cmd1"); // Execute cmd1 again

      expect(recentCommands.value[0].id).toBe("cmd1");
      expect(recentCommands.value[1].id).toBe("cmd2");
    });

    it("limits recent commands to 10", () => {
      const commands: Command[] = Array.from({ length: 15 }, (_, i) => ({
        id: `cmd-${i}`,
        label: `Command ${i}`,
        handler: mockHandler,
      }));
      const { executeCommand, recentCommands } = useCommandPalette(commands);

      // Execute all 15 commands
      for (let i = 0; i < 15; i++) {
        executeCommand(`cmd-${i}`);
      }

      expect(recentCommands.value.length).toBeLessThanOrEqual(10);
    });

    it("saves recent commands to localStorage", () => {
      const commands: Command[] = [
        { id: "test", label: "Test", handler: mockHandler },
      ];
      const { executeCommand } = useCommandPalette(commands);

      executeCommand("test");

      expect(localStorageMock.setItem).toHaveBeenCalledWith(
        "command-palette-recent",
        JSON.stringify(["test"]),
      );
    });

    it("handles non-existent command gracefully", () => {
      const consoleWarn = vi
        .spyOn(console, "warn")
        .mockImplementation(() => {});
      const { executeCommand } = useCommandPalette(testCommands);

      expect(() => executeCommand("non-existent")).not.toThrow();
      expect(consoleWarn).toHaveBeenCalledWith(
        'Command "non-existent" not found',
      );

      consoleWarn.mockRestore();
    });
  });

  describe("clearRecent", () => {
    it("clears recent commands", () => {
      const commands: Command[] = [
        { id: "test", label: "Test", handler: mockHandler },
      ];
      const { executeCommand, clearRecent, recentCommands } =
        useCommandPalette(commands);

      executeCommand("test");
      expect(recentCommands.value.length).toBe(1);

      clearRecent();
      expect(recentCommands.value.length).toBe(0);
    });

    it("removes recent commands from localStorage", () => {
      const commands: Command[] = [
        { id: "test", label: "Test", handler: mockHandler },
      ];
      const { executeCommand, clearRecent } = useCommandPalette(commands);

      executeCommand("test");
      clearRecent();

      expect(localStorageMock.removeItem).toHaveBeenCalledWith(
        "command-palette-recent",
      );
    });
  });

  describe("localStorage persistence", () => {
    it("loads recent commands from localStorage on init", () => {
      localStorageMock.getItem.mockReturnValueOnce(
        JSON.stringify(["dashboard", "settings"]),
      );

      const { recentCommands } = useCommandPalette(testCommands);

      expect(recentCommands.value.length).toBe(2);
      expect(recentCommands.value[0].id).toBe("dashboard");
      expect(recentCommands.value[1].id).toBe("settings");
    });

    it("handles invalid localStorage data gracefully", () => {
      const consoleError = vi
        .spyOn(console, "error")
        .mockImplementation(() => {});
      localStorageMock.getItem.mockReturnValueOnce("invalid json");

      const { recentCommands } = useCommandPalette(testCommands);

      expect(recentCommands.value.length).toBe(0);
      expect(consoleError).toHaveBeenCalled();

      consoleError.mockRestore();
    });

    it("filters out non-existent command IDs from localStorage", () => {
      localStorageMock.getItem.mockReturnValueOnce(
        JSON.stringify(["dashboard", "non-existent-command"]),
      );

      const { recentCommands } = useCommandPalette(testCommands);

      expect(recentCommands.value.length).toBe(1);
      expect(recentCommands.value[0].id).toBe("dashboard");
    });
  });
});
