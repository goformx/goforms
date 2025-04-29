import { z } from 'zod';

interface ValidationRule {
  field: string;
  type: string;
  params: Record<string, any>;
  message: string;
}

interface ValidationSchema {
  name: string;
  rules: ValidationRule[];
}

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

export async function getValidationSchema(schemaName: string): Promise<z.ZodType<any>> {
  const response = await fetch(`/api/validation/${schemaName}`);
  if (!response.ok) {
    throw new Error('Failed to fetch validation schema');
  }
  
  const schema: ApiValidationSchema = await response.json();
  return generateZodSchema(schema);
}

function generateZodSchema(schema: ApiValidationSchema): z.ZodType<any> {
  const shape: Record<string, z.ZodType<any>> = {};
  
  for (const [field, rule] of Object.entries(schema)) {
    shape[field] = generateZodRule(rule);
  }
  
  return z.object(shape);
}

function generateZodRule(rule: ApiValidationRule): z.ZodType<any> {
  let zodRule: z.ZodString | z.ZodEffects<z.ZodString, string, string> = z.string();
  
  switch (rule.type) {
    case 'string':
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
      
    case 'email':
      zodRule = zodRule.email(rule.message);
      break;
      
    case 'password':
      zodRule = zodRule
        .min(rule.min || 8, rule.message)
        .regex(/[A-Z]/, 'Password must contain at least one uppercase letter')
        .regex(/[a-z]/, 'Password must contain at least one lowercase letter')
        .regex(/[0-9]/, 'Password must contain at least one number')
        .regex(/[^A-Za-z0-9]/, 'Password must contain at least one special character');
      break;
      
    case 'match':
      if (rule.matchField) {
        zodRule = zodRule.refine((val: string) => {
          const matchField = document.getElementById(rule.matchField!) as HTMLInputElement;
          return val === matchField?.value;
        }, {
          message: rule.message
        });
      }
      break;
      
    case 'required':
      zodRule = zodRule.min(1, rule.message);
      break;
  }
  
  return zodRule;
} 