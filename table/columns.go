package table

import (
	"fmt"
	"reflect"
	"time"

	"github.com/a-h/templ"
)

// TextColumn represents a text column.
type TextColumn struct {
	colKey       string
	LabelStr     string
	SortableFlag bool
	SearchFlag   bool
	CopyFlag     bool
	ValueFunc    func(item any) string // optional: replaces reflect-based lookup
	// Filament-inspired enrichments
	IsBadge       bool
	DescField     string       // field name for sub-text below value
	IconName      string       // Material icon name shown before value
	ColorEval     Eval[string] // Eval[string]: Static("green") or Dynamic(fn) — color based on value/record
	WeightStr     string       // "bold", "semibold", "medium"
	PrefixStr     string
	SuffixStr     string
	LimitChars    int  // truncate value to N chars (0 = no limit)
	StrikeThrough bool // render with line-through style
	Wrap          bool // allow text wrapping (default: nowrap)
	Prose         bool // render as prose (markdown-like paragraph)
	Bulleted      bool // render list items as bullet points
	// State transforms (applied in Render before display)
	MoneySymbol string // non-empty = format as money with this currency symbol
	NumericDec  int    // -1 = disabled, >=0 = format with N decimal places
	DateFormat  string // non-empty = parse RFC3339/date and reformat with this Go layout
	SinceFlag   bool   // true = show relative time ("2h ago")
}

// Text creates a new text column.
func Text(key string) *TextColumn {
	return &TextColumn{
		colKey:     key,
		LabelStr:   key,
		NumericDec: -1, // disabled by default
	}
}

// Money formats the value as money with the given currency symbol (e.g. "€", "$").
func (c *TextColumn) Money(symbol string) *TextColumn {
	c.MoneySymbol = symbol
	return c
}

// Numeric formats the value with N decimal places and thousands separator.
func (c *TextColumn) Numeric(decimals int) *TextColumn {
	c.NumericDec = decimals
	return c
}

// Date parses the raw value and reformats it with the given Go time layout.
// Example: Date("02/01/2006") formats "2024-03-15" as "15/03/2024".
func (c *TextColumn) Date(layout string) *TextColumn {
	c.DateFormat = layout
	return c
}

// Since renders the value as a relative time string ("2h ago", "3d ago").
func (c *TextColumn) Since() *TextColumn {
	c.SinceFlag = true
	return c
}

// Using sets a custom accessor function, bypassing reflection.
func (c *TextColumn) Using(fn func(item any) string) *TextColumn {
	c.ValueFunc = fn
	return c
}

