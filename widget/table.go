package widget

import (
	"fmt"
	"reflect"

	"github.com/a-h/templ"
)

// TableWidgetColumn defines a column in a TableWidget.
type TableWidgetColumn struct {
	Label     string
	Key       string                // struct field name (case-sensitive) for reflection
	Align     string                // "left" (default), "center", "right"
	ValueFunc func(item any) string // overrides reflection when set
}

// TableWidget renders a compact data table inside a card.
// Use it on dashboards to display recent records without full list-page features
// (no search, filters, pagination, or bulk actions).
type TableWidget struct {
	Title        string
	Description  string
	Columns      []TableWidgetColumn
	Items        []any
	Striped      bool   // alternate row background shading
	ViewAllURL   string // optional "View all" footer link
	ViewAllLabel string // label for the footer link (default: "View all")
	EmptyMessage string // shown when Items is empty (default: "No data available")
}

// NewTable creates a TableWidget with the given title and items.
func NewTable(title string, items []any) *TableWidget {
	return &TableWidget{
		Title: title,
		Items: items,
	}
}

// WithDescription sets an optional subtitle shown below the title.
func (w *TableWidget) WithDescription(desc string) *TableWidget {
	w.Description = desc
	return w
}

// WithColumns sets the column definitions.
func (w *TableWidget) WithColumns(cols ...TableWidgetColumn) *TableWidget {
	w.Columns = cols
	return w
}

// WithStriped enables alternating row background shading.
func (w *TableWidget) WithStriped() *TableWidget {
	w.Striped = true
	return w
}

// WithViewAll sets the footer "View all" link and its label.
func (w *TableWidget) WithViewAll(url, label string) *TableWidget {
	w.ViewAllURL = url
	w.ViewAllLabel = label
	return w
}

// WithEmptyMessage sets the message shown when Items is empty.
func (w *TableWidget) WithEmptyMessage(msg string) *TableWidget {
	w.EmptyMessage = msg
	return w
}

// Rows returns a 2D slice of string values ready for template rendering.
// Each inner slice corresponds to one item and contains one value per column.
func (w *TableWidget) Rows() [][]string {
	result := make([][]string, len(w.Items))
	for i, item := range w.Items {
		row := make([]string, len(w.Columns))
		for j, col := range w.Columns {
			if col.ValueFunc != nil {
				row[j] = col.ValueFunc(item)
			} else {
				row[j] = tableExtractField(item, col.Key)
			}
		}
		result[i] = row
	}
	return result
}

// tableExtractField extracts a string value from a struct field via reflection.
func tableExtractField(item any, key string) string {
	if item == nil || key == "" {
		return ""
	}
	v := reflect.ValueOf(item)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return fmt.Sprintf("%v", item)
	}
	f := v.FieldByName(key)
	if f.IsValid() {
		return fmt.Sprintf("%v", f.Interface())
	}
	return ""
}

func (w *TableWidget) GetType() string { return "table" }

var tableRenderFunc func(*TableWidget) templ.Component

// SetTableRenderer registers the render function (called from views/widgets init).
func SetTableRenderer(fn func(*TableWidget) templ.Component) {
	tableRenderFunc = fn
}

func (w *TableWidget) Render() templ.Component {
	if tableRenderFunc != nil {
		return tableRenderFunc(w)
	}
	return templ.NopComponent
}
