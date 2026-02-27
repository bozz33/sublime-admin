package form

import (
	"testing"
)

// ---------------------------------------------------------------------------
// BaseField tests
// ---------------------------------------------------------------------------

func TestBaseField_Name(t *testing.T) {
	f := Text("username")
	if f.Name() != "username" {
		t.Errorf("expected Name()='username', got '%s'", f.Name())
	}
}

func TestBaseField_Label(t *testing.T) {
	f := Text("username")
	// Default label equals the field name.
	// Use GetLabel() since TextInput.Label(string) shadows BaseField.Label().
	if f.GetLabel() != "username" {
		t.Errorf("expected GetLabel()='username', got '%s'", f.GetLabel())
	}
}

func TestBaseField_Value_nil(t *testing.T) {
	f := Text("field")
	if f.Value() != nil {
		t.Errorf("expected Value()=nil, got %v", f.Value())
	}
}

func TestBaseField_IsRequired_default(t *testing.T) {
	f := Text("field")
	if f.IsRequired() {
		t.Error("expected IsRequired()=false by default")
	}
}

func TestBaseField_IsDisabled_default(t *testing.T) {
	f := Text("field")
	if f.IsDisabled() {
		t.Error("expected IsDisabled()=false by default")
	}
}

func TestBaseField_IsVisible_default(t *testing.T) {
	f := Text("field")
	if !f.IsVisible() {
		t.Error("expected IsVisible()=true by default")
	}
}

func TestBaseField_ValueString_nil(t *testing.T) {
	f := Text("field")
	if f.ValueString() != "" {
		t.Errorf("expected ValueString()='', got '%s'", f.ValueString())
	}
}

func TestBaseField_ValueString_string(t *testing.T) {
	f := Text("field").Default("hello")
	if f.ValueString() != "hello" {
		t.Errorf("expected ValueString()='hello', got '%s'", f.ValueString())
	}
}

func TestBaseField_ValueString_int(t *testing.T) {
	f := Number("age").Default(42)
	if f.ValueString() != "42" {
		t.Errorf("expected ValueString()='42', got '%s'", f.ValueString())
	}
}

func TestBaseField_IsChecked_nilValue(t *testing.T) {
	f := Checkbox("active")
	if f.IsChecked() {
		t.Error("expected IsChecked()=false when value is nil")
	}
}

func TestBaseField_IsChecked_true(t *testing.T) {
	f := Checkbox("active").Default(true)
	if !f.IsChecked() {
		t.Error("expected IsChecked()=true")
	}
}

func TestBaseField_IsChecked_false(t *testing.T) {
	f := Checkbox("active").Default(false)
	if f.IsChecked() {
		t.Error("expected IsChecked()=false")
	}
}

// ---------------------------------------------------------------------------
// TextInput tests
// ---------------------------------------------------------------------------

func TestTextInput_Text_constructor(t *testing.T) {
	f := Text("title")

	if f.Name() != "title" {
		t.Errorf("expected Name()='title', got '%s'", f.Name())
	}
	if f.Type != "text" {
		t.Errorf("expected Type='text', got '%s'", f.Type)
	}
	// Default label equals the field name
	if f.LabelStr != "title" {
		t.Errorf("expected default LabelStr='title', got '%s'", f.LabelStr)
	}
}

func TestTextInput_Email_constructor(t *testing.T) {
	f := Email("email")

	if f.Type != "email" {
		t.Errorf("expected Type='email', got '%s'", f.Type)
	}
	// Email auto-adds "email" validation rule
	rules := f.Rules()
	if len(rules) == 0 {
		t.Fatal("expected at least one validation rule for Email field")
	}
	if rules[0] != "email" {
		t.Errorf("expected first rule='email', got '%s'", rules[0])
	}
}

func TestTextInput_Password_constructor(t *testing.T) {
	f := Password("pass")
	if f.Type != "password" {
		t.Errorf("expected Type='password', got '%s'", f.Type)
	}
}

func TestTextInput_Number_constructor(t *testing.T) {
	f := Number("qty")
	if f.Type != "number" {
		t.Errorf("expected Type='number', got '%s'", f.Type)
	}
}

func TestTextInput_Label(t *testing.T) {
	f := Text("name").Label("Full Name")
	if f.LabelStr != "Full Name" {
		t.Errorf("expected LabelStr='Full Name', got '%s'", f.LabelStr)
	}
}

