import { z } from "zod";

interface ApiValidationRule {
  type: string;
  min?: number;
  max?: number;
  pattern?: string;
  message: string;
  matchField?: string;
}

interface ApiValidationSchema {
  [field: string]: ApiValidationRule;
}

export async function getValidationSchema(
  schemaName: string,
): Promise<z.ZodType<Record<string, string>>> {
  try {
    const response = await fetch(`/api/v1/validation/${schemaName}`);
    if (!response.ok) {
      console.error(
        `Failed to fetch validation schema for ${schemaName}:`,
        response.status,
      );
      throw new Error("Failed to fetch validation schema");
    }

    const schema: ApiValidationSchema = await response.json();
    return generateZodSchema(schema);
  } catch (error) {
    console.error(`Error fetching validation schema for ${schemaName}:`, error);
    throw error;
  }
}

function generateZodSchema(
  schema: ApiValidationSchema,
): z.ZodType<Record<string, string>> {
  const shape: Record<string, z.ZodType<string>> = {};

  for (const [field, rule] of Object.entries(schema)) {
    shape[field] = generateZodRule(rule);
  }

  return z.object(shape);
}

function generateZodRule(rule: ApiValidationRule): z.ZodType<string> {
  let zodRule: z.ZodString | z.ZodEffects<z.ZodString, string, string> =
    z.string();

  switch (rule.type) {
    case "string":
      if (rule.min !== undefined) {
        zodRule = zodRule.min(rule.min, rule.message);
      }
      if (rule.max !== undefined) {
        zodRule = zodRule.max(rule.max, rule.message);
      }
      if (rule.pattern !== undefined) {
        zodRule = zodRule.regex(new RegExp(rule.pattern), rule.message);
      }
      break;

    case "email":
      zodRule = zodRule.email(rule.message);
      break;

    case "required":
      zodRule = zodRule.min(1, rule.message);
      break;

    case "password":
      zodRule = zodRule
        .min(rule.min || 8, rule.message)
        .regex(/[A-Z]/, "Password must contain at least one uppercase letter")
        .regex(/[a-z]/, "Password must contain at least one lowercase letter")
        .regex(/[0-9]/, "Password must contain at least one number")
        .regex(
          /[^A-Za-z0-9]/,
          "Password must contain at least one special character",
        );
      break;

    case "match":
      if (rule.matchField) {
        if (rule.min !== undefined) {
          zodRule = zodRule.min(rule.min, "Password is too short");
        }
        zodRule = zodRule.refine(
          (val: string) => {
            const matchField = document.getElementById(
              rule.matchField!,
            ) as HTMLInputElement;
            return matchField && val === matchField.value;
          },
          {
            message: rule.message,
          },
        );
      }
      break;
  }

  return zodRule;
}
