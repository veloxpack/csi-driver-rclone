// Package metadata defines the structured metadata.
//
// https://cloudinary.com/documentation/metadata_api
package metadata

// Field is a single metadata field.
type Field struct {
	Type         FieldType   `json:"type"`
	ExternalID   string      `json:"external_id"`
	Label        string      `json:"label"`
	Mandatory    bool        `json:"mandatory"`
	DefaultValue interface{} `json:"default_value,omitempty"`
	Validation   interface{} `json:"validation,omitempty"`
	DataSource   DataSource  `json:"datasource,omitempty"`
}

// FieldType is the type of the metadata field.
type FieldType string

const (
	// StringFieldType is a string metadata field type.
	StringFieldType FieldType = "string"
	// IntegerFieldType is an integer metadata field type.
	IntegerFieldType FieldType = "integer"
	// DateFieldType is a date metadata field type.
	DateFieldType FieldType = "date"
	// EnumFieldType is an enum metadata field type.
	EnumFieldType FieldType = "enum"
	// SetFieldType is a set metadata field type.
	SetFieldType FieldType = "set"
)

// Validation is the validation of the metadata field.
type Validation interface{}

// ValidationType is the type of the metadata field validation.
type ValidationType string

const (
	greaterThanValidationType  ValidationType = "greater_than"
	lessThanValidationType     ValidationType = "less_than"
	stringLengthValidationType ValidationType = "strlen"
	andValidationType          ValidationType = "and"
)

type compositeValidation struct {
	Type  ValidationType `json:"type"`
	Rules []interface{}  `json:"rules"`
}

type comparisonValidationRule struct {
	Type   ValidationType `json:"type"`
	Value  interface{}    `json:"value"`
	Equals bool           `json:"equals"`
}

type valueLengthValidationRule struct {
	Type ValidationType `json:"type"`
	Min  int            `json:"min,omitempty"`
	Max  int            `json:"max,omitempty"`
}

// AndValidation is relevant for all field types. Allows to include more than one validation rule to be evaluated.
func AndValidation(rules []interface{}) *compositeValidation {
	return &compositeValidation{Type: andValidationType, Rules: rules}
}

// GreaterThanValidation rule for integers.
func GreaterThanValidation(value interface{}, equals bool) *comparisonValidationRule {
	return &comparisonValidationRule{Type: greaterThanValidationType, Value: value, Equals: equals}
}

// LessThanValidation rule for integers.
func LessThanValidation(value interface{}, equals bool) *comparisonValidationRule {
	return &comparisonValidationRule{Type: lessThanValidationType, Value: value, Equals: equals}
}

// StringLengthValidation rule for strings.
func StringLengthValidation(min, max int) *valueLengthValidationRule {
	return &valueLengthValidationRule{Type: stringLengthValidationType, Min: min, Max: max}
}

// DataSource is the data source definition.
type DataSource struct {
	Values []DataSourceValue `json:"values"`
}

// DataSourceValue is the value of a single data source.
type DataSourceValue struct {
	ExternalID string `json:"external_id"`
	Value      string `json:"value"`
	State      string `json:"state"`
}
