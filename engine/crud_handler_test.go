package engine

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/a-h/templ"
)

// ---------------------------------------------------------------------------
// mockResource is a minimal Resource implementation for testing.
// It uses BaseResource for most defaults and overrides what is needed.
// ---------------------------------------------------------------------------

type mockResource struct {
	*BaseResource
	deleteCalledWith    string
	bulkDeleteCalledIDs []string
}

func newMockResource(slug string) *mockResource {
	return &mockResource{
		BaseResource: NewBaseResource(slug, slug, slug+"s"),
	}
}

// Table and Form must return a non-nil templ.Component.
func (m *mockResource) Table(ctx context.Context) templ.Component {
	return emptyComponent()
}

func (m *mockResource) Form(ctx context.Context, item any) templ.Component {
	return emptyComponent()
}

// Override Delete to track calls.
func (m *mockResource) Delete(ctx context.Context, id string) error {
	m.deleteCalledWith = id
	return nil
}

// Override BulkDelete to track calls.
func (m *mockResource) BulkDelete(ctx context.Context, ids []string) error {
	m.bulkDeleteCalledIDs = ids
	return nil
}

// ---------------------------------------------------------------------------
// mockValidatorResource additionally implements ResourceValidator.
// ---------------------------------------------------------------------------

type mockValidatorResource struct {
	*mockResource
	// fieldErrors maps field name -> error message (empty = valid)
	fieldErrors map[string]string
}

func newMockValidatorResource(slug string) *mockValidatorResource {
	return &mockValidatorResource{
		mockResource: newMockResource(slug),
		fieldErrors:  make(map[string]string),
	}
}

func (v *mockValidatorResource) ValidateField(ctx context.Context, field, value string) error {
	if msg, bad := v.fieldErrors[field]; bad {
		return errors.New(msg)
	}
	return nil
}

// ---------------------------------------------------------------------------
// Helper: build a CRUDHandler whose URL base is "/{slug}"
// ---------------------------------------------------------------------------

func newHandler(r Resource) *CRUDHandler {
	return NewCRUDHandler(r)
}

// serveWith sends a request through the handler and returns the recorder.
func serveWith(h http.Handler, method, path string, body url.Values) *httptest.ResponseRecorder {
	var req *http.Request
	if body != nil {
		req = httptest.NewRequest(method, path, strings.NewReader(body.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, req)
	return rw
}

// ---------------------------------------------------------------------------
// routeGET: List route (GET /)
// ---------------------------------------------------------------------------

func TestCRUDHandler_GET_list(t *testing.T) {
	res := newMockResource("items")
	h := newHandler(res)

	rw := serveWith(h, http.MethodGet, "/items", nil)

	// List renders the table — expects a 200 response (render writes 200 implicitly)
	if rw.Code != http.StatusOK {
		t.Errorf("expected status 200 for GET /items, got %d", rw.Code)
	}
}

func TestCRUDHandler_GET_list_empty_path(t *testing.T) {
	res := newMockResource("users")
	h := newHandler(res)

	// When URL is exactly "/{slug}" with no trailing content
	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, req)

	if rw.Code != http.StatusOK {
		t.Errorf("expected status 200 for GET /users (list), got %d", rw.Code)
	}
}

// ---------------------------------------------------------------------------
// routeGET: Create route (GET /create)
// ---------------------------------------------------------------------------

func TestCRUDHandler_GET_create(t *testing.T) {
	res := newMockResource("items")
	h := newHandler(res)

	rw := serveWith(h, http.MethodGet, "/items/create", nil)

	// Create renders the form — expects 200
	if rw.Code != http.StatusOK {
		t.Errorf("expected status 200 for GET /items/create, got %d", rw.Code)
	}
}

func TestCRUDHandler_GET_create_forbidden(t *testing.T) {
	res := newMockResource("items")
	res.BaseResource.SetSlug("items")
	// Override CanCreate to deny
	h := &CRUDHandler{Resource: &noCreateResource{BaseResource: res.BaseResource}}

	rw := serveWith(h, http.MethodGet, "/items/create", nil)

	if rw.Code != http.StatusForbidden {
		t.Errorf("expected status 403 for forbidden Create, got %d", rw.Code)
	}
}

// noCreateResource denies CanCreate.
type noCreateResource struct {
	*BaseResource
}

func (n *noCreateResource) CanCreate(ctx context.Context) bool { return false }
func (n *noCreateResource) Table(ctx context.Context) templ.Component {
	return emptyComponent()
}
func (n *noCreateResource) Form(ctx context.Context, item any) templ.Component {
	return emptyComponent()
}

// ---------------------------------------------------------------------------
// routeGET: ValidateField — resource does NOT implement ResourceValidator
// ---------------------------------------------------------------------------

func TestCRUDHandler_GET_validateField_no_validator_returns_204(t *testing.T) {
	res := newMockResource("items")
	h := newHandler(res)

	rw := serveWith(h, http.MethodGet, "/items/validate-field?field=name&value=test", nil)

	// Resource does not implement ResourceValidator → 204 No Content
	if rw.Code != http.StatusNoContent {
		t.Errorf("expected status 204 when resource has no validator, got %d", rw.Code)
	}
}

// ---------------------------------------------------------------------------
// routeGET: ValidateField — resource IMPLEMENTS ResourceValidator
// ---------------------------------------------------------------------------

func TestCRUDHandler_GET_validateField_with_validator_valid_field(t *testing.T) {
	res := newMockValidatorResource("users")
	// No error for "name" field → valid
	h := newHandler(res)

	rw := serveWith(h, http.MethodGet, "/users/validate-field?field=name&value=Alice", nil)

	// Validator returned nil → SSE response with a clear fragment
	// The SSE handler sets text/event-stream
	ct := rw.Header().Get("Content-Type")
	if ct != "text/event-stream" {
		t.Errorf("expected Content-Type='text/event-stream', got '%s'", ct)
	}
	body := rw.Body.String()
	// Should contain the field name in the HTML fragment
	if !strings.Contains(body, "field-error-name") {
		t.Errorf("expected response body to contain 'field-error-name', got: %s", body)
	}
}

func TestCRUDHandler_GET_validateField_with_validator_invalid_field(t *testing.T) {
	res := newMockValidatorResource("users")
	res.fieldErrors["email"] = "invalid email address"
	h := newHandler(res)

	rw := serveWith(h, http.MethodGet, "/users/validate-field?field=email&value=bad", nil)

	ct := rw.Header().Get("Content-Type")
	if ct != "text/event-stream" {
		t.Errorf("expected Content-Type='text/event-stream', got '%s'", ct)
	}
	body := rw.Body.String()
	if !strings.Contains(body, "field-error-email") {
		t.Errorf("expected 'field-error-email' in body, got: %s", body)
	}
	if !strings.Contains(body, "invalid email address") {
		t.Errorf("expected error message in body, got: %s", body)
	}
}

func TestCRUDHandler_GET_validateField_missing_field_param(t *testing.T) {
	res := newMockValidatorResource("users")
	h := newHandler(res)

	// No "field" query param → 400 Bad Request
	rw := serveWith(h, http.MethodGet, "/users/validate-field", nil)

	if rw.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 when 'field' param is missing, got %d", rw.Code)
	}
}

