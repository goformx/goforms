package validation

// SchemaParser handles parsing and extracting validation rules from form schemas
type SchemaParser struct{}

// NewSchemaParser creates a new schema parser
func NewSchemaParser() *SchemaParser {
	return &SchemaParser{}
}

// ExtractValidationRules extracts validation rules from a component
func (p *SchemaParser) ExtractValidationRules(component map[string]any) FieldValidation {
	validation := FieldValidation{
		Required:    false,
		Type:        "",
		MinLength:   0,
		MaxLength:   0,
		Min:         0,
		Max:         0,
		Pattern:     "",
		Options:     []string{},
		CustomRules: []Rule{},
		Conditional: map[string]any{},
	}

	// Extract basic validation properties
	p.extractBasicValidation(component, &validation)

	// Extract component type
	p.extractComponentType(component, &validation)

	// Extract options for select/radio/checkbox components
	p.extractComponentOptions(component, &validation)

	return validation
}

// extractBasicValidation extracts basic validation properties
func (p *SchemaParser) extractBasicValidation(component map[string]any, validation *FieldValidation) {
	validate, validateOk := component["validate"].(map[string]any)
	if !validateOk {
		return
	}

	p.extractRequired(validate, validation)
	p.extractLengthValidation(validate, validation)
	p.extractNumericValidation(validate, validation)
	p.extractPattern(validate, validation)
}

// extractRequired extracts required field validation
func (p *SchemaParser) extractRequired(validate map[string]any, validation *FieldValidation) {
	if required, requiredOk := validate["required"].(bool); requiredOk {
		validation.Required = required
	}
}

// extractLengthValidation extracts length validation rules
func (p *SchemaParser) extractLengthValidation(validate map[string]any, validation *FieldValidation) {
	if minLength, minLengthOk := validate["minLength"].(float64); minLengthOk {
		validation.MinLength = int(minLength)
	}

	if maxLength, maxLengthOk := validate["maxLength"].(float64); maxLengthOk {
		validation.MaxLength = int(maxLength)
	}
}

// extractNumericValidation extracts numeric validation rules
func (p *SchemaParser) extractNumericValidation(validate map[string]any, validation *FieldValidation) {
	if minVal, minOk := validate["min"].(float64); minOk {
		validation.Min = minVal
	}

	if maxVal, maxOk := validate["max"].(float64); maxOk {
		validation.Max = maxVal
	}
}

// extractPattern extracts pattern validation
func (p *SchemaParser) extractPattern(validate map[string]any, validation *FieldValidation) {
	if pattern, patternOk := validate["pattern"].(string); patternOk {
		validation.Pattern = pattern
	}
}

// extractComponentType extracts the component type
func (p *SchemaParser) extractComponentType(component map[string]any, validation *FieldValidation) {
	if componentType, typeOk := component["type"].(string); typeOk {
		validation.Type = componentType
	}
}

// extractComponentOptions extracts options for select/radio/checkbox components
func (p *SchemaParser) extractComponentOptions(component map[string]any, validation *FieldValidation) {
	data, dataOk := component["data"].(map[string]any)
	if !dataOk {
		return
	}

	values, valuesOk := data["values"].([]any)
	if !valuesOk {
		return
	}

	for _, value := range values {
		p.extractOptionValue(value, validation)
	}
}

// extractOptionValue extracts a single option value
func (p *SchemaParser) extractOptionValue(value any, validation *FieldValidation) {
	valueMap, valueMapOk := value.(map[string]any)
	if !valueMapOk {
		return
	}

	if label, labelOk := valueMap["label"].(string); labelOk {
		validation.Options = append(validation.Options, label)
	}
}

// ExtractComponents extracts components from a form schema
func (p *SchemaParser) ExtractComponents(schema map[string]any) ([]any, bool) {
	components, ok := schema["components"].([]any)

	return components, ok
}

// ExtractComponentKey extracts the key from a component
func (p *SchemaParser) ExtractComponentKey(component map[string]any) (string, bool) {
	key, ok := component["key"].(string)

	return key, ok
}

// ConvertToClientRules converts server-side validation rules to client-side format
func (p *SchemaParser) ConvertToClientRules(validation *FieldValidation) map[string]any {
	clientRules := make(map[string]any)

	if validation.Required {
		clientRules["required"] = true
	}

	if validation.Type != "" {
		clientRules["type"] = validation.Type
	}

	if validation.MinLength > 0 {
		clientRules["minLength"] = validation.MinLength
	}

	if validation.MaxLength > 0 {
		clientRules["maxLength"] = validation.MaxLength
	}

	if validation.Min != 0 {
		clientRules["min"] = validation.Min
	}

	if validation.Max != 0 {
		clientRules["max"] = validation.Max
	}

	if validation.Pattern != "" {
		clientRules["pattern"] = validation.Pattern
	}

	if len(validation.Options) > 0 {
		clientRules["options"] = validation.Options
	}

	return clientRules
}
