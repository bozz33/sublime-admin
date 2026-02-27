package form

import (
	"context"
	"io"

	"github.com/a-h/templ"
)

// layoutEmptyComponent returns an empty templ component (layouts render via parent template).
func layoutEmptyComponent() templ.Component {
	return templ.ComponentFunc(func(_ context.Context, _ io.Writer) error { return nil })
}

// Section represents a form section.
type Section struct {
	Heading     string
	Description string
	Components  []Component
	CanCollapse bool
	Collapsed   bool
}

// NewSection creates a new section.
func NewSection(heading string) *Section {
	return &Section{
		Heading:    heading,
		Components: make([]Component, 0),
	}
}

// SetSchema sets the section components.
func (s *Section) SetSchema(components ...Component) *Section {
	s.Components = components
	return s
}

// Desc sets the description.
func (s *Section) Desc(desc string) *Section {
	s.Description = desc
	return s
}

// Collapsible makes the section collapsible.
func (s *Section) Collapsible() *Section {
	s.CanCollapse = true
	return s
}

// IsVisible returns true if the section is visible.
func (s *Section) IsVisible() bool { return true }

// GetComponentType returns the component type.
func (s *Section) ComponentType() string    { return "layout_section" }
func (s *Section) GetComponentType() string { return s.ComponentType() }
func (s *Section) Render() templ.Component  { return SectionRender(s) }

// Schema returns the section components.
func (s *Section) Schema() []Component    { return s.Components }
func (s *Section) GetSchema() []Component { return s.Components }

// Grid represents a column grid.
type Grid struct {
	Columns    int
	Components []Component
}

// NewGrid creates a new grid.
func NewGrid(columns int) *Grid {
	return &Grid{
		Columns:    columns,
		Components: make([]Component, 0),
	}
}

// SetSchema sets the grid components.
func (g *Grid) SetSchema(components ...Component) *Grid {
	g.Components = components
	return g
}

// IsVisible returns true if the grid is visible.
func (g *Grid) IsVisible() bool { return true }

// GetComponentType returns the component type.
func (g *Grid) ComponentType() string    { return "layout_grid" }
func (g *Grid) GetComponentType() string { return g.ComponentType() }
func (g *Grid) Render() templ.Component  { return GridRender(g) }

// Schema returns the grid components.
func (g *Grid) Schema() []Component    { return g.Components }
func (g *Grid) GetSchema() []Component { return g.Components }

// Tabs represents a tab system.
type Tabs struct {
	TabItems []Tab
}

// Tab represents a tab.
type Tab struct {
	Label      string
	Icon       string
	Components []Component
}

// NewTabs creates a new tab system.
func NewTabs() *Tabs {
	return &Tabs{
		TabItems: make([]Tab, 0),
	}
}

// AddTab adds a tab.
func (t *Tabs) AddTab(label string, components ...Component) *Tabs {
	t.TabItems = append(t.TabItems, Tab{
		Label:      label,
		Components: components,
	})
	return t
}

// IsVisible returns true if the tabs are visible.
func (t *Tabs) IsVisible() bool { return true }

// GetComponentType returns the component type.
func (t *Tabs) ComponentType() string    { return "layout_tabs" }
func (t *Tabs) GetComponentType() string { return t.ComponentType() }
func (t *Tabs) Render() templ.Component  { return TabsRender(t) }

// Schema returns all components from all tabs.
func (t *Tabs) Schema() []Component {
	var all []Component
	for _, tab := range t.TabItems {
		all = append(all, tab.Components...)
	}
	return all
}
func (t *Tabs) GetSchema() []Component { return t.Schema() }

// WizardStep represents a single step in a Wizard.
type WizardStep struct {
	Label       string
	Description string
	Icon        string
	Components  []Component
}

// Wizard represents a multi-step form wizard.
type Wizard struct {
	Steps []*WizardStep
}

