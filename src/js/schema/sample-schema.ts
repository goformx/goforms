import type { FormSchema } from './form-schema';

export const sampleFormSchema: FormSchema = {
  id: 0,
  name: "contact-form",
  title: "Contact Information",
  description: "Please provide your contact information",
  pages: [
    {
      id: "contact-page",
      name: "contact",
      title: "Contact Details",
      elements: [
        {
          id: "name-panel",
          name: "nameInfo",
          title: "Name Information",
          elements: [
            {
              id: "first-name",
              name: "firstName",
              type: "text",
              title: "First Name",
              isRequired: true,
              validators: [
                {
                  type: "text",
                  minLength: 1,
                  maxLength: 50,
                  text: "First name must be between 1 and 50 characters"
                }
              ]
            },
            {
              id: "last-name", 
              name: "lastName",
              type: "text",
              title: "Last Name",
              isRequired: true,
              validators: [
                {
                  type: "text",
                  minLength: 1,
                  maxLength: 50,
                  text: "Last name must be between 1 and 50 characters"
                }
              ]
            }
          ]
        },
        {
          id: "email",
          name: "email",
          type: "text",
          title: "Email",
          isRequired: true,
          validators: [
            {
              type: "email",
              text: "Please enter a valid email address"
            }
          ]
        },
        {
          id: "phone",
          name: "phone",
          type: "text",
          title: "Phone Number",
          validators: [
            {
              type: "regex",
              regex: "^[0-9]{10}$",
              text: "Please enter a valid 10-digit phone number"
            }
          ]
        }
      ]
    }
  ],
  showProgressBar: "top",
  progressBarType: "pages",
  mode: "edit",
  checkErrorsMode: "onValueChanged"
}; 