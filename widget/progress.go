package widget

import "github.com/a-h/templ"

// ProgressItem represents a single progress bar entry.
type ProgressItem struct {
	Label  string
	Value  int    // current value (0-Max)
	Max    int    // maximum value, default 100
	Color  string // "primary", "success", "danger", "warning", "info"
	Suffix string // optional suffix shown after percentage (e.g. " tasks")
}

// ProgressWidget displays one or more labeled progress bars.
type ProgressWidget struct {
	Title string
	Items []ProgressItem
}

// NewProgress creates a new progress widget.
func NewProgress(items ...ProgressItem) *ProgressWidget {
	return &ProgressWidget{Items: items}
}

// WithTitle sets the widget title.
func (w *ProgressWidget) WithTitle(title string) *ProgressWidget {
	w.Title = title
	return w
}

func (w *ProgressWidget) GetType() string { return "progress" }

var progressRenderFunc func(*ProgressWidget) templ.Component

// SetProgressRenderer registers the render function (called from views/widgets init).
func SetProgressRenderer(fn func(*ProgressWidget) templ.Component) {
	progressRenderFunc = fn
}

func (w *ProgressWidget) Render() templ.Component {
	if progressRenderFunc != nil {
		return progressRenderFunc(w)
	}
	return templ.NopComponent
}