// NewWizard creates a new Wizard.
func NewWizard() *Wizard {
	return &Wizard{Steps: make([]*WizardStep, 0)}
}

// AddStep adds a step to the wizard.
func (w *Wizard) AddStep(label string, components ...Component) *Wizard {
	w.Steps = append(w.Steps, &WizardStep{
		Label:      label,
		Components: components,
	})
	return w
}

// WithDescription sets the description of the last added step.
func (w *Wizard) WithDescription(desc string) *Wizard {
	if len(w.Steps) > 0 {
		w.Steps[len(w.Steps)-1].Description = desc
	}
	return w
}

// WithIcon sets the icon of the last added step.
func (w *Wizard) WithIcon(icon string) *Wizard {
	if len(w.Steps) > 0 {
		w.Steps[len(w.Steps)-1].Icon = icon
	}
	return w
}

// IsVisible returns true.
func (w *Wizard) IsVisible() bool { return true }

// ComponentType returns the component type.
func (w *Wizard) ComponentType() string    { return "layout_wizard" }
func (w *Wizard) GetComponentType() string { return w.ComponentType() }
func (w *Wizard) Render() templ.Component  { return WizardRender(w) }

// Schema returns all components from all steps.
func (w *Wizard) Schema() []Component {
	var all []Component
	for _, step := range w.Steps {
		all = append(all, step.Components...)
	}
	return all
}
func (w *Wizard) GetSchema() []Component { return w.Schema() }

// CalloutColor defines the color/intent of a Callout.
type CalloutColor string

const (
	CalloutInfo    CalloutColor = "info"
	CalloutSuccess CalloutColor = "success"
	CalloutWarning CalloutColor = "warning"
	CalloutDanger  CalloutColor = "danger"
)

// Callout renders an informational banner inside a form.
type Callout struct {
	Heading    string
	Body       string
	Icon       string
	Color      CalloutColor
	Components []Component
}

// NewCallout creates a new Callout.
func NewCallout(heading string) *Callout {
	return &Callout{
		Heading:    heading,
		Color:      CalloutInfo,
		Components: make([]Component, 0),
	}
}

// WithBody sets the callout body text.
func (c *Callout) WithBody(body string) *Callout {
	c.Body = body
	return c
}

// WithIcon sets the callout icon.
func (c *Callout) WithIcon(icon string) *Callout {
	c.Icon = icon
	return c
}

// WithColor sets the callout color.
func (c *Callout) WithColor(color CalloutColor) *Callout {
	c.Color = color
	return c
}

// IsVisible returns true.
func (c *Callout) IsVisible() bool { return true }

// ComponentType returns the component type.
func (c *Callout) ComponentType() string    { return "layout_callout" }
func (c *Callout) GetComponentType() string { return c.ComponentType() }
func (c *Callout) Render() templ.Component  { return CalloutRender(c) }

// Schema returns the callout's nested components.
func (c *Callout) Schema() []Component    { return c.Components }
func (c *Callout) GetSchema() []Component { return c.Components }

// ---------------------------------------------------------------------------
// Fieldset — HTML fieldset with legend (like Filament's Fieldset layout).
// ---------------------------------------------------------------------------

// Fieldset groups form components inside a native HTML <fieldset> element.
type Fieldset struct {
	Legend     string
	Components []Component
}

// NewFieldset creates a new Fieldset with the given legend text.
func NewFieldset(legend string) *Fieldset {
	return &Fieldset{
		Legend:     legend,
		Components: make([]Component, 0),
	}
}

// SetSchema sets the nested components.
func (f *Fieldset) SetSchema(components ...Component) *Fieldset {
	f.Components = components
	return f
}

// IsVisible returns true.
func (f *Fieldset) IsVisible() bool { return true }

// ComponentType returns the component type identifier.
func (f *Fieldset) ComponentType() string    { return "layout_fieldset" }
func (f *Fieldset) GetComponentType() string { return "layout_fieldset" }
func (f *Fieldset) Render() templ.Component  { return FieldsetRender(f) }

