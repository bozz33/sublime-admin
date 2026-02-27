package actions

import "context"

// ActionGroup groups multiple actions into a collapsible dropdown menu.
// Renders as a "..." (or custom icon) button that reveals child actions on click.
// Mirrors Filament's ActionGroup::make([...]).
type ActionGroup struct {
	Label         string
	Icon          string // Material icon name for the trigger button (default: "more_vert")
	Color         string // Tailwind color key: "gray", "primary", "danger", etc.
	items         []*Action
	AuthorizeFunc func(ctx context.Context, item any) bool
}

// NewGroup creates an ActionGroup with the given label.
// Default icon: "more_vert", default color: "gray".
func NewGroup(label string) *ActionGroup {
	return &ActionGroup{
		Label: label,
		Icon:  "more_vert",
		Color: "gray",
	}
}

// Add appends one or more actions to the group.
func (g *ActionGroup) Add(acts ...*Action) *ActionGroup {
	g.items = append(g.items, acts...)
	return g
}

// SetIcon sets the trigger button icon (Material icon name).
func (g *ActionGroup) SetIcon(icon string) *ActionGroup {
	g.Icon = icon
	return g
}

// SetColor sets the button color key.
func (g *ActionGroup) SetColor(color string) *ActionGroup {
	g.Color = color
	return g
}

// Authorize sets an authorization function.
// Returns false → the whole group is hidden for that item.
func (g *ActionGroup) Authorize(fn func(ctx context.Context, item any) bool) *ActionGroup {
	g.AuthorizeFunc = fn
	return g
}

// IsAuthorized returns true if the group is allowed for the given context and item.
func (g *ActionGroup) IsAuthorized(ctx context.Context, item any) bool {
	if g.AuthorizeFunc == nil {
		return true
	}
	return g.AuthorizeFunc(ctx, item)
}

// Items returns the actions inside the group.
func (g *ActionGroup) Items() []*Action {
	return g.items
}

// MoreActionsGroup creates a standard "More" group with a vertical ellipsis icon.
// Usage: actions.MoreActionsGroup(actions.RestoreAction(url), actions.ForceDeleteAction(url))
func MoreActionsGroup(acts ...*Action) *ActionGroup {
	return NewGroup("More").SetIcon("more_vert").Add(acts...)
}
