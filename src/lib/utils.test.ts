import { describe, it, expect } from "vitest";
import { cn } from "./utils";

describe("cn", () => {
  it("merges class names", () => {
    expect(cn("foo", "bar")).toBe("foo bar");
  });

  it("handles single class", () => {
    expect(cn("single")).toBe("single");
  });

  it("handles empty inputs", () => {
    expect(cn()).toBe("");
    expect(cn("")).toBe("");
  });

  it("handles undefined and null values", () => {
    expect(cn("base", undefined, "extra")).toBe("base extra");
    expect(cn("base", null, "extra")).toBe("base extra");
  });

  it("handles conditional classes with false", () => {
    expect(cn("base", false && "hidden")).toBe("base");
  });

  it("handles conditional classes with true", () => {
    expect(cn("base", true && "visible")).toBe("base visible");
  });

  it("handles mixed conditional classes", () => {
    const isHidden = false;
    const isActive = true;
    expect(cn("base", isHidden && "hidden", isActive && "active")).toBe(
      "base active",
    );
  });

  it("deduplicates conflicting Tailwind classes", () => {
    expect(cn("p-4", "p-6")).toBe("p-6");
    expect(cn("mt-2", "mt-4")).toBe("mt-4");
    expect(cn("text-red-500", "text-blue-500")).toBe("text-blue-500");
  });

  it("preserves non-conflicting Tailwind classes", () => {
    expect(cn("p-4", "m-2")).toBe("p-4 m-2");
    expect(cn("text-lg", "font-bold")).toBe("text-lg font-bold");
  });

  it("handles arrays of classes", () => {
    expect(cn(["foo", "bar"])).toBe("foo bar");
    expect(cn("base", ["extra", "classes"])).toBe("base extra classes");
  });

  it("handles objects with boolean values", () => {
    expect(cn({ hidden: true, visible: false })).toBe("hidden");
    expect(cn("base", { active: true, disabled: false })).toBe("base active");
  });

  it("handles complex Tailwind class combinations", () => {
    expect(
      cn(
        "px-4 py-2 rounded",
        "bg-blue-500 hover:bg-blue-600",
        "text-white font-medium",
      ),
    ).toBe("px-4 py-2 rounded bg-blue-500 hover:bg-blue-600 text-white font-medium");
  });

  it("handles responsive and state variants", () => {
    expect(cn("md:p-4", "lg:p-6", "hover:bg-gray-100")).toBe(
      "md:p-4 lg:p-6 hover:bg-gray-100",
    );
  });
});
