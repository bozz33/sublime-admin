package table

import (
	"testing"
)

func TestNewTable(t *testing.T) {
	data := []any{"item1", "item2"}
	tbl := New(data)

	if tbl == nil {
		t.Error("Expected table to be created")
	}
	if len(tbl.Rows) != 2 {
		t.Errorf("Expected 2 rows, got %d", len(tbl.Rows))
	}
	if !tbl.Searchable {
		t.Error("Expected Searchable to be true by default")
	}
	if !tbl.Pagination {
		t.Error("Expected Pagination to be true by default")
	}
	if tbl.PerPage != 15 {
		t.Errorf("Expected PerPage 15, got %d", tbl.PerPage)
	}
}

func TestTableAddColumn(t *testing.T) {
	tbl := New(nil).
		AddColumn("id", "ID").
		AddColumn("name", "Name")

	if len(tbl.Columns) != 2 {
		t.Errorf("Expected 2 columns, got %d", len(tbl.Columns))
	}
}

func TestTableWithColumns(t *testing.T) {
	col1 := Text("id").WithLabel("ID")
	col2 := Text("name").WithLabel("Name")

	tbl := New(nil).WithColumns(col1, col2)

	if len(tbl.Columns) != 2 {
		t.Errorf("Expected 2 columns, got %d", len(tbl.Columns))
	}
}

func TestTableWithBaseURL(t *testing.T) {
	tbl := New(nil).WithBaseURL("/users")

	if tbl.BaseURL != "/users" {
		t.Errorf("Expected BaseURL '/users', got '%s'", tbl.BaseURL)
	}
}

func TestTableSearch(t *testing.T) {
	tbl := New(nil).Search(false)

	if tbl.Searchable {
		t.Error("Expected Searchable to be false")
	}
}

func TestTablePaginate(t *testing.T) {
	tbl := New(nil).Paginate(false)

	if tbl.Pagination {
		t.Error("Expected Pagination to be false")
	}
}

func TestTextColumn(t *testing.T) {
	col := Text("name").
		WithLabel("Full Name").
		Sortable().
		Searchable().
		Copyable()

	if col.Key() != "name" {
		t.Errorf("Expected key 'name', got '%s'", col.Key())
	}
	if col.LabelStr != "Full Name" {
		t.Errorf("Expected label 'Full Name', got '%s'", col.LabelStr)
	}
	if !col.SortableFlag {
		t.Error("Expected Sortable to be true")
	}
	if !col.SearchFlag {
		t.Error("Expected Searchable to be true")
	}
	if !col.CopyFlag {
		t.Error("Expected Copyable to be true")
	}
}

func TestTextColumnInterface(t *testing.T) {
	col := Text("test")

	if col.Key() != "test" {
		t.Errorf("Expected Key() 'test', got '%s'", col.Key())
	}
	if col.Type() != "text" {
		t.Errorf("Expected Type() 'text', got '%s'", col.Type())
	}
}

func TestBadgeColumn(t *testing.T) {
	col := Badge("status").
		WithLabel("Status").
		Colors(map[string]string{
			"active":   "success",
			"inactive": "danger",
		})

	if col.Key() != "status" {
		t.Errorf("Expected key 'status', got '%s'", col.Key())
	}
	if col.Type() != "badge" {
		t.Errorf("Expected type 'badge', got '%s'", col.Type())
	}
	if len(col.ColorMap) != 2 {
		t.Errorf("Expected 2 color mappings, got %d", len(col.ColorMap))
	}
}

func TestImageColumn(t *testing.T) {
	col := Image("avatar").
		WithLabel("Profile Picture").
		Round()

	if col.Key() != "avatar" {
		t.Errorf("Expected key 'avatar', got '%s'", col.Key())
	}
	if col.Type() != "image" {
		t.Errorf("Expected type 'image', got '%s'", col.Type())
	}
	if !col.Rounded {
		t.Error("Expected Rounded to be true")
	}
}

func TestSelectFilter(t *testing.T) {
	filter := Select("status").
		WithLabel("Status").
		WithOptions([]FilterOption{
			{Value: "active", Label: "Active"},
			{Value: "inactive", Label: "Inactive"},
		})

	if filter.Key() != "status" {
		t.Errorf("Expected key 'status', got '%s'", filter.Key())
	}
	if filter.Type() != "select" {
		t.Errorf("Expected type 'select', got '%s'", filter.Type())
	}
	if len(filter.FilterOptions()) != 2 {
		t.Errorf("Expected 2 options, got %d", len(filter.FilterOptions()))
	}
}

func TestBooleanFilter(t *testing.T) {
	filter := Boolean("active").WithLabel("Is Active")

	if filter.Key() != "active" {
		t.Errorf("Expected key 'active', got '%s'", filter.Key())
	}
	if filter.Type() != "boolean" {
		t.Errorf("Expected type 'boolean', got '%s'", filter.Type())
	}

	opts := filter.FilterOptions()
	if len(opts) != 2 {
		t.Errorf("Expected 2 options (Yes/No), got %d", len(opts))
	}
}

// ---------------------------------------------------------------------------
// TableTab tests
// ---------------------------------------------------------------------------

func TestTableTab_fields(t *testing.T) {
	tab := TableTab{
		Label:      "Active",
		Key:        "active",
		Value:      "1",
		Icon:       "check_circle",
		Color:      "green",
		BadgeCount: "42",
	}

	if tab.Label != "Active" {
		t.Errorf("expected Label='Active', got '%s'", tab.Label)
	}
	if tab.Key != "active" {
		t.Errorf("expected Key='active', got '%s'", tab.Key)
	}
	if tab.BadgeCount != "42" {
		t.Errorf("expected BadgeCount='42', got '%s'", tab.BadgeCount)
	}
}

func TestWithTabs_appends(t *testing.T) {
	tbl := New(nil).WithTabs(
		TableTab{Label: "All", Key: "all"},
		TableTab{Label: "Active", Key: "active"},
		TableTab{Label: "Archived", Key: "archived"},
	)

	if len(tbl.Tabs) != 3 {
		t.Fatalf("expected 3 tabs, got %d", len(tbl.Tabs))
	}
	if tbl.Tabs[0].Key != "all" {
		t.Errorf("expected Tabs[0].Key='all', got '%s'", tbl.Tabs[0].Key)
	}
	if tbl.Tabs[2].Key != "archived" {
		t.Errorf("expected Tabs[2].Key='archived', got '%s'", tbl.Tabs[2].Key)
	}
}

func TestWithTabs_empty_by_default(t *testing.T) {
	tbl := New(nil)
	if len(tbl.Tabs) != 0 {
		t.Errorf("expected 0 tabs by default, got %d", len(tbl.Tabs))
	}
}

func TestWithTabs_chained(t *testing.T) {
	tbl := New(nil).
		WithTabs(TableTab{Label: "All", Key: "all"}).
		WithTabs(TableTab{Label: "Active", Key: "active"})

	if len(tbl.Tabs) != 2 {
		t.Errorf("expected 2 tabs after chained WithTabs, got %d", len(tbl.Tabs))
	}
}

// ---------------------------------------------------------------------------
// Grouping tests
// ---------------------------------------------------------------------------

func TestGroupBy_defaults(t *testing.T) {
	g := GroupBy("status")

	if g.ColumnKey() != "status" {
		t.Errorf("expected ColumnKey='status', got '%s'", g.ColumnKey())
	}
	if g.Label() != "status" {
		t.Errorf("expected default Label='status', got '%s'", g.Label())
	}
	if g.IsCollapsible() {
		t.Error("expected IsCollapsible()=false by default")
	}
	if g.IsCollapsed() {
		t.Error("expected IsCollapsed()=false by default")
	}
}

func TestGroupBy_WithLabel(t *testing.T) {
	g := GroupBy("status").WithLabel("Status")
	if g.Label() != "Status" {
		t.Errorf("expected Label='Status', got '%s'", g.Label())
	}
}

func TestGroupBy_Collapsible(t *testing.T) {
	g := GroupBy("category").Collapsible()
	if !g.IsCollapsible() {
		t.Error("expected IsCollapsible()=true after Collapsible()")
	}
	if g.IsCollapsed() {
		t.Error("expected IsCollapsed()=false (not collapsed by default after Collapsible())")
	}
}

