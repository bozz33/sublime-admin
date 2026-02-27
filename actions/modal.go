package actions

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

// ModalSize defines the size of a modal dialog.
type ModalSize string

const (
	ModalSizeSM   ModalSize = "sm"
	ModalSizeMD   ModalSize = "md"
	ModalSizeLG   ModalSize = "lg"
	ModalSizeXL   ModalSize = "xl"
	ModalSizeFull ModalSize = "full"
)

// ModalAction is an action that opens a modal dialog.
// It extends Action with modal-specific configuration.
type ModalAction struct {
	*Action

	// Modal display
	Size                ModalSize
	CloseOnEsc          bool
	CloseOnBack         bool
	IsSlideOver         bool // renders as a slide-over panel instead of centered modal
	DatabaseTransaction bool // wraps the action handler in a DB transaction

	// Form inside modal (optional)
	FormFields []ModalField
	FormAction string // POST URL for the form
	FormMethod string // default "POST"

	// HTMX: load modal content from a URL
	HTMXLoadURL string // if set, modal content is fetched via HTMX
}

// ModalField describes a simple field rendered inside a modal form.
type ModalField struct {
	Name        string
	Label       string
	Type        string // "text", "email", "select", "textarea", "hidden"
	Placeholder string
	Required    bool
	Options     []ModalFieldOption
	Value       string
}

// ModalFieldOption is an option for a select field inside a modal.
type ModalFieldOption struct {
	Value string
	Label string
}

// NewModal creates a new ModalAction wrapping a base Action.
func NewModal(name string) *ModalAction {
	return &ModalAction{
		Action:      New(name),
		Size:        ModalSizeMD,
		CloseOnEsc:  true,
		CloseOnBack: true,
		FormMethod:  "POST",
	}
}

// WithSize sets the modal size.
func (m *ModalAction) WithSize(size ModalSize) *ModalAction {
	m.Size = size
	return m
}

// SlideOver makes the modal render as a slide-over panel from the right.
func (m *ModalAction) SlideOver() *ModalAction {
	m.IsSlideOver = true
	return m
}

// WithDatabaseTransaction wraps the action handler in a database transaction.
func (m *ModalAction) WithDatabaseTransaction() *ModalAction {
	m.DatabaseTransaction = true
	return m
}

// WithHTMXLoad sets a URL to load the modal content via HTMX.
func (m *ModalAction) WithHTMXLoad(url string) *ModalAction {
	m.HTMXLoadURL = url
	return m
}

// WithForm adds a simple inline form to the modal.
func (m *ModalAction) WithForm(action string, fields ...ModalField) *ModalAction {
	m.FormAction = action
	m.FormFields = fields
	return m
}

// SetLabel sets the label (fluent, returns *ModalAction).
func (m *ModalAction) SetLabel(label string) *ModalAction {
	m.Action.Label = label
	return m
}

// SetIcon sets the icon (fluent, returns *ModalAction).
func (m *ModalAction) SetIcon(icon string) *ModalAction {
	m.Action.Icon = icon
	return m
}

// SetColor sets the color (fluent, returns *ModalAction).
func (m *ModalAction) SetColor(color string) *ModalAction {
	m.Action.Color = color
	return m
}

// Authorize sets the authorization function (fluent, returns *ModalAction).
func (m *ModalAction) Authorize(fn func(ctx context.Context, item any) bool) *ModalAction {
	m.Action.AuthorizeFunc = fn
	return m
}

// OnSuccess sets the success callback (fluent, returns *ModalAction).
func (m *ModalAction) OnSuccess(fn func(ctx context.Context, item any)) *ModalAction {
	m.Action.OnSuccessFunc = fn
	return m
}

// WithSuccessMessage sets the flash message shown on success (fluent, returns *ModalAction).
func (m *ModalAction) WithSuccessMessage(msg string) *ModalAction {
	m.Action.SuccessMessage = msg
	return m
}