func TestTextInput_Required(t *testing.T) {
	f := Text("name").Required()

	if !f.IsRequired() {
		t.Error("expected IsRequired()=true after Required()")
	}
	// Required() also appends "required" to the rules
	found := false
	for _, r := range f.Rules() {
		if r == "required" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'required' rule to be in Rules() after Required()")
	}
}

func TestTextInput_Default(t *testing.T) {
	f := Text("name").Default("Alice")
	if f.Value() != "Alice" {
		t.Errorf("expected Value()='Alice', got '%v'", f.Value())
	}
}

func TestTextInput_Disabled(t *testing.T) {
	f := Text("name").Disabled()
	if !f.IsDisabled() {
		t.Error("expected IsDisabled()=true after Disabled()")
	}
}

func TestTextInput_WithLiveValidation(t *testing.T) {
	url := "/users/validate-field"
	f := Text("email").WithLiveValidation(url)
	if f.LiveValidateURL != url {
		t.Errorf("expected LiveValidateURL='%s', got '%s'", url, f.LiveValidateURL)
	}
}

func TestTextInput_Render_not_nil(t *testing.T) {
	f := Text("name")
	if f.Render() == nil {
		t.Error("expected Render() to return a non-nil component")
	}
}

// ---------------------------------------------------------------------------
// TextareaInput tests
// ---------------------------------------------------------------------------

func TestTextareaInput_Textarea_constructor(t *testing.T) {
	f := Textarea("bio")

	if f.Name() != "bio" {
		t.Errorf("expected Name()='bio', got '%s'", f.Name())
	}
	// Default row count is 3
	if f.RowCount != 3 {
		t.Errorf("expected default RowCount=3, got %d", f.RowCount)
	}
}

func TestTextareaInput_Rows(t *testing.T) {
	f := Textarea("bio").Rows(8)
	if f.RowCount != 8 {
		t.Errorf("expected RowCount=8, got %d", f.RowCount)
	}
}

func TestTextareaInput_Required(t *testing.T) {
	f := Textarea("bio").Required()
	if !f.IsRequired() {
		t.Error("expected IsRequired()=true after Required()")
	}
}

func TestTextareaInput_Render_not_nil(t *testing.T) {
	f := Textarea("bio")
	if f.Render() == nil {
		t.Error("expected Render() to return a non-nil component")
	}
}

// ---------------------------------------------------------------------------
// SelectInput tests
// ---------------------------------------------------------------------------

func TestSelectInput_Select_constructor(t *testing.T) {
	f := Select("role")

	if f.Name() != "role" {
		t.Errorf("expected Name()='role', got '%s'", f.Name())
	}
	if len(f.SelectOptions()) != 0 {
		t.Errorf("expected 0 options by default, got %d", len(f.SelectOptions()))
	}
}

func TestSelectInput_Options(t *testing.T) {
	opts := map[string]string{
		"admin": "Administrator",
		"user":  "User",
	}
	f := Select("role").Options(opts)

	if len(f.SelectOptions()) != 2 {
		t.Errorf("expected 2 options, got %d", len(f.SelectOptions()))
	}
}

func TestSelectInput_Required(t *testing.T) {
	f := Select("role").Required()
	if !f.IsRequired() {
		t.Error("expected IsRequired()=true after Required()")
	}
}

func TestSelectInput_Default(t *testing.T) {
	f := Select("role").Default("admin")
	if f.Value() != "admin" {
		t.Errorf("expected Value()='admin', got '%v'", f.Value())
	}
}

func TestSelectInput_SelectOptions_returns_slice(t *testing.T) {
	f := Select("status").Options(map[string]string{
		"active":   "Active",
		"inactive": "Inactive",
		"pending":  "Pending",
	})

	opts := f.SelectOptions()
	if len(opts) != 3 {
		t.Errorf("expected 3 SelectOptions, got %d", len(opts))
	}

	// Every option must have non-empty Value and Label
	for _, o := range opts {
		if o.Value == "" {
			t.Error("option Value must not be empty")
		}
		if o.Label == "" {
			t.Error("option Label must not be empty")
		}
	}
}

func TestSelectInput_Render_not_nil(t *testing.T) {
	f := Select("role")
	if f.Render() == nil {
		t.Error("expected Render() to return a non-nil component")
	}
}

// ---------------------------------------------------------------------------
// CheckboxInput tests
// ---------------------------------------------------------------------------

