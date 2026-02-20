package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/a-h/templ"
	"github.com/bozz33/sublimeadmin/ui/layouts"
)

// -------------------------
// PaginationConfig
// -------------------------

// PaginationConfig holds global pagination settings.
type PaginationConfig struct {
	DefaultSize  int      // default items per page (default: 15)
	MaxSize      int      // maximum allowed per page (default: 100)
	PageStart    int      // page numbering start: 0 or 1 (default: 1)
	SortParams   []string // query param names for sort (default: ["sort"])
	PageParams   []string // query param names for page (default: ["page"])
	SizeParams   []string // query param names for size (default: ["per_page","size"])
	SearchParams []string // query param names for search (default: ["search","q"])
	FilterParams []string // query param names for filters (default: ["filters"])
	OrderParams  []string // query param names for order (default: ["order"])
	ErrorEnabled bool     // include error info in Page response
}

func defaultPaginationConfig() PaginationConfig {
	return PaginationConfig{
		DefaultSize:  15,
		MaxSize:      100,
		PageStart:    1,
		SortParams:   []string{"sort"},
		PageParams:   []string{"page"},
		SizeParams:   []string{"per_page", "size"},
		SearchParams: []string{"search", "q"},
		FilterParams: []string{"filters"},
		OrderParams:  []string{"order"},
	}
}

// SortField represents a single sort column with direction.
// Prefix "-" means DESC, no prefix means ASC. Example: "-name,id" → [{name,DESC},{id,ASC}]
type SortField struct {
	Field string
	Desc  bool
}

// ParseSortFields parses a comma-separated sort string.
// Prefix "-" means DESC, no prefix means ASC.
// Example: "-name,id,+price" → [{name,DESC},{id,ASC},{price,ASC}]
func ParseSortFields(s string) []SortField {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	result := make([]SortField, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if strings.HasPrefix(p, "-") {
			result = append(result, SortField{Field: p[1:], Desc: true})
		} else if strings.HasPrefix(p, "+") {
			result = append(result, SortField{Field: p[1:], Desc: false})
		} else {
			result = append(result, SortField{Field: p, Desc: false})
		}
	}
	return result
}

// FilterOperator represents a filter comparison operator.
type FilterOperator string

const (
	FilterEq        FilterOperator = "eq"
	FilterNeq       FilterOperator = "neq"
	FilterLike      FilterOperator = "like"
	FilterNotLike   FilterOperator = "not like"
	FilterGt        FilterOperator = "gt"
	FilterGte       FilterOperator = "gte"
	FilterLt        FilterOperator = "lt"
	FilterLte       FilterOperator = "lte"
	FilterBetween   FilterOperator = "between"
	FilterIn        FilterOperator = "in"
	FilterNotIn     FilterOperator = "not in"
	FilterIsNull    FilterOperator = "is null"
	FilterIsNotNull FilterOperator = "is not null"
)

// FilterExpr represents a single filter condition.
// Format: ["field","operator","value"]
type FilterExpr struct {
	Field    string
	Operator FilterOperator
	Value    any
	// For compound filters (AND/OR groups)
	And []*FilterExpr
	Or  []*FilterExpr
}

// ParseFiltersJSON parses JSON filter expressions.
// Supports:
//
//	["name","john"]                        → name = 'john'
//	["name","like","john"]                 → name LIKE '%john%'
//	["age","between",[20,25]]              → age BETWEEN 20 AND 25
//	[["name","like","john"],["OR"],["age","gt",18]]
func ParseFiltersJSON(raw string) ([]*FilterExpr, error) {
	if raw == "" {
		return nil, nil
	}
	var data any
	if err := json.Unmarshal([]byte(raw), &data); err != nil {
		return nil, fmt.Errorf("invalid filters JSON: %w", err)
	}
	return parseFilterNode(data)
}

