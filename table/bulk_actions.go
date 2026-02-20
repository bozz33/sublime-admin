package table

import "fmt"

// BulkAction defines an action that can be applied to multiple selected rows.
type BulkAction struct {
	Name                 string
	Label                string
	Icon                 string
	Color                string
	Method               string
	URL                  string
	RequiresConfirmation bool
	ModalTitle           string
	ModalDescription     string
}

// NewBulkAction creates a new bulk action.
func NewBulkAction(name string) *BulkAction {
	return &BulkAction{
		Name:   name,
		Label:  name,
		Color:  "gray",
		Method: "POST",
	}
}

// SetLabel sets the display label.
func (b *BulkAction) SetLabel(label string) *BulkAction {
	b.Label = label
	return b
}

// SetIcon sets the icon name.
func (b *BulkAction) SetIcon(icon string) *BulkAction {
	b.Icon = icon
	return b
}

// SetColor sets the color variant.
func (b *BulkAction) SetColor(color string) *BulkAction {
	b.Color = color
	return b
}

// SetURL sets the endpoint URL for the bulk action.
func (b *BulkAction) SetURL(url string) *BulkAction {
	b.URL = url
	return b
}

// SetMethod sets the HTTP method (default POST).
func (b *BulkAction) SetMethod(method string) *BulkAction {
	b.Method = method
	return b
}

// RequiresDialog enables a confirmation modal before executing the action.
func (b *BulkAction) RequiresDialog(title, desc string) *BulkAction {
	b.RequiresConfirmation = true
	b.ModalTitle = title
	b.ModalDescription = desc
	return b
}

// BulkDelete creates a standard bulk delete action.
func BulkDelete(baseURL string) *BulkAction {
	return NewBulkAction("bulk-delete").
		SetLabel("Delete selected").
		SetIcon("trash").
		SetColor("danger").
		SetURL(fmt.Sprintf("%s/bulk-delete", baseURL)).
		SetMethod("DELETE").
		RequiresDialog("Delete selected items?", "This action cannot be undone.")
}

// BulkExport creates a standard bulk export action (CSV).
func BulkExport(baseURL string) *BulkAction {
	return NewBulkAction("bulk-export").
		SetLabel("Export selected").
		SetIcon("download").
		SetColor("secondary").
		SetURL(fmt.Sprintf("%s/bulk-export?format=csv", baseURL)).
		SetMethod("POST")
}

// BulkRestore creates a standard bulk restore action (soft-delete recovery).
func BulkRestore(baseURL string) *BulkAction {
	return NewBulkAction("bulk-restore").
		SetLabel("Restore selected").
		SetIcon("refresh").
		SetColor("success").
		SetURL(fmt.Sprintf("%s/bulk-restore", baseURL)).
		SetMethod("POST")
}

// WithBulkActions adds bulk actions to the table.
func (t *Table) WithBulkActions(actions ...*BulkAction) *Table {
	t.BulkActions = append(t.BulkActions, actions...)
	return t
}
