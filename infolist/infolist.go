package infolist

import (
	"fmt"

	"github.com/a-h/templ"
	"github.com/bozz33/sublimeadmin/table"
)

// EntryType defines the display type of an infolist entry.
type EntryType string

const (
	EntryTypeText       EntryType = "text"
	EntryTypeBadge      EntryType = "badge"
	EntryTypeBoolean    EntryType = "boolean"
	EntryTypeDate       EntryType = "date"
	EntryTypeImage      EntryType = "image"
	EntryTypeColor      EntryType = "color"
	EntryTypeKeyValue   EntryType = "keyvalue"
	EntryTypeIcon       EntryType = "icon"
	EntryTypeList       EntryType = "list"
	EntryTypeLink       EntryType = "link"
	EntryTypeCode       EntryType = "code"       // syntax-highlighted code block
	EntryTypeRepeatable EntryType = "repeatable" // list of sub-entries (nested)
)

// KeyValuePair is a single key→value pair for EntryTypeKeyValue.
type KeyValuePair struct {
	Key   string
	Value string
}

// Entry is a single read-only field in an Infolist.
type Entry struct {
	Name       string
	LabelStr   string
	Value      any
	Type       EntryType
	BadgeColor string   // for EntryTypeBadge
	Format     string   // for EntryTypeDate (Go time format)
	IconColor  string   // for EntryTypeIcon (Tailwind color name)
	ListItems  []string // for EntryTypeList
	LinkURL    string   // for EntryTypeLink
	LinkTarget string   // for EntryTypeLink ("_blank" etc.)
	IsCopyable bool
	Hidden     bool
	HelpText   string
	// Filament-inspired enrichments (TextEntry)
	IsBadge      bool               // render value as a badge pill
	ColorEval    table.Eval[string] // Eval[string]: Static("green") or Dynamic(fn)
	IconName     string             // Material icon shown before value
	IconPosition string             // "before" (default) or "after"
	WeightStr    string             // "bold", "semibold", "medium"
	LimitChars   int                // truncate to N chars
	// CodeEntry
	Language string // for EntryTypeCode: "go", "json", "sql", etc.
	// KeyValueEntry
	KeyValues []KeyValuePair // for EntryTypeKeyValue
	// RepeatableEntry
	RepeatItems  [][]KeyValuePair // for EntryTypeRepeatable: list of key-value sets
	RepeatLabels []string         // column labels for repeatable
	// HasTooltip
	Tooltip string
	// HasAlignment
	Alignment string // "left", "center", "right"
	// HasPlaceholder
	Placeholder string // shown when value is empty
}

// Label returns the display label.
func (e *Entry) Label() string { return e.LabelStr }

// ValueStr returns the value as a string.
func (e *Entry) ValueStr() string {
	if e.Value == nil {
		return ""
	}
	return fmt.Sprintf("%v", e.Value)
}

// IsVisible returns true if the entry should be displayed.
func (e *Entry) IsVisible() bool { return !e.Hidden }

// Badge renders the entry value as a colored badge pill.
func (e *Entry) Badge() *Entry {
	e.IsBadge = true
	return e
}

// Color sets a static Tailwind color name for the value.
func (e *Entry) Color(color string) *Entry {
	e.ColorEval = table.Static[string](color)
	return e
}

// DynamicColor sets a function that returns a Tailwind color based on the value.
func (e *Entry) DynamicColor(fn func(value string, record any) string) *Entry {
	e.ColorEval = table.Dynamic[string](fn)
	return e
}

// Icon sets a Material Icons Outlined icon shown next to the value.
func (e *Entry) Icon(icon string) *Entry {
	e.IconName = icon
	return e
}

// IconAfter positions the icon after the value instead of before.
func (e *Entry) IconAfter() *Entry {
	e.IconPosition = "after"
	return e
}

// Weight sets the font weight: "bold", "semibold", or "medium".
func (e *Entry) Weight(w string) *Entry {
	e.WeightStr = w
	return e
}

// Limit truncates the displayed value to n characters.
func (e *Entry) Limit(n int) *Entry {
	e.LimitChars = n
	return e
}

// Render returns the polymorphic templ component for this entry's value.
func (e *Entry) Render() templ.Component {
	return EntryRender(e)
}

// Section groups entries under a heading.
type Section struct {
	Heading     string
	Description string
	Columns     int // 1, 2, or 3 — default 2
	Entries     []*Entry
}

// NewSection creates a new section with a heading.
func NewSection(heading string) *Section {
	return &Section{Heading: heading, Columns: 2}
}

// WithDescription sets the section description.
func (s *Section) WithDescription(desc string) *Section {
	s.Description = desc
	return s
}

// WithColumns sets the number of columns (1, 2, or 3).
func (s *Section) WithColumns(n int) *Section {
	s.Columns = n
	return s
}

