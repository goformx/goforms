import { describe, it, expect, beforeEach } from "vitest";
import { z } from "zod";
import {
  useFormValidation,
  loginSchema,
  signupSchema,
  forgotPasswordSchema,
} from "./useFormValidation";

describe("useFormValidation", () => {
  const testSchema = z.object({
    name: z.string().min(1, "Name is required"),
    email: z.string().email("Invalid email"),
    age: z.number().min(18, "Must be 18 or older"),
  });

  describe("validate", () => {
    it("returns valid result for correct data", () => {
      const { validate } = useFormValidation(testSchema);
      const result = validate({ name: "John", email: "john@example.com", age: 25 });

      expect(result.valid).toBe(true);
      expect(result.errors).toEqual({});
    });

    it("returns invalid result with errors for incorrect data", () => {
      const { validate } = useFormValidation(testSchema);
      const result = validate({ name: "", email: "invalid", age: 16 });

      expect(result.valid).toBe(false);
      expect(result.errors.name).toBe("Name is required");
      expect(result.errors.email).toBe("Invalid email");
      expect(result.errors.age).toBe("Must be 18 or older");
    });

    it("updates errors ref on validation failure", () => {
      const { validate, errors, hasErrors } = useFormValidation(testSchema);

      validate({ name: "", email: "invalid", age: 16 });

      expect(hasErrors.value).toBe(true);
      expect(errors.value.name).toBe("Name is required");
    });

    it("clears errors ref on validation success", () => {
      const { validate, errors, hasErrors } = useFormValidation(testSchema);

      // First fail
      validate({ name: "", email: "invalid", age: 16 });
      expect(hasErrors.value).toBe(true);

      // Then succeed
      validate({ name: "John", email: "john@example.com", age: 25 });
      expect(hasErrors.value).toBe(false);
      expect(errors.value).toEqual({});
    });
  });

  describe("validateField", () => {
    it("returns null for valid field", () => {
      const { validateField } = useFormValidation(testSchema);
      const result = validateField("email", "john@example.com");

      expect(result).toBeNull();
    });

    it("returns error message for invalid field", () => {
      const { validateField } = useFormValidation(testSchema);
      const result = validateField("email", "invalid");

      expect(result).toBe("Invalid email");
    });

    it("updates errors ref for invalid field", () => {
      const { validateField, errors, hasErrors } = useFormValidation(testSchema);

      validateField("email", "invalid");

      expect(hasErrors.value).toBe(true);
      expect(errors.value.email).toBe("Invalid email");
    });

    it("clears field error when valid", () => {
      const { validateField, errors, hasErrors } = useFormValidation(testSchema);

      // First fail
      validateField("email", "invalid");
      expect(errors.value.email).toBe("Invalid email");

      // Then succeed
      validateField("email", "john@example.com");
      expect(errors.value.email).toBeUndefined();
      expect(hasErrors.value).toBe(false);
    });
  });

  describe("clearErrors", () => {
    it("clears all errors", () => {
      const { validate, clearErrors, errors, hasErrors } =
        useFormValidation(testSchema);

      validate({ name: "", email: "invalid", age: 16 });
      expect(hasErrors.value).toBe(true);

      clearErrors();

      expect(errors.value).toEqual({});
      expect(hasErrors.value).toBe(false);
    });
  });

  describe("clearFieldError", () => {
    it("clears specific field error", () => {
      const { validate, clearFieldError, errors } = useFormValidation(testSchema);

      validate({ name: "", email: "invalid", age: 16 });
      expect(errors.value.email).toBeDefined();

      clearFieldError("email");

      expect(errors.value.email).toBeUndefined();
      expect(errors.value.name).toBeDefined(); // Other errors remain
    });

    it("updates hasErrors when last error is cleared", () => {
      const { validateField, clearFieldError, hasErrors } =
        useFormValidation(testSchema);

      validateField("email", "invalid");
      expect(hasErrors.value).toBe(true);

      clearFieldError("email");
      expect(hasErrors.value).toBe(false);
    });
  });

  describe("setFieldError", () => {
    it("sets error for a field", () => {
      const { setFieldError, errors, hasErrors } = useFormValidation(testSchema);

      setFieldError("email", "Custom error message");

      expect(errors.value.email).toBe("Custom error message");
      expect(hasErrors.value).toBe(true);
    });
  });
});

