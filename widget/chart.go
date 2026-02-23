package widget

import (
	"encoding/json"

	"github.com/a-h/templ"
)

// ChartType defines the supported chart type.
type ChartType string

const (
	Line    ChartType = "area"
	Bar     ChartType = "bar"
	Donut   ChartType = "donut"
	Pie     ChartType = "pie"
	Radar   ChartType = "radar"
	Scatter ChartType = "scatter"
	HeatMap ChartType = "heatmap"
)

// ChartDataSet represents a data series.
type ChartDataSet struct {
	Name string `json:"name"`
	Data []int  `json:"data"`
}

// ChartWidget configures a complete chart.
type ChartWidget struct {
	ID          string
	Label       string
	Type        ChartType
	Series      []ChartDataSet
	Labels      []string
	Colors      []string
	Height      string
	ColumnSpan  int    // dashboard grid column span (1-4, 0 = full width)
	IsLazy      bool   // CanBeLazy: defer rendering until visible
	Description string // optional subtitle
	Footer      string // optional footer text
}

// NewChart creates a new chart.
func NewChart(id, label string, t ChartType) *ChartWidget {
	return &ChartWidget{
		ID:     id,
		Label:  label,
		Type:   t,
		Height: "300",
		Colors: []string{"#22c55e", "#3b82f6", "#eab308", "#ef4444"},
	}
}

// WithHeight sets the chart height in pixels.
func (c *ChartWidget) WithHeight(h string) *ChartWidget {
	c.Height = h
	return c
}

// WithColumnSpan sets the dashboard grid column span (1–4).
func (c *ChartWidget) WithColumnSpan(span int) *ChartWidget {
	c.ColumnSpan = span
	return c
}

// Lazy enables deferred rendering (CanBeLazy).
func (c *ChartWidget) Lazy() *ChartWidget {
	c.IsLazy = true
	return c
}

// WithDescription sets an optional subtitle.
func (c *ChartWidget) WithDescription(desc string) *ChartWidget {
	c.Description = desc
	return c
}

// WithFooter sets an optional footer text.
func (c *ChartWidget) WithFooter(footer string) *ChartWidget {
	c.Footer = footer
	return c
}

func (c *ChartWidget) SetLabels(labels []string) *ChartWidget {
	c.Labels = labels
	return c
}

func (c *ChartWidget) AddSeries(name string, data []int) *ChartWidget {
	c.Series = append(c.Series, ChartDataSet{Name: name, Data: data})
	return c
}

// GetSeriesJSON returns the series as JSON for JavaScript.
func (c *ChartWidget) GetSeriesJSON() string {
	b, err := json.Marshal(c.Series)
	if err != nil {
		return "[]"
	}
	if c.Type == Donut {
		simpleData := make([]int, len(c.Series))
		for i, s := range c.Series {
			if len(s.Data) > 0 {
				simpleData[i] = s.Data[0]
			}
		}
		if b2, e2 := json.Marshal(simpleData); e2 == nil {
			b = b2
		}
	}
	return string(b)
}

// GetLabelsJSON returns the labels as JSON.
func (c *ChartWidget) GetLabelsJSON() string {
	b, err := json.Marshal(c.Labels)
	if err != nil {
		return "[]"
	}
	return string(b)
}

// GetColorsJSON returns the colors as JSON.
func (c *ChartWidget) GetColorsJSON() string {
	b, err := json.Marshal(c.Colors)
	if err != nil {
		return "[]"
	}
	return string(b)
}

func (c *ChartWidget) GetType() string { return "chart" }

// chartRenderFunc is set by views/widgets to avoid import cycles.
var chartRenderFunc func(*ChartWidget) templ.Component

// SetChartRenderer registers the render function.
func SetChartRenderer(fn func(*ChartWidget) templ.Component) {
	chartRenderFunc = fn
}

func (c *ChartWidget) Render() templ.Component {
	if chartRenderFunc != nil {
		return chartRenderFunc(c)
	}
	return templ.NopComponent
}