// WithLabel sets the column label.
func (c *TextColumn) WithLabel(label string) *TextColumn {
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

// Badge renders the value as a colored badge pill.
func (c *TextColumn) Badge() *TextColumn {
	c.IsBadge = true
	return c
}

// Description sets a field name whose value is shown as sub-text below the main value.
func (c *TextColumn) Description(field string) *TextColumn {
	c.DescField = field
	return c
}

// WithIcon sets a Material Icons Outlined icon shown before the value.
func (c *TextColumn) WithIcon(icon string) *TextColumn {
	c.IconName = icon
	return c
}

// WithColor sets a static Tailwind color name for the cell.
func (c *TextColumn) WithColor(color string) *TextColumn {
	c.ColorEval = Static[string](color)
	return c
}

// WithColorFunc sets a dynamic color function based on the cell value and record.
func (c *TextColumn) WithColorFunc(fn func(value string, record any) string) *TextColumn {
	c.ColorEval = Dynamic[string](fn)
	return c
}

// Weight sets the font weight: "bold", "semibold", or "medium".
func (c *TextColumn) Weight(w string) *TextColumn {
	c.WeightStr = w
	return c
}

// Prefix sets a static prefix string prepended to the value.
func (c *TextColumn) Prefix(s string) *TextColumn {
	c.PrefixStr = s
	return c
}

// Suffix sets a static suffix string appended to the value.
func (c *TextColumn) Suffix(s string) *TextColumn {
	c.SuffixStr = s
	return c
}

// Limit truncates the displayed value to n characters.
func (c *TextColumn) Limit(n int) *TextColumn {
	c.LimitChars = n
	return c
}

// WithStrikeThrough renders the value with a line-through style.
func (c *TextColumn) WithStrikeThrough() *TextColumn {
	c.StrikeThrough = true
	return c
}

// WithWrap allows the cell text to wrap (default is nowrap).
func (c *TextColumn) WithWrap() *TextColumn {
	c.Wrap = true
	return c
}

// WithProse renders the value as a prose paragraph.
func (c *TextColumn) WithProse() *TextColumn {
	c.Prose = true
	return c
}

// WithBulleted renders the value as a bulleted list (newline-separated items).
func (c *TextColumn) WithBulleted() *TextColumn {
	c.Bulleted = true
	return c
}

// Column interface implementation
func (c *TextColumn) Key() string        { return c.colKey }
func (c *TextColumn) Label() string      { return c.LabelStr }
func (c *TextColumn) Type() string       { return "text" }
func (c *TextColumn) IsSortable() bool   { return c.SortableFlag }
func (c *TextColumn) IsSearchable() bool { return c.SearchFlag }
func (c *TextColumn) IsCopyable() bool   { return c.CopyFlag }
func (c *TextColumn) Render(value string, record any) templ.Component {
	v := applyTextTransforms(value, c)
	color := c.ColorEval.Resolve(v, record)
	if c.IsBadge {
		return TextCellBadgeView(v, color)
	}
	if c.IconName != "" {
		return TextCellWithIconView(v, c.IconName, color, c.PrefixStr, c.SuffixStr)
	}
	if c.DescField != "" {
		desc := ""
		if record != nil {
			desc = extractField(record, c.DescField)
		}
		return TextCellWithDescView(v, desc, c.PrefixStr, c.SuffixStr)
	}
	return TextCellView(v, c.PrefixStr, c.SuffixStr)
}

// applyTextTransforms applies state transforms (money, numeric, date, since, limit) to a raw value.
func applyTextTransforms(v string, c *TextColumn) string {
	if v == "" {
		return v
	}
	if c.SinceFlag {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			return sinceStr(t)
		}
		if t, err := time.Parse("2006-01-02T15:04:05Z", v); err == nil {
			return sinceStr(t)
		}
	}
	if c.DateFormat != "" {
		for _, layout := range []string{time.RFC3339, "2006-01-02T15:04:05Z", "2006-01-02 15:04:05", "2006-01-02"} {
			if t, err := time.Parse(layout, v); err == nil {
				v = t.Format(c.DateFormat)
				break
			}
		}
	}
	if c.MoneySymbol != "" {
		v = formatMoney(v, c.MoneySymbol)
	}
	if c.NumericDec >= 0 && c.MoneySymbol == "" {
		v = formatNumeric(v, c.NumericDec)
	}
	if c.LimitChars > 0 && len([]rune(v)) > c.LimitChars {
		v = string([]rune(v)[:c.LimitChars]) + "…"
	}
	return v
}

// sinceStr returns a human-readable relative time string.
func sinceStr(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return fmt.Sprintf("%ds ago", int(d.Seconds()))
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	case d < 30*24*time.Hour:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	case d < 365*24*time.Hour:
		return fmt.Sprintf("%dmo ago", int(d.Hours()/(24*30)))
	default:
		return fmt.Sprintf("%dy ago", int(d.Hours()/(24*365)))
	}
}

// formatMoney formats a numeric string as money (e.g. "1234.56" → "1 234,56 €").
func formatMoney(v string, symbol string) string {
	var f float64
	if _, err := fmt.Sscanf(v, "%f", &f); err != nil {
		return v
	}
	intPart := int64(f)
	decPart := int(f*100) % 100
	if decPart < 0 {
		decPart = -decPart
	}
	formatted := formatIntWithSep(intPart, " ")
	return fmt.Sprintf("%s,%02d %s", formatted, decPart, symbol)
}

// formatNumeric formats a numeric string with N decimal places and thousands separator.
func formatNumeric(v string, decimals int) string {
	var f float64
	if _, err := fmt.Sscanf(v, "%f", &f); err != nil {
		return v
	}
	if decimals == 0 {
		return formatIntWithSep(int64(f), ",")
	}
	return fmt.Sprintf("%."+fmt.Sprintf("%d", decimals)+"f", f)
}

// formatIntWithSep formats an integer with a thousands separator.
func formatIntWithSep(n int64, sep string) string {
	s := fmt.Sprintf("%d", n)
	if n < 0 {
		s = s[1:]
	}
	result := ""
	for i, ch := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			result += sep
		}
		result += string(ch)
	}
	if n < 0 {
		return "-" + result
	}
	return result
}

