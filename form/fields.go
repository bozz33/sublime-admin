package form

import (
	"fmt"
	"html/template"
)

// BaseField contains common logic.
type BaseField struct {
	Name        string
	LabelStr    string
	Value       any
	Placeholder string
	HelpText    string
	Required    bool
	Disabled    bool
	Hidden      bool
	Rules       []string
}

func (b *BaseField) GetName() string                  { return b.Name }
func (b *BaseField) GetLabel() string                 { return b.LabelStr }
func (b *BaseField) GetValue() any                    { return b.Value }
func (b *BaseField) GetPlaceholder() string           { return b.Placeholder }
func (b *BaseField) GetHelp() string                  { return b.HelpText }
func (b *BaseField) IsRequired() bool                 { return b.Required }
func (b *BaseField) IsDisabled() bool                 { return b.Disabled }
func (b *BaseField) IsVisible() bool                  { return !b.Hidden }
func (b *BaseField) GetComponentType() string         { return "field" }
func (b *BaseField) GetAttributes() template.HTMLAttr { return "" }
func (b *BaseField) GetRules() []string               { return b.Rules }

// HasValue returns true if the field has a non-nil value.
func (b *BaseField) HasValue() bool { return b.Value != nil }

// GetValueString returns the value as a string.
func (b *BaseField) GetValueString() string {
	if b.Value == nil {
		return ""
	}
	return fmt.Sprintf("%v", b.Value)
}

// IsChecked returns true if the value is a bool true (for checkbox).
func (b *BaseField) IsChecked() bool {
	if b.Value == nil {
		return false
	}
	if val, ok := b.Value.(bool); ok {
		return val
	}
	return false
}

// TextInput represents a text input field.
type TextInput struct {
	BaseField
	Type string
}

// Text creates a standard text field.
func Text(name string) *TextInput {
	return &TextInput{
		BaseField: BaseField{Name: name, LabelStr: name},
		Type:      "text",
	}
}

// Email creates an email field.
func Email(name string) *TextInput {
	t := Text(name)
	t.Type = "email"
	t.Rules = append(t.Rules, "email")
	return t
}

// Password creates a password field.
func Password(name string) *TextInput {
	t := Text(name)
	t.Type = "password"
	return t
}

// Number creates a numeric field.
func Number(name string) *TextInput {
	t := Text(name)
	t.Type = "number"
	return t
}

// Label sets the field label.
func (f *TextInput) Label(label string) *TextInput {
	f.LabelStr = label
	return f
}

// Placeholder sets the placeholder.
func (f *TextInput) Placeholder(text string) *TextInput {
	f.BaseField.Placeholder = text
	return f
}

// HelperText sets the help text.
func (f *TextInput) HelperText(text string) *TextInput {
	f.HelpText = text
	return f
}

// Required makes the field required.
func (f *TextInput) Required() *TextInput {
	f.BaseField.Required = true
	f.Rules = append(f.Rules, "required")
	return f
}

// Disabled disables the field.
func (f *TextInput) Disabled() *TextInput {
	f.BaseField.Disabled = true
	return f
}

// Default sets the default value.
func (f *TextInput) Default(val any) *TextInput {
	f.Value = val
	return f
}

// TextareaInput represents a textarea field.
type TextareaInput struct {
	BaseField
	RowCount int
}

// Textarea creates a textarea field.
func Textarea(name string) *TextareaInput {
	return &TextareaInput{
		BaseField: BaseField{Name: name, LabelStr: name},
		RowCount:  3,
	}
}

// Label sets the label.
func (t *TextareaInput) Label(label string) *TextareaInput {
	t.LabelStr = label
	return t
}

// Rows sets the number of rows.
func (t *TextareaInput) Rows(rows int) *TextareaInput {
	t.RowCount = rows
	return t
}

// Required makes the field required.
func (t *TextareaInput) Required() *TextareaInput {
	t.BaseField.Required = true
	t.Rules = append(t.Rules, "required")
	return t
}

// SelectOption represents a select option.
type SelectOption struct {
	Label string
	Value string
}

// SelectInput represents a select field.
type SelectInput struct {
	BaseField
	Options []SelectOption
}

// Select creates a select field.
func Select(name string) *SelectInput {
	return &SelectInput{
		BaseField: BaseField{Name: name, LabelStr: name},
		Options:   make([]SelectOption, 0),
	}
}