describe("loginSchema", () => {
  it("validates correct login data", () => {
    const result = loginSchema.safeParse({
      email: "user@example.com",
      password: "password123",
    });

    expect(result.success).toBe(true);
  });

  it("rejects invalid email", () => {
    const result = loginSchema.safeParse({
      email: "invalid-email",
      password: "password123",
    });

    expect(result.success).toBe(false);
    if (!result.success) {
      expect(result.error.issues[0].message).toBe(
        "Please enter a valid email address",
      );
    }
  });

  it("rejects empty password", () => {
    const result = loginSchema.safeParse({
      email: "user@example.com",
      password: "",
    });

    expect(result.success).toBe(false);
    if (!result.success) {
      expect(result.error.issues[0].message).toBe("Password is required");
    }
  });
});

describe("signupSchema", () => {
  const validSignup = {
    email: "user@example.com",
    password: "Password1!",
    confirmPassword: "Password1!",
  };

  it("validates correct signup data", () => {
    const result = signupSchema.safeParse(validSignup);
    expect(result.success).toBe(true);
  });

  it("rejects password without uppercase", () => {
    const result = signupSchema.safeParse({
      ...validSignup,
      password: "password1!",
      confirmPassword: "password1!",
    });

    expect(result.success).toBe(false);
    if (!result.success) {
      expect(result.error.issues[0].message).toContain("uppercase");
    }
  });

  it("rejects password without lowercase", () => {
    const result = signupSchema.safeParse({
      ...validSignup,
      password: "PASSWORD1!",
      confirmPassword: "PASSWORD1!",
    });

    expect(result.success).toBe(false);
    if (!result.success) {
      expect(result.error.issues[0].message).toContain("lowercase");
    }
  });

  it("rejects password without number", () => {
    const result = signupSchema.safeParse({
      ...validSignup,
      password: "Password!",
      confirmPassword: "Password!",
    });

    expect(result.success).toBe(false);
    if (!result.success) {
      expect(result.error.issues[0].message).toContain("number");
    }
  });

  it("rejects password without special character", () => {
    const result = signupSchema.safeParse({
      ...validSignup,
      password: "Password1",
      confirmPassword: "Password1",
    });

    expect(result.success).toBe(false);
    if (!result.success) {
      expect(result.error.issues[0].message).toContain("special character");
    }
  });

  it("rejects password shorter than 8 characters", () => {
    const result = signupSchema.safeParse({
      ...validSignup,
      password: "Pass1!",
      confirmPassword: "Pass1!",
    });

    expect(result.success).toBe(false);
    if (!result.success) {
      expect(result.error.issues[0].message).toContain("8 characters");
    }
  });

  it("rejects mismatched passwords", () => {
    const result = signupSchema.safeParse({
      ...validSignup,
      confirmPassword: "DifferentPassword1!",
    });

    expect(result.success).toBe(false);
    if (!result.success) {
      expect(result.error.issues.some((i) => i.message.includes("match"))).toBe(
        true,
      );
    }
  });
});

describe("forgotPasswordSchema", () => {
  it("validates correct email", () => {
    const result = forgotPasswordSchema.safeParse({
      email: "user@example.com",
    });

    expect(result.success).toBe(true);
  });

  it("rejects invalid email", () => {
    const result = forgotPasswordSchema.safeParse({
      email: "invalid-email",
    });

    expect(result.success).toBe(false);
    if (!result.success) {
      expect(result.error.issues[0].message).toBe(
        "Please enter a valid email address",
      );
    }
  });
});