// ---------------------------------------------------------------------------
// routePOST: _method=DELETE → calls Delete
// ---------------------------------------------------------------------------

func TestCRUDHandler_POST_method_override_DELETE(t *testing.T) {
	res := newMockResource("items")
	h := newHandler(res)

	form := url.Values{}
	form.Set("_method", "DELETE")

	rw := serveWith(h, http.MethodPost, "/items/42", form)

	// Delete triggers a redirect to /{slug}
	if rw.Code != http.StatusSeeOther {
		t.Errorf("expected 303 redirect after DELETE, got %d", rw.Code)
	}
	if res.deleteCalledWith != "42" {
		t.Errorf("expected Delete called with id='42', got '%s'", res.deleteCalledWith)
	}
	location := rw.Header().Get("Location")
	if !strings.HasSuffix(location, "/items") {
		t.Errorf("expected redirect to '/items', got '%s'", location)
	}
}

func TestCRUDHandler_POST_method_override_DELETE_forbidden(t *testing.T) {
	res := &noDeleteResource{BaseResource: newMockResource("items").BaseResource}
	h := &CRUDHandler{Resource: res}

	form := url.Values{}
	form.Set("_method", "DELETE")

	rw := serveWith(h, http.MethodPost, "/items/99", form)

	if rw.Code != http.StatusForbidden {
		t.Errorf("expected 403 when CanDelete=false, got %d", rw.Code)
	}
}

// noDeleteResource denies CanDelete.
type noDeleteResource struct {
	*BaseResource
}

func (n *noDeleteResource) CanDelete(ctx context.Context) bool { return false }
func (n *noDeleteResource) Table(ctx context.Context) templ.Component {
	return emptyComponent()
}
func (n *noDeleteResource) Form(ctx context.Context, item any) templ.Component {
	return emptyComponent()
}

// ---------------------------------------------------------------------------
// routePOST: bulk-delete route
// ---------------------------------------------------------------------------

