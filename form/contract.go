package form

import (
	"html/template"

	"github.com/a-h/templ"
)

// Component is the base interface for any form element.
type Component interface {
	IsVisible() bool
	ComponentType() string
	GetComponentType() string // Alias for templates
	Render() templ.Component
}

// Field defines the contract for an input field.
type Field interface {
	Component
	Name() string
	Label() string
	Value() any
	ValueString() string
	Placeholder() string
	Help() string
	IsRequired() bool
	IsDisabled() bool
	Attributes() template.HTMLAttr
	Rules() []string
}

// Layout defines the contract for layout.
type Layout interface {
	Component
	Schema() []Component
	GetSchema() []Component // Alias for templates
}
