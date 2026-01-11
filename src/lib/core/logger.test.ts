import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import type { Logger as LoggerType } from "./logger";

// We need to test the Logger class with different DEV states
// Since import.meta.env.DEV is read at class definition time,
// we'll test the actual behavior in the test environment

describe("Logger", () => {
  let consoleSpy: {
    log: ReturnType<typeof vi.spyOn>;
    error: ReturnType<typeof vi.spyOn>;
    warn: ReturnType<typeof vi.spyOn>;
    info: ReturnType<typeof vi.spyOn>;
    group: ReturnType<typeof vi.spyOn>;
    groupEnd: ReturnType<typeof vi.spyOn>;
    table: ReturnType<typeof vi.spyOn>;
  };

  beforeEach(() => {
    consoleSpy = {
      log: vi.spyOn(console, "log").mockImplementation(() => {}),
      error: vi.spyOn(console, "error").mockImplementation(() => {}),
      warn: vi.spyOn(console, "warn").mockImplementation(() => {}),
      info: vi.spyOn(console, "info").mockImplementation(() => {}),
      group: vi.spyOn(console, "group").mockImplementation(() => {}),
      groupEnd: vi.spyOn(console, "groupEnd").mockImplementation(() => {}),
      table: vi.spyOn(console, "table").mockImplementation(() => {}),
    };
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  // In test environment (VITEST=true), import.meta.env.DEV is typically true
  // We test that the methods work correctly when called

  describe("in development mode", () => {
    // Import fresh to get the test environment state
    let Logger: typeof LoggerType;

    beforeEach(async () => {
      // Dynamic import to get fresh module
      const module = await import("./logger");
      Logger = module.Logger;
    });

    it("Logger.log calls console.log with arguments", () => {
      Logger.log("test message", { data: 123 });

      // In test mode, DEV should be true
      if (import.meta.env.DEV) {
        expect(consoleSpy.log).toHaveBeenCalledWith("test message", {
          data: 123,
        });
      }
    });

    it("Logger.error calls console.error with arguments", () => {
      Logger.error("error message", new Error("test"));

      if (import.meta.env.DEV) {
        expect(consoleSpy.error).toHaveBeenCalled();
      }
    });

    it("Logger.warn calls console.warn with arguments", () => {
      Logger.warn("warning message");

      if (import.meta.env.DEV) {
        expect(consoleSpy.warn).toHaveBeenCalledWith("warning message");
      }
    });

    it("Logger.debug calls console.log with arguments", () => {
      Logger.debug("debug info", 123, true);

      if (import.meta.env.DEV) {
        expect(consoleSpy.log).toHaveBeenCalledWith("debug info", 123, true);
      }
    });

    it("Logger.info calls console.info with arguments", () => {
      Logger.info("info message");

      if (import.meta.env.DEV) {
        expect(consoleSpy.info).toHaveBeenCalledWith("info message");
      }
    });

    it("Logger.group calls console.group with label", () => {
      Logger.group("Test Group");

      if (import.meta.env.DEV) {
        expect(consoleSpy.group).toHaveBeenCalledWith("Test Group");
      }
    });

    it("Logger.groupEnd calls console.groupEnd", () => {
      Logger.groupEnd();

      if (import.meta.env.DEV) {
        expect(consoleSpy.groupEnd).toHaveBeenCalled();
      }
    });

    it("Logger.table calls console.table with data", () => {
      const data = [
        { name: "Alice", age: 30 },
        { name: "Bob", age: 25 },
      ];
      Logger.table(data);

      if (import.meta.env.DEV) {
        expect(consoleSpy.table).toHaveBeenCalledWith(data);
      }
    });

    it("handles multiple arguments correctly", () => {
      Logger.log("message", 1, 2, 3, { nested: { value: true } });

      if (import.meta.env.DEV) {
        expect(consoleSpy.log).toHaveBeenCalledWith("message", 1, 2, 3, {
          nested: { value: true },
        });
      }
    });

    it("handles no arguments", () => {
      Logger.log();

      if (import.meta.env.DEV) {
        expect(consoleSpy.log).toHaveBeenCalledWith();
      }
    });
  });
});

describe("Logger static class structure", () => {
  it("exports Logger as a class", async () => {
    const { Logger } = await import("./logger");

    expect(Logger).toBeDefined();
    expect(typeof Logger.log).toBe("function");
    expect(typeof Logger.error).toBe("function");
    expect(typeof Logger.warn).toBe("function");
    expect(typeof Logger.debug).toBe("function");
    expect(typeof Logger.info).toBe("function");
    expect(typeof Logger.group).toBe("function");
    expect(typeof Logger.groupEnd).toBe("function");
    expect(typeof Logger.table).toBe("function");
  });

  it("all methods are static", async () => {
    const { Logger } = await import("./logger");

    // All methods should be callable without instantiation
    expect(() => {
      Logger.log("test");
      Logger.error("test");
      Logger.warn("test");
      Logger.debug("test");
      Logger.info("test");
      Logger.group("test");
      Logger.groupEnd();
      Logger.table([]);
    }).not.toThrow();
  });
});
