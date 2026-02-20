package engine

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/a-h/templ"
	"github.com/bozz33/sublimeadmin/ui/layouts"
)

// PaginationParams represents pagination query parameters.
type PaginationParams struct {
	Page    int
	PerPage int
	Search  string
	Sort    string
	Order   string // "asc" or "desc"
}

// ParsePaginationParams extracts pagination from HTTP request.
func ParsePaginationParams(r *http.Request) PaginationParams {
	params := PaginationParams{
		Page:    1,
		PerPage: 15,
		Search:  "",
		Sort:    "",
		Order:   "asc",
	}

	// Parse page
	if page := r.URL.Query().Get("page"); page != "" {
		if p, err := strconv.Atoi(page); err == nil && p > 0 {
			params.Page = p
		}
	}

	// Parse per_page
	if perPage := r.URL.Query().Get("per_page"); perPage != "" {
		if pp, err := strconv.Atoi(perPage); err == nil && pp > 0 && pp <= 100 {
			params.PerPage = pp
		}
	}

	// Parse search
	params.Search = r.URL.Query().Get("search")

	// Parse sort
	params.Sort = r.URL.Query().Get("sort")

	// Parse order
	if order := r.URL.Query().Get("order"); strings.ToLower(order) == "desc" {
		params.Order = "desc"
	}

	return params
}

// PaginatedResult represents a paginated list result.
type PaginatedResult struct {
	Items      []any
	Total      int
	Page       int
	PerPage    int
	TotalPages int
	HasNext    bool
	HasPrev    bool
}

// NewPaginatedResult creates a paginated result.
func NewPaginatedResult(items []any, total, page, perPage int) *PaginatedResult {
	totalPages := (total + perPage - 1) / perPage
	if totalPages == 0 {
		totalPages = 1
	}

	return &PaginatedResult{
		Items:      items,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
	}
}

// PaginatedListFunc is a function that returns a paginated list of items.
type PaginatedListFunc func(ctx context.Context, params PaginationParams) (*PaginatedResult, error)

// PaginatedResource extends Resource with server-side pagination support.
type PaginatedResource interface {
	Resource
	// ListPaginated returns a paginated list of items.
	ListPaginated(ctx context.Context, params PaginationParams) (*PaginatedResult, error)
}

// PaginatedCRUDHandler handles CRUD operations with server-side pagination.
type PaginatedCRUDHandler struct {
	Resource Resource
}

// NewPaginatedCRUDHandler creates a paginated CRUD handler.
func NewPaginatedCRUDHandler(r Resource) *PaginatedCRUDHandler {
	return &PaginatedCRUDHandler{Resource: r}
}

// List displays the paginated list of items.
func (h *PaginatedCRUDHandler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	params := ParsePaginationParams(r)

	var component templ.Component
	var title string

	if paginated, ok := h.Resource.(PaginatedResource); ok {
		// Use server-side pagination
		result, err := paginated.ListPaginated(ctx, params)
		if err != nil {
			http.Error(w, "List error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Create table with pagination info
		tableComponent := h.Resource.Table(ctx)
		title = h.Resource.PluralLabel()

		// Wrap with pagination metadata in context for the table component
		ctx = context.WithValue(ctx, "pagination_result", result)
		component = tableComponent
	} else {
		// Fallback to regular list (loads all items)
		component = h.Resource.Table(ctx)
		title = h.Resource.PluralLabel()
	}

	renderPage(w, r, title, component)
}

// Create displays the creation form.
func (h *PaginatedCRUDHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if !h.Resource.CanCreate(ctx) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	component := h.Resource.Form(ctx, nil)
	render(w, r, "Create "+h.Resource.Label(), component)
}

// Edit displays the edit form.
func (h *PaginatedCRUDHandler) Edit(w http.ResponseWriter, r *http.Request, id string) {
	ctx := r.Context()

	item, err := h.Resource.Get(ctx, id)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	component := h.Resource.Form(ctx, item)
	render(w, r, "Edit "+h.Resource.Label(), component)
}

// Store handles creation.
func (h *PaginatedCRUDHandler) Store(w http.ResponseWriter, r *http.Request) {
	if !h.Resource.CanCreate(r.Context()) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	if err := h.Resource.Create(r.Context(), r); err != nil {
		http.Error(w, "Creation error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Preserve pagination params when redirecting
	redirectURL := "/" + h.Resource.Slug()
	if r.URL.Query().Get("page") != "" {
		redirectURL += "?page=" + r.URL.Query().Get("page")
		if r.URL.Query().Get("per_page") != "" {
			redirectURL += "&per_page=" + r.URL.Query().Get("per_page")
		}
	}

	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}

// Update handles updates.
func (h *PaginatedCRUDHandler) Update(w http.ResponseWriter, r *http.Request, id string) {
	if !h.Resource.CanUpdate(r.Context()) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	if err := h.Resource.Update(r.Context(), id, r); err != nil {
		http.Error(w, "Update error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Preserve pagination params when redirecting
	redirectURL := "/" + h.Resource.Slug()
	if r.URL.Query().Get("page") != "" {
		redirectURL += "?page=" + r.URL.Query().Get("page")
	}

	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}

// Delete handles deletion.
func (h *PaginatedCRUDHandler) Delete(w http.ResponseWriter, r *http.Request, id string) {
	ctx := r.Context()

	if !h.Resource.CanDelete(ctx) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	if err := h.Resource.Delete(ctx, id); err != nil {
		http.Error(w, "Delete error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Preserve pagination params when redirecting
	redirectURL := "/" + h.Resource.Slug()
	if r.URL.Query().Get("page") != "" {
		redirectURL += "?page=" + r.URL.Query().Get("page")
	}

	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}

// BulkDelete handles bulk deletion.
func (h *PaginatedCRUDHandler) BulkDelete(w http.ResponseWriter, r *http.Request) {
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

	// Preserve pagination params when redirecting
	redirectURL := "/" + h.Resource.Slug()
	if r.URL.Query().Get("page") != "" {
		redirectURL += "?page=" + r.URL.Query().Get("page")
	}

	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}

// ServeHTTP implements http.Handler with automatic routing.
func (h *PaginatedCRUDHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/"+h.Resource.Slug())
	path = strings.TrimPrefix(path, "/")

	parts := strings.Split(path, "/")

	switch r.Method {
	case http.MethodGet:
		if path == "" || path == "/" {
			h.List(w, r)
		} else if path == "create" {
			h.Create(w, r)
		} else if len(parts) == 2 && parts[1] == "edit" {
			h.Edit(w, r, parts[0])
		} else {
			http.Redirect(w, r, fmt.Sprintf("/%s/%s/edit", h.Resource.Slug(), parts[0]), http.StatusSeeOther)
		}

	case http.MethodPost:
		r.ParseForm()
		methodOverride := r.FormValue("_method")

		if methodOverride == "DELETE" && len(parts) >= 1 {
			h.Delete(w, r, parts[0])
			return
		}

		if path == "bulk-delete" {
			h.BulkDelete(w, r)
			return
		}

		if path == "" || path == "/" {
			h.Store(w, r)
		} else if len(parts) >= 1 {
			h.Update(w, r, parts[0])
		}

	case http.MethodDelete:
		if len(parts) >= 1 {
			h.Delete(w, r, parts[0])
		}
	}
}

// renderPage is a helper to display a component in the layout.
func renderPage(w http.ResponseWriter, r *http.Request, title string, content templ.Component) {
	fullPage := layouts.Page(title, content)
	fullPage.Render(r.Context(), w)
}

// SimplePaginatedResource extends SimpleResource with pagination support.
type SimplePaginatedResource struct {
	*SimpleResource
	paginatedListFunc PaginatedListFunc
}

// NewSimplePaginatedResource creates a simple paginated resource.
func NewSimplePaginatedResource(slug, label, pluralLabel string) *SimplePaginatedResource {
	return &SimplePaginatedResource{
		SimpleResource: NewSimpleResource(slug, label, pluralLabel),
	}
}

// WithPaginatedList sets the paginated list function.
func (s *SimplePaginatedResource) WithPaginatedList(fn PaginatedListFunc) *SimplePaginatedResource {
	s.paginatedListFunc = fn
	return s
}

// ListPaginated implements PaginatedResource interface.
func (s *SimplePaginatedResource) ListPaginated(ctx context.Context, params PaginationParams) (*PaginatedResult, error) {
	if s.paginatedListFunc != nil {
		return s.paginatedListFunc(ctx, params)
	}
	return nil, fmt.Errorf("paginated list function not configured")
}
