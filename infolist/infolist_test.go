package infolist

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	il := New()
	assert.NotNil(t, il)
	assert.Empty(t, il.Sections)
}

func TestAddSection(t *testing.T) {
	il := New()
	s := &Section{Heading: "Details", Entries: []*Entry{}}
	il.AddSection(s)

	assert.Len(t, il.Sections, 1)
	assert.Equal(t, "Details", il.Sections[0].Heading)
}

func TestAddSectionDefaultColumns(t *testing.T) {
	il := New()
	s := &Section{Heading: "Test"}
	il.AddSection(s)
	assert.Equal(t, 2, il.Sections[0].Columns)
}

func TestAddSectionChaining(t *testing.T) {
	il := New().
		AddSection(&Section{Heading: "A"}).
		AddSection(&Section{Heading: "B"})
	assert.Len(t, il.Sections, 2)
}

func TestTextEntry(t *testing.T) {
	e := TextEntry("name", "Full Name", "John Doe")
	assert.Equal(t, "name", e.Name)
	assert.Equal(t, "Full Name", e.Label())
	assert.Equal(t, "John Doe", e.ValueStr())
	assert.Equal(t, EntryTypeText, e.Type)
}

func TestBadgeEntry(t *testing.T) {
	e := BadgeEntry("status", "Status", "active", "green")
	assert.Equal(t, EntryTypeBadge, e.Type)
	assert.Equal(t, "green", e.BadgeColor)
	assert.Equal(t, "active", e.ValueStr())
}

func TestBooleanEntry(t *testing.T) {
	e := BooleanEntry("active", "Active", true)
	assert.Equal(t, EntryTypeBoolean, e.Type)
	assert.Equal(t, "true", e.ValueStr())
}

func TestDateEntry(t *testing.T) {
	e := DateEntry("created_at", "Created", "2024-01-15", "")
	assert.Equal(t, EntryTypeDate, e.Type)
	assert.Equal(t, "2006-01-02", e.Format)
}

func TestDateEntryCustomFormat(t *testing.T) {
	e := DateEntry("created_at", "Created", "2024-01-15", "02/01/2006")
	assert.Equal(t, "02/01/2006", e.Format)
}

func TestImageEntry(t *testing.T) {
	e := ImageEntry("avatar", "Avatar", "/uploads/avatar.png")
	assert.Equal(t, EntryTypeImage, e.Type)
	assert.Equal(t, "/uploads/avatar.png", e.ValueStr())
}

func TestColorEntry(t *testing.T) {
	e := ColorEntry("color", "Color", "#FF5733")
	assert.Equal(t, EntryTypeColor, e.Type)
	assert.Equal(t, "#FF5733", e.ValueStr())
}

func TestEntryGetValueStrNil(t *testing.T) {
	e := TextEntry("x", "X", nil)
	assert.Equal(t, "", e.ValueStr())
}

func TestEntryIsVisible(t *testing.T) {
	e := TextEntry("x", "X", "val")
	assert.True(t, e.IsVisible())

	e.Hide(true)
	assert.False(t, e.IsVisible())
}

func TestEntryWithCopy(t *testing.T) {
	e := TextEntry("x", "X", "val").WithCopy()
	assert.True(t, e.IsCopyable)
}

func TestEntryHelp(t *testing.T) {
	e := TextEntry("x", "X", "val").Help("Some help text")
	assert.Equal(t, "Some help text", e.HelpText)
}

// ---------------------------------------------------------------------------
// IconEntry tests
// ---------------------------------------------------------------------------

func TestIconEntry(t *testing.T) {
	e := IconEntry("icon_field", "Icon", "star", "yellow")
	assert.Equal(t, EntryTypeIcon, e.Type)
	assert.Equal(t, "star", e.ValueStr())
	assert.Equal(t, "yellow", e.IconColor)
}

func TestIconEntry_EmptyColor(t *testing.T) {
	e := IconEntry("icon_field", "Icon", "home", "")
	assert.Equal(t, "home", e.ValueStr())
	assert.Equal(t, "", e.IconColor)
}

// ---------------------------------------------------------------------------
// ListEntry tests
// ---------------------------------------------------------------------------

func TestListEntry(t *testing.T) {
	items := []string{"Go", "Rust", "Python"}
	e := ListEntry("langs", "Languages", items)
	assert.Equal(t, EntryTypeList, e.Type)
	assert.Equal(t, items, e.ListItems)
	assert.Len(t, e.ListItems, 3)
}

func TestListEntry_Empty(t *testing.T) {
	e := ListEntry("tags", "Tags", nil)
	assert.Equal(t, EntryTypeList, e.Type)
	assert.Nil(t, e.ListItems)
}

// ---------------------------------------------------------------------------
// LinkEntry tests
// ---------------------------------------------------------------------------

func TestLinkEntry(t *testing.T) {
	e := LinkEntry("website", "Website", "https://example.com", "Visit")
	assert.Equal(t, EntryTypeLink, e.Type)
	assert.Equal(t, "https://example.com", e.LinkURL)
	assert.Equal(t, "Visit", e.ValueStr())
}

func TestLinkEntry_OpenInNewTab(t *testing.T) {
	e := LinkEntry("website", "Website", "https://example.com", "Visit").OpenInNewTab()
	assert.Equal(t, "_blank", e.LinkTarget)
}

func TestLinkEntry_NoNewTab_default(t *testing.T) {
	e := LinkEntry("website", "Website", "https://example.com", "Visit")
	assert.Equal(t, "", e.LinkTarget)
}

// ---------------------------------------------------------------------------
// CodeEntry tests
// ---------------------------------------------------------------------------

func TestCodeEntry(t *testing.T) {
	e := CodeEntry("snippet", "Code", "fmt.Println(\"hello\")", "go")
	assert.Equal(t, EntryTypeCode, e.Type)
	assert.Equal(t, "go", e.Language)
	assert.Equal(t, "fmt.Println(\"hello\")", e.ValueStr())
}

func TestCodeEntry_DefaultLanguage(t *testing.T) {
	e := CodeEntry("snippet", "Code", "SELECT * FROM users", "")
	assert.Equal(t, "text", e.Language)
}

// ---------------------------------------------------------------------------
// KeyValueEntry tests
// ---------------------------------------------------------------------------

func TestKeyValueEntry(t *testing.T) {
	e := KeyValueEntry("meta", "Metadata",
		KeyValuePair{Key: "author", Value: "Alice"},
		KeyValuePair{Key: "version", Value: "1.0"},
	)
	assert.Equal(t, EntryTypeKeyValue, e.Type)
	assert.Len(t, e.KeyValues, 2)
	assert.Equal(t, "author", e.KeyValues[0].Key)
	assert.Equal(t, "Alice", e.KeyValues[0].Value)
}

func TestKeyValueEntry_Empty(t *testing.T) {
	e := KeyValueEntry("meta", "Metadata")
	assert.Equal(t, EntryTypeKeyValue, e.Type)
	assert.Empty(t, e.KeyValues)
}

// ---------------------------------------------------------------------------
// RepeatableEntry tests
// ---------------------------------------------------------------------------

func TestRepeatableEntry(t *testing.T) {
	labels := []string{"Name", "Role"}
	row1 := []KeyValuePair{{Key: "Name", Value: "Alice"}, {Key: "Role", Value: "Admin"}}
	row2 := []KeyValuePair{{Key: "Name", Value: "Bob"}, {Key: "Role", Value: "Editor"}}

	e := RepeatableEntry("team", "Team Members", labels, row1, row2)

	assert.Equal(t, EntryTypeRepeatable, e.Type)
	assert.Equal(t, labels, e.RepeatLabels)
	assert.Len(t, e.RepeatItems, 2)
	assert.Equal(t, "Alice", e.RepeatItems[0][0].Value)
}

func TestRepeatableEntry_NoRows(t *testing.T) {
	e := RepeatableEntry("items", "Items", []string{"A", "B"})
	assert.Equal(t, EntryTypeRepeatable, e.Type)
	assert.Empty(t, e.RepeatItems)
}

// ---------------------------------------------------------------------------
// Entry fluent modifier tests
// ---------------------------------------------------------------------------

func TestEntry_Badge(t *testing.T) {
	e := TextEntry("status", "Status", "active").Badge()
	assert.True(t, e.IsBadge)
}

