package widgets

import (
	"github.com/a-h/templ"
	"github.com/bozz33/sublimeadmin/widget"
)

func init() {
	widget.SetStatsRenderer(func(w *widget.StatsWidget) templ.Component {
		return Stats(w)
	})
	widget.SetChartRenderer(func(w *widget.ChartWidget) templ.Component {
		return Chart(w)
	})
	widget.SetTimelineRenderer(func(w *widget.TimelineWidget) templ.Component {
		return Timeline(w)
	})
	widget.SetProgressRenderer(func(w *widget.ProgressWidget) templ.Component {
		return Progress(w)
	})
}