func TestCRUDHandler_POST_bulk_delete(t *testing.T) {
	res := newMockResource("products")
	h := newHandler(res)

	form := url.Values{}
	form.Add("ids[]", "1")
	form.Add("ids[]", "2")
	form.Add("ids[]", "5")

	rw := serveWith(h, http.MethodPost, "/products/bulk-delete", form)

	// BulkDelete should redirect to /{slug}
	if rw.Code != http.StatusSeeOther {
		t.Errorf("expected 303 redirect after bulk-delete, got %d", rw.Code)
	}
	if len(res.bulkDeleteCalledIDs) != 3 {
		t.Errorf("expected BulkDelete called with 3 ids, got %d", len(res.bulkDeleteCalledIDs))
	}
	expectedIDs := map[string]bool{"1": true, "2": true, "5": true}
	for _, id := range res.bulkDeleteCalledIDs {
		if !expectedIDs[id] {
			t.Errorf("unexpected id '%s' in BulkDelete call", id)
		}
	}
}

func TestCRUDHandler_POST_bulk_delete_no_ids_returns_400(t *testing.T) {
	res := newMockResource("products")
	h := newHandler(res)

	// Post with no ids[]
	form := url.Values{}
	rw := serveWith(h, http.MethodPost, "/products/bulk-delete", form)

	if rw.Code != http.StatusBadRequest {
		t.Errorf("expected 400 when no ids provided, got %d", rw.Code)
	}
}

func TestCRUDHandler_POST_bulk_delete_forbidden(t *testing.T) {
	res := &noDeleteResource{BaseResource: newMockResource("products").BaseResource}
	h := &CRUDHandler{Resource: res}

	form := url.Values{}
	form.Add("ids[]", "1")

	rw := serveWith(h, http.MethodPost, "/products/bulk-delete", form)

	if rw.Code != http.StatusForbidden {
		t.Errorf("expected 403 when CanDelete=false for bulk delete, got %d", rw.Code)
	}
}

// ---------------------------------------------------------------------------
// NewCRUDHandler smoke test
// ---------------------------------------------------------------------------

func TestNewCRUDHandler_returns_handler(t *testing.T) {
	res := newMockResource("things")
	h := NewCRUDHandler(res)

	if h == nil {
		t.Fatal("expected NewCRUDHandler to return a non-nil handler")
	}
	if h.Resource != res {
		t.Error("expected handler.Resource to be the supplied resource")
	}
}

// ---------------------------------------------------------------------------
// GetListQuery helper
// ---------------------------------------------------------------------------

func TestGetListQuery_returns_nil_when_not_set(t *testing.T) {
	ctx := context.Background()
	if GetListQuery(ctx) != nil {
		t.Error("expected GetListQuery to return nil when not set in context")
	}
}

func TestGetListQuery_returns_value_when_set(t *testing.T) {
	// Trigger a List request which injects ListQuery into the context
	req := httptest.NewRequest(http.MethodGet, "/items?page=2&per_page=50&search=foo&sort=name&dir=desc", nil)
	rw := httptest.NewRecorder()

	// Intercept by wrapping the List handler to inspect the context
	var capturedQuery *ListQuery

	// We use a spy resource to capture the query from context inside Table()
	spy := &spyResource{
		mockResource: newMockResource("items"),
		captureFunc: func(ctx context.Context) {
			capturedQuery = GetListQuery(ctx)
		},
	}
	spyHandler := NewCRUDHandler(spy)
	spyHandler.ServeHTTP(rw, req)

	if capturedQuery == nil {
		t.Fatal("expected ListQuery to be injected into context during List()")
	}
	if capturedQuery.Page != 2 {
		t.Errorf("expected Page=2, got %d", capturedQuery.Page)
	}
	if capturedQuery.PerPage != 50 {
		t.Errorf("expected PerPage=50, got %d", capturedQuery.PerPage)
	}
	if capturedQuery.Search != "foo" {
		t.Errorf("expected Search='foo', got '%s'", capturedQuery.Search)
	}
	if capturedQuery.SortKey != "name" {
		t.Errorf("expected SortKey='name', got '%s'", capturedQuery.SortKey)
	}
	if capturedQuery.SortDir != "desc" {
		t.Errorf("expected SortDir='desc', got '%s'", capturedQuery.SortDir)
	}
}

// spyResource captures the context passed to Table() for inspection.
type spyResource struct {
	*mockResource
	captureFunc func(ctx context.Context)
}

func (s *spyResource) Table(ctx context.Context) templ.Component {
	if s.captureFunc != nil {
		s.captureFunc(ctx)
	}
	return emptyComponent()
}

// ---------------------------------------------------------------------------
// PATCH /{slug}/{id} — inline edit column updates
// ---------------------------------------------------------------------------

// patchableResource implements ResourcePatchable for testing.
type patchableResource struct {
	*mockResource
	patchCalledID     string
	patchCalledFields map[string]string
	patchErr          error
}

func newPatchableResource(slug string) *patchableResource {
	return &patchableResource{mockResource: newMockResource(slug)}
}