// extractField extracts a string field from a struct by field name using reflection.
func extractField(record any, field string) string {
	v := reflect.ValueOf(record)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	f := v.FieldByName(field)
	if f.IsValid() {
		return fmt.Sprintf("%v", f.Interface())
	}
	return ""
}
func (c *TextColumn) Value(item any) string {
	if c.ValueFunc != nil {
		return c.ValueFunc(item)
	}
	v := reflect.ValueOf(item)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	field := v.FieldByName(c.colKey)
	if field.IsValid() {
		return fmt.Sprintf("%v", field.Interface())
	}
	return ""
}

// BadgeColumn represents a badge column.
type BadgeColumn struct {
	colKey       string
	LabelStr     string
	SortableFlag bool
	ColorMap     map[string]string
	ValueFunc    func(item any) string // optional: replaces reflect-based lookup
	IconFunc     func(string) string   // returns Material icon name based on value
}

// Badge creates a new badge column.
func Badge(key string) *BadgeColumn {
	return &BadgeColumn{
		colKey:   key,
		LabelStr: key,
		ColorMap: make(map[string]string),
	}
}

// WithLabel sets the column label.
func (c *BadgeColumn) WithLabel(label string) *BadgeColumn {
	c.LabelStr = label
	return c
}

// Sortable makes the column sortable.
func (c *BadgeColumn) Sortable() *BadgeColumn {
	c.SortableFlag = true
	return c
}

