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

export async function getValidationSchema(schemaName: string): Promise<z.ZodType<any>> {
  const response = await fetch(`/api/validation/${schemaName}`);
  if (!response.ok) {
    throw new Error('Failed to fetch validation schema');
  }
  
  const schema: ValidationSchema = await response.json();
  return generateZodSchema(schema);
}

function generateZodSchema(schema: ValidationSchema): z.ZodType<any> {
  const shape: Record<string, z.ZodType<any>> = {};
  
  for (const rule of schema.rules) {
    shape[rule.field] = generateZodRule(rule);
  }
  
  return z.object(shape);
}

function generateZodRule(rule: ValidationRule): z.ZodType<any> {
  let zodRule = z.string();
  
  switch (rule.type) {
    case 'string':
      if (rule.params.min !== undefined) {
        zodRule = zodRule.min(rule.params.min, rule.message);
      }
      if (rule.params.max !== undefined) {
        zodRule = zodRule.max(rule.params.max, rule.message);
      }
      if (rule.params.pattern !== undefined) {
        zodRule = zodRule.regex(new RegExp(rule.params.pattern), rule.message);
      }
      break;
      
    case 'email':
      zodRule = zodRule.email(rule.message);
      break;
      
    case 'password':
      zodRule = zodRule
        .min(rule.params.min, rule.message)
        .regex(/[A-Z]/, 'Password must contain at least one uppercase letter')
        .regex(/[a-z]/, 'Password must contain at least one lowercase letter')
        .regex(/[0-9]/, 'Password must contain at least one number')
        .regex(/[^A-Za-z0-9]/, 'Password must contain at least one special character');
      break;
      
    case 'required':
      zodRule = zodRule.min(1, rule.message);
      break;
  }
  
  return zodRule;
} 