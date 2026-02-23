package widget

import "github.com/a-h/templ"

// TimelineItem represents a single event in a timeline.
type TimelineItem struct {
	Title       string
	Date        string
	Description string
	Icon        string // Material Icons Outlined name, default "radio_button_checked"
	Color       string // "primary", "success", "danger", "warning", "info", "gray"
}

// TimelineWidget displays a list of chronological events.
type TimelineWidget struct {
	Title      string
	Items      []TimelineItem
	Horizontal bool
}

// NewTimeline creates a new vertical timeline widget.
func NewTimeline(items ...TimelineItem) *TimelineWidget {
	return &TimelineWidget{Items: items}
}

// WithTitle sets the widget title.
func (w *TimelineWidget) WithTitle(title string) *TimelineWidget {
	w.Title = title
	return w
}

// AsHorizontal renders the timeline horizontally.
func (w *TimelineWidget) AsHorizontal() *TimelineWidget {
	w.Horizontal = true
	return w
}

func (w *TimelineWidget) GetType() string { return "timeline" }

var timelineRenderFunc func(*TimelineWidget) templ.Component

// SetTimelineRenderer registers the render function (called from views/widgets init).
func SetTimelineRenderer(fn func(*TimelineWidget) templ.Component) {
	timelineRenderFunc = fn
}

func (w *TimelineWidget) Render() templ.Component {
	if timelineRenderFunc != nil {
		return timelineRenderFunc(w)
	}
	return templ.NopComponent
}
