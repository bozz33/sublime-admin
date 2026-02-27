package table

import (
	"testing"
	"time"
)

// testRecord is a simple struct used to exercise reflect-based field lookup.
type testRecord struct {
	ID        int
	Name      string
	Email     string
	Active    bool
	Score     float64
	CreatedAt time.Time
	AvatarURL string
	Status    string
	Color     string
}

// ---------------------------------------------------------------------------
// TextColumn tests
// ---------------------------------------------------------------------------

func TestText_Key(t *testing.T) {
	col := Text("name")
	if col.Key() != "name" {
		t.Errorf("expected Key()='name', got '%s'", col.Key())
	}
}

func TestText_Label_default(t *testing.T) {
	col := Text("name")
	if col.Label() != "name" {
		t.Errorf("expected default Label()='name', got '%s'", col.Label())
	}
}

func TestText_Label_custom(t *testing.T) {
	col := Text("name").WithLabel("Full Name")
	if col.Label() != "Full Name" {
		t.Errorf("expected Label()='Full Name', got '%s'", col.Label())
	}
}

func TestText_Type(t *testing.T) {
	col := Text("name")
	if col.Type() != "text" {
		t.Errorf("expected Type()='text', got '%s'", col.Type())
	}
}

func TestText_IsSortable_default(t *testing.T) {
	col := Text("name")
	if col.IsSortable() {
		t.Error("expected IsSortable()=false by default")
	}
}

func TestText_IsSortable_after_Sortable(t *testing.T) {
	col := Text("name").Sortable()
	if !col.IsSortable() {
		t.Error("expected IsSortable()=true after Sortable()")
	}
}

func TestText_IsSearchable_default(t *testing.T) {
	col := Text("name")
	if col.IsSearchable() {
		t.Error("expected IsSearchable()=false by default")
	}
}

func TestText_IsSearchable_after_Searchable(t *testing.T) {
	col := Text("name").Searchable()
	if !col.IsSearchable() {
		t.Error("expected IsSearchable()=true after Searchable()")
	}
}

func TestText_IsCopyable_default(t *testing.T) {
	col := Text("name")
	if col.IsCopyable() {
		t.Error("expected IsCopyable()=false by default")
	}
}

func TestText_IsCopyable_after_Copyable(t *testing.T) {
	col := Text("name").Copyable()
	if !col.IsCopyable() {
		t.Error("expected IsCopyable()=true after Copyable()")
	}
}

func TestText_Value_reflect_struct(t *testing.T) {
	col := Text("Name")
	rec := testRecord{Name: "Alice"}

	got := col.Value(rec)
	if got != "Alice" {
		t.Errorf("expected Value()='Alice', got '%s'", got)
	}
}

func TestText_Value_reflect_ptr_struct(t *testing.T) {
	col := Text("Name")
	rec := &testRecord{Name: "Bob"}

	got := col.Value(rec)
	if got != "Bob" {
		t.Errorf("expected Value()='Bob', got '%s'", got)
	}
}

func TestText_Value_missing_field(t *testing.T) {
	col := Text("DoesNotExist")
	rec := testRecord{Name: "Alice"}

	got := col.Value(rec)
	if got != "" {
		t.Errorf("expected Value()='' for missing field, got '%s'", got)
	}
}

func TestText_Value_using_custom_func(t *testing.T) {
	col := Text("Name").Using(func(item any) string {
		r := item.(testRecord)
		return "custom:" + r.Name
	})
	rec := testRecord{Name: "Charlie"}

	got := col.Value(rec)
	if got != "custom:Charlie" {
		t.Errorf("expected 'custom:Charlie', got '%s'", got)
	}
}

func TestText_Render_not_nil(t *testing.T) {
	col := Text("name")
	if col.Render("hello", nil) == nil {
		t.Error("expected Render() to return a non-nil component")
	}
}

// ---------------------------------------------------------------------------
// BadgeColumn tests
// ---------------------------------------------------------------------------

func TestBadge_Key(t *testing.T) {
	col := Badge("Status")
	if col.Key() != "Status" {
		t.Errorf("expected Key()='Status', got '%s'", col.Key())
	}
}