func parseFilterNode(node any) ([]*FilterExpr, error) {
	arr, ok := node.([]any)
	if !ok || len(arr) == 0 {
		return nil, nil
	}

	// Simple filter: ["field","value"] or ["field","op","value"]
	if s, ok := arr[0].(string); ok && !isLogicalOp(s) {
		expr, err := parseSingleFilter(arr)
		if err != nil {
			return nil, err
		}
		return []*FilterExpr{expr}, nil
	}

	// Compound filter: [expr, ["OR"], expr, ...]
	var result []*FilterExpr
	var pending []*FilterExpr
	isOr := false

	for _, item := range arr {
		sub, ok := item.([]any)
		if !ok {
			continue
		}
		if len(sub) == 1 {
			if op, ok := sub[0].(string); ok {
				switch strings.ToUpper(op) {
				case "OR":
					isOr = true
				case "AND":
					isOr = false
				}
				continue
			}
		}
		exprs, err := parseFilterNode(item)
		if err != nil {
			return nil, err
		}
		if isOr && len(pending) > 0 {
			group := &FilterExpr{Or: append(pending, exprs...)}
			result = append(result, group)
			pending = nil
			isOr = false
		} else {
			pending = append(pending, exprs...)
		}
	}
	result = append(result, pending...)
	return result, nil
}

func parseSingleFilter(arr []any) (*FilterExpr, error) {
	if len(arr) < 2 {
		return nil, fmt.Errorf("filter needs at least 2 elements")
	}
	field, ok := arr[0].(string)
	if !ok {
		return nil, fmt.Errorf("filter field must be a string")
	}

	// ["field","value"] → eq
	if len(arr) == 2 {
		return &FilterExpr{Field: field, Operator: FilterEq, Value: arr[1]}, nil
	}

	// ["field","op","value"]
	opStr, ok := arr[1].(string)
	if !ok {
		return nil, fmt.Errorf("filter operator must be a string")
	}
	op := FilterOperator(strings.ToLower(opStr))
	var value any
	if len(arr) >= 3 {
		value = arr[2]
	}
	return &FilterExpr{Field: field, Operator: op, Value: value}, nil
}

func isLogicalOp(s string) bool {
	up := strings.ToUpper(s)
	return up == "OR" || up == "AND"
}

// PaginationParams represents all pagination, sort, search and filter params.
type PaginationParams struct {
	Page    int
	PerPage int
	Search  string
	Sorts   []SortField   // multi-column sort, parsed from sort param
	Filters []*FilterExpr // structured filters, parsed from filters param
	// Legacy single-sort fields (backward compat)
	Sort  string
	Order string
}

// ParsePaginationParams extracts pagination from an HTTP request.
// Supports GET query params and POST JSON body.
func ParsePaginationParams(r *http.Request) PaginationParams {
	return ParsePaginationParamsWithConfig(r, defaultPaginationConfig())
}

// ParsePaginationParamsWithConfig parses with a custom config.
func ParsePaginationParamsWithConfig(r *http.Request, cfg PaginationConfig) PaginationParams {
	params := PaginationParams{
		Page:    cfg.PageStart,
		PerPage: cfg.DefaultSize,
		Order:   "asc",
	}

	q := r.URL.Query()

	// Support POST JSON body
	if r.Method == http.MethodPost {
		body, _ := io.ReadAll(io.LimitReader(r.Body, 64*1024))
		if len(body) > 0 {
			var req struct {
				Page    *int   `json:"page"`
				Size    *int   `json:"size"`
				PerPage *int   `json:"per_page"`
				Sort    string `json:"sort"`
				Order   string `json:"order"`
				Search  string `json:"search"`
				Filters any    `json:"filters"`
			}
			if json.Unmarshal(body, &req) == nil {
				if req.Page != nil {
					params.Page = *req.Page
				}
				if req.Size != nil {
					params.PerPage = *req.Size
				}
				if req.PerPage != nil {
					params.PerPage = *req.PerPage
				}
				params.Sort = req.Sort
				params.Order = req.Order
				params.Search = req.Search
				if req.Filters != nil {
					b, _ := json.Marshal(req.Filters)
					params.Filters, _ = ParseFiltersJSON(string(b))
				}
			}
		}
	}

	// Page
	for _, pname := range cfg.PageParams {
		if v := q.Get(pname); v != "" {
			if p, err := strconv.Atoi(v); err == nil && p >= cfg.PageStart {
				params.Page = p
				break
			}
		}
	}

	// Size / PerPage
	for _, pname := range cfg.SizeParams {
		if v := q.Get(pname); v != "" {
			if pp, err := strconv.Atoi(v); err == nil && pp > 0 {
				if pp > cfg.MaxSize {
					pp = cfg.MaxSize
				}
				params.PerPage = pp
				break
			}
		}
	}

	// Search
	for _, pname := range cfg.SearchParams {
		if v := q.Get(pname); v != "" {
			params.Search = v
			break
		}
	}

	// Sort — multi-column, e.g. "-name,id"
	for _, pname := range cfg.SortParams {
		if v := q.Get(pname); v != "" {
			params.Sort = v
			params.Sorts = ParseSortFields(v)
			break
		}
	}

	// Order (legacy single-sort)
	for _, pname := range cfg.OrderParams {
		if v := q.Get(pname); v != "" {
			if strings.ToLower(v) == "desc" {
				params.Order = "desc"
			}
			break
		}
	}

	// Filters JSON
	for _, pname := range cfg.FilterParams {
		if v := q.Get(pname); v != "" {
			params.Filters, _ = ParseFiltersJSON(v)
			break
		}
	}

	return params
}

