package hitl

import (
	"testing"
)

func TestCLIFormRenderer_createField_Text(t *testing.T) {
	r := NewCLIFormRenderer()

	field := Field{
		Name:    "name",
		Type:    FieldTypeText,
		Label:   "Your Name",
		Default: "Alice",
	}

	f, ptr := r.createField(field)
	if f == nil {
		t.Fatal("expected field")
	}
	if ptr == nil {
		t.Fatal("expected value pointer")
	}

	strPtr, ok := ptr.(*string)
	if !ok {
		t.Fatalf("expected *string, got %T", ptr)
	}
	if *strPtr != "Alice" {
		t.Errorf("expected default value 'Alice', got '%s'", *strPtr)
	}
}

func TestCLIFormRenderer_createField_Select(t *testing.T) {
	r := NewCLIFormRenderer()

	field := Field{
		Name:    "choice",
		Type:    FieldTypeSelect,
		Label:   "Choose",
		Default: "b",
		Options: []Option{
			{Value: "a", Label: "A"},
			{Value: "b", Label: "B"},
			{Value: "c", Label: "C"},
		},
	}

	f, ptr := r.createField(field)
	if f == nil {
		t.Fatal("expected field")
	}

	strPtr, ok := ptr.(*string)
	if !ok {
		t.Fatalf("expected *string, got %T", ptr)
	}
	if *strPtr != "b" {
		t.Errorf("expected default value 'b', got '%s'", *strPtr)
	}
}

func TestCLIFormRenderer_createField_Multiselect(t *testing.T) {
	r := NewCLIFormRenderer()

	field := Field{
		Name:    "tags",
		Type:    FieldTypeMultiselect,
		Label:   "Select tags",
		Default: []any{"a", "c"},
		Options: []Option{
			{Value: "a", Label: "A"},
			{Value: "b", Label: "B"},
			{Value: "c", Label: "C"},
		},
	}

	f, ptr := r.createField(field)
	if f == nil {
		t.Fatal("expected field")
	}

	arrPtr, ok := ptr.(*[]string)
	if !ok {
		t.Fatalf("expected *[]string, got %T", ptr)
	}
	if len(*arrPtr) != 2 {
		t.Errorf("expected 2 default values, got %d", len(*arrPtr))
	}
}

func TestCLIFormRenderer_createField_Confirm(t *testing.T) {
	r := NewCLIFormRenderer()

	field := Field{
		Name:    "agree",
		Type:    FieldTypeConfirm,
		Label:   "Do you agree?",
		Default: true,
	}

	f, ptr := r.createField(field)
	if f == nil {
		t.Fatal("expected field")
	}

	boolPtr, ok := ptr.(*bool)
	if !ok {
		t.Fatalf("expected *bool, got %T", ptr)
	}
	if !*boolPtr {
		t.Error("expected default value true")
	}
}

func TestCLIFormRenderer_createField_Number(t *testing.T) {
	r := NewCLIFormRenderer()

	field := Field{
		Name:    "count",
		Type:    FieldTypeNumber,
		Label:   "Count",
		Default: 42,
	}

	f, ptr := r.createField(field)
	if f == nil {
		t.Fatal("expected field")
	}

	strPtr, ok := ptr.(*string)
	if !ok {
		t.Fatalf("expected *string, got %T", ptr)
	}
	if *strPtr != "42" {
		t.Errorf("expected default value '42', got '%s'", *strPtr)
	}
}

func TestCLIFormRenderer_textValidator(t *testing.T) {
	r := NewCLIFormRenderer()

	minLen := 3
	maxLen := 10
	field := Field{
		Name:     "text",
		Label:    "Text",
		Required: true,
		Validation: &Validation{
			MinLength: &minLen,
			MaxLength: &maxLen,
		},
	}

	validator := r.textValidator(field)

	tests := []struct {
		value   string
		wantErr bool
	}{
		{"", true},           // Required
		{"ab", true},         // Too short
		{"abc", false},       // Exactly min
		{"hello", false},     // Valid
		{"0123456789", false}, // Exactly max
		{"01234567890", true}, // Too long
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			err := validator(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("validator(%q) error = %v, wantErr %v", tt.value, err, tt.wantErr)
			}
		})
	}
}

func TestCLIFormRenderer_numberValidator(t *testing.T) {
	r := NewCLIFormRenderer()

	min := 0
	max := 100
	field := Field{
		Name:     "number",
		Label:    "Number",
		Required: true,
		Validation: &Validation{
			Min: &min,
			Max: &max,
		},
	}

	validator := r.numberValidator(field)

	tests := []struct {
		value   string
		wantErr bool
	}{
		{"", true},      // Required
		{"abc", true},   // Not a number
		{"-1", true},    // Below min
		{"0", false},    // At min
		{"50", false},   // Valid
		{"100", false},  // At max
		{"101", true},   // Above max
		{"3.14", false}, // Float is valid
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			err := validator(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("validator(%q) error = %v, wantErr %v", tt.value, err, tt.wantErr)
			}
		})
	}
}

func TestCLIFormRenderer_dateValidator(t *testing.T) {
	r := NewCLIFormRenderer()

	field := Field{
		Name:     "date",
		Label:    "Date",
		Required: true,
	}

	validator := r.dateValidator(field)

	tests := []struct {
		value   string
		wantErr bool
	}{
		{"", true},           // Required
		{"2024-01-15", false}, // Valid
		{"01-15-2024", true},  // Wrong format
		{"2024/01/15", true},  // Wrong format
		{"invalid", true},     // Invalid
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			err := validator(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("validator(%q) error = %v, wantErr %v", tt.value, err, tt.wantErr)
			}
		})
	}
}

func TestExtractValue(t *testing.T) {
	// String pointer
	s := "hello"
	if extractValue(&s) != "hello" {
		t.Error("string extraction failed")
	}

	// Bool pointer
	b := true
	if extractValue(&b) != true {
		t.Error("bool extraction failed")
	}

	// String slice pointer
	arr := []string{"a", "b"}
	result := extractValue(&arr)
	anyArr, ok := result.([]any)
	if !ok {
		t.Fatalf("expected []any, got %T", result)
	}
	if len(anyArr) != 2 {
		t.Errorf("expected 2 elements, got %d", len(anyArr))
	}
}

func TestCLIFormRenderer_SetAccessible(t *testing.T) {
	r := NewCLIFormRenderer()
	if r.accessible {
		t.Error("expected accessible to be false by default")
	}

	r.SetAccessible(true)
	if !r.accessible {
		t.Error("expected accessible to be true after setting")
	}
}
