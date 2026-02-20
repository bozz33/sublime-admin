package table

import (
	"fmt"
	"reflect"
	"time"
)

// TextColumn represents a text column.
type TextColumn struct {
	Key          string
	LabelStr     string
	SortableFlag bool
	SearchFlag   bool
	CopyFlag     bool
}

// Text creates a new text column.
func Text(key string) *TextColumn {
	return &TextColumn{
		Key:      key,
		LabelStr: key,
	}
}

// Label sets the column label.
func (c *TextColumn) Label(label string) *TextColumn {
	c.LabelStr = label
	return c
}

// Sortable makes the column sortable.
func (c *TextColumn) Sortable() *TextColumn {
	c.SortableFlag = true
	return c
}

// Searchable makes the column searchable.
func (c *TextColumn) Searchable() *TextColumn {
	c.SearchFlag = true
	return c
}

// Copyable makes the column copyable.
func (c *TextColumn) Copyable() *TextColumn {
	c.CopyFlag = true
	return c
}

// Column interface implementation
func (c *TextColumn) GetKey() string     { return c.Key }
func (c *TextColumn) GetLabel() string   { return c.LabelStr }
func (c *TextColumn) GetType() string    { return "text" }
func (c *TextColumn) IsSortable() bool   { return c.SortableFlag }
func (c *TextColumn) IsSearchable() bool { return c.SearchFlag }
func (c *TextColumn) IsCopyable() bool   { return c.CopyFlag }
func (c *TextColumn) GetValue(item any) string {
	v := reflect.ValueOf(item)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	field := v.FieldByName(c.Key)
	if field.IsValid() {
		return fmt.Sprintf("%v", field.Interface())
	}
	return ""
}

// BadgeColumn represents a badge column.
type BadgeColumn struct {
	Key          string
	LabelStr     string
	SortableFlag bool
	ColorMap     map[string]string
}

// Badge creates a new badge column.
func Badge(key string) *BadgeColumn {
	return &BadgeColumn{
		Key:      key,
		LabelStr: key,
		ColorMap: make(map[string]string),
	}
}

// Label sets the column label.
func (c *BadgeColumn) Label(label string) *BadgeColumn {
	c.LabelStr = label
	return c
}

// Sortable makes the column sortable.
func (c *BadgeColumn) Sortable() *BadgeColumn {
	c.SortableFlag = true
	return c
}

// Colors sets the colors by value.
func (c *BadgeColumn) Colors(colors map[string]string) *BadgeColumn {
	c.ColorMap = colors
	return c
}

// GetColor returns the color for a value.
func (c *BadgeColumn) GetColor(value string) string {
	if color, ok := c.ColorMap[value]; ok {
		return color
	}
	return "primary"
}

// Column interface implementation
func (c *BadgeColumn) GetKey() string     { return c.Key }
func (c *BadgeColumn) GetLabel() string   { return c.LabelStr }
func (c *BadgeColumn) GetType() string    { return "badge" }
func (c *BadgeColumn) IsSortable() bool   { return c.SortableFlag }
func (c *BadgeColumn) IsSearchable() bool { return false }
func (c *BadgeColumn) IsCopyable() bool   { return false }
func (c *BadgeColumn) GetValue(item any) string {
	v := reflect.ValueOf(item)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	field := v.FieldByName(c.Key)
	if field.IsValid() {
		return fmt.Sprintf("%v", field.Interface())
	}
	return ""
}

// ImageColumn represents an image column.
type ImageColumn struct {
	Key      string
	LabelStr string
	Rounded  bool
}

// Image creates a new image column.
func Image(key string) *ImageColumn {
	return &ImageColumn{
		Key:      key,
		LabelStr: key,
	}
}

// Label sets the column label.
func (c *ImageColumn) Label(label string) *ImageColumn {
	c.LabelStr = label
	return c
}

// Round makes the image round.
func (c *ImageColumn) Round() *ImageColumn {
	c.Rounded = true
	return c
}

// Column interface implementation
func (c *ImageColumn) GetKey() string     { return c.Key }
func (c *ImageColumn) GetLabel() string   { return c.LabelStr }
func (c *ImageColumn) GetType() string    { return "image" }
func (c *ImageColumn) IsSortable() bool   { return false }
func (c *ImageColumn) IsSearchable() bool { return false }
func (c *ImageColumn) IsCopyable() bool   { return false }
func (c *ImageColumn) GetValue(item any) string {
	v := reflect.ValueOf(item)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	field := v.FieldByName(c.Key)
	if field.IsValid() {
		return fmt.Sprintf("%v", field.Interface())
	}
	return ""
}