func (p *patchableResource) PatchField(ctx context.Context, id string, fields map[string]string) error {
	p.patchCalledID = id
	p.patchCalledFields = fields
	return p.patchErr
}

func TestCRUDHandler_PATCH_with_ResourcePatchable_success(t *testing.T) {
	res := newPatchableResource("products")
	h := newHandler(res)

	form := url.Values{}
	form.Set("status", "active")

	rw := serveWith(h, http.MethodPatch, "/products/7", form)

	// PATCH returns Datastar SSE (text/event-stream)
	ct := rw.Header().Get("Content-Type")
	if ct != "text/event-stream" {
		t.Errorf("expected Content-Type='text/event-stream', got '%s'", ct)
	}
	if res.patchCalledID != "7" {
		t.Errorf("expected PatchField called with id='7', got '%s'", res.patchCalledID)
	}
	if res.patchCalledFields["status"] != "active" {
		t.Errorf("expected field status='active', got '%s'", res.patchCalledFields["status"])
	}
	body := rw.Body.String()
	if !strings.Contains(body, "Saved") {
		t.Errorf("expected 'Saved' toast in SSE response, got: %s", body)
	}
}

func TestCRUDHandler_PATCH_with_ResourcePatchable_error(t *testing.T) {
	res := newPatchableResource("products")
	res.patchErr = errors.New("validation failed")
	h := newHandler(res)

	form := url.Values{}
	form.Set("status", "invalid")

	rw := serveWith(h, http.MethodPatch, "/products/7", form)

	ct := rw.Header().Get("Content-Type")
	if ct != "text/event-stream" {
		t.Errorf("expected Content-Type='text/event-stream', got '%s'", ct)
	}
	body := rw.Body.String()
	if !strings.Contains(body, "validation failed") {
		t.Errorf("expected error message in SSE response, got: %s", body)
	}
}

func TestCRUDHandler_PATCH_forbidden(t *testing.T) {
	res := &noUpdateResource{BaseResource: newMockResource("products").BaseResource}
	h := &CRUDHandler{Resource: res}

	form := url.Values{}
	form.Set("status", "active")

	rw := serveWith(h, http.MethodPatch, "/products/7", form)

	if rw.Code != http.StatusForbidden {
		t.Errorf("expected 403 when CanUpdate=false, got %d", rw.Code)
	}
}

func TestCRUDHandler_PATCH_missing_id_returns_404(t *testing.T) {
	res := newMockResource("products")
	h := newHandler(res)

	// PATCH /{slug}/ with empty id segment
	req := httptest.NewRequest(http.MethodPatch, "/products/", nil)
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, req)

	if rw.Code != http.StatusNotFound {
		t.Errorf("expected 404 for PATCH with empty id, got %d", rw.Code)
	}
}

// noUpdateResource denies CanUpdate.
type noUpdateResource struct {
	*BaseResource
}

func (n *noUpdateResource) CanUpdate(ctx context.Context) bool { return false }
func (n *noUpdateResource) Table(ctx context.Context) templ.Component {
	return emptyComponent()
}
func (n *noUpdateResource) Form(ctx context.Context, item any) templ.Component {
	return emptyComponent()
}

// ---------------------------------------------------------------------------
// Store errors — form re-render on Create/Update failure
// ---------------------------------------------------------------------------

// storeErrorResource returns an error on Create and Update.
type storeErrorResource struct {
	*mockResource
	storeErr error
}

func newStoreErrorResource(slug string, err error) *storeErrorResource {
	return &storeErrorResource{mockResource: newMockResource(slug), storeErr: err}
}

func (s *storeErrorResource) Create(ctx context.Context, r *http.Request) error {
	return s.storeErr
}

func (s *storeErrorResource) Update(ctx context.Context, id string, r *http.Request) error {
	return s.storeErr
}

func TestCRUDHandler_POST_create_store_error_rerenders_form(t *testing.T) {
	res := newStoreErrorResource("items", errors.New("duplicate entry"))
	h := newHandler(res)

	form := url.Values{}
	form.Set("name", "Test")

	rw := serveWith(h, http.MethodPost, "/items/create", form)

	// On validation/store error the form is re-rendered with 422 Unprocessable Entity.
	if rw.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected 422 re-render on Create error, got %d", rw.Code)
	}
}

func TestCRUDHandler_POST_update_store_error_rerenders_form(t *testing.T) {
	res := newStoreErrorResource("items", errors.New("constraint violation"))
	h := newHandler(res)

	form := url.Values{}
	form.Set("name", "Updated")

	rw := serveWith(h, http.MethodPost, "/items/42", form)

	// On validation/store error the edit form is re-rendered with 422 Unprocessable Entity.
	if rw.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected 422 re-render on Update error, got %d", rw.Code)
	}
}
