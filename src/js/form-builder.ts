console.log('form-builder.ts');

import { Formio } from '@formio/js';
import type { FormSchema } from './schema/form-schema';
import { validation } from './validation';

// Import Form.io styles
import '@formio/js/dist/formio.full.min.css';

export interface FormBuilderOptions {
  disabled?: string[];
  noNewEdit?: boolean;
  noDefaultSubmitButton?: boolean;
  alwaysConfirmComponentRemoval?: boolean;
  formConfig?: object;
  resourceTag?: string;
  editForm?: any;
  language?: string;
  builder?: object;
  display?: 'form' | 'wizard' | 'pdf';
  resourceFilter?: string;
  noSource?: boolean;
  showFullJsonSchema?: boolean;
}

export class FormBuilder {
  private container: HTMLElement;
  private builder: any; // Form.io builder instance
  private formId: number;
  private currentSchema: any = {
    display: 'form',
    components: []
  };

  constructor(containerId: string, formId: number) {
    console.log('FormBuilder: constructor called with formId:', formId);
    const container = document.getElementById(containerId);
    if (!container) throw new Error(`Container ${containerId} not found`);
    this.container = container;
    this.formId = formId;
    this.init();
  }

  private init() {
    const builderOptions: FormBuilderOptions = {
      display: 'form',
      noDefaultSubmitButton: true,
      builder: {
        basic: {},
        advanced: {},
        layout: {},
        data: false,
        premium: false,
        resource: {}
      }
    };

    Formio.builder(this.container, {}, builderOptions).then((builder: any) => {
      this.builder = builder;
      this.loadExistingSchema();
    });
  }

  private async loadExistingSchema() {
    try {
      if (this.formId === 0) {
        return;
      }
      console.log('Loading form schema for form ID:', this.formId);
      const response = await validation.fetchWithCSRF(`/dashboard/forms/${this.formId}/schema`, {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json'
        }
      });
      if (response.ok) {
        const schema = await response.json();
        console.log('Loaded form schema:', schema);
        this.builder.setForm(schema);
        this.currentSchema = schema;
      } else {
        if (response.status === 401) {
          console.error('Not authenticated, redirecting to login');
          window.location.href = '/login';
        } else {
          console.error('Failed to load form schema:', response.status, response.statusText);
        }
      }
    } catch (error) {
      console.error('Failed to load form schema:', error);
    }
  }

  public async saveSchema(): Promise<boolean> {
    try {
      const formioSchema = this.builder.schema;
      const response = await validation.fetchWithCSRF(`/dashboard/forms/${this.formId}/schema`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(formioSchema)
      });
      if (response.ok) {
        console.log('Schema saved successfully');
        this.currentSchema = formioSchema;
        return true;
      } else {
        throw new Error('Failed to save schema');
      }
    } catch (error) {
      console.error('Failed to save form schema:', error);
      return false;
    }
  }
}

// Initialize form builder when the module is loaded
const formSchemaBuilder = document.getElementById('form-schema-builder');
if (formSchemaBuilder) {
  const formIdAttr = formSchemaBuilder.getAttribute('data-form-id');
  if (formIdAttr) {
    const formId = parseInt(formIdAttr, 10);
    if (!isNaN(formId)) {
      (window as any).formBuilderInstance = new FormBuilder('form-schema-builder', formId);
    } else {
      console.error('FormBuilder: Invalid form ID:', formIdAttr);
    }
  }
}

console.log('Basic components in Form.io:', Object.values(Formio.Components.components)
  .filter(c => (c as any).builderInfo && (c as any).builderInfo.group === 'basic')
  .map(c => (c as any).builderInfo)); 