func TestGroupBy_CollapsedByDefault(t *testing.T) {
	g := GroupBy("category").CollapsedByDefault()
	if !g.IsCollapsible() {
		t.Error("expected IsCollapsible()=true after CollapsedByDefault()")
	}
	if !g.IsCollapsed() {
		t.Error("expected IsCollapsed()=true after CollapsedByDefault()")
	}
}

func TestGroupBy_WithTitleFn(t *testing.T) {
	g := GroupBy("status").WithTitleFn(func(v string) string {
		return "Status: " + v
	})

	got := g.Title("active")
	if got != "Status: active" {
		t.Errorf("expected Title='Status: active', got '%s'", got)
	}
}

func TestGroupBy_Title_default(t *testing.T) {
	g := GroupBy("status")
	if g.Title("pending") != "pending" {
		t.Errorf("expected default Title='pending', got '%s'", g.Title("pending"))
	}
}

// ---------------------------------------------------------------------------
// GroupRows tests
// ---------------------------------------------------------------------------

type groupTestRecord struct {
	Name   string
	Status string
}

func TestGroupRows_partitions_by_column(t *testing.T) {
	rows := []any{
		groupTestRecord{Name: "Alice", Status: "active"},
		groupTestRecord{Name: "Bob", Status: "inactive"},
		groupTestRecord{Name: "Charlie", Status: "active"},
		groupTestRecord{Name: "Dave", Status: "pending"},
	}

	col := Text("Status")
	g := GroupBy("Status")

	groups := GroupRows(rows, col, g)

	if len(groups) != 3 {
		t.Fatalf("expected 3 groups, got %d", len(groups))
	}

	// First group encountered is "active" (first row)
	if groups[0].Key != "active" {
		t.Errorf("expected first group Key='active', got '%s'", groups[0].Key)
	}
	if len(groups[0].Rows) != 2 {
		t.Errorf("expected 2 rows in 'active' group, got %d", len(groups[0].Rows))
	}
}

func TestGroupRows_preserves_insertion_order(t *testing.T) {
	rows := []any{
		groupTestRecord{Name: "X", Status: "c"},
		groupTestRecord{Name: "Y", Status: "a"},
		groupTestRecord{Name: "Z", Status: "b"},
	}
	col := Text("Status")
	g := GroupBy("Status")
	groups := GroupRows(rows, col, g)

	keys := []string{groups[0].Key, groups[1].Key, groups[2].Key}
	expected := []string{"c", "a", "b"}
	for i, k := range expected {
		if keys[i] != k {
			t.Errorf("expected groups[%d].Key='%s', got '%s'", i, k, keys[i])
		}
	}
}

func TestGroupRows_empty_rows(t *testing.T) {
	col := Text("Status")
	g := GroupBy("Status")
	groups := GroupRows(nil, col, g)
	if len(groups) != 0 {
		t.Errorf("expected 0 groups for empty rows, got %d", len(groups))
	}
}

func TestGroupRows_collapsible_propagated(t *testing.T) {
	rows := []any{
		groupTestRecord{Name: "A", Status: "x"},
	}
	col := Text("Status")
	g := GroupBy("Status").CollapsedByDefault()
	groups := GroupRows(rows, col, g)

	if !groups[0].Collapsible {
		t.Error("expected Collapsible=true on group")
	}
	if !groups[0].Collapsed {
		t.Error("expected Collapsed=true on group")
	}
}

// ---------------------------------------------------------------------------
// Summary tests
// ---------------------------------------------------------------------------

func TestNewSummary_fields(t *testing.T) {
	s := NewSummary("total", SummarySum)

	if s.ColumnKey() != "total" {
		t.Errorf("expected ColumnKey='total', got '%s'", s.ColumnKey())
	}
	if s.Type() != SummarySum {
		t.Errorf("expected Type=SummarySum, got '%s'", s.Type())
	}
}

