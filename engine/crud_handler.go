package engine

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/a-h/templ"
	datastarPkg "github.com/bozz33/sublimeadmin/datastar"
	formPkg "github.com/bozz33/sublimeadmin/form"
	"github.com/bozz33/sublimeadmin/ui/layouts"
)

const contextKeyListQuery contextKey = "list_query"

// GetListQuery retrieves the ListQuery from context (set by CRUDHandler.List).
func GetListQuery(ctx context.Context) *ListQuery {
	if q, ok := ctx.Value(contextKeyListQuery).(*ListQuery); ok {
		return q
	}
	return nil
}

// CRUDHandler automatically handles CRUD operations for a resource.
type CRUDHandler struct {
	Resource Resource
}

// NewCRUDHandler creates a CRUD handler for a given resource.
func NewCRUDHandler(r Resource) *CRUDHandler {
	return &CRUDHandler{Resource: r}
}

// List displays the list of items.
// Extracts filter_*, search, sort, dir, page, per_page from query params
// and injects them into context as both ActiveFilters and ListQuery.
func (h *CRUDHandler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	q := r.URL.Query()

	// Build ListQuery from all relevant params
	lq := &ListQuery{
		Filters: make(map[string]string),
		Search:  q.Get("search"),
		SortKey: q.Get("sort"),
		SortDir: q.Get("dir"),
		Page:    1,
		PerPage: 20,
	}
	if lq.SortDir != "desc" {
		lq.SortDir = "asc"
	}
	if p, err := strconv.Atoi(q.Get("page")); err == nil && p > 0 {
		lq.Page = p
	}
	if pp, err := strconv.Atoi(q.Get("per_page")); err == nil && pp > 0 && pp <= 200 {
		lq.PerPage = pp
	}
	for key, vals := range q {
		if strings.HasPrefix(key, "filter_") && len(vals) > 0 && vals[0] != "" {
			lq.Filters[strings.TrimPrefix(key, "filter_")] = vals[0]
		}
	}

	// Inject into context
	ctx = context.WithValue(ctx, contextKeyListQuery, lq)
	if len(lq.Filters) > 0 {
		ctx = context.WithValue(ctx, ContextKeyActiveFilters, lq.Filters)
	}

	component := h.Resource.Table(ctx)
	render(w, r, h.Resource.PluralLabel(), component)
}

// Create displays the creation form.
func (h *CRUDHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if !h.Resource.CanCreate(ctx) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	component := h.Resource.Form(ctx, nil)
	render(w, r, "Create "+h.Resource.Label(), component)
}

// View displays the read-only detail view (Infolist) for a resource.
// Only available if the resource implements ResourceViewable.
func (h *CRUDHandler) View(w http.ResponseWriter, r *http.Request, id string) {
	ctx := r.Context()

	if !h.Resource.CanRead(ctx) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	viewable, ok := h.Resource.(ResourceViewable)
	if !ok {
		// Resource has no View — redirect to edit
		http.Redirect(w, r, fmt.Sprintf("/%s/%s/edit", h.Resource.Slug(), id), http.StatusSeeOther)
		return
	}

	item, err := h.Resource.Get(ctx, id)
	if err != nil || item == nil {
		http.NotFound(w, r)
		return
	}

	component := viewable.View(ctx, item)
	render(w, r, h.Resource.Label(), component)
}

// Edit displays the edit form.
func (h *CRUDHandler) Edit(w http.ResponseWriter, r *http.Request, id string) {
	ctx := r.Context()

	item, err := h.Resource.Get(ctx, id)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// Inject relation managers into context if the resource supports them.
	if rwr, ok := h.Resource.(RelationManagerAware); ok {
		ctx = context.WithValue(ctx, contextKeyRelationManagers, rwr.GetRelationManagers())
	}

	component := h.Resource.Form(ctx, item)
	render(w, r, "Edit "+h.Resource.Label(), component)
}

// Store handles creation.
// If the resource returns a ValidationErrors error, the form is re-rendered
// with inline field errors instead of returning HTTP 500.
func (h *CRUDHandler) Store(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if !h.Resource.CanCreate(ctx) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	if err := h.Resource.Create(ctx, r); err != nil {
		ctx2 := injectFormErrors(ctx, err)
		w.WriteHeader(http.StatusUnprocessableEntity)
		component := h.Resource.Form(ctx2, nil)
		render(w, r.WithContext(ctx2), "Create "+h.Resource.Label(), component)
		return
	}

	http.Redirect(w, r, "/"+h.Resource.Slug(), http.StatusSeeOther)
}

