// Form schema types
export interface FormField {
  // Core field properties
  id: string;
  name: string;
  type: string;
  title: string;
  description?: string;

  // Validation and behavior
  isRequired?: boolean;
  isReadOnly?: boolean;
  isHidden?: boolean;
  validators?: FieldValidator[];

  // UI/UX properties
  placeholder?: string;
  cssClasses?: string;
  defaultValue?: any;

  // Options for choice-based fields
  choices?: FieldChoice[];
  hasOther?: boolean;
  otherText?: string;

  // Layout properties
  startWithNewLine?: boolean;
  indent?: number;
  width?: string | number;

  // Advanced properties
  visibleIf?: string;
  enableIf?: string;
  requiredIf?: string;
  defaultValueExpression?: string;
  calculatedValue?: string;

  // Custom properties
  customProperties?: Record<string, any>;
}

export interface FieldValidator {
  type: string;
  text?: string;
  regex?: string;
  minValue?: number;
  maxValue?: number;
  minLength?: number;
  maxLength?: number;
  allowDigits?: boolean;
  expression?: string;
}

export interface FieldChoice {
  value: string | number;
  text: string;
  imageUrl?: string;
  description?: string;
  customProperties?: Record<string, any>;
}

export interface FormPanel {
  id: string;
  name: string;
  title: string;
  description?: string;
  elements: (FormField | FormPanel)[];
  state?: "expanded" | "collapsed";
  visibleIf?: string;
  customProperties?: Record<string, any>;
}

export interface FormPage {
  id: string;
  name: string;
  title: string;
  description?: string;
  elements: (FormField | FormPanel)[];
  visibleIf?: string;
  customProperties?: Record<string, any>;
}

export interface FormSchema {
  id: number;
  name: string;
  title: string;
  description?: string;

  // Core properties
  pages: FormPage[];
  showProgressBar?: "top" | "bottom" | "both" | "off";
  progressBarType?: "pages" | "questions";

  // Behavior
  mode?: "edit" | "display" | "preview";
  completedHtml?: string;

  // Validation
  checkErrorsMode?: "onNextPage" | "onValueChanged" | "onComplete";

  // Customization
  questionStartIndex?: string | number;
  requiredText?: string;
  questionErrorLocation?: "top" | "bottom";

  // Custom properties
  customProperties?: Record<string, any>;

  // Metadata
  createdAt?: string;
  updatedAt?: string;
  version?: number;
  tags?: string[];
}