func TestSummary_WithLabel(t *testing.T) {
	s := NewSummary("revenue", SummarySum).WithLabel("Total Revenue")
	if s.Label() != "Total Revenue" {
		t.Errorf("expected Label='Total Revenue', got '%s'", s.Label())
	}
}

func TestSummary_WithFormat(t *testing.T) {
	s := NewSummary("price", SummaryAverage).WithFormat("%.2f")
	if s.Format() != "%.2f" {
		t.Errorf("expected Format='%%.2f', got '%s'", s.Format())
	}
}

func TestSummary_WithCompute_custom(t *testing.T) {
	s := NewSummary("count", SummaryCount).WithCompute(func(rows []any) string {
		return "42"
	})

	result := s.Compute([]any{"a", "b", "c"})
	if result != "42" {
		t.Errorf("expected Compute()='42', got '%s'", result)
	}
}

func TestSummary_Compute_nil_fn(t *testing.T) {
	s := NewSummary("count", SummaryCount)
	// No compute fn → returns ""
	if s.Compute([]any{"a"}) != "" {
		t.Error("expected empty string when no compute fn is set")
	}
}

func TestSummaryTypes(t *testing.T) {
	types := []SummaryType{SummarySum, SummaryAverage, SummaryMin, SummaryMax, SummaryCount}
	expected := []string{"sum", "average", "min", "max", "count"}
	for i, st := range types {
		if string(st) != expected[i] {
			t.Errorf("expected SummaryType '%s', got '%s'", expected[i], st)
		}
	}
}

func TestWithSummaries_appends(t *testing.T) {
	s1 := NewSummary("revenue", SummarySum)
	s2 := NewSummary("orders", SummaryCount)
	tbl := New(nil).WithSummaries(s1, s2)

	if len(tbl.Summaries) != 2 {
		t.Errorf("expected 2 summaries, got %d", len(tbl.Summaries))
	}
}

// ---------------------------------------------------------------------------
// Table builder misc tests
// ---------------------------------------------------------------------------

func TestWithStriped(t *testing.T) {
	tbl := New(nil).WithStriped()
	if !tbl.Striped {
		t.Error("expected Striped=true after WithStriped()")
	}
}

func TestWithDeferred(t *testing.T) {
	tbl := New(nil).WithDeferred()
	if !tbl.Deferred {
		t.Error("expected Deferred=true after WithDeferred()")
	}
}

func TestWithEmptyState(t *testing.T) {
	tbl := New(nil).WithEmptyState("No results", "Try adjusting your search", "search_off")

	if tbl.EmptyHeading != "No results" {
		t.Errorf("expected EmptyHeading='No results', got '%s'", tbl.EmptyHeading)
	}
	if tbl.EmptyDesc != "Try adjusting your search" {
		t.Errorf("expected EmptyDesc='Try adjusting your search', got '%s'", tbl.EmptyDesc)
	}
	if tbl.EmptyIcon != "search_off" {
		t.Errorf("expected EmptyIcon='search_off', got '%s'", tbl.EmptyIcon)
	}
}

func TestWithEmptyState_default_icon_preserved(t *testing.T) {
	// Passing empty icon string should keep "inbox" default
	tbl := New(nil).WithEmptyState("No data", "Empty", "")
	if tbl.EmptyIcon != "inbox" {
		t.Errorf("expected EmptyIcon='inbox' (default), got '%s'", tbl.EmptyIcon)
	}
}

func TestWithRecordUrl(t *testing.T) {
	tbl := New(nil).WithRecordUrl(func(item any) string {
		return "/custom/" + item.(string)
	})

	if tbl.RecordUrlFn == nil {
		t.Fatal("expected RecordUrlFn to be set")
	}
	if got := tbl.RecordUrlFn("42"); got != "/custom/42" {
		t.Errorf("expected RecordUrlFn('42')='/custom/42', got '%s'", got)
	}
}

func TestWithGroups_appends(t *testing.T) {
	g1 := *GroupBy("status")
	g2 := *GroupBy("category").Collapsible()
	tbl := New(nil).WithGroups(g1, g2)

	if len(tbl.Groups) != 2 {
		t.Errorf("expected 2 groups, got %d", len(tbl.Groups))
	}
}
