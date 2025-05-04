console.log('form-builder.ts');

import { z } from 'zod';
import { validation } from './validation';

// Types for form fields
export type FieldType = 'text' | 'email' | 'number' | 'textarea' | 'select' | 'checkbox' | 'radio' | 'submit';

export interface FieldConfig {
  type: FieldType;
  label: string;
  hasOptions?: boolean;
  hasPlaceholder?: boolean;
  hasRequired?: boolean;
  hasButtonText?: boolean;
}

const FIELD_CONFIGS: FieldConfig[] = [
  {
    type: 'text',
    label: 'Text',
    hasPlaceholder: true,
    hasRequired: true
  },
  {
    type: 'email',
    label: 'Email',
    hasPlaceholder: true,
    hasRequired: true
  },
  {
    type: 'number',
    label: 'Number',
    hasPlaceholder: true,
    hasRequired: true
  },
  {
    type: 'textarea',
    label: 'Textarea',
    hasPlaceholder: true,
    hasRequired: true
  },
  {
    type: 'select',
    label: 'Select',
    hasOptions: true,
    hasRequired: true
  },
  {
    type: 'checkbox',
    label: 'Checkbox',
    hasRequired: true
  },
  {
    type: 'radio',
    label: 'Radio',
    hasOptions: true,
    hasRequired: true
  },
  {
    type: 'submit',
    label: 'Submit Button',
    hasButtonText: true
  }
];

export interface FormField {
  id: string;
  name: string;
  label: string;
  type: FieldType;
  required: boolean;
  options?: string[]; // For select/radio fields
  placeholder?: string;
  buttonText?: string; // For submit buttons
}

export interface FormSchema {
  id: number;
  fields: FormField[];
}

// Zod schema for field validation
const formFieldSchema = z.object({
  id: z.string(),
  name: z.string().min(1, "Field name is required"),
  label: z.string().min(1, "Field label is required"),
  type: z.enum(['text', 'email', 'number', 'textarea', 'select', 'checkbox', 'radio']),
  required: z.boolean(),
  options: z.array(z.string()).optional(),
  placeholder: z.string().optional()
});

export const formSchemaSchema = z.object({
  id: z.number(),
  fields: z.array(formFieldSchema)
});

export class FormBuilder {
  private container: HTMLElement;
  private fields: FormField[] = [];
  private formId: number;

  constructor(containerId: string, formId: number) {
    console.log('FormBuilder: constructor called with formId:', formId);
    const container = document.getElementById(containerId);
    if (!container) throw new Error(`Container ${containerId} not found`);
    
    this.container = container;
    this.formId = formId;
    this.init();
  }

  private init() {
    console.log('init');
    this.renderBuilder();
    this.loadExistingSchema();
  }