// ServeHTTP handles modal rendering (GET) and form submission (POST).
//
//   - GET  → returns the modal HTML fragment (form or confirmation dialog)
//   - POST → processes the form submission and redirects back
func (m *ModalAction) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := m.renderModalFragment(w, r); err != nil {
			http.Error(w, "failed to render modal", http.StatusInternalServerError)
		}

	case http.MethodPost:
		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		// Execute BeforeFunc if configured
		if m.Action.BeforeFunc != nil {
			if err := m.Action.BeforeFunc(r.Context(), nil); err != nil {
				http.Error(w, m.Action.FailureMessage, http.StatusInternalServerError)
				return
			}
		}
		// Redirect to FormAction or Referer
		redirectTo := m.FormAction
		if redirectTo == "" {
			redirectTo = r.Header.Get("Referer")
		}
		if redirectTo == "" {
			redirectTo = "/"
		}
		http.Redirect(w, r, redirectTo, http.StatusSeeOther)

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// renderModalFragment writes the modal HTML fragment to w.
func (m *ModalAction) renderModalFragment(w http.ResponseWriter, r *http.Request) error {
	action := m.FormAction
	if action == "" {
		action = r.URL.Path
	}
	method := m.FormMethod
	if method == "" {
		method = "POST"
	}
	sizeClass := modalSizeClass(m.Size)

	// Determine if this is a slide-over or centered modal
	wrapperClass := "fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4"
	panelClass := sizeClass + " w-full bg-white dark:bg-gray-800 rounded-2xl shadow-xl overflow-hidden"
	if m.IsSlideOver {
		wrapperClass = "fixed inset-0 z-50 flex items-end justify-end bg-black/50"
		panelClass = "h-full " + sizeClass + " bg-white dark:bg-gray-800 shadow-xl overflow-y-auto"
	}

	fmt.Fprintf(w, `<div id="modal-fragment" class="%s">`, wrapperClass)
	fmt.Fprintf(w, `<div class="%s">`, panelClass)

	// Header
	title := m.Action.ModalTitle
	if title == "" {
		title = m.Action.Label
	}
	fmt.Fprintf(w, `<div class="flex items-center justify-between px-6 py-4 border-b border-gray-200 dark:border-gray-700">`)
	fmt.Fprintf(w, `<h2 class="text-lg font-semibold text-gray-900 dark:text-white">%s</h2>`, htmlEscape(title))
	fmt.Fprintf(w, `<button type="button" onclick="document.getElementById('modal-fragment').remove()" class="p-2 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-700 text-gray-400 hover:text-gray-600 transition-colors"><span class="material-icons-outlined">close</span></button>`)
	fmt.Fprintf(w, `</div>`)

	// Body
	fmt.Fprintf(w, `<div class="px-6 py-4">`)
	if m.Action.ModalDescription != "" {
		fmt.Fprintf(w, `<p class="text-sm text-gray-600 dark:text-gray-400 mb-4">%s</p>`, htmlEscape(m.Action.ModalDescription))
	}

	// Form with fields
	if len(m.FormFields) > 0 {
		fmt.Fprintf(w, `<form method="POST" action="%s" class="space-y-4">`, htmlEscape(action))
		if method != "POST" {
			fmt.Fprintf(w, `<input type="hidden" name="_method" value="%s"/>`, htmlEscape(method))
		}
		for _, field := range m.FormFields {
			renderModalFieldHTML(w, field)
		}
		fmt.Fprintf(w, `<div class="flex items-center justify-end gap-3 pt-4 border-t border-gray-200 dark:border-gray-700">`)
		cancelLabel := m.Action.CancelLabel
		if cancelLabel == "" {
			cancelLabel = "Cancel"
		}
		confirmLabel := m.Action.ConfirmLabel
		if confirmLabel == "" {
			confirmLabel = "Submit"
		}
		fmt.Fprintf(w, `<button type="button" onclick="document.getElementById('modal-fragment').remove()" class="px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-300 border border-gray-300 dark:border-gray-600 rounded-xl hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors">%s</button>`, htmlEscape(cancelLabel))
		fmt.Fprintf(w, `<button type="submit" class="px-4 py-2 text-sm font-semibold text-white bg-primary-600 hover:bg-primary-700 rounded-xl transition-colors">%s</button>`, htmlEscape(confirmLabel))
		fmt.Fprintf(w, `</div></form>`)
	} else if m.Action.RequiresConfirmation {
		// Pure confirmation dialog (no form fields)
		fmt.Fprintf(w, `<div class="flex items-start gap-4">`)
		fmt.Fprintf(w, `<div class="w-12 h-12 bg-red-100 dark:bg-red-900/30 rounded-full flex items-center justify-center flex-shrink-0"><span class="material-icons-outlined text-red-500 text-2xl">warning</span></div>`)
		fmt.Fprintf(w, `<div class="flex-1"><p class="text-sm text-gray-600 dark:text-gray-400">%s</p></div>`, htmlEscape(m.Action.ModalDescription))
		fmt.Fprintf(w, `</div>`)
		fmt.Fprintf(w, `<div class="flex items-center justify-end gap-3 mt-6">`)
		cancelLabel := m.Action.CancelLabel
		if cancelLabel == "" {
			cancelLabel = "Cancel"
		}
		confirmLabel := m.Action.ConfirmLabel
		if confirmLabel == "" {
			confirmLabel = "Confirm"
		}
		colorClass := "bg-red-600 hover:bg-red-700"
		if m.Action.Color != "danger" {
			colorClass = "bg-primary-600 hover:bg-primary-700"
		}
		fmt.Fprintf(w, `<button type="button" onclick="document.getElementById('modal-fragment').remove()" class="px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-300 border border-gray-300 dark:border-gray-600 rounded-xl hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors">%s</button>`, htmlEscape(cancelLabel))
		fmt.Fprintf(w, `<form method="POST" action="%s" style="display:inline">`, htmlEscape(action))
		if method != "POST" {
			fmt.Fprintf(w, `<input type="hidden" name="_method" value="%s"/>`, htmlEscape(method))
		}
		fmt.Fprintf(w, `<button type="submit" class="px-4 py-2 text-sm font-semibold text-white %s rounded-xl transition-colors">%s</button>`, colorClass, htmlEscape(confirmLabel))
		fmt.Fprintf(w, `</form></div>`)
	}

	fmt.Fprintf(w, `</div></div></div>`)
	return nil
}