// SetOptions sets the options.
func (s *SelectInput) SetOptions(options map[string]string) *SelectInput {
	for v, l := range options {
		s.Options = append(s.Options, SelectOption{Value: v, Label: l})
	}
	return s
}

// Label sets the label.
func (s *SelectInput) Label(label string) *SelectInput {
	s.LabelStr = label
	return s
}

// Required makes the field required.
func (s *SelectInput) Required() *SelectInput {
	s.BaseField.Required = true
	s.Rules = append(s.Rules, "required")
	return s
}

// Default sets the default value.
func (s *SelectInput) Default(val any) *SelectInput {
	s.Value = val
	return s
}

// CheckboxInput represents a checkbox field.
type CheckboxInput struct {
	BaseField
}

// Checkbox creates a checkbox field.
func Checkbox(name string) *CheckboxInput {
	return &CheckboxInput{
		BaseField: BaseField{Name: name, LabelStr: name},
	}
}

// Label sets the label.
func (c *CheckboxInput) Label(label string) *CheckboxInput {
	c.LabelStr = label
	return c
}

// Default sets the default value.
func (c *CheckboxInput) Default(val bool) *CheckboxInput {
	c.Value = val
	return c
}

// FileUploadInput represents a file upload field.
type FileUploadInput struct {
	BaseField
	AcceptTypes   string
	MaxFileSize   int64
	AllowMultiple bool
}

// FileUpload creates a file upload field.
func FileUpload(name string) *FileUploadInput {
	return &FileUploadInput{
		BaseField: BaseField{Name: name, LabelStr: name},
	}
}

// Label sets the label.
func (f *FileUploadInput) Label(label string) *FileUploadInput {
	f.LabelStr = label
	return f
}

// Accept sets the accepted file types.
func (f *FileUploadInput) Accept(accept string) *FileUploadInput {
	f.AcceptTypes = accept
	return f
}

// MaxSize sets the maximum size in bytes.
func (f *FileUploadInput) MaxSize(size int64) *FileUploadInput {
	f.MaxFileSize = size
	return f
}

// Multiple allows multiple files.
func (f *FileUploadInput) Multiple() *FileUploadInput {
	f.AllowMultiple = true
	return f
}

// Required makes the field required.
func (f *FileUploadInput) Required() *FileUploadInput {
	f.BaseField.Required = true
	f.Rules = append(f.Rules, "required")
	return f
}

// ---------------------------------------------------------------------------
// Toggle — boolean toggle switch.
// ---------------------------------------------------------------------------

// ToggleInput represents a toggle switch field (boolean, rendered differently from Checkbox).
type ToggleInput struct {
	BaseField
	OnLabel  string
	OffLabel string
}

// Toggle creates a toggle switch field.
func Toggle(name string) *ToggleInput {
	return &ToggleInput{
		BaseField: BaseField{Name: name, LabelStr: name},
		OnLabel:   "Yes",
		OffLabel:  "No",
	}
}

// Label sets the label.
func (t *ToggleInput) Label(label string) *ToggleInput {
	t.LabelStr = label
	return t
}

// Labels sets the on/off labels.
func (t *ToggleInput) Labels(on, off string) *ToggleInput {
	t.OnLabel = on
	t.OffLabel = off
	return t
}

// Default sets the default boolean value.
func (t *ToggleInput) Default(val bool) *ToggleInput {
	t.Value = val
	return t
}

// ---------------------------------------------------------------------------
// Repeater — dynamic multi-entry field.
// ---------------------------------------------------------------------------

// RepeaterField represents a dynamic multi-value field (list of sub-fields).
type RepeaterField struct {
	BaseField
	SubFields []Field
	MinItems  int
	MaxItems  int
	AddLabel  string
}

// Repeater creates a repeater field with the given sub-fields.
func Repeater(name string, subFields ...Field) *RepeaterField {
	return &RepeaterField{
		BaseField: BaseField{Name: name, LabelStr: name},
		SubFields: subFields,
		MinItems:  0,
		MaxItems:  0,
		AddLabel:  "Add item",
	}
}

// Label sets the label.
func (r *RepeaterField) Label(label string) *RepeaterField {
	r.LabelStr = label
	return r
}

// Min sets the minimum number of items.
func (r *RepeaterField) Min(n int) *RepeaterField {
	r.MinItems = n
	return r
}

// Max sets the maximum number of items.
func (r *RepeaterField) Max(n int) *RepeaterField {
	r.MaxItems = n
	return r
}

// AddButtonLabel sets the label for the "add item" button.
func (r *RepeaterField) AddButtonLabel(label string) *RepeaterField {
	r.AddLabel = label
	return r
}

// ---------------------------------------------------------------------------
// RichEditor — WYSIWYG rich-text editor.
// ---------------------------------------------------------------------------

// RichEditorInput represents a rich-text / WYSIWYG editor field.
type RichEditorInput struct {
	BaseField
	Toolbar   []string
	MaxLength int
}

// RichEditor creates a rich editor field.
func RichEditor(name string) *RichEditorInput {
	return &RichEditorInput{
		BaseField: BaseField{Name: name, LabelStr: name},
		Toolbar:   []string{"bold", "italic", "underline", "link", "heading", "list", "image", "code"},
	}
}

// Label sets the label.
func (r *RichEditorInput) Label(label string) *RichEditorInput {
	r.LabelStr = label
	return r
}

// WithToolbar overrides the default toolbar buttons.
func (r *RichEditorInput) WithToolbar(items ...string) *RichEditorInput {
	r.Toolbar = items
	return r
}

// WithMaxLength sets the maximum character count.
func (r *RichEditorInput) WithMaxLength(n int) *RichEditorInput {
	r.MaxLength = n
	return r
}

// Required makes the field required.
func (r *RichEditorInput) Required() *RichEditorInput {
	r.BaseField.Required = true
	r.Rules = append(r.Rules, "required")
	return r
}

// Default sets the default HTML value.
func (r *RichEditorInput) Default(val string) *RichEditorInput {
	r.Value = val
	return r
}

// ---------------------------------------------------------------------------
// MarkdownEditor — Markdown editor with live preview.
// ---------------------------------------------------------------------------

// MarkdownEditorInput represents a Markdown editor field with live preview.
type MarkdownEditorInput struct {
	BaseField
	RowCount int
}

// MarkdownEditor creates a Markdown editor field.
func MarkdownEditor(name string) *MarkdownEditorInput {
	return &MarkdownEditorInput{
		BaseField: BaseField{Name: name, LabelStr: name},
		RowCount:  10,
	}
}

// Label sets the label.
func (m *MarkdownEditorInput) Label(label string) *MarkdownEditorInput {
	m.LabelStr = label
	return m
}

// Rows sets the number of visible rows.
func (m *MarkdownEditorInput) Rows(rows int) *MarkdownEditorInput {
	m.RowCount = rows
	return m
}

// Required makes the field required.
func (m *MarkdownEditorInput) Required() *MarkdownEditorInput {
	m.BaseField.Required = true
	m.Rules = append(m.Rules, "required")
	return m
}

// Default sets the default Markdown value.
func (m *MarkdownEditorInput) Default(val string) *MarkdownEditorInput {
	m.Value = val
	return m
}

// ---------------------------------------------------------------------------
// Tags — multi-value tag/chip input.
// ---------------------------------------------------------------------------

// TagsField represents a tag/chip input field that stores multiple string values.
type TagsField struct {
	BaseField
	Suggestions []string
	MaxTags     int
	Separator   string
}

// Tags creates a tags input field.
func Tags(name string) *TagsField {
	return &TagsField{
		BaseField: BaseField{Name: name, LabelStr: name},
		Separator: ",",
	}
}

// Label sets the label.
func (t *TagsField) Label(label string) *TagsField {
	t.LabelStr = label
	return t
}

// WithSuggestions sets the autocomplete suggestions.
func (t *TagsField) WithSuggestions(suggestions ...string) *TagsField {
	t.Suggestions = suggestions
	return t
}

// WithMaxTags limits the number of tags.
func (t *TagsField) WithMaxTags(n int) *TagsField {
	t.MaxTags = n
	return t
}

// WithSeparator sets the delimiter used in form submission (default ",").
func (t *TagsField) WithSeparator(sep string) *TagsField {
	t.Separator = sep
	return t
}

// Required makes the field required.
func (t *TagsField) Required() *TagsField {
	t.BaseField.Required = true
	t.Rules = append(t.Rules, "required")
	return t
}

// Default sets the default tags.
func (t *TagsField) Default(tags []string) *TagsField {
	t.Value = tags
	return t
}

// TagValues returns the current value as a string slice.
func (t *TagsField) TagValues() []string {
	if v, ok := t.Value.([]string); ok {
		return v
	}
	return nil
}

// ---------------------------------------------------------------------------
// KeyValue — dynamic key-value pair input.
// ---------------------------------------------------------------------------

// KeyValuePair represents a single key-value entry.
type KeyValuePair struct {
	Key   string
	Value string
}

// KeyValueInput represents a dynamic key-value pair input field.
type KeyValueInput struct {
	BaseField
	KeyLabel   string
	ValueLabel string
	MaxPairs   int
	AddLabel   string
}

// KeyValue creates a key-value input field.
func KeyValue(name string) *KeyValueInput {
	return &KeyValueInput{
		BaseField:  BaseField{Name: name, LabelStr: name},
		KeyLabel:   "Key",
		ValueLabel: "Value",
		AddLabel:   "Add pair",
	}
}

// Label sets the label.
func (kv *KeyValueInput) Label(label string) *KeyValueInput {
	kv.LabelStr = label
	return kv
}

// WithLabels sets the key and value column labels.
func (kv *KeyValueInput) WithLabels(keyLabel, valueLabel string) *KeyValueInput {
	kv.KeyLabel = keyLabel
	kv.ValueLabel = valueLabel
	return kv
}

// WithMaxPairs limits the number of pairs.
func (kv *KeyValueInput) WithMaxPairs(n int) *KeyValueInput {
	kv.MaxPairs = n
	return kv
}

// AddButtonLabel sets the label for the "add pair" button.
func (kv *KeyValueInput) AddButtonLabel(label string) *KeyValueInput {
	kv.AddLabel = label
	return kv
}

// Default sets the default pairs.
func (kv *KeyValueInput) Default(pairs []KeyValuePair) *KeyValueInput {
	kv.Value = pairs
	return kv
}

// ---------------------------------------------------------------------------
// ColorPicker — color selection input.
// ---------------------------------------------------------------------------

// ColorPickerInput represents a color picker input field.
type ColorPickerInput struct {
	BaseField
	Swatches []string
}

// ColorPicker creates a color picker field.
func ColorPicker(name string) *ColorPickerInput {
	return &ColorPickerInput{
		BaseField: BaseField{Name: name, LabelStr: name},
	}
}

// Label sets the label.
func (c *ColorPickerInput) Label(label string) *ColorPickerInput {
	c.LabelStr = label
	return c
}

// WithSwatches sets predefined color swatches.
func (c *ColorPickerInput) WithSwatches(colors ...string) *ColorPickerInput {
	c.Swatches = colors
	return c
}

// Required makes the field required.
func (c *ColorPickerInput) Required() *ColorPickerInput {
	c.BaseField.Required = true
	c.Rules = append(c.Rules, "required")
	return c
}

// Default sets the default color (hex string, e.g. "#22c55e").
func (c *ColorPickerInput) Default(hex string) *ColorPickerInput {
	c.Value = hex
	return c
}

// ---------------------------------------------------------------------------
// Slider — range slider input.
// ---------------------------------------------------------------------------

// SliderInput represents a range slider input field.
type SliderInput struct {
	BaseField
	Min  float64
	Max  float64
	Step float64
	Unit string
}

// Slider creates a slider field.
func Slider(name string) *SliderInput {
	return &SliderInput{
		BaseField: BaseField{Name: name, LabelStr: name},
		Min:       0,
		Max:       100,
		Step:      1,
	}
}

// Label sets the label.
func (s *SliderInput) Label(label string) *SliderInput {
	s.LabelStr = label
	return s
}

// Range sets the min and max values.
func (s *SliderInput) Range(min, max float64) *SliderInput {
	s.Min = min
	s.Max = max
	return s
}

// WithStep sets the step increment.
func (s *SliderInput) WithStep(step float64) *SliderInput {
	s.Step = step
	return s
}

// WithUnit sets the display unit suffix.
func (s *SliderInput) WithUnit(unit string) *SliderInput {
	s.Unit = unit
	return s
}

// Default sets the default value.
func (s *SliderInput) Default(val float64) *SliderInput {
	s.Value = val
	return s
}