// Add appends entries to the section.
func (s *Section) Add(entries ...*Entry) *Section {
	s.Entries = append(s.Entries, entries...)
	return s
}

// Infolist is the top-level container for a read-only detail view.
type Infolist struct {
	Sections []*Section
}

// New creates an empty Infolist.
func New() *Infolist {
	return &Infolist{}
}

// AddSection appends a section and returns the Infolist for chaining.
func (il *Infolist) AddSection(s *Section) *Infolist {
	if s.Columns == 0 {
		s.Columns = 2
	}
	il.Sections = append(il.Sections, s)
	return il
}

// --- Fluent entry builders ---

// TextEntry creates a plain text entry.
func TextEntry(name, label string, value any) *Entry {
	return &Entry{Name: name, LabelStr: label, Value: value, Type: EntryTypeText}
}

// BadgeEntry creates a badge entry with a color.
func BadgeEntry(name, label string, value any, color string) *Entry {
	return &Entry{Name: name, LabelStr: label, Value: value, Type: EntryTypeBadge, BadgeColor: color}
}

// BooleanEntry creates a boolean (✓/✗) entry.
func BooleanEntry(name, label string, value any) *Entry {
	return &Entry{Name: name, LabelStr: label, Value: value, Type: EntryTypeBoolean}
}

// DateEntry creates a date entry with a Go time format string.
func DateEntry(name, label string, value any, format string) *Entry {
	if format == "" {
		format = "2006-01-02"
	}
	return &Entry{Name: name, LabelStr: label, Value: value, Type: EntryTypeDate, Format: format}
}

// ImageEntry creates an image entry (renders an <img> tag).
func ImageEntry(name, label string, value any) *Entry {
	return &Entry{Name: name, LabelStr: label, Value: value, Type: EntryTypeImage}
}

// ColorEntry creates a color swatch entry.
func ColorEntry(name, label string, value any) *Entry {
	return &Entry{Name: name, LabelStr: label, Value: value, Type: EntryTypeColor}
}

// IconEntry creates a Material Icons Outlined entry.
func IconEntry(name, label string, icon string, color string) *Entry {
	return &Entry{Name: name, LabelStr: label, Value: icon, Type: EntryTypeIcon, IconColor: color}
}

// ListEntry creates a list entry (renders as a bullet list).
func ListEntry(name, label string, items []string) *Entry {
	return &Entry{Name: name, LabelStr: label, Type: EntryTypeList, ListItems: items}
}

// LinkEntry creates a clickable link entry.
func LinkEntry(name, label string, url string, displayText string) *Entry {
	return &Entry{Name: name, LabelStr: label, Value: displayText, Type: EntryTypeLink, LinkURL: url}
}

// OpenInNewTab makes a LinkEntry open in a new tab.
func (e *Entry) OpenInNewTab() *Entry {
	e.LinkTarget = "_blank"
	return e
}

// WithCopy marks the entry as copyable (adds a copy button).
func (e *Entry) WithCopy() *Entry {
	e.IsCopyable = true
	return e
}

// Help adds a help text below the entry.
func (e *Entry) Help(text string) *Entry {
	e.HelpText = text
	return e
}

// Hide hides the entry conditionally.
func (e *Entry) Hide(hidden bool) *Entry {
	e.Hidden = hidden
	return e
}

// WithTooltip sets a tooltip shown on hover (HasTooltip).
func (e *Entry) WithTooltip(tip string) *Entry {
	e.Tooltip = tip
	return e
}

// Align sets the text alignment: "left", "center", "right" (HasAlignment).
func (e *Entry) Align(alignment string) *Entry {
	e.Alignment = alignment
	return e
}

// WithPlaceholder sets the text shown when the value is empty (HasPlaceholder).
func (e *Entry) WithPlaceholder(placeholder string) *Entry {
	e.Placeholder = placeholder
	return e
}

// CodeEntry creates a code block entry with optional language for syntax highlighting.
func CodeEntry(name, label string, code string, language string) *Entry {
	if language == "" {
		language = "text"
	}
	return &Entry{Name: name, LabelStr: label, Value: code, Type: EntryTypeCode, Language: language}
}

// KeyValueEntry creates a key→value table entry.
func KeyValueEntry(name, label string, pairs ...KeyValuePair) *Entry {
	return &Entry{Name: name, LabelStr: label, Type: EntryTypeKeyValue, KeyValues: pairs}
}

// RepeatableEntry creates a repeatable entry (list of key-value rows).
// labels: column headers; rows: each row is a slice of KeyValuePair.
func RepeatableEntry(name, label string, labels []string, rows ...[]KeyValuePair) *Entry {
	return &Entry{
		Name:         name,
		LabelStr:     label,
		Type:         EntryTypeRepeatable,
		RepeatLabels: labels,
		RepeatItems:  rows,
	}
}