// Offset returns the SQL OFFSET value for the current page.
func (p PaginationParams) Offset() int {
	page := p.Page - 1 // convert 1-based to 0-based
	if page < 0 {
		page = 0
	}
	return page * p.PerPage
}

// -----------
// PageResult
// ----------

// PageResult is the paginated result wrapper.
// Mirrors morkid/paginate's Page struct with all fields.
type PageResult struct {
	Items        []any  `json:"items"`
	Total        int64  `json:"total"`
	Page         int    `json:"page"`
	Size         int    `json:"size"`
	TotalPages   int64  `json:"total_pages"`
	MaxPage      int64  `json:"max_page"`
	First        bool   `json:"first"`
	Last         bool   `json:"last"`
	Visible      int    `json:"visible"`
	Error        bool   `json:"error,omitempty"`
	ErrorMessage string `json:"error_message,omitempty"`
}

// NewPage creates a PageResult from items, total count, page and size.
func NewPage(items []any, total int64, page, size int) *PageResult {
	if size <= 0 {
		size = 15
	}
	totalPages := (total + int64(size) - 1) / int64(size)
	if totalPages == 0 {
		totalPages = 1
	}
	maxPage := totalPages - 1
	if maxPage < 0 {
		maxPage = 0
	}
	return &PageResult{
		Items:      items,
		Total:      total,
		Page:       page,
		Size:       size,
		TotalPages: totalPages,
		MaxPage:    maxPage,
		First:      page <= 1,
		Last:       int64(page) >= totalPages,
		Visible:    len(items),
	}
}

// NewPageError creates a PageResult with an error.
func NewPageError(err error) *PageResult {
	return &PageResult{Error: true, ErrorMessage: err.Error()}
}

// HasNext returns true if there is a next page.
func (p *PageResult) HasNext() bool { return !p.Last }

// HasPrev returns true if there is a previous page.
func (p *PageResult) HasPrev() bool { return !p.First }

// PaginatedResult is an alias for PageResult for backward compatibility.
type PaginatedResult = PageResult

// NewPaginatedResult creates a PageResult (backward compat alias).
func NewPaginatedResult(items []any, total, page, perPage int) *PageResult {
	return NewPage(items, int64(total), page, perPage)
}

// ---------------------------------
// PaginationCache — in-memory cache
// ----------------------------------

// PaginationCacheAdapter is the interface for pagination caching.
type PaginationCacheAdapter interface {
	Get(key string) (*PageResult, bool)
	Set(key string, page *PageResult, ttl time.Duration)
	Delete(key string)
	Flush()
}

// MemoryPaginationCache is a simple in-memory cache for paginated results.
type MemoryPaginationCache struct {
	mu      sync.RWMutex
	entries map[string]*pageCacheEntry
}

type pageCacheEntry struct {
	page *PageResult
	exp  time.Time
}

func NewMemoryPaginationCache() *MemoryPaginationCache {
	return &MemoryPaginationCache{entries: make(map[string]*pageCacheEntry)}
}