// renderModalFieldHTML writes a single form field HTML for a modal.
func renderModalFieldHTML(w http.ResponseWriter, f ModalField) {
	fmt.Fprintf(w, `<div class="space-y-1">`)
	if f.Label != "" {
		req := ""
		if f.Required {
			req = `<span class="text-red-500 ml-1">*</span>`
		}
		fmt.Fprintf(w, `<label for="%s" class="block text-sm font-medium text-gray-700 dark:text-gray-300">%s%s</label>`,
			htmlEscape(f.Name), htmlEscape(f.Label), req)
	}

	inputClass := "block w-full rounded-xl border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 px-3 py-2 text-sm text-gray-900 dark:text-white placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-primary-500 transition-colors"

	switch f.Type {
	case "select":
		fmt.Fprintf(w, `<select id="%s" name="%s" class="%s">`, htmlEscape(f.Name), htmlEscape(f.Name), inputClass)
		if f.Placeholder != "" {
			fmt.Fprintf(w, `<option value="">%s</option>`, htmlEscape(f.Placeholder))
		}
		for _, opt := range f.Options {
			selected := ""
			if opt.Value == f.Value {
				selected = ` selected`
			}
			fmt.Fprintf(w, `<option value="%s"%s>%s</option>`, htmlEscape(opt.Value), selected, htmlEscape(opt.Label))
		}
		fmt.Fprintf(w, `</select>`)
	case "textarea":
		fmt.Fprintf(w, `<textarea id="%s" name="%s" placeholder="%s" class="%s resize-y min-h-[80px]">%s</textarea>`,
			htmlEscape(f.Name), htmlEscape(f.Name), htmlEscape(f.Placeholder), inputClass, htmlEscape(f.Value))
	case "hidden":
		fmt.Fprintf(w, `<input type="hidden" name="%s" value="%s"/>`, htmlEscape(f.Name), htmlEscape(f.Value))
	default:
		fieldType := f.Type
		if fieldType == "" {
			fieldType = "text"
		}
		req := ""
		if f.Required {
			req = ` required`
		}
		fmt.Fprintf(w, `<input type="%s" id="%s" name="%s" value="%s" placeholder="%s" class="%s"%s/>`,
			htmlEscape(fieldType), htmlEscape(f.Name), htmlEscape(f.Name),
			htmlEscape(f.Value), htmlEscape(f.Placeholder), inputClass, req)
	}
	fmt.Fprintf(w, `</div>`)
}

// modalSizeClass returns the Tailwind width class for a modal size.
func modalSizeClass(size ModalSize) string {
	switch size {
	case ModalSizeSM:
		return "max-w-sm"
	case ModalSizeLG:
		return "max-w-2xl"
	case ModalSizeXL:
		return "max-w-4xl"
	case ModalSizeFull:
		return "max-w-full"
	default:
		return "max-w-lg"
	}
}

// htmlEscape escapes a string for safe HTML attribute/content usage.
func htmlEscape(s string) string {
	r := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&#34;", "'", "&#39;")
	return r.Replace(s)
}

// ConfirmAction creates a pre-configured delete confirmation modal action.
func ConfirmAction(name, title, description string) *ModalAction {
	m := NewModal(name)
	m.Action.RequiresConfirmation = true
	m.Action.ModalTitle = title
	m.Action.ModalDescription = description
	m.Action.ConfirmLabel = "Confirm"
	m.Action.CancelLabel = "Cancel"
	m.Action.Type = Button
	return m
}

// DeleteModalAction creates a standard delete confirmation modal.
func DeleteModalAction(baseURL string) *ModalAction {
	m := ConfirmAction("delete", "Delete this item?", "This action cannot be undone.")
	m.Action.SetLabel("Delete").
		SetIcon("delete").
		SetColor("danger").
		SetUrl(func(item any) string {
			return baseURL + "/" + getItemID(item)
		})
	m.Action.Method = "DELETE"
	m.Action.SuccessMessage = "Item deleted successfully."
	m.Action.FailureMessage = "Failed to delete item."
	return m
}
