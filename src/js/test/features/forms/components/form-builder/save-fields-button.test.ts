import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { setupSaveFieldsButton } from "@/features/forms/components/form-builder/save-fields-button";
import { FormService } from "@/features/forms/services/form-service";
import { formState } from "@/features/forms/state/form-state";
import { FormBuilderError, ErrorCode } from "@/core/errors/form-builder-error";

// Mock dependencies
vi.mock("@/features/forms/services/form-service");
vi.mock("@/features/forms/state/form-state");
vi.mock("@/core/logger");

describe("setupSaveFieldsButton", () => {
  let mockSaveFieldsBtn: HTMLButtonElement;
  let mockFeedbackSpan: HTMLSpanElement;
  let mockFormService: any;
  let mockBuilder: any;

  beforeEach(() => {
    // Create mock DOM elements
    mockSaveFieldsBtn = document.createElement("button");
    mockSaveFieldsBtn.id = "save-fields-btn";
    mockSaveFieldsBtn.innerHTML = `
      <span class="spinner" style="display:none;"></span>
      <span>Save Fields</span>
    `;

    mockFeedbackSpan = document.createElement("span");
    mockFeedbackSpan.id = "schema-save-feedback";

    // Add elements to document
    document.body.appendChild(mockSaveFieldsBtn);
    document.body.appendChild(mockFeedbackSpan);

    // Mock FormService
    mockFormService = {
      saveSchema: vi.fn(),
    };
    vi.mocked(FormService.getInstance).mockReturnValue(mockFormService);

    // Mock form state
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
        ],
      },
    };
    vi.mocked(formState.get).mockReturnValue(mockBuilder);
  });

  afterEach(() => {
    // Clean up DOM
    document.body.removeChild(mockSaveFieldsBtn);
    document.body.removeChild(mockFeedbackSpan);
    vi.clearAllMocks();
  });

  it("should set up click event handler for save fields button", () => {
    setupSaveFieldsButton("test-form-id");

    // Verify the button exists and event handler is set up
    expect(mockSaveFieldsBtn).toBeDefined();
    expect(mockSaveFieldsBtn.id).toBe("save-fields-btn");
  });

  it("should save schema when button is clicked", async () => {
    const formId = "test-form-id";
    setupSaveFieldsButton(formId);

    // Mock successful save
    mockFormService.saveSchema.mockResolvedValue(mockBuilder.form);

    // Click the button
    mockSaveFieldsBtn.click();

    // Wait for async operation
    await new Promise((resolve) => setTimeout(resolve, 0));

    // Verify saveSchema was called with correct parameters
    expect(mockFormService.saveSchema).toHaveBeenCalledWith(
      formId,
      mockBuilder.form,
    );

    // Verify success feedback
    expect(mockFeedbackSpan.textContent).toBe("Fields saved successfully!");
    expect(mockFeedbackSpan.style.color).toBe("rgb(40, 167, 69)");
  });

  it("should show loading state during save", async () => {
    setupSaveFieldsButton("test-form-id");

    // Mock delayed save
    mockFormService.saveSchema.mockImplementation(
      () =>
        new Promise((resolve) =>
          setTimeout(() => resolve(mockBuilder.form), 100),
        ),
    );

    // Click the button
    mockSaveFieldsBtn.click();

    // Check loading state immediately
    const spinner = mockSaveFieldsBtn.querySelector(".spinner") as HTMLElement;
    const buttonText = mockSaveFieldsBtn.querySelector(
      "span:not(.spinner)",
    ) as HTMLElement;

    expect(spinner?.style.display).toBe("inline-block");
    expect(buttonText?.textContent).toBe("Saving...");
    expect(mockSaveFieldsBtn.disabled).toBe(true);

    // Wait for save to complete
    await new Promise((resolve) => setTimeout(resolve, 150));

    // Check final state
    expect(spinner?.style.display).toBe("none");
    expect(buttonText?.textContent).toBe("Save Fields");
    expect(mockSaveFieldsBtn.disabled).toBe(false);
  });

  it("should handle save errors", async () => {
    setupSaveFieldsButton("test-form-id");

    // Mock save error
    const error = new FormBuilderError(
      "Save failed",
      ErrorCode.SAVE_FAILED,
      "Failed to save fields",
    );
    mockFormService.saveSchema.mockRejectedValue(error);

    // Click the button
    mockSaveFieldsBtn.click();

    // Wait for async operation
    await new Promise((resolve) => setTimeout(resolve, 0));

    // Verify error feedback
    expect(mockFeedbackSpan.textContent).toBe("Failed to save fields");
    expect(mockFeedbackSpan.style.color).toBe("rgb(220, 53, 69)");
  });

  it("should handle missing builder", async () => {
    setupSaveFieldsButton("test-form-id");

    // Mock missing builder
    vi.mocked(formState.get).mockReturnValue(undefined);

    // Click the button
    mockSaveFieldsBtn.click();

    // Wait for async operation
    await new Promise((resolve) => setTimeout(resolve, 0));

    // Verify error feedback
    expect(mockFeedbackSpan.textContent).toBe(
      "Form builder not found. Please refresh the page.",
    );
    expect(mockFeedbackSpan.style.color).toBe("rgb(220, 53, 69)");
  });

  it("should clear feedback messages after timeout", async () => {
    setupSaveFieldsButton("test-form-id");

    // Mock successful save
    mockFormService.saveSchema.mockResolvedValue(mockBuilder.form);

    // Click the button
    mockSaveFieldsBtn.click();

    // Wait for save to complete
    await new Promise((resolve) => setTimeout(resolve, 0));

    // Verify success message is shown
    expect(mockFeedbackSpan.textContent).toBe("Fields saved successfully!");

    // Wait for timeout (3 seconds for success)
    await new Promise((resolve) => setTimeout(resolve, 3100));

    // Verify message is cleared
    expect(mockFeedbackSpan.textContent).toBe("");
  });
});