func (c *MemoryPaginationCache) Get(key string) (*PageResult, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	e, ok := c.entries[key]
	if !ok || time.Now().After(e.exp) {
		return nil, false
	}
	return e.page, true
}

func (c *MemoryPaginationCache) Set(key string, page *PageResult, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[key] = &pageCacheEntry{page: page, exp: time.Now().Add(ttl)}
}

func (c *MemoryPaginationCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.entries, key)
}

func (c *MemoryPaginationCache) Flush() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries = make(map[string]*pageCacheEntry)
}

// ---------------------------
// PaginatedListFunc — function type
// ---------------------------

// PaginatedListFunc is a function that returns a PageResult.
type PaginatedListFunc func(ctx context.Context, params PaginationParams) (*PageResult, error)

// ---------------------------
// PaginatedResource — interface
// ---------------------------

// PaginatedResource extends Resource with server-side pagination support.
type PaginatedResource interface {
	Resource
	ListPaginated(ctx context.Context, params PaginationParams) (*PageResult, error)
}

// ---------------------------
// Paginator — fluent builder
// ---------------------------

// Paginator is a fluent builder for paginated queries.
type Paginator struct {
	cfg      PaginationConfig
	cache    PaginationCacheAdapter
	cacheTTL time.Duration
	listFn   PaginatedListFunc
}

// NewPaginator creates a Paginator with default config.
func NewPaginator() *Paginator {
	return &Paginator{
		cfg:      defaultPaginationConfig(),
		cacheTTL: 30 * time.Second,
	}
}

// WithConfig sets a custom pagination config.
func (p *Paginator) WithConfig(cfg PaginationConfig) *Paginator {
	p.cfg = cfg
	return p
}

// WithCache enables caching of paginated results.
func (p *Paginator) WithCache(adapter PaginationCacheAdapter, ttl time.Duration) *Paginator {
	p.cache = adapter
	p.cacheTTL = ttl
	return p
}

// With sets the list function.
func (p *Paginator) With(fn PaginatedListFunc) *PaginatorRequest {
	return &PaginatorRequest{paginator: p, fn: fn}
}

// PaginatorRequest holds the list function and waits for a request.
type PaginatorRequest struct {
	paginator *Paginator
	fn        PaginatedListFunc
}

// Request attaches an HTTP request.
func (pr *PaginatorRequest) Request(r *http.Request) *PaginatorResponse {
	params := ParsePaginationParamsWithConfig(r, pr.paginator.cfg)
	return &PaginatorResponse{paginator: pr.paginator, fn: pr.fn, params: params, r: r}
}

// PaginatorResponse executes the query and returns results.
type PaginatorResponse struct {
	paginator *Paginator
	fn        PaginatedListFunc
	params    PaginationParams
	r         *http.Request
}

// Response executes the paginated query.
func (pr *PaginatorResponse) Response(ctx context.Context) *PageResult {
	// Build cache key
	cacheKey := fmt.Sprintf("page:%d:size:%d:search:%s:sort:%s:order:%s",
		pr.params.Page, pr.params.PerPage, pr.params.Search, pr.params.Sort, pr.params.Order)

	if pr.paginator.cache != nil {
		if cached, ok := pr.paginator.cache.Get(cacheKey); ok {
			return cached
		}
	}

	pageResult, err := pr.fn(ctx, pr.params)
	if err != nil {
		if pr.paginator.cfg.ErrorEnabled {
			return NewPageError(err)
		}
		return NewPage(nil, 0, pr.params.Page, pr.params.PerPage)
	}

	if pr.paginator.cache != nil && pageResult != nil {
		pr.paginator.cache.Set(cacheKey, pageResult, pr.paginator.cacheTTL)
	}

	return pageResult
}

// ---------------------------
// PaginatedCRUDHandler — handler CRUD avec pagination
// ---------------------------
// ...

// PaginatedCRUDHandler handles CRUD operations with server-side pagination.
type PaginatedCRUDHandler struct {
	Resource  Resource
	Paginator *Paginator
}

// NewPaginatedCRUDHandler creates a paginated CRUD handler.
func NewPaginatedCRUDHandler(r Resource) *PaginatedCRUDHandler {
	return &PaginatedCRUDHandler{Resource: r, Paginator: NewPaginator()}
}