// Using sets a custom accessor function, bypassing reflection.
func (c *BadgeColumn) Using(fn func(item any) string) *BadgeColumn {
	c.ValueFunc = fn
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

// WithIcon sets a function that returns a Material icon name based on the badge value.
func (c *BadgeColumn) WithIcon(fn func(string) string) *BadgeColumn {
	c.IconFunc = fn
	return c
}

// Column interface implementation
func (c *BadgeColumn) Key() string        { return c.colKey }
func (c *BadgeColumn) Label() string      { return c.LabelStr }
func (c *BadgeColumn) Type() string       { return "badge" }
func (c *BadgeColumn) IsSortable() bool   { return c.SortableFlag }
func (c *BadgeColumn) IsSearchable() bool { return false }
func (c *BadgeColumn) IsCopyable() bool   { return false }
func (c *BadgeColumn) Render(value string, _ any) templ.Component {
	return BadgeCellView(value, c.GetColor(value))
}
func (c *BadgeColumn) Value(item any) string {
	if c.ValueFunc != nil {
		return c.ValueFunc(item)
	}
	v := reflect.ValueOf(item)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	field := v.FieldByName(c.colKey)
	if field.IsValid() {
		return fmt.Sprintf("%v", field.Interface())
	}
	return ""
}

// ImageColumn represents an image column.
type ImageColumn struct {
	colKey    string
	LabelStr  string
	Rounded   bool
	ValueFunc func(item any) string // optional: replaces reflect-based lookup
	// Filament-inspired enrichments
	IsCircular bool
	IsSquare   bool
	WidthPx    int
	HeightPx   int
}

// Image creates a new image column.
func Image(key string) *ImageColumn {
	return &ImageColumn{
		colKey:   key,
		LabelStr: key,
	}
}

// WithLabel sets the column label.
func (c *ImageColumn) WithLabel(label string) *ImageColumn {
	c.LabelStr = label
	return c
}

// Round makes the image round.
func (c *ImageColumn) Round() *ImageColumn {
	c.Rounded = true
	return c
}

// Using sets a custom accessor function, bypassing reflection.
func (c *ImageColumn) Using(fn func(item any) string) *ImageColumn {
	c.ValueFunc = fn
	return c
}

// Circular makes the image display as a circle.
func (c *ImageColumn) Circular() *ImageColumn {
	c.IsCircular = true
	c.Rounded = true
	return c
}

// Square makes the image display as a square (no rounding).
func (c *ImageColumn) Square() *ImageColumn {
	c.IsSquare = true
	c.Rounded = false
	return c
}

// Size sets custom width and height in pixels.
func (c *ImageColumn) Size(w, h int) *ImageColumn {
	c.WidthPx = w
	c.HeightPx = h
	return c
}

// Column interface implementation
func (c *ImageColumn) Key() string        { return c.colKey }
func (c *ImageColumn) Label() string      { return c.LabelStr }
func (c *ImageColumn) Type() string       { return "image" }
func (c *ImageColumn) IsSortable() bool   { return false }
func (c *ImageColumn) IsSearchable() bool { return false }
func (c *ImageColumn) IsCopyable() bool   { return false }
func (c *ImageColumn) Render(value string, _ any) templ.Component {
	return ImageCellView(value, c.IsCircular || c.Rounded)
}
func (c *ImageColumn) Value(item any) string {
	if c.ValueFunc != nil {
		return c.ValueFunc(item)
	}
	v := reflect.ValueOf(item)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	field := v.FieldByName(c.colKey)
	if field.IsValid() {
		return fmt.Sprintf("%v", field.Interface())
	}
	return ""
}

// BooleanColumn displays a boolean value as a ✓ or ✗ icon.
type BooleanColumn struct {
	colKey        string
	LabelStr      string
	SortableFlag  bool
	TrueLabel     string
	FalseLabel    string
	ValueFunc     func(item any) string // optional: replaces reflect-based lookup
	TrueIconStr   string
	FalseIconStr  string
	TrueColorStr  string
	FalseColorStr string
}

// BoolCol creates a new boolean column.
func BoolCol(key string) *BooleanColumn {
	return &BooleanColumn{
		colKey:     key,
		LabelStr:   key,
		TrueLabel:  "Yes",
		FalseLabel: "No",
	}
}

// WithLabel sets the column label.
func (c *BooleanColumn) WithLabel(label string) *BooleanColumn {
	c.LabelStr = label
	return c
}

// Sortable makes the column sortable.
func (c *BooleanColumn) Sortable() *BooleanColumn {
	c.SortableFlag = true
	return c
}

// Using sets a custom accessor function, bypassing reflection.
func (c *BooleanColumn) Using(fn func(item any) string) *BooleanColumn {
	c.ValueFunc = fn
	return c
}

// Labels sets custom true/false display labels.
func (c *BooleanColumn) Labels(trueLabel, falseLabel string) *BooleanColumn {
	c.TrueLabel = trueLabel
	c.FalseLabel = falseLabel
	return c
}

// TrueIcon sets the Material icon shown for true values.
func (c *BooleanColumn) TrueIcon(icon string) *BooleanColumn {
	c.TrueIconStr = icon
	return c
}

// FalseIcon sets the Material icon shown for false values.
func (c *BooleanColumn) FalseIcon(icon string) *BooleanColumn {
	c.FalseIconStr = icon
	return c
}

// TrueColor sets the Tailwind color name for true values.
func (c *BooleanColumn) TrueColor(color string) *BooleanColumn {
	c.TrueColorStr = color
	return c
}

// FalseColor sets the Tailwind color name for false values.
func (c *BooleanColumn) FalseColor(color string) *BooleanColumn {
	c.FalseColorStr = color
	return c
}

// Column interface implementation
func (c *BooleanColumn) Key() string        { return c.colKey }
func (c *BooleanColumn) Label() string      { return c.LabelStr }
func (c *BooleanColumn) Type() string       { return "boolean" }
func (c *BooleanColumn) IsSortable() bool   { return c.SortableFlag }
func (c *BooleanColumn) IsSearchable() bool { return false }
func (c *BooleanColumn) IsCopyable() bool   { return false }
func (c *BooleanColumn) Render(value string, _ any) templ.Component {
	trueIcon := c.TrueIconStr
	if trueIcon == "" {
		trueIcon = "check_circle"
	}
	falseIcon := c.FalseIconStr
	if falseIcon == "" {
		falseIcon = "cancel"
	}
	trueColor := c.TrueColorStr
	if trueColor == "" {
		trueColor = "green"
	}
	return BooleanCellView(value, trueIcon, falseIcon, trueColor, c.FalseColorStr)
}
func (c *BooleanColumn) Value(item any) string {
	if c.ValueFunc != nil {
		return c.ValueFunc(item)
	}
	v := reflect.ValueOf(item)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	field := v.FieldByName(c.colKey)
	if !field.IsValid() {
		return c.FalseLabel
	}
	switch val := field.Interface().(type) {
	case bool:
		if val {
			return c.TrueLabel
		}
		return c.FalseLabel
	case int, int8, int16, int32, int64:
		if field.Int() != 0 {
			return c.TrueLabel
		}
		return c.FalseLabel
	}
	return fmt.Sprintf("%v", field.Interface())
}

// DateColumn displays a time.Time value with a configurable format.
type DateColumn struct {
	colKey       string
	LabelStr     string
	SortableFlag bool
	Format       string           // Go time format string, default "2006-01-02"
	Relative     bool             // Show relative time ("2 hours ago")
	ValueFunc    func(any) string // optional: replaces reflect-based lookup
}

// DateCol creates a new date column.
func DateCol(key string) *DateColumn {
	return &DateColumn{
		colKey:   key,
		LabelStr: key,
		Format:   "2006-01-02",
	}
}

// WithLabel sets the column label.
func (c *DateColumn) WithLabel(label string) *DateColumn {
	c.LabelStr = label
	return c
}

// Sortable makes the column sortable.
func (c *DateColumn) Sortable() *DateColumn {
	c.SortableFlag = true
	return c
}

// DateFormat sets a custom Go time format string.
func (c *DateColumn) DateFormat(format string) *DateColumn {
	c.Format = format
	return c
}

// ShowRelative displays relative time ("2 hours ago") instead of absolute.
func (c *DateColumn) ShowRelative() *DateColumn {
	c.Relative = true
	return c
}

// Using sets a custom accessor function, bypassing reflection.
func (c *DateColumn) Using(fn func(any) string) *DateColumn {
	c.ValueFunc = fn
	return c
}

// Column interface implementation
func (c *DateColumn) Key() string        { return c.colKey }
func (c *DateColumn) Label() string      { return c.LabelStr }
func (c *DateColumn) Type() string       { return "date" }
func (c *DateColumn) IsSortable() bool   { return c.SortableFlag }
func (c *DateColumn) IsSearchable() bool { return false }
func (c *DateColumn) IsCopyable() bool   { return false }
func (c *DateColumn) Render(value string, _ any) templ.Component {
	return DateCellView(value)
}
func (c *DateColumn) Value(item any) string {
	if c.ValueFunc != nil {
		return c.ValueFunc(item)
	}
	v := reflect.ValueOf(item)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	field := v.FieldByName(c.colKey)
	if !field.IsValid() {
		return ""
	}

	var t time.Time
	switch val := field.Interface().(type) {
	case time.Time:
		t = val
	case *time.Time:
		if val == nil {
			return ""
		}
		t = *val
	default:
		return fmt.Sprintf("%v", field.Interface())
	}

	if t.IsZero() {
		return ""
	}

	if c.Relative {
		return relativeTime(t)
	}
	return t.Format(c.Format)
}

// ---------------------------------------------------------------------------
// AvatarColumn — colored circle with initials + name beside it
// ---------------------------------------------------------------------------

// AvatarColumn displays a colored avatar circle with initials and the full name.
type AvatarColumn struct {
	colKey    string
	LabelStr  string
	ValueFunc func(item any) string
}

// Avatar creates a new avatar column.
func Avatar(key string) *AvatarColumn {
	return &AvatarColumn{
		colKey:   key,
		LabelStr: key,
	}
}

// WithLabel sets the column label.
func (c *AvatarColumn) WithLabel(label string) *AvatarColumn {
	c.LabelStr = label
	return c
}

// Using sets a custom accessor function.
func (c *AvatarColumn) Using(fn func(item any) string) *AvatarColumn {
	c.ValueFunc = fn
	return c
}

func (c *AvatarColumn) Key() string        { return c.colKey }
func (c *AvatarColumn) Label() string      { return c.LabelStr }
func (c *AvatarColumn) Type() string       { return "avatar" }
func (c *AvatarColumn) IsSortable() bool   { return false }
func (c *AvatarColumn) IsSearchable() bool { return false }
func (c *AvatarColumn) IsCopyable() bool   { return false }
func (c *AvatarColumn) Value(item any) string {
	if c.ValueFunc != nil {
		return c.ValueFunc(item)
	}
	v := reflect.ValueOf(item)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	field := v.FieldByName(c.colKey)
	if field.IsValid() {
		return fmt.Sprintf("%v", field.Interface())
	}
	return ""
}
func (c *AvatarColumn) Render(value string, _ any) templ.Component {
	initials := avatarCellInitials(value)
	bg := avatarCellBgColor(value)
	return AvatarCellView(value, initials, bg)
}

// ---------------------------------------------------------------------------
// IconColumn — standalone Material icon with optional conditional color
// ---------------------------------------------------------------------------

// IconColumn displays a Material Icons Outlined icon with an optional color.
type IconColumn struct {
	colKey    string
	LabelStr  string
	IconName  string
	ColorEval Eval[string] // Eval[string]: Static("green") or Dynamic(fn)
	IconEval  Eval[string] // Eval[string]: override icon per row
	ValueFunc func(item any) string
}

// Icon creates a new icon column with a fixed icon name.
func Icon(key, icon string) *IconColumn {
	return &IconColumn{
		colKey:   key,
		LabelStr: key,
		IconName: icon,
	}
}

// WithLabel sets the column label.
func (c *IconColumn) WithLabel(label string) *IconColumn {
	c.LabelStr = label
	return c
}

// WithColor sets a static Tailwind color name for the icon.
func (c *IconColumn) WithColor(color string) *IconColumn {
	c.ColorEval = Static[string](color)
	return c
}

// WithColorFunc sets a dynamic color function based on the cell value and record.
func (c *IconColumn) WithColorFunc(fn func(value string, record any) string) *IconColumn {
	c.ColorEval = Dynamic[string](fn)
	return c
}

// Using sets a custom accessor function.
func (c *IconColumn) Using(fn func(item any) string) *IconColumn {
	c.ValueFunc = fn
	return c
}

func (c *IconColumn) Key() string        { return c.colKey }
func (c *IconColumn) Label() string      { return c.LabelStr }
func (c *IconColumn) Type() string       { return "icon" }
func (c *IconColumn) IsSortable() bool   { return false }
func (c *IconColumn) IsSearchable() bool { return false }
func (c *IconColumn) IsCopyable() bool   { return false }
func (c *IconColumn) Value(item any) string {
	if c.ValueFunc != nil {
		return c.ValueFunc(item)
	}
	v := reflect.ValueOf(item)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	field := v.FieldByName(c.colKey)
	if field.IsValid() {
		return fmt.Sprintf("%v", field.Interface())
	}
	return ""
}
func (c *IconColumn) Render(value string, record any) templ.Component {
	color := c.ColorEval.Resolve(value, record)
	icon := c.IconName
	if c.IconEval.IsSet() {
		icon = c.IconEval.Resolve(value, record)
	}
	return IconCellView(icon, color)
}

// ---------------------------------------------------------------------------
// ColorColumn — color swatch (hex/named)
// ---------------------------------------------------------------------------

// ColorColumn displays a color swatch for a hex or named color value.
type ColorColumn struct {
	colKey    string
	LabelStr  string
	ValueFunc func(item any) string
}

// Color creates a new color column.
func Color(key string) *ColorColumn {
	return &ColorColumn{
		colKey:   key,
		LabelStr: key,
	}
}

// WithLabel sets the column label.
func (c *ColorColumn) WithLabel(label string) *ColorColumn {
	c.LabelStr = label
	return c
}

// Using sets a custom accessor function.
func (c *ColorColumn) Using(fn func(item any) string) *ColorColumn {
	c.ValueFunc = fn
	return c
}

func (c *ColorColumn) Key() string        { return c.colKey }
func (c *ColorColumn) Label() string      { return c.LabelStr }
func (c *ColorColumn) Type() string       { return "color" }
func (c *ColorColumn) IsSortable() bool   { return false }
func (c *ColorColumn) IsSearchable() bool { return false }
func (c *ColorColumn) IsCopyable() bool   { return false }
func (c *ColorColumn) Value(item any) string {
	if c.ValueFunc != nil {
		return c.ValueFunc(item)
	}
	v := reflect.ValueOf(item)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	field := v.FieldByName(c.colKey)
	if field.IsValid() {
		return fmt.Sprintf("%v", field.Interface())
	}
	return ""
}
func (c *ColorColumn) Render(value string, _ any) templ.Component {
	return ColorCellView(value)
}

// relativeTime returns a human-readable relative time string.
func relativeTime(t time.Time) string {
	diff := time.Since(t)
	switch {
	case diff < time.Minute:
		return "just now"
	case diff < time.Hour:
		mins := int(diff.Minutes())
		if mins == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", mins)
	case diff < 24*time.Hour:
		hours := int(diff.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	case diff < 7*24*time.Hour:
		days := int(diff.Hours() / 24)
		if days == 1 {
			return "yesterday"
		}
		return fmt.Sprintf("%d days ago", days)
	default:
		return t.Format("2006-01-02")
	}
}