// Schema returns the nested components.
func (f *Fieldset) Schema() []Component    { return f.Components }
func (f *Fieldset) GetSchema() []Component { return f.Components }

// ---------------------------------------------------------------------------
// Flex — flexible row layout (like Filament's Flex layout).
// ---------------------------------------------------------------------------

// FlexAlign defines alignment along the cross axis.
type FlexAlign string

const (
	FlexAlignStart   FlexAlign = "start"
	FlexAlignCenter  FlexAlign = "center"
	FlexAlignEnd     FlexAlign = "end"
	FlexAlignStretch FlexAlign = "stretch"
)

// Flex arranges form components in a flexible row.
type Flex struct {
	Components []Component
	GapSize    int       // 1-12, default 4
	Wrap       bool      // allow wrapping to next line
	Align      FlexAlign // cross-axis alignment
}

// NewFlex creates a new Flex layout with sensible defaults.
func NewFlex() *Flex {
	return &Flex{
		Components: make([]Component, 0),
		GapSize:    4,
		Wrap:       true,
		Align:      FlexAlignStart,
	}
}

// SetSchema sets the nested components.
func (f *Flex) SetSchema(components ...Component) *Flex {
	f.Components = components
	return f
}

// WithGap sets the gap between items (Tailwind gap-N, 1-12).
func (f *Flex) WithGap(gap int) *Flex {
	f.GapSize = gap
	return f
}

// NoWrap disables line wrapping.
func (f *Flex) NoWrap() *Flex {
	f.Wrap = false
	return f
}

// WithAlign sets the cross-axis alignment.
func (f *Flex) WithAlign(align FlexAlign) *Flex {
	f.Align = align
	return f
}

// IsVisible returns true.
func (f *Flex) IsVisible() bool { return true }

// ComponentType returns the component type identifier.
func (f *Flex) ComponentType() string    { return "layout_flex" }
func (f *Flex) GetComponentType() string { return "layout_flex" }
func (f *Flex) Render() templ.Component  { return FlexRender(f) }

// Schema returns the nested components.
func (f *Flex) Schema() []Component    { return f.Components }
func (f *Flex) GetSchema() []Component { return f.Components }

// ---------------------------------------------------------------------------
// Split — two-column fixed layout (like Filament's Split layout).
// ---------------------------------------------------------------------------

// Split divides the form into two columns (left and right).
type Split struct {
	Left       []Component
	Right      []Component
	LeftWidth  int // Tailwind col-span value for left column (default 1, right also 1 — equal)
	RightWidth int
}

// NewSplit creates a new Split layout with equal columns.
func NewSplit() *Split {
	return &Split{
		Left:       make([]Component, 0),
		Right:      make([]Component, 0),
		LeftWidth:  1,
		RightWidth: 1,
	}
}

// WithLeft sets the left column components.
func (s *Split) WithLeft(components ...Component) *Split {
	s.Left = components
	return s
}

// WithRight sets the right column components.
func (s *Split) WithRight(components ...Component) *Split {
	s.Right = components
	return s
}

// WithRatio sets the column ratio (e.g. WithRatio(2, 1) → 2/3 left, 1/3 right).
func (s *Split) WithRatio(left, right int) *Split {
	s.LeftWidth = left
	s.RightWidth = right
	return s
}

// IsVisible returns true.
func (s *Split) IsVisible() bool { return true }

// ComponentType returns the component type identifier.
func (s *Split) ComponentType() string    { return "layout_split" }
func (s *Split) GetComponentType() string { return "layout_split" }
func (s *Split) Render() templ.Component  { return SplitRender(s) }

// Schema returns all components from both columns.
func (s *Split) Schema() []Component {
	all := make([]Component, 0, len(s.Left)+len(s.Right))
	all = append(all, s.Left...)
	all = append(all, s.Right...)
	return all
}
func (s *Split) GetSchema() []Component { return s.Schema() }