// WithPaginator sets a custom Paginator.
func (h *PaginatedCRUDHandler) WithPaginator(p *Paginator) *PaginatedCRUDHandler {
	h.Paginator = p
	return h
}

// List displays the paginated list of items.
func (h *PaginatedCRUDHandler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	params := ParsePaginationParamsWithConfig(r, h.Paginator.cfg)

	var component templ.Component
	title := h.Resource.PluralLabel()

	if paginated, ok := h.Resource.(PaginatedResource); ok {
		pageResult, err := paginated.ListPaginated(ctx, params)
		if err != nil {
			http.Error(w, "List error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		ctx = context.WithValue(ctx, paginationContextKey, pageResult)
		component = h.Resource.Table(ctx)
	} else {
		component = h.Resource.Table(ctx)
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
	renderPage(w, r, "Create "+h.Resource.Label(), h.Resource.Form(ctx, nil))
}

// Edit displays the edit form.
func (h *PaginatedCRUDHandler) Edit(w http.ResponseWriter, r *http.Request, id string) {
	ctx := r.Context()
	item, err := h.Resource.Get(ctx, id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	renderPage(w, r, "Edit "+h.Resource.Label(), h.Resource.Form(ctx, item))
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
	http.Redirect(w, r, "/"+h.Resource.Slug()+preservePaginationQuery(r), http.StatusSeeOther)
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
	http.Redirect(w, r, "/"+h.Resource.Slug()+preservePaginationQuery(r), http.StatusSeeOther)
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
	http.Redirect(w, r, "/"+h.Resource.Slug()+preservePaginationQuery(r), http.StatusSeeOther)
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
	http.Redirect(w, r, "/"+h.Resource.Slug()+preservePaginationQuery(r), http.StatusSeeOther)
}

// ServeHTTP implements http.Handler with automatic routing.
func (h *PaginatedCRUDHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/"+h.Resource.Slug())
	path = strings.TrimPrefix(path, "/")
	parts := strings.Split(path, "/")

	switch r.Method {
	case http.MethodGet:
		switch {
		case path == "" || path == "/":
			h.List(w, r)
		case path == "create":
			h.Create(w, r)
		case len(parts) == 2 && parts[1] == "edit":
			h.Edit(w, r, parts[0])
		default:
			http.Redirect(w, r, fmt.Sprintf("/%s/%s/edit", h.Resource.Slug(), parts[0]), http.StatusSeeOther)
		}
	case http.MethodPost:
		r.ParseForm()
		if r.FormValue("_method") == "DELETE" && len(parts) >= 1 {
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

// ListPaginated implements PaginatedResource.
func (s *SimplePaginatedResource) ListPaginated(ctx context.Context, params PaginationParams) (*PageResult, error) {
	if s.paginatedListFunc != nil {
		return s.paginatedListFunc(ctx, params)
	}
	return nil, fmt.Errorf("paginated list function not configured for resource %q", s.Slug())
}

// ---------------------------------------------------------------------------
// Context helpers
// ---------------------------------------------------------------------------

type paginationCtxKey string

const paginationContextKey paginationCtxKey = "pagination_page"

// PageFromContext retrieves the current PageResult from context (set by PaginatedCRUDHandler).
func PageFromContext(ctx context.Context) (*PageResult, bool) {
	p, ok := ctx.Value(paginationContextKey).(*PageResult)
	return p, ok
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// renderPage renders a templ component inside the base layout.
func renderPage(w http.ResponseWriter, r *http.Request, title string, content templ.Component) {
	fullPage := layouts.Page(title, content)
	fullPage.Render(r.Context(), w)
}

// preservePaginationQuery builds a query string preserving page/size/search/sort params.
func preservePaginationQuery(r *http.Request) string {
	q := r.URL.Query()
	var parts []string
	for _, key := range []string{"page", "per_page", "size", "search", "sort", "order"} {
		if v := q.Get(key); v != "" {
			parts = append(parts, key+"="+v)
		}
	}
	if len(parts) == 0 {
		return ""
	}
	return "?" + strings.Join(parts, "&")
}