func TestBadge_Type(t *testing.T) {
	col := Badge("Status")
	if col.Type() != "badge" {
		t.Errorf("expected Type()='badge', got '%s'", col.Type())
	}
}

func TestBadge_GetColor_mapped(t *testing.T) {
	col := Badge("Status").Colors(map[string]string{
		"active":   "green",
		"inactive": "red",
		"pending":  "yellow",
	})

	if col.GetColor("active") != "green" {
		t.Errorf("expected GetColor('active')='green', got '%s'", col.GetColor("active"))
	}
	if col.GetColor("inactive") != "red" {
		t.Errorf("expected GetColor('inactive')='red', got '%s'", col.GetColor("inactive"))
	}
	if col.GetColor("pending") != "yellow" {
		t.Errorf("expected GetColor('pending')='yellow', got '%s'", col.GetColor("pending"))
	}
}

func TestBadge_GetColor_fallback(t *testing.T) {
	col := Badge("Status")
	// Unmapped value falls back to "primary"
	if col.GetColor("unknown") != "primary" {
		t.Errorf("expected fallback color='primary', got '%s'", col.GetColor("unknown"))
	}
}

func TestBadge_IsSortable_default(t *testing.T) {
	col := Badge("Status")
	if col.IsSortable() {
		t.Error("expected IsSortable()=false by default")
	}
}

func TestBadge_IsSearchable(t *testing.T) {
	col := Badge("Status")
	if col.IsSearchable() {
		t.Error("expected IsSearchable()=false for BadgeColumn")
	}
}

func TestBadge_IsCopyable(t *testing.T) {
	col := Badge("Status")
	if col.IsCopyable() {
		t.Error("expected IsCopyable()=false for BadgeColumn")
	}
}

func TestBadge_Render_not_nil(t *testing.T) {
	col := Badge("Status")
	if col.Render("active", nil) == nil {
		t.Error("expected Render() to return a non-nil component")
	}
}

// ---------------------------------------------------------------------------
// BooleanColumn tests
// ---------------------------------------------------------------------------

func TestBoolCol_Type(t *testing.T) {
	col := BoolCol("Active")
	if col.Type() != "boolean" {
		t.Errorf("expected Type()='boolean', got '%s'", col.Type())
	}
}

func TestBoolCol_DefaultLabels(t *testing.T) {
	col := BoolCol("Active")
	if col.TrueLabel != "Yes" {
		t.Errorf("expected TrueLabel='Yes', got '%s'", col.TrueLabel)
	}
	if col.FalseLabel != "No" {
		t.Errorf("expected FalseLabel='No', got '%s'", col.FalseLabel)
	}
}

func TestBoolCol_CustomLabels(t *testing.T) {
	col := BoolCol("Active").Labels("Enabled", "Disabled")
	if col.TrueLabel != "Enabled" {
		t.Errorf("expected TrueLabel='Enabled', got '%s'", col.TrueLabel)
	}
	if col.FalseLabel != "Disabled" {
		t.Errorf("expected FalseLabel='Disabled', got '%s'", col.FalseLabel)
	}
}

func TestBoolCol_Value_true(t *testing.T) {
	col := BoolCol("Active")
	rec := testRecord{Active: true}

	got := col.Value(rec)
	if got != "Yes" {
		t.Errorf("expected Value()='Yes' for true, got '%s'", got)
	}
}

func TestBoolCol_Value_false(t *testing.T) {
	col := BoolCol("Active")
	rec := testRecord{Active: false}

	got := col.Value(rec)
	if got != "No" {
		t.Errorf("expected Value()='No' for false, got '%s'", got)
	}
}

func TestBoolCol_Render_not_nil(t *testing.T) {
	col := BoolCol("Active")
	if col.Render("Yes", nil) == nil {
		t.Error("expected Render() to return a non-nil component")
	}
}

// ---------------------------------------------------------------------------
// DateColumn tests
// ---------------------------------------------------------------------------

func TestDateCol_Type(t *testing.T) {
	col := DateCol("CreatedAt")
	if col.Type() != "date" {
		t.Errorf("expected Type()='date', got '%s'", col.Type())
	}
}