// Update handles updates.
// If the resource returns a ValidationErrors error, the form is re-rendered
// with inline field errors instead of returning HTTP 500.
func (h *CRUDHandler) Update(w http.ResponseWriter, r *http.Request, id string) {
	ctx := r.Context()

	if !h.Resource.CanUpdate(ctx) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	if err := h.Resource.Update(ctx, id, r); err != nil {
		// Re-fetch item to pre-populate the form with submitted values.
		item, _ := h.Resource.Get(ctx, id)
		ctx2 := injectFormErrors(ctx, err)
		w.WriteHeader(http.StatusUnprocessableEntity)
		component := h.Resource.Form(ctx2, item)
		render(w, r.WithContext(ctx2), "Edit "+h.Resource.Label(), component)
		return
	}

	http.Redirect(w, r, "/"+h.Resource.Slug(), http.StatusSeeOther)
}

// Delete handles deletion.
// If the resource implements SoftDeletable, this performs a soft delete.
// Otherwise it permanently deletes the item.
func (h *CRUDHandler) Delete(w http.ResponseWriter, r *http.Request, id string) {
	ctx := r.Context()

	if !h.Resource.CanDelete(ctx) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Use soft delete when resource supports it.
	if sd, ok := h.Resource.(SoftDeletable); ok {
		if err := sd.SoftDelete(ctx, id); err != nil {
			http.Error(w, "Soft delete error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/"+h.Resource.Slug(), http.StatusSeeOther)
		return
	}

	if err := h.Resource.Delete(ctx, id); err != nil {
		http.Error(w, "Delete error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/"+h.Resource.Slug(), http.StatusSeeOther)
}

// Restore handles restoring a soft-deleted item.
// Only available when the resource implements SoftDeletable.
func (h *CRUDHandler) Restore(w http.ResponseWriter, r *http.Request, id string) {
	ctx := r.Context()

	if !h.Resource.CanDelete(ctx) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	sd, ok := h.Resource.(SoftDeletable)
	if !ok {
		http.Error(w, "Resource does not support soft delete", http.StatusBadRequest)
		return
	}

	if err := sd.Restore(ctx, id); err != nil {
		http.Error(w, "Restore error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/"+h.Resource.Slug(), http.StatusSeeOther)
}

// ForceDelete permanently deletes a (possibly soft-deleted) item.
// Only available when the resource implements SoftDeletable.
func (h *CRUDHandler) ForceDelete(w http.ResponseWriter, r *http.Request, id string) {
	ctx := r.Context()

	if !h.Resource.CanDelete(ctx) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	sd, ok := h.Resource.(SoftDeletable)
	if !ok {
		http.Error(w, "Resource does not support force delete", http.StatusBadRequest)
		return
	}

	if err := sd.ForceDelete(ctx, id); err != nil {
		http.Error(w, "Force delete error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/"+h.Resource.Slug(), http.StatusSeeOther)
}

// BulkDelete handles bulk deletion.
func (h *CRUDHandler) BulkDelete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if !h.Resource.CanDelete(ctx) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Form parsing error", http.StatusBadRequest)
		return
	}

	ids := r.Form["ids[]"]
	if len(ids) == 0 {
		http.Error(w, "No items selected", http.StatusBadRequest)
		return
	}

	if err := h.Resource.BulkDelete(ctx, ids); err != nil {
		http.Error(w, "Bulk delete error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/"+h.Resource.Slug(), http.StatusSeeOther)
}

// ServeHTTP implements http.Handler with automatic routing.
func (h *CRUDHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/"+h.Resource.Slug())
	path = strings.TrimPrefix(path, "/")
	parts := strings.Split(path, "/")

	switch r.Method {
	case http.MethodGet:
		h.routeGET(w, r, path, parts)
	case http.MethodPost:
		h.routePOST(w, r, path, parts)
	case http.MethodPatch:
		if len(parts) >= 1 && parts[0] != "" {
			h.Patch(w, r, parts[0])
		} else {
			http.NotFound(w, r)
		}
	case http.MethodDelete:
		if len(parts) >= 1 {
			h.Delete(w, r, parts[0])
		}
	}
}

// Patch handles partial updates from inline-edit table columns.
// Route: PATCH /{slug}/{id}
//
// Expects a single field=value pair in the request body.
// If the resource implements ResourcePatchable, its PatchField method is called.
// Otherwise the full Update path is used as a fallback.
// On success returns a Datastar SSE toast notification.
func (h *CRUDHandler) Patch(w http.ResponseWriter, r *http.Request, id string) {
	ctx := r.Context()

	if !h.Resource.CanUpdate(ctx) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	sse := datastarPkg.NewSSE(w)

	// Prefer ResourcePatchable when available (more efficient — single field update).
	if patchable, ok := h.Resource.(ResourcePatchable); ok {
		// Collect all submitted form fields as a map.
		fields := make(map[string]string, len(r.Form))
		for k, vals := range r.Form {
			if len(vals) > 0 {
				fields[k] = vals[0]
			}
		}
		if err := patchable.PatchField(ctx, id, fields); err != nil {
			sse.Toast(htmlEscapeString(err.Error()), "error")
			return
		}
		sse.Toast("Saved", "success")
		return
	}

	// Fallback: full update.
	if err := h.Resource.Update(ctx, id, r); err != nil {
		sse.Toast(htmlEscapeString(err.Error()), "error")
		return
	}
	sse.Toast("Saved", "success")
}

// ValidateField handles real-time per-field validation via Datastar SSE.
//
// Route: GET /{slug}/validate-field?field=name&value=xxx
//
// If the resource implements ResourceValidator, the field is validated and a
// Datastar SSE fragment is returned updating the #field-error-{name} element.
// If the resource does not implement ResourceValidator, 204 No Content is returned.
func (h *CRUDHandler) ValidateField(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	field := r.URL.Query().Get("field")
	value := r.URL.Query().Get("value")
	if field == "" {
		http.Error(w, "missing field parameter", http.StatusBadRequest)
		return
	}

	validator, ok := h.Resource.(ResourceValidator)
	if !ok {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	sse := datastarPkg.NewSSE(w)
	if err := validator.ValidateField(ctx, field, value); err != nil {
		// Send error fragment
		errHTML := fmt.Sprintf(
			`<p id="field-error-%s" class="mt-1 text-sm text-red-600 dark:text-red-400">%s</p>`,
			field, htmlEscapeString(err.Error()),
		)
		sse.MergeFragment(errHTML)
	} else {
		// Clear any previous error
		clearHTML := fmt.Sprintf(`<p id="field-error-%s" class="mt-1 text-sm"></p>`, field)
		sse.MergeFragment(clearHTML)
	}
}

// htmlEscapeString escapes a string for safe HTML insertion.
func htmlEscapeString(s string) string {
	r := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
		`"`, "&#34;",
		"'", "&#39;",
	)
	return r.Replace(s)
}

// routeGET dispatches GET requests.
func (h *CRUDHandler) routeGET(w http.ResponseWriter, r *http.Request, path string, parts []string) {
	switch {
	case path == "" || path == "/":
		h.List(w, r)
	case path == "create":
		h.Create(w, r)
	case path == "validate-field":
		h.ValidateField(w, r)
	case len(parts) == 2 && parts[1] == "edit":
		h.Edit(w, r, parts[0])
	case len(parts) == 1 && parts[0] != "":
		h.View(w, r, parts[0])
	default:
		http.NotFound(w, r)
	}
}

// routePOST dispatches POST requests (including _method override).
func (h *CRUDHandler) routePOST(w http.ResponseWriter, r *http.Request, path string, parts []string) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	if r.FormValue("_method") == "DELETE" && len(parts) >= 1 {
		h.Delete(w, r, parts[0])
		return
	}
	switch {
	case path == "bulk-delete":
		h.BulkDelete(w, r)
	case path == "" || path == "/":
		h.Store(w, r)
	case len(parts) >= 1:
		h.Update(w, r, parts[0])
	}
}

// injectFormErrors converts an error into FormErrors and injects it into context.
// If the error implements ValidationErrors (with FieldErrors()), per-field errors
// are used. Otherwise, the error message is stored under the "_error" key.
func injectFormErrors(ctx context.Context, err error) context.Context {
	var fe formPkg.FormErrors
	var ve formPkg.ValidationErrors
	switch {
	case errors.As(err, &fe):
		return formPkg.WithFormErrors(ctx, fe)
	case errors.As(err, &ve):
		return formPkg.WithFormErrors(ctx, formPkg.FormErrors(ve.FieldErrors()))
	default:
		return formPkg.WithFormErrors(ctx, formPkg.FormErrors{"_error": err.Error()})
	}
}

// render is a helper to display a component in the layout.
func render(w http.ResponseWriter, r *http.Request, title string, content templ.Component) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fullPage := layouts.Page(title, content)
	_ = fullPage.Render(r.Context(), w)
}
