console.log('form-builder.ts');

import { Formio } from '@formio/js';
import type { FormSchema } from './schema/form-schema';
import { validation } from './validation';

// Import Form.io styles
import '@formio/js/dist/formio.full.min.css';

export class FormBuilder {
  private container: HTMLElement;
  private builder: any; // Form.io builder instance
  private formId: number;
  private currentSchema: FormSchema = {
    id: 0,
    name: 'form',
    title: 'Form',
    pages: [],
    version: 1
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
    // Create the form builder with options
    const builderOptions = {
      display: 'form',
      noDefaultSubmitButton: true,
      builder: {
        basic: {
          title: 'Basic Components',
          default: true,
          weight: 0,
          components: {
            textfield: true,
            textarea: true,
            email: true,
            phoneNumber: true,
            number: true,
            password: true,
            checkbox: true,
            selectboxes: true,
            select: true,
            radio: true,
            button: true,
          }
        },
        advanced: false,
        premium: false,
        resource: false
      }
    };

    // Initialize Form.io builder
    Formio.builder(this.container, {}, builderOptions).then((builder: any) => {
      this.builder = builder;
      
      // Load existing schema if any
      this.loadExistingSchema();

      // Handle save events
      this.builder.on('saveComponent', this.saveSchema.bind(this));
      this.builder.on('editComponent', this.saveSchema.bind(this));
      this.builder.on('deleteComponent', this.saveSchema.bind(this));
    });
  }

  private async loadExistingSchema() {
    try {
      if (this.formId === 0) {
        // New form, use empty schema
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
        const schema = await response.json() as FormSchema;
        console.log('Loaded form schema:', schema);
        
        // Convert our schema format to Form.io format and load it
        const formioSchema = this.convertToFormio(schema);
        this.builder.setForm(formioSchema);
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

  private async saveSchema(): Promise<boolean> {
    try {
      // Get the schema from Form.io builder
      const formioSchema = this.builder.schema;
      
      // Convert Form.io schema to our format
      const schema = this.convertFromFormio(formioSchema);
      schema.id = this.formId;
      schema.updatedAt = new Date().toISOString();

      const response = await validation.fetchWithCSRF(`/dashboard/forms/${this.formId}/schema`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(schema)
      });

      if (response.ok) {
        console.log('Schema saved successfully');
        this.currentSchema = schema;
        return true;
      } else {
        throw new Error('Failed to save schema');
      }
    } catch (error) {
      console.error('Failed to save form schema:', error);
      return false;
    }
  }

  private convertToFormio(schema: FormSchema): any {
    // Convert our schema format to Form.io format
    return {
      display: 'form',
      components: schema.pages.flatMap(page => {
        return [
          {
            type: 'panel',
            title: page.title,
            key: page.name,
            components: page.elements.map(element => {
              if ('elements' in element) {
                // It's a panel
                return {
                  type: 'panel',
                  title: element.title,
                  key: element.name,
                  components: this.convertElementsToFormio(element.elements)
                };
              } else {
                // It's a field
                return this.convertFieldToFormio(element);
              }
            })
          }
        ];
      })
    };
  }

  private convertElementsToFormio(elements: any[]): any[] {
    return elements.map(element => {
      if ('elements' in element) {
        return {
          type: 'panel',
          title: element.title,
          key: element.name,
          components: this.convertElementsToFormio(element.elements)
        };
      } else {
        return this.convertFieldToFormio(element);
      }
    });
  }

  private convertFieldToFormio(field: any): any {
    const baseField = {
      type: this.mapFieldType(field.type),
      label: field.title,
      key: field.name,
      input: true,
      validate: {
        required: field.isRequired
      }
    };

    if (field.choices) {
      return {
        ...baseField,
        data: {
          values: field.choices.map((choice: any) => ({
            value: choice.value,
            label: choice.text
          }))
        }
      };
    }

    return baseField;
  }

  private mapFieldType(type: string): string {
    // Map our field types to Form.io types
    const typeMap: { [key: string]: string } = {
      text: 'textfield',
      textarea: 'textarea',
      email: 'email',
      number: 'number',
      checkbox: 'checkbox',
      radio: 'radio',
      select: 'select',
      date: 'datetime',
      file: 'file',
      panel: 'panel'
    };

    return typeMap[type] || 'textfield';
  }

  private convertFromFormio(formioSchema: any): FormSchema {
    // Convert Form.io format to our schema format
    return {
      id: this.formId,
      name: formioSchema.name || 'form',
      title: formioSchema.title || 'Form',
      pages: formioSchema.components
        .filter((component: any) => component.type === 'panel')
        .map((panel: any) => ({
          id: panel.key,
          name: panel.key,
          title: panel.title,
          elements: this.convertFormioElements(panel.components)
        })),
      version: 1
    };
  }

  private convertFormioElements(components: any[]): any[] {
    return components.map(component => {
      if (component.type === 'panel') {
        return {
          id: component.key,
          name: component.key,
          title: component.title,
          elements: this.convertFormioElements(component.components)
        };
      } else {
        return {
          id: component.key,
          name: component.key,
          type: this.mapFormioType(component.type),
          title: component.label,
          isRequired: component.validate?.required || false,
          choices: component.data?.values?.map((value: any) => ({
            value: value.value,
            text: value.label
          }))
        };
      }
    });
  }

  private mapFormioType(formioType: string): string {
    // Map Form.io types to our types
    const typeMap: { [key: string]: string } = {
      textfield: 'text',
      textarea: 'textarea',
      email: 'email',
      number: 'number',
      checkbox: 'checkbox',
      radio: 'radio',
      select: 'select',
      datetime: 'date',
      file: 'file',
      panel: 'panel'
    };

    return typeMap[formioType] || 'text';
  }
}

// Initialize form builder when the module is loaded
const formSchemaBuilder = document.getElementById('form-schema-builder');
if (formSchemaBuilder) {
  const formIdAttr = formSchemaBuilder.getAttribute('data-form-id');
  if (formIdAttr) {
    const formId = parseInt(formIdAttr, 10);
    if (!isNaN(formId)) {
      new FormBuilder('form-schema-builder', formId);
    } else {
      console.error('FormBuilder: Invalid form ID:', formIdAttr);
    }
  }
} 