package table

// SelectFilter represents a select type filter
type SelectFilter struct {
	filterKey string
	LabelStr  string
	Options   []FilterOption
}

// Select creates a new select filter
func Select(key string) *SelectFilter {
	return &SelectFilter{
		filterKey: key,
		LabelStr:  key,
		Options:   make([]FilterOption, 0),
	}
}

// WithLabel sets the filter label
func (f *SelectFilter) WithLabel(label string) *SelectFilter {
	f.LabelStr = label
	return f
}

// WithOptions sets the filter options
func (f *SelectFilter) WithOptions(options []FilterOption) *SelectFilter {
	f.Options = options
	return f
}

// Implementation of the Filter interface
func (f *SelectFilter) Key() string                   { return f.filterKey }
func (f *SelectFilter) Label() string                 { return f.LabelStr }
func (f *SelectFilter) Type() string                  { return "select" }
func (f *SelectFilter) FilterOptions() []FilterOption { return f.Options }

// BooleanFilter represents a boolean filter
type BooleanFilter struct {
	filterKey string
	LabelStr  string
}

// Boolean creates a new boolean filter
func Boolean(key string) *BooleanFilter {
	return &BooleanFilter{
		filterKey: key,
		LabelStr:  key,
	}
}

// WithLabel sets the filter label
func (f *BooleanFilter) WithLabel(label string) *BooleanFilter {
	f.LabelStr = label
	return f
}

// Implementation of the Filter interface
func (f *BooleanFilter) Key() string   { return f.filterKey }
func (f *BooleanFilter) Label() string { return f.LabelStr }
func (f *BooleanFilter) Type() string  { return "boolean" }
func (f *BooleanFilter) FilterOptions() []FilterOption {
	return []FilterOption{
		{Value: "true", Label: "Yes"},
		{Value: "false", Label: "No"},
	}
}

// DateFilter represents a date range filter with from/until inputs.
type DateFilter struct {
	filterKey string
	LabelStr  string
}

// DateRange creates a new date range filter.
func DateRange(key string) *DateFilter {
	return &DateFilter{
		filterKey: key,
		LabelStr:  key,
	}
}

// WithLabel sets the filter label.
func (f *DateFilter) WithLabel(label string) *DateFilter {
	f.LabelStr = label
	return f
}

func (f *DateFilter) Key() string                   { return f.filterKey }
func (f *DateFilter) Label() string                 { return f.LabelStr }
func (f *DateFilter) Type() string                  { return "date" }
func (f *DateFilter) FilterOptions() []FilterOption { return nil }

// TextFilter represents a free-text search filter on a specific field.
type TextFilter struct {
	filterKey string
	LabelStr  string
}

// TextSearch creates a new text search filter.
func TextSearch(key string) *TextFilter {
	return &TextFilter{
		filterKey: key,
		LabelStr:  key,
	}
}

// WithLabel sets the filter label.
func (f *TextFilter) WithLabel(label string) *TextFilter {
	f.LabelStr = label
	return f
}

func (f *TextFilter) Key() string                   { return f.filterKey }
func (f *TextFilter) Label() string                 { return f.LabelStr }
func (f *TextFilter) Type() string                  { return "text" }
func (f *TextFilter) FilterOptions() []FilterOption { return nil }

// CustomFilterField describes a single field inside a custom filter form.
type CustomFilterField struct {
	Name        string
	Label       string
	Type        string // "text", "select", "date", "number"
	Placeholder string
	Options     []FilterOption // for Type="select"
}

// CustomFilter represents a filter with a fully custom form (multiple fields).
type CustomFilter struct {
	filterKey string
	LabelStr  string
	Fields    []CustomFilterField
	ApplyFunc func(params map[string]string) string // returns a query condition string (optional)
}

// Custom creates a new custom filter.
func Custom(key string) *CustomFilter {
	return &CustomFilter{
		filterKey: key,
		LabelStr:  key,
		Fields:    make([]CustomFilterField, 0),
	}
}

// WithLabel sets the filter label.
func (f *CustomFilter) WithLabel(label string) *CustomFilter {
	f.LabelStr = label
	return f
}

// WithFields sets the custom form fields.
func (f *CustomFilter) WithFields(fields ...CustomFilterField) *CustomFilter {
	f.Fields = append(f.Fields, fields...)
	return f
}

// WithApply sets a function that converts filter params to a query condition.
func (f *CustomFilter) WithApply(fn func(map[string]string) string) *CustomFilter {
	f.ApplyFunc = fn
	return f
}

func (f *CustomFilter) Key() string                   { return f.filterKey }
func (f *CustomFilter) Label() string                 { return f.LabelStr }
func (f *CustomFilter) Type() string                  { return "custom" }
func (f *CustomFilter) FilterOptions() []FilterOption { return nil }

// TrashedFilter adds a "trashed records" filter with Active / Only / All options.
// Use it with resources that implement SoftDeletable.
type TrashedFilter struct {
	LabelStr string
}

// Trashed creates a new trashed filter.
func Trashed() *TrashedFilter {
	return &TrashedFilter{LabelStr: "Deleted records"}
}

// WithLabel sets the filter label.
func (f *TrashedFilter) WithLabel(label string) *TrashedFilter {
	f.LabelStr = label
	return f
}

func (f *TrashedFilter) Key() string   { return "trashed" }
func (f *TrashedFilter) Label() string { return f.LabelStr }
func (f *TrashedFilter) Type() string  { return "trashed" }
func (f *TrashedFilter) FilterOptions() []FilterOption {
	return []FilterOption{
		{Value: "active", Label: "Active only"},
		{Value: "only", Label: "Deleted only"},
		{Value: "all", Label: "All records"},
	}
}