func TestCheckboxInput_Checkbox_constructor(t *testing.T) {
	f := Checkbox("agree")

	if f.Name() != "agree" {
		t.Errorf("expected Name()='agree', got '%s'", f.Name())
	}
}

func TestCheckboxInput_Default_true(t *testing.T) {
	f := Checkbox("active").Default(true)
	if !f.IsChecked() {
		t.Error("expected IsChecked()=true after Default(true)")
	}
}

func TestCheckboxInput_Default_false(t *testing.T) {
	f := Checkbox("active").Default(false)
	if f.IsChecked() {
		t.Error("expected IsChecked()=false after Default(false)")
	}
}

func TestCheckboxInput_Render_not_nil(t *testing.T) {
	f := Checkbox("agree")
	if f.Render() == nil {
		t.Error("expected Render() to return a non-nil component")
	}
}

// ---------------------------------------------------------------------------
// HiddenField tests
// ---------------------------------------------------------------------------

func TestHiddenField_Hidden_constructor(t *testing.T) {
	f := Hidden("user_id", 42)

	if f.Name() != "user_id" {
		t.Errorf("expected Name()='user_id', got '%s'", f.Name())
	}
	if f.Value() != 42 {
		t.Errorf("expected Value()=42, got '%v'", f.Value())
	}
}

func TestHiddenField_IsVisible_false(t *testing.T) {
	f := Hidden("token", "abc123")
	if f.IsVisible() {
		t.Error("expected IsVisible()=false for HiddenField")
	}
}

func TestHiddenField_ValueString(t *testing.T) {
	f := Hidden("token", "abc123")
	if f.ValueString() != "abc123" {
		t.Errorf("expected ValueString()='abc123', got '%s'", f.ValueString())
	}
}

func TestHiddenField_Render_not_nil(t *testing.T) {
	f := Hidden("id", 1)
	if f.Render() == nil {
		t.Error("expected Render() to return a non-nil component")
	}
}

// ---------------------------------------------------------------------------
// ToggleInput tests
// ---------------------------------------------------------------------------

func TestToggleInput_Toggle_constructor(t *testing.T) {
	f := Toggle("published")

	if f.Name() != "published" {
		t.Errorf("expected Name()='published', got '%s'", f.Name())
	}
	// Default labels
	if f.OnLabel != "Yes" {
		t.Errorf("expected OnLabel='Yes', got '%s'", f.OnLabel)
	}
	if f.OffLabel != "No" {
		t.Errorf("expected OffLabel='No', got '%s'", f.OffLabel)
	}
}

func TestToggleInput_Labels(t *testing.T) {
	f := Toggle("active").Labels("Enabled", "Disabled")

	if f.OnLabel != "Enabled" {
		t.Errorf("expected OnLabel='Enabled', got '%s'", f.OnLabel)
	}
	if f.OffLabel != "Disabled" {
		t.Errorf("expected OffLabel='Disabled', got '%s'", f.OffLabel)
	}
}

func TestToggleInput_Default(t *testing.T) {
	f := Toggle("published").Default(true)
	if !f.IsChecked() {
		t.Error("expected IsChecked()=true after Default(true)")
	}
}

func TestToggleInput_Render_not_nil(t *testing.T) {
	f := Toggle("active")
	if f.Render() == nil {
		t.Error("expected Render() to return a non-nil component")
	}
}

// ---------------------------------------------------------------------------
// RadioInput tests
// ---------------------------------------------------------------------------

func TestRadioInput_Radio_constructor(t *testing.T) {
	f := Radio("gender")

	if f.Name() != "gender" {
		t.Errorf("expected Name()='gender', got '%s'", f.Name())
	}
	if len(f.RadioOptions()) != 0 {
		t.Errorf("expected 0 radio options by default, got %d", len(f.RadioOptions()))
	}
}

func TestRadioInput_Options(t *testing.T) {
	f := Radio("gender").Options(map[string]string{
		"m": "Male",
		"f": "Female",
	})

	if len(f.RadioOptions()) != 2 {
		t.Errorf("expected 2 radio options, got %d", len(f.RadioOptions()))
	}
}