// BooleanColumn displays a boolean value as configurable true/false labels.
type BooleanColumn struct {
	Key          string
	LabelStr     string
	SortableFlag bool
	TrueLabel    string
	FalseLabel   string
}

// BoolCol creates a new boolean column.
func BoolCol(key string) *BooleanColumn {
	return &BooleanColumn{
		Key:        key,
		LabelStr:   key,
		TrueLabel:  "Yes",
		FalseLabel: "No",
	}
}

// Label sets the column label.
func (c *BooleanColumn) Label(label string) *BooleanColumn {
	c.LabelStr = label
	return c
}

// Sortable makes the column sortable.
func (c *BooleanColumn) Sortable() *BooleanColumn {
	c.SortableFlag = true
	return c
}

// Labels sets custom true/false display labels.
func (c *BooleanColumn) Labels(trueLabel, falseLabel string) *BooleanColumn {
	c.TrueLabel = trueLabel
	c.FalseLabel = falseLabel
	return c
}

// Column interface implementation
func (c *BooleanColumn) GetKey() string     { return c.Key }
func (c *BooleanColumn) GetLabel() string   { return c.LabelStr }
func (c *BooleanColumn) GetType() string    { return "boolean" }
func (c *BooleanColumn) IsSortable() bool   { return c.SortableFlag }
func (c *BooleanColumn) IsSearchable() bool { return false }
func (c *BooleanColumn) IsCopyable() bool   { return false }
func (c *BooleanColumn) GetValue(item any) string {
	v := reflect.ValueOf(item)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	field := v.FieldByName(c.Key)
	if !field.IsValid() {
		return c.FalseLabel
	}
	switch field.Kind() {
	case reflect.Bool:
		if field.Bool() {
			return c.TrueLabel
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if field.Int() != 0 {
			return c.TrueLabel
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if field.Uint() != 0 {
			return c.TrueLabel
		}
	}
	return c.FalseLabel
}

// DateColumn displays a time.Time value with a configurable format.
type DateColumn struct {
	Key          string
	LabelStr     string
	SortableFlag bool
	Format       string
	Relative     bool
}

// DateCol creates a new date column with default format "2006-01-02".
func DateCol(key string) *DateColumn {
	return &DateColumn{
		Key:      key,
		LabelStr: key,
		Format:   "2006-01-02",
	}
}

// Label sets the column label.
func (c *DateColumn) Label(label string) *DateColumn {
	c.LabelStr = label
	return c
}

// Sortable makes the column sortable.
func (c *DateColumn) Sortable() *DateColumn {
	c.SortableFlag = true
	return c
}

// WithFormat sets a custom time format string (Go reference time layout).
func (c *DateColumn) WithFormat(format string) *DateColumn {
	c.Format = format
	return c
}

// ShowRelative displays relative time ("2 hours ago") instead of a fixed format.
func (c *DateColumn) ShowRelative() *DateColumn {
	c.Relative = true
	return c
}

// Column interface implementation
func (c *DateColumn) GetKey() string     { return c.Key }
func (c *DateColumn) GetLabel() string   { return c.LabelStr }
func (c *DateColumn) GetType() string    { return "date" }
func (c *DateColumn) IsSortable() bool   { return c.SortableFlag }
func (c *DateColumn) IsSearchable() bool { return false }
func (c *DateColumn) IsCopyable() bool   { return false }
func (c *DateColumn) GetValue(item any) string {
	v := reflect.ValueOf(item)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	field := v.FieldByName(c.Key)
	if !field.IsValid() {
		return ""
	}
	t, ok := field.Interface().(time.Time)
	if !ok {
		return fmt.Sprintf("%v", field.Interface())
	}
	if c.Relative {
		return relativeTime(t)
	}
	return t.Format(c.Format)
}

// relativeTime returns a human-readable relative time string.
func relativeTime(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		m := int(d.Minutes())
		if m == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", m)
	case d < 24*time.Hour:
		h := int(d.Hours())
		if h == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", h)
	case d < 7*24*time.Hour:
		days := int(d.Hours() / 24)
		if days == 1 {
			return "yesterday"
		}
		return fmt.Sprintf("%d days ago", days)
	default:
		return t.Format("2006-01-02")
	}
}