func TestDateCol_DefaultFormat(t *testing.T) {
	col := DateCol("CreatedAt")
	if col.Format != "2006-01-02" {
		t.Errorf("expected default Format='2006-01-02', got '%s'", col.Format)
	}
}

func TestDateCol_CustomFormat(t *testing.T) {
	col := DateCol("CreatedAt").DateFormat("02/01/2006")
	if col.Format != "02/01/2006" {
		t.Errorf("expected Format='02/01/2006', got '%s'", col.Format)
	}
}

func TestDateCol_Value_time_Time(t *testing.T) {
	col := DateCol("CreatedAt")
	ts := time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)
	rec := testRecord{CreatedAt: ts}

	got := col.Value(rec)
	if got != "2024-03-15" {
		t.Errorf("expected Value()='2024-03-15', got '%s'", got)
	}
}

func TestDateCol_Value_zero_time(t *testing.T) {
	col := DateCol("CreatedAt")
	rec := testRecord{CreatedAt: time.Time{}}

	got := col.Value(rec)
	if got != "" {
		t.Errorf("expected Value()='' for zero time, got '%s'", got)
	}
}

func TestDateCol_IsSortable_default(t *testing.T) {
	col := DateCol("CreatedAt")
	if col.IsSortable() {
		t.Error("expected IsSortable()=false by default")
	}
}

func TestDateCol_Render_not_nil(t *testing.T) {
	col := DateCol("CreatedAt")
	if col.Render("2024-03-15", nil) == nil {
		t.Error("expected Render() to return a non-nil component")
	}
}

// ---------------------------------------------------------------------------
// ImageColumn tests
// ---------------------------------------------------------------------------

func TestImage_Type(t *testing.T) {
	col := Image("AvatarURL")
	if col.Type() != "image" {
		t.Errorf("expected Type()='image', got '%s'", col.Type())
	}
}

func TestImage_Rounded_default(t *testing.T) {
	col := Image("AvatarURL")
	if col.Rounded {
		t.Error("expected Rounded=false by default")
	}
}

func TestImage_Round(t *testing.T) {
	col := Image("AvatarURL").Round()
	if !col.Rounded {
		t.Error("expected Rounded=true after Round()")
	}
}

func TestImage_Circular(t *testing.T) {
	col := Image("AvatarURL").Circular()
	if !col.IsCircular {
		t.Error("expected IsCircular=true after Circular()")
	}
	if !col.Rounded {
		t.Error("expected Rounded=true after Circular()")
	}
}

func TestImage_IsSortable(t *testing.T) {
	col := Image("AvatarURL")
	if col.IsSortable() {
		t.Error("expected IsSortable()=false for ImageColumn")
	}
}

func TestImage_IsSearchable(t *testing.T) {
	col := Image("AvatarURL")
	if col.IsSearchable() {
		t.Error("expected IsSearchable()=false for ImageColumn")
	}
}

func TestImage_IsCopyable(t *testing.T) {
	col := Image("AvatarURL")
	if col.IsCopyable() {
		t.Error("expected IsCopyable()=false for ImageColumn")
	}
}

func TestImage_Render_not_nil(t *testing.T) {
	col := Image("AvatarURL")
	if col.Render("/img/avatar.jpg", nil) == nil {
		t.Error("expected Render() to return a non-nil component")
	}
}

// ---------------------------------------------------------------------------
// Inline editing column type tests
// ---------------------------------------------------------------------------

func TestTextInputColumn_Type(t *testing.T) {
	col := TextInput("Score")
	if col.Type() != "text_input_col" {
		t.Errorf("expected Type()='text_input_col', got '%s'", col.Type())
	}
}

func TestTextInputColumn_IsSortable_false(t *testing.T) {
	col := TextInput("Score")
	if col.IsSortable() {
		t.Error("expected IsSortable()=false by default")
	}
}

func TestTextInputColumn_IsSortable_true(t *testing.T) {
	col := TextInput("Score").Sortable()
	if !col.IsSortable() {
		t.Error("expected IsSortable()=true after Sortable()")
	}
}

