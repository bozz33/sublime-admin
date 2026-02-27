package widget

import "github.com/a-h/templ"

// GridItem holds one or more stacked widgets for a single grid column.
type GridItem struct {
	Widgets []Widget // widgets stacked vertically in this column
	Span    int      // CSS col-span (1..Columns). 0 = auto (span 1)
}

// GridWidget renders a responsive multi-column grid of widgets.
// It implements the Widget interface so it can be added to the dashboard
// widget list alongside StatsWidget, ChartWidget, etc.
//
// Example — faithfully reproducing the dashboard Charts Row (7-col grid):
//
//	widget.NewGrid(7).
//	    Add(4, lineChart).
//	    Add(3, donutChart)
//
// Example — Table + Activity sidebar (3-col grid, with stacked sidebar):
//
//	widget.NewGrid(3).
//	    Add(2, recentOffersWidget).
//	    Add(1, moderationWidget, activityWidget) // two widgets stacked
type GridWidget struct {
	Columns int        // total CSS grid columns (e.g., 7 → grid-cols-7)
	Items   []GridItem // ordered column items
}

// NewGrid creates a new grid widget with the given number of CSS columns.
// Typical values: 2, 3, 4, 7, 12.
func NewGrid(columns int) *GridWidget {
	if columns < 1 {
		columns = 1
	}
	return &GridWidget{Columns: columns}
}

// Add adds a column to the grid with the given col-span.
// Multiple widgets passed to a single Add() call are rendered
// as a vertical stack (space-y-4) inside that column.
func (g *GridWidget) Add(span int, widgets ...Widget) *GridWidget {
	if span < 1 {
		span = 1
	}
	if span > g.Columns {
		span = g.Columns
	}
	g.Items = append(g.Items, GridItem{
		Widgets: widgets,
		Span:    span,
	})
	return g
}

func (g *GridWidget) GetType() string { return "grid" }

// gridRenderFunc is set by views/widgets to avoid import cycles.
var gridRenderFunc func(*GridWidget) templ.Component

// SetGridRenderer registers the render function (called from views/widgets init).
func SetGridRenderer(fn func(*GridWidget) templ.Component) {
	gridRenderFunc = fn
}

func (g *GridWidget) Render() templ.Component {
	if gridRenderFunc != nil {
		return gridRenderFunc(g)
	}
	return templ.NopComponent
}

// gridColsClass maps a column count to a Tailwind grid-cols-N class.
// Returns the CSS class string for use in templates.
func GridColsClass(n int) string {
	switch n {
	case 1:
		return "grid-cols-1"
	case 2:
		return "grid-cols-1 lg:grid-cols-2"
	case 3:
		return "grid-cols-1 lg:grid-cols-3"
	case 4:
		return "grid-cols-1 sm:grid-cols-2 lg:grid-cols-4"
	case 5:
		return "grid-cols-1 sm:grid-cols-2 lg:grid-cols-5"
	case 6:
		return "grid-cols-1 sm:grid-cols-2 lg:grid-cols-6"
	case 7:
		return "grid-cols-1 lg:grid-cols-7"
	case 8:
		return "grid-cols-1 sm:grid-cols-2 lg:grid-cols-8"
	case 12:
		return "grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-12"
	default:
		return "grid-cols-1"
	}
}

// GridSpanClass maps a col-span value to a Tailwind lg:col-span-N class.
func GridSpanClass(span int) string {
	switch span {
	case 1:
		return "lg:col-span-1"
	case 2:
		return "lg:col-span-2"
	case 3:
		return "lg:col-span-3"
	case 4:
		return "lg:col-span-4"
	case 5:
		return "lg:col-span-5"
	case 6:
		return "lg:col-span-6"
	case 7:
		return "lg:col-span-7"
	case 8:
		return "lg:col-span-8"
	case 9:
		return "lg:col-span-9"
	case 10:
		return "lg:col-span-10"
	case 11:
		return "lg:col-span-11"
	case 12:
		return "lg:col-span-12"
	default:
		return "lg:col-span-1"
	}
}