  private async loadExistingSchema() {
    try {
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
        this.fields = schema.fields;
        this.renderFields();
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

  private renderBuilder() {
    this.container.innerHTML = `
      <div class="form-builder">
        <div class="form-builder-toolbar">
          <button type="button" class="btn btn-primary" id="add-field-btn">Add Field</button>
          <button type="button" class="btn btn-primary" id="save-fields-btn">Save Fields</button>
        </div>
        <div class="form-builder-fields"></div>
        <div class="form-builder-preview">
          <h3>Form Preview</h3>
          <div id="form-preview"></div>
        </div>
      </div>
    `;

    const addButton = this.container.querySelector('#add-field-btn');
    if (addButton) {
      addButton.addEventListener('click', () => this.showAddFieldDialog());
    }

    const saveButton = this.container.querySelector('#save-fields-btn');
    if (saveButton) {
      saveButton.addEventListener('click', () => this.saveSchema());
    }
  }

  private showAddFieldDialog() {
    const dialog = document.createElement('div');
    dialog.className = 'form-builder-dialog';
    dialog.innerHTML = `
      <div class="dialog-content">
        <h3>Add New Field</h3>
        <form id="add-field-form">
          <div class="form-group">
            <label>Field Name</label>
            <input type="text" name="name" required class="form-input" />
          </div>
          <div class="form-group">
            <label>Label</label>
            <input type="text" name="label" required class="form-input" />
          </div>
          <div class="form-group">
            <label>Type</label>
            <select name="type" required class="form-input" id="field-type">
              ${FIELD_CONFIGS.map(config => `
                <option value="${config.type}">${config.label}</option>
              `).join('')}
            </select>
          </div>
          <div class="form-group" id="button-text-group" style="display: none;">
            <label>Button Text</label>
            <input type="text" name="buttonText" class="form-input" placeholder="Submit" />
          </div>
          <div class="form-group" id="options-group" style="display: none;">
            <label>Options (one per line)</label>
            <textarea name="options" class="form-input" rows="4"></textarea>
          </div>
          <div class="form-group" id="required-group">
            <label>
              <input type="checkbox" name="required" />
              Required
            </label>
          </div>
          <div class="form-group" id="placeholder-group">
            <label>Placeholder</label>
            <input type="text" name="placeholder" class="form-input" />
          </div>
          <div class="dialog-actions">
            <button type="button" class="btn btn-outline" onclick="this.closest('.form-builder-dialog').remove()">Cancel</button>
            <button type="submit" class="btn btn-primary">Add Field</button>
          </div>
        </form>
      </div>
    `;

    this.container.appendChild(dialog);

    // Show/hide fields based on type selection
    const typeSelect = dialog.querySelector('#field-type');
    const buttonTextGroup = dialog.querySelector('#button-text-group') as HTMLDivElement;
    const optionsGroup = dialog.querySelector('#options-group') as HTMLDivElement;
    const placeholderGroup = dialog.querySelector('#placeholder-group') as HTMLDivElement;
    const requiredGroup = dialog.querySelector('#required-group') as HTMLDivElement;
    
    if (typeSelect && buttonTextGroup && optionsGroup && placeholderGroup && requiredGroup) {
      typeSelect.addEventListener('change', (e) => {
        const target = e.target as HTMLSelectElement;
        const config = FIELD_CONFIGS.find(c => c.type === target.value);
        if (config) {
          buttonTextGroup.style.display = config.hasButtonText ? 'block' : 'none';
          optionsGroup.style.display = config.hasOptions ? 'block' : 'none';
          placeholderGroup.style.display = config.hasPlaceholder ? 'block' : 'none';
          requiredGroup.style.display = config.hasRequired ? 'block' : 'none';
        }
      });
    }

    const form = dialog.querySelector('form');
    if (form) {
      form.addEventListener('submit', (e) => {
        e.preventDefault();
        const formData = new FormData(form);
        const fieldType = formData.get('type') as FieldType;
        const config = FIELD_CONFIGS.find(c => c.type === fieldType);
        
        const field: FormField = {
          id: crypto.randomUUID(),
          name: formData.get('name') as string,
          label: formData.get('label') as string,
          type: fieldType,
          required: config?.hasRequired ? formData.get('required') === 'on' : false
        };

        // Add options for select/radio fields
        if (config?.hasOptions) {
          const optionsText = formData.get('options') as string;
          field.options = optionsText.split('\n')
            .map(opt => opt.trim())
            .filter(opt => opt.length > 0);
        }

        // Add placeholder if supported
        if (config?.hasPlaceholder) {
          field.placeholder = formData.get('placeholder') as string;
        }

        // Add button text for submit buttons
        if (config?.hasButtonText) {
          field.buttonText = formData.get('buttonText') as string || 'Submit';
        }

        this.addField(field);
        dialog.remove();
      });
    }
  }

  private addField(field: FormField) {
    this.fields.push(field);
    this.renderFields();
  }

  private renderFields() {
    const fieldsContainer = this.container.querySelector('.form-builder-fields');
    if (!fieldsContainer) return;

    fieldsContainer.innerHTML = this.fields.map(field => `
      <div class="form-builder-field" data-field-id="${field.id}">
        <div class="field-info">
          <span class="field-label">${field.label}</span>
          <span class="field-type">${field.type}</span>
        </div>
        <div class="field-actions">
          <button type="button" class="btn btn-outline btn-sm" onclick="this.closest('.form-builder-field').querySelector('.edit-field-form').classList.toggle('hidden')">Edit</button>
          <button type="button" class="btn btn-danger btn-sm" onclick="document.dispatchEvent(new CustomEvent('deleteField', {detail: '${field.id}'}))">Delete</button>
        </div>
        <form class="edit-field-form hidden">
          <!-- Edit form fields here -->
        </form>
      </div>
    `).join('');

    this.renderPreview();
  }

  private renderPreview() {
    const previewContainer = document.getElementById('form-preview');
    if (!previewContainer) return;

    previewContainer.innerHTML = this.fields.map(field => {
      let input = '';
      switch (field.type) {
        case 'textarea':
          input = `<textarea name="${field.name}" class="form-input" ${field.required ? 'required' : ''} placeholder="${field.placeholder || ''}"></textarea>`;
          break;
        case 'select':
          input = `
            <select name="${field.name}" class="form-input" ${field.required ? 'required' : ''}>
              <option value="">Select...</option>
              ${(field.options || []).map(opt => `<option value="${opt}">${opt}</option>`).join('')}
            </select>
          `;
          break;
        case 'checkbox':
          input = `<input type="checkbox" name="${field.name}" ${field.required ? 'required' : ''} />`;
          break;
        case 'radio':
          input = (field.options || []).map(opt => `
            <label>
              <input type="radio" name="${field.name}" value="${opt}" ${field.required ? 'required' : ''} />
              ${opt}
            </label>
          `).join('');
          break;
        case 'submit':
          input = `<button type="submit" class="btn btn-primary">${field.buttonText || 'Submit'}</button>`;
          break;
        default:
          input = `<input type="${field.type}" name="${field.name}" class="form-input" ${field.required ? 'required' : ''} placeholder="${field.placeholder || ''}" />`;
      }

      // Don't wrap submit buttons in form-group
      if (field.type === 'submit') {
        return `<div class="form-actions">${input}</div>`;
      }

      return `
        <div class="form-group">
          <label class="form-label">
            ${field.label}
            ${field.required ? '<span class="required">*</span>' : ''}
          </label>
          ${input}
        </div>
      `;
    }).join('');
  }

  private async saveSchema() {
    try {
      const schema = { id: this.formId, fields: this.fields };
      const validSchema = formSchemaSchema.parse(schema);
      
      const saveButton = this.container.querySelector('#save-fields-btn') as HTMLButtonElement;
      if (saveButton) {
        saveButton.disabled = true;
        saveButton.textContent = 'Saving...';
      }

      const response = await validation.fetchWithCSRF(`/dashboard/forms/${this.formId}/schema`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(validSchema)
      });

      if (response.ok) {
        if (saveButton) {
          saveButton.textContent = 'Saved!';
          setTimeout(() => {
            saveButton.disabled = false;
            saveButton.textContent = 'Save Fields';
          }, 2000);
        }
      } else {
        throw new Error('Failed to save schema');
      }
    } catch (error) {
      console.error('Failed to save form schema:', error);
      const saveButton = this.container.querySelector('#save-fields-btn') as HTMLButtonElement;
      if (saveButton) {
        saveButton.disabled = false;
        saveButton.textContent = 'Save Failed';
        setTimeout(() => {
          saveButton.textContent = 'Save Fields';
        }, 2000);
      }
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
      new FormBuilder('form-schema-builder', formId);
    } else {
      console.error('FormBuilder: Invalid form ID:', formIdAttr);
    }
  }
} 