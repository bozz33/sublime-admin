package widget

import "github.com/a-h/templ"

// ListItem represents a single entry in a ListWidget.
type ListItem struct {
	Title       string // main label (required)
	Description string // optional subtitle shown below the title
	Icon        string // Material Icons Outlined name (e.g. "person", "shopping_cart")
	Color       string // icon color: "primary","success","danger","warning","info","purple","gray"
	Badge       string // optional badge text shown on the right
	BadgeColor  string // badge color (same palette as Color)
	URL         string // optional: wraps the item in an anchor tag
	Meta        string // optional right-side metadata text (e.g. "2h ago", "€ 1 234")
	Avatar      string // optional image URL — shown instead of the icon when set
}

// ListWidget renders a vertical list of items inside a card.
// Ideal for activity feeds, recent records, or simple ranked lists.
type ListWidget struct {
	Title        string
	Items        []ListItem
	Divided      bool   // show a separator line between items
	ViewAllURL   string // optional footer "View all" link
	ViewAllLabel string // label for the footer link (default: "View all")
	EmptyMessage string // shown when Items is empty (default: "No items")
}

// NewList creates a ListWidget with the given title and items.
func NewList(title string, items ...ListItem) *ListWidget {
	return &ListWidget{
		Title: title,
		Items: items,
	}
}

// WithDivided adds a separator line between each item.
func (w *ListWidget) WithDivided() *ListWidget {
	w.Divided = true
	return w
}

// WithViewAll sets the footer "View all" link and its label.
func (w *ListWidget) WithViewAll(url, label string) *ListWidget {
	w.ViewAllURL = url
	w.ViewAllLabel = label
	return w
}

// WithEmptyMessage sets the message shown when Items is empty.
func (w *ListWidget) WithEmptyMessage(msg string) *ListWidget {
	w.EmptyMessage = msg
	return w
}

func (w *ListWidget) GetType() string { return "list" }

var listRenderFunc func(*ListWidget) templ.Component

// SetListRenderer registers the render function (called from views/widgets init).
func SetListRenderer(fn func(*ListWidget) templ.Component) {
	listRenderFunc = fn
}

func (w *ListWidget) Render() templ.Component {
	if listRenderFunc != nil {
		return listRenderFunc(w)
	}
	return templ.NopComponent
}
