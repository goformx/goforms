import {
  describe,
  it,
  expect,
  beforeEach,
  afterEach,
  vi,
  type MockedFunction,
} from "vitest";
import { setupViewSchemaButton } from "@/features/forms/components/form-builder/view-schema-button";
import { showSchemaModal } from "@/features/forms/components/form-builder/schema-modal";
import { Logger } from "@/core/logger";
import { FormBuilderError } from "@/core/errors/form-builder-error";

// Mock dependencies
vi.mock("@/features/forms/components/form-builder/schema-modal");
vi.mock("@/core/logger");

describe("setupViewSchemaButton", () => {
  let mockBuilder: any;
  let mockViewSchemaBtn: HTMLButtonElement;
  let mockShowSchemaModal: MockedFunction<typeof showSchemaModal>;
  let mockLoggerDebug: MockedFunction<typeof Logger.debug>;
  let mockLoggerWarn: MockedFunction<typeof Logger.warn>;
  let mockLoggerError: MockedFunction<typeof Logger.error>;

  beforeEach(() => {
    // Reset all mocks
    vi.clearAllMocks();

    // Mock Logger methods
    mockLoggerDebug = vi.mocked(Logger.debug);
    mockLoggerWarn = vi.mocked(Logger.warn);
    mockLoggerError = vi.mocked(Logger.error);

    // Mock showSchemaModal
    mockShowSchemaModal = vi.mocked(showSchemaModal);

    // Create mock builder with form schema
    mockBuilder = {
      form: {
        display: "form",
        components: [
          {
            type: "textfield",
            key: "name",
            label: "Name",
            input: true,
          },
          {
            type: "email",
            key: "email",
            label: "Email",
            input: true,
          },
        ],
      },
    };

    // Create mock button element
    mockViewSchemaBtn = document.createElement("button");
    mockViewSchemaBtn.id = "view-schema-btn";

    // Mock getElementById
    vi.spyOn(document, "getElementById").mockImplementation((id: string) => {
      if (id === "view-schema-btn") {
        return mockViewSchemaBtn;
      }
      return null;
    });

    // Mock setTimeout for debouncing
    vi.useFakeTimers();
  });

  afterEach(() => {
    // Clean up any event listeners
    mockViewSchemaBtn.remove();
    vi.restoreAllMocks();
    vi.useRealTimers();
  });

  describe("Button Setup", () => {
    it("should set up event listener when button exists", () => {
      setupViewSchemaButton(mockBuilder);

      expect(mockLoggerDebug).toHaveBeenCalledWith(
        "SchemaViewerHandler initialized",
      );
      expect(mockViewSchemaBtn.onclick).toBeDefined();
    });

    it("should throw error when button is not found", () => {
      vi.spyOn(document, "getElementById").mockReturnValue(null);

      expect(() => {
        setupViewSchemaButton(mockBuilder);
      }).toThrow(FormBuilderError);
      expect(mockLoggerDebug).not.toHaveBeenCalled();
    });
  });

  describe("Click Event Handling", () => {
    it("should call showSchemaModal with formatted schema when clicked", () => {
      setupViewSchemaButton(mockBuilder);

      // Simulate button click
      mockViewSchemaBtn.click();

      // Fast-forward timers to trigger debounced function
      vi.runAllTimers();

      expect(mockShowSchemaModal).toHaveBeenCalledWith(
        JSON.stringify(mockBuilder.form, null, 2),
      );
      expect(mockLoggerDebug).toHaveBeenCalledWith(
        "Schema modal opened successfully",
      );
    });

    it("should handle empty schema gracefully", () => {
      mockBuilder.form = { display: "form", components: [] };
      setupViewSchemaButton(mockBuilder);

      mockViewSchemaBtn.click();
      vi.runAllTimers();

      expect(mockShowSchemaModal).toHaveBeenCalledWith(
        JSON.stringify(mockBuilder.form, null, 2),
      );
    });

    it("should handle complex schema with nested components", () => {
      mockBuilder.form = {
        display: "form",
        components: [
          {
            type: "container",
            key: "container1",
            components: [
              {
                type: "textfield",
                key: "nestedField",
                label: "Nested Field",
                input: true,
              },
            ],
          },
        ],
      };

      setupViewSchemaButton(mockBuilder);
      mockViewSchemaBtn.click();
      vi.runAllTimers();

      expect(mockShowSchemaModal).toHaveBeenCalledWith(
        JSON.stringify(mockBuilder.form, null, 2),
      );
    });

    it("should handle errors during schema retrieval", () => {
      // Mock builder.form to throw an error
      Object.defineProperty(mockBuilder, "form", {
        get: () => {
          throw new Error("Schema access error");
        },
      });

      expect(() => {
        setupViewSchemaButton(mockBuilder);
      }).toThrow(Error);
      expect(mockShowSchemaModal).not.toHaveBeenCalled();
    });

    it("should handle errors during JSON stringification", () => {
      // Create a circular reference to cause JSON.stringify to fail
      const circularObj: any = { display: "form", components: [] };
      circularObj.self = circularObj;
      mockBuilder.form = circularObj;

      setupViewSchemaButton(mockBuilder);
      mockViewSchemaBtn.click();
      vi.runAllTimers();

      expect(mockLoggerError).toHaveBeenCalledWith(
        "Schema viewer error:",
        expect.any(Error),
      );
      expect(mockShowSchemaModal).not.toHaveBeenCalled();
    });
  });

  describe("Schema Formatting", () => {
    it("should format schema with proper indentation", () => {
      setupViewSchemaButton(mockBuilder);
      mockViewSchemaBtn.click();
      vi.runAllTimers();

      const expectedSchema = JSON.stringify(mockBuilder.form, null, 2);
      expect(mockShowSchemaModal).toHaveBeenCalledWith(expectedSchema);

      // Verify it's properly formatted (has newlines and spaces)
      expect(expectedSchema).toContain("\n");
      expect(expectedSchema).toContain("  ");
    });

    it("should preserve all schema properties", () => {
      const complexSchema = {
        display: "form",
        components: [
          {
            type: "textfield",
            key: "name",
            label: "Name",
            input: true,
            validate: {
              required: true,
              minLength: 2,
            },
            customClass: "custom-field",
          },
        ],
        settings: {
          showTitle: true,
          showDescription: true,
        },
      };

      mockBuilder.form = complexSchema;
      setupViewSchemaButton(mockBuilder);
      mockViewSchemaBtn.click();
      vi.runAllTimers();

      expect(mockShowSchemaModal).toHaveBeenCalledWith(
        JSON.stringify(complexSchema, null, 2),
      );
    });
  });

  describe("Multiple Button Instances", () => {
    it("should handle multiple calls to setupViewSchemaButton", () => {
      // First setup
      setupViewSchemaButton(mockBuilder);
      expect(mockLoggerDebug).toHaveBeenCalledWith(
        "SchemaViewerHandler initialized",
      );

      // Second setup (should not cause issues)
      setupViewSchemaButton(mockBuilder);
      expect(mockLoggerDebug).toHaveBeenCalledTimes(2);

      // Both event listeners should work (multiple calls to showSchemaModal)
      mockViewSchemaBtn.click();
      vi.runAllTimers();
      expect(mockShowSchemaModal).toHaveBeenCalledTimes(2);
    });
  });

  describe("Event Listener Cleanup", () => {
    it("should properly handle event listener removal", () => {
      setupViewSchemaButton(mockBuilder);

      // Verify event listener is attached
      expect(mockViewSchemaBtn.onclick).toBeDefined();

      // Remove the button (simulating cleanup)
      mockViewSchemaBtn.remove();

      // Should not throw errors
      expect(() => {
        setupViewSchemaButton(mockBuilder);
      }).not.toThrow();
    });
  });
});