func TestTextInputColumn_IsSearchable(t *testing.T) {
	col := TextInput("Score")
	if col.IsSearchable() {
		t.Error("expected IsSearchable()=false for TextInputColumn")
	}
}

func TestTextInputColumn_IsCopyable(t *testing.T) {
	col := TextInput("Score")
	if col.IsCopyable() {
		t.Error("expected IsCopyable()=false for TextInputColumn")
	}
}

func TestSelectColumn_Type(t *testing.T) {
	col := SelectCol("Status")
	if col.Type() != "select_col" {
		t.Errorf("expected Type()='select_col', got '%s'", col.Type())
	}
}

func TestSelectColumn_IsSearchable(t *testing.T) {
	col := SelectCol("Status")
	if col.IsSearchable() {
		t.Error("expected IsSearchable()=false for SelectColumn")
	}
}

func TestSelectColumn_IsCopyable(t *testing.T) {
	col := SelectCol("Status")
	if col.IsCopyable() {
		t.Error("expected IsCopyable()=false for SelectColumn")
	}
}

func TestToggleColumn_Type(t *testing.T) {
	col := Toggle("Active")
	if col.Type() != "toggle_col" {
		t.Errorf("expected Type()='toggle_col', got '%s'", col.Type())
	}
}

func TestToggleColumn_IsSearchable(t *testing.T) {
	col := Toggle("Active")
	if col.IsSearchable() {
		t.Error("expected IsSearchable()=false for ToggleColumn")
	}
}

func TestToggleColumn_IsCopyable(t *testing.T) {
	col := Toggle("Active")
	if col.IsCopyable() {
		t.Error("expected IsCopyable()=false for ToggleColumn")
	}
}

func TestCheckboxColumn_Type(t *testing.T) {
	col := Checkbox("Selected")
	if col.Type() != "checkbox_col" {
		t.Errorf("expected Type()='checkbox_col', got '%s'", col.Type())
	}
}

func TestCheckboxColumn_IsSearchable(t *testing.T) {
	col := Checkbox("Selected")
	if col.IsSearchable() {
		t.Error("expected IsSearchable()=false for CheckboxColumn")
	}
}

func TestCheckboxColumn_IsCopyable(t *testing.T) {
	col := Checkbox("Selected")
	if col.IsCopyable() {
		t.Error("expected IsCopyable()=false for CheckboxColumn")
	}
}

// ---------------------------------------------------------------------------
// Text column transformation tests
// ---------------------------------------------------------------------------

func TestText_Limit_truncation(t *testing.T) {
	col := Text("Name").Limit(5)

	// Supply a value longer than 5 chars via ValueFunc so we can test the
	// transform on applyTextTransforms directly.
	raw := "Hello World"
	result := applyTextTransforms(raw, col)
	// Should be truncated to 5 runes + ellipsis
	runes := []rune(result)
	if len(runes) != 6 { // 5 chars + "…"
		t.Errorf("expected 6 runes after Limit(5), got %d: %s", len(runes), result)
	}
}

func TestText_Prefix_Suffix(t *testing.T) {
	col := Text("Score").Prefix("$").Suffix("USD")

	if col.PrefixStr != "$" {
		t.Errorf("expected PrefixStr='$', got '%s'", col.PrefixStr)
	}
	if col.SuffixStr != "USD" {
		t.Errorf("expected SuffixStr='USD', got '%s'", col.SuffixStr)
	}
}

func TestText_Badge_flag(t *testing.T) {
	col := Text("Status").Badge()
	if !col.IsBadge {
		t.Error("expected IsBadge=true after Badge()")
	}
}

func TestText_StrikeThrough_flag(t *testing.T) {
	col := Text("Name").WithStrikeThrough()
	if !col.StrikeThrough {
		t.Error("expected StrikeThrough=true after WithStrikeThrough()")
	}
}

func TestText_Wrap_flag(t *testing.T) {
	col := Text("Description").WithWrap()
	if !col.Wrap {
		t.Error("expected Wrap=true after WithWrap()")
	}
}