func TestEntry_Icon_modifier(t *testing.T) {
	e := TextEntry("name", "Name", "Alice").Icon("person")
	assert.Equal(t, "person", e.IconName)
}

func TestEntry_IconAfter(t *testing.T) {
	e := TextEntry("name", "Name", "Alice").Icon("person").IconAfter()
	assert.Equal(t, "after", e.IconPosition)
}

func TestEntry_Weight(t *testing.T) {
	e := TextEntry("title", "Title", "Hello").Weight("bold")
	assert.Equal(t, "bold", e.WeightStr)
}

func TestEntry_Limit(t *testing.T) {
	e := TextEntry("bio", "Bio", "Long text here").Limit(50)
	assert.Equal(t, 50, e.LimitChars)
}

func TestEntry_Align(t *testing.T) {
	e := TextEntry("score", "Score", "100").Align("right")
	assert.Equal(t, "right", e.Alignment)
}

func TestEntry_WithTooltip(t *testing.T) {
	e := TextEntry("info", "Info", "value").WithTooltip("More details here")
	assert.Equal(t, "More details here", e.Tooltip)
}

func TestEntry_WithPlaceholder(t *testing.T) {
	e := TextEntry("notes", "Notes", nil).WithPlaceholder("No notes yet")
	assert.Equal(t, "No notes yet", e.Placeholder)
}

func TestEntry_Color_static(t *testing.T) {
	e := TextEntry("status", "Status", "active").Color("green")
	assert.NotNil(t, e.ColorEval)
}

func TestEntry_DynamicColor(t *testing.T) {
	e := TextEntry("status", "Status", "active").DynamicColor(func(v string, _ any) string {
		if v == "active" {
			return "green"
		}
		return "red"
	})
	assert.NotNil(t, e.ColorEval)
}

func TestEntry_Render_not_nil(t *testing.T) {
	e := TextEntry("name", "Name", "Alice")
	assert.NotNil(t, e.Render())
}

// ---------------------------------------------------------------------------
// Section builder tests
// ---------------------------------------------------------------------------

func TestNewSection_defaults(t *testing.T) {
	s := NewSection("Personal Info")
	assert.Equal(t, "Personal Info", s.Heading)
	assert.Equal(t, 2, s.Columns)
	assert.Empty(t, s.Entries)
}

func TestSection_WithDescription(t *testing.T) {
	s := NewSection("Info").WithDescription("Basic user information")
	assert.Equal(t, "Basic user information", s.Description)
}

func TestSection_WithColumns(t *testing.T) {
	s := NewSection("Wide").WithColumns(3)
	assert.Equal(t, 3, s.Columns)
}

func TestSection_Add(t *testing.T) {
	s := NewSection("Details").Add(
		TextEntry("name", "Name", "Alice"),
		BadgeEntry("status", "Status", "active", "green"),
	)
	assert.Len(t, s.Entries, 2)
}

func TestSection_Add_chained(t *testing.T) {
	s := NewSection("Details").
		Add(TextEntry("a", "A", "v1")).
		Add(TextEntry("b", "B", "v2")).
		Add(TextEntry("c", "C", "v3"))
	assert.Len(t, s.Entries, 3)
}

// ---------------------------------------------------------------------------
// Full Infolist builder integration test
// ---------------------------------------------------------------------------

func TestInfolist_FullBuild(t *testing.T) {
	il := New().
		AddSection(
			NewSection("Identity").
				WithColumns(2).
				Add(TextEntry("name", "Name", "Alice")).
				Add(BadgeEntry("role", "Role", "admin", "blue")).
				Add(BooleanEntry("active", "Active", true)),
		).
		AddSection(
			NewSection("Details").
				WithColumns(1).
				Add(DateEntry("created_at", "Created", "2024-03-15", "")).
				Add(ImageEntry("avatar", "Avatar", "/img/alice.png")).
				Add(ColorEntry("accent", "Accent", "#3b82f6")),
		)

	assert.Len(t, il.Sections, 2)
	assert.Len(t, il.Sections[0].Entries, 3)
	assert.Len(t, il.Sections[1].Entries, 3)
	assert.Equal(t, 2, il.Sections[0].Columns)
	assert.Equal(t, 1, il.Sections[1].Columns)
}