func TestRadioInput_OptionsOrdered(t *testing.T) {
	opts := []RadioOption{
		{Value: "low", Label: "Low"},
		{Value: "medium", Label: "Medium"},
		{Value: "high", Label: "High"},
	}
	f := Radio("priority").OptionsOrdered(opts)

	got := f.RadioOptions()
	if len(got) != 3 {
		t.Fatalf("expected 3 options, got %d", len(got))
	}
	if got[0].Value != "low" {
		t.Errorf("expected first option Value='low', got '%s'", got[0].Value)
	}
	if got[2].Value != "high" {
		t.Errorf("expected last option Value='high', got '%s'", got[2].Value)
	}
}

func TestRadioInput_Required(t *testing.T) {
	f := Radio("size").Required()
	if !f.IsRequired() {
		t.Error("expected IsRequired()=true after Required()")
	}
}

func TestRadioInput_Render_not_nil(t *testing.T) {
	f := Radio("size")
	if f.Render() == nil {
		t.Error("expected Render() to return a non-nil component")
	}
}

// ---------------------------------------------------------------------------
// CheckboxListInput tests
// ---------------------------------------------------------------------------

func TestCheckboxListInput_CheckboxList_constructor(t *testing.T) {
	f := CheckboxList("permissions")

	if f.Name() != "permissions" {
		t.Errorf("expected Name()='permissions', got '%s'", f.Name())
	}
	if len(f.CheckboxOptions()) != 0 {
		t.Errorf("expected 0 options by default, got %d", len(f.CheckboxOptions()))
	}
}

func TestCheckboxListInput_Options(t *testing.T) {
	f := CheckboxList("roles").Options(map[string]string{
		"admin":  "Admin",
		"editor": "Editor",
		"viewer": "Viewer",
	})

	if len(f.CheckboxOptions()) != 3 {
		t.Errorf("expected 3 options, got %d", len(f.CheckboxOptions()))
	}
}

func TestCheckboxListInput_Default_IsSelected(t *testing.T) {
	f := CheckboxList("tags").
		Options(map[string]string{"go": "Go", "rust": "Rust", "py": "Python"}).
		Default([]string{"go", "rust"})

	if !f.IsSelected("go") {
		t.Error("expected 'go' to be selected")
	}
	if !f.IsSelected("rust") {
		t.Error("expected 'rust' to be selected")
	}
	if f.IsSelected("py") {
		t.Error("expected 'py' to NOT be selected")
	}
}

func TestCheckboxListInput_Required(t *testing.T) {
	f := CheckboxList("items").Required()
	if !f.IsRequired() {
		t.Error("expected IsRequired()=true after Required()")
	}
}

func TestCheckboxListInput_Render_not_nil(t *testing.T) {
	f := CheckboxList("roles")
	if f.Render() == nil {
		t.Error("expected Render() to return a non-nil component")
	}
}

// ---------------------------------------------------------------------------
// Fluent chaining / combined tests
// ---------------------------------------------------------------------------

func TestTextInput_FluentChain(t *testing.T) {
	f := Text("email").
		Label("Email Address").
		Required().
		Disabled().
		Default("user@example.com").
		WithLiveValidation("/validate")

	if f.GetLabel() != "Email Address" {
		t.Errorf("expected GetLabel()='Email Address', got '%s'", f.GetLabel())
	}
	if !f.IsRequired() {
		t.Error("expected IsRequired()=true")
	}
	if !f.IsDisabled() {
		t.Error("expected IsDisabled()=true")
	}
	if f.ValueString() != "user@example.com" {
		t.Errorf("expected ValueString()='user@example.com', got '%s'", f.ValueString())
	}
	if f.LiveValidateURL != "/validate" {
		t.Errorf("expected LiveValidateURL='/validate', got '%s'", f.LiveValidateURL)
	}
}

func TestRulesString(t *testing.T) {
	f := Email("email").Required()
	rs := f.RulesString()

	if rs == "" {
		t.Error("expected non-empty RulesString()")
	}
	// Both "email" and "required" rules should appear
	hasEmail := false
	hasRequired := false
	for _, r := range f.Rules() {
		if r == "email" {
			hasEmail = true
		}
		if r == "required" {
			hasRequired = true
		}
	}
	if !hasEmail {
		t.Error("expected 'email' rule")
	}
	if !hasRequired {
		t.Error("expected 'required' rule")
	}
}

func TestBaseField_HasValue(t *testing.T) {
	f := Text("x")
	if f.HasValue() {
		t.Error("expected HasValue()=false when no default is set")
	}

	f.Default("something")
	if !f.HasValue() {
		t.Error("expected HasValue()=true after Default()")
	}
}
