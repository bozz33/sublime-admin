package engine

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/a-h/templ"
)

// ---------------------------------------------------------------------------
// mockRelationManager — full implementation of RelationManager for tests.
// ---------------------------------------------------------------------------

type mockRelationManager struct {
	*BaseRelationManager
	listErr        error
	listItems      []any
	createErr      error
	attachErr      error
	detachErr      error
	deleteErr      error
	formComponent  templ.Component
	canCreate      bool
	canAttach      bool
	canDelete      bool
	createCalled   bool
	attachCalled   bool
	detachCalled   bool
	deleteCalled   bool
}

func newMockRM(name string) *mockRelationManager {
	return &mockRelationManager{
		BaseRelationManager: NewBaseRelationManager(name, name+" Label", name, RelationHasMany),
		canCreate:           true,
		canAttach:           true,
		canDelete:           true,
	}
}

func (m *mockRelationManager) ListRelated(_ context.Context, _ string) ([]any, error) {
	return m.listItems, m.listErr
}

func (m *mockRelationManager) CreateRelated(_ context.Context, _ string, _ *http.Request) error {
	m.createCalled = true
	return m.createErr
}

func (m *mockRelationManager) AttachRelated(_ context.Context, _, _ string) error {
	m.attachCalled = true
	return m.attachErr
}

func (m *mockRelationManager) DetachRelated(_ context.Context, _, _ string) error {
	m.detachCalled = true
	return m.detachErr
}

func (m *mockRelationManager) DeleteRelated(_ context.Context, _, _ string) error {
	m.deleteCalled = true
	return m.deleteErr
}

func (m *mockRelationManager) Form(_ context.Context, _ string) templ.Component {
	if m.formComponent != nil {
		return m.formComponent
	}
	return m.BaseRelationManager.Form(context.Background(), "")
}

func (m *mockRelationManager) CanCreate(_ context.Context) bool { return m.canCreate }
func (m *mockRelationManager) CanAttach(_ context.Context) bool { return m.canAttach }
func (m *mockRelationManager) CanDelete(_ context.Context) bool { return m.canDelete }

// ---------------------------------------------------------------------------
// TestBaseRelationManager_Defaults
// ---------------------------------------------------------------------------

func TestBaseRelationManager_Defaults(t *testing.T) {
	b := NewBaseRelationManager("posts", "Posts", "posts", RelationHasMany)

	if b.Name() != "posts" {
		t.Errorf("Name() = %q, want %q", b.Name(), "posts")
	}
	if b.Label() != "Posts" {
		t.Errorf("Label() = %q, want %q", b.Label(), "Posts")
	}
	if b.Icon() != "link" {
		t.Errorf("Icon() = %q, want %q", b.Icon(), "link")
	}
	if b.RelationName() != "posts" {
		t.Errorf("RelationName() = %q, want %q", b.RelationName(), "posts")
	}
	if b.RelationType() != RelationHasMany {
		t.Errorf("RelationType() = %q, want %q", b.RelationType(), RelationHasMany)
	}

	items, err := b.ListRelated(context.Background(), "1")
	if err != nil {
		t.Errorf("ListRelated() error = %v", err)
	}
	if len(items) != 0 {
		t.Errorf("ListRelated() len = %d, want 0", len(items))
	}

	ctx := context.Background()
	if !b.CanAttach(ctx) {
		t.Error("CanAttach() should return true by default")
	}
	if !b.CanCreate(ctx) {
		t.Error("CanCreate() should return true by default")
	}
	if !b.CanDelete(ctx) {
		t.Error("CanDelete() should return true by default")
	}
}

func TestBaseRelationManager_Form_ReturnsEmptyComponent(t *testing.T) {
	b := NewBaseRelationManager("posts", "Posts", "posts", RelationHasMany)
	comp := b.Form(context.Background(), "42")
	if comp == nil {
		t.Fatal("Form() should not return nil")
	}
	// Render should produce no output and no error.
	var buf strings.Builder
	if err := comp.Render(context.Background(), &buf); err != nil {
		t.Errorf("Form() render error = %v", err)
	}
	if buf.String() != "" {
		t.Errorf("Form() default render output = %q, want empty", buf.String())
	}
}

// ---------------------------------------------------------------------------
// TestExtractRelatedID / TestSetRelatedID
// ---------------------------------------------------------------------------

type testItem struct {
	AuthorID int
	Name     string
}

func TestExtractRelatedID(t *testing.T) {
	item := testItem{AuthorID: 42, Name: "Alice"}

	got := ExtractRelatedID(item, "AuthorID")
	if got != 42 {
		t.Errorf("ExtractRelatedID() = %v, want 42", got)
	}

	got = ExtractRelatedID(item, "NoSuchField")
	if got != nil {
		t.Errorf("ExtractRelatedID() for missing field = %v, want nil", got)
	}

	got = ExtractRelatedID(&item, "AuthorID")
	if got != 42 {
		t.Errorf("ExtractRelatedID() via pointer = %v, want 42", got)
	}

	got = ExtractRelatedID("not a struct", "AuthorID")
	if got != nil {
		t.Errorf("ExtractRelatedID() for non-struct = %v, want nil", got)
	}
}

func TestSetRelatedID(t *testing.T) {
	item := &testItem{AuthorID: 0}

	err := SetRelatedID(item, "AuthorID", 99)
	if err != nil {
		t.Errorf("SetRelatedID() error = %v", err)
	}
	if item.AuthorID != 99 {
		t.Errorf("SetRelatedID() did not set value: got %d, want 99", item.AuthorID)
	}

	err = SetRelatedID(item, "NoSuchField", 1)
	if err == nil {
		t.Error("SetRelatedID() should return error for missing field")
	}
}

// ---------------------------------------------------------------------------
// TestGetRelationManagers context helpers
// ---------------------------------------------------------------------------

func TestGetRelationManagers_FromContext(t *testing.T) {
	rm := newMockRM("posts")
	managers := []RelationManager{rm}
	ctx := context.WithValue(context.Background(), contextKeyRelationManagers, managers)

	got := GetRelationManagers(ctx)
	if len(got) != 1 {
		t.Errorf("GetRelationManagers() len = %d, want 1", len(got))
	}
	if got[0].Name() != "posts" {
		t.Errorf("GetRelationManagers()[0].Name() = %q, want %q", got[0].Name(), "posts")
	}
}

func TestGetRelationManagers_MissingFromContext(t *testing.T) {
	got := GetRelationManagers(context.Background())
	if got != nil {
		t.Errorf("GetRelationManagers() = %v, want nil", got)
	}
}

// ---------------------------------------------------------------------------
// TestParseRelationPath
// ---------------------------------------------------------------------------

func TestParseRelationPath(t *testing.T) {
	h := &RelationManagerHandler{}

	cases := []struct {
		path        string
		wantParent  string
		wantRel     string
		wantSub     string
		wantRelID   string
		wantOK      bool
	}{
		{"42/relations/posts", "42", "posts", "", "", true},
		{"42/relations/posts/form", "42", "posts", "form", "", true},
		{"42/relations/posts/attach", "42", "posts", "attach", "", true},
		{"42/relations/posts/detach/7", "42", "posts", "detach", "7", true},
		{"42/relations/posts/99", "42", "posts", "", "99", true},
		{"bad/path", "", "", "", "", false},
		{"only", "", "", "", "", false},
	}

	for _, tc := range cases {
		r := httptest.NewRequest(http.MethodGet, "/"+tc.path, nil)
		parentID, rel, sub, relID, ok := h.parseRelationPath(r)
		if ok != tc.wantOK {
			t.Errorf("path=%q: ok=%v, want %v", tc.path, ok, tc.wantOK)
			continue
		}
		if !ok {
			continue
		}
		if parentID != tc.wantParent {
			t.Errorf("path=%q: parentID=%q, want %q", tc.path, parentID, tc.wantParent)
		}
		if rel != tc.wantRel {
			t.Errorf("path=%q: relationName=%q, want %q", tc.path, rel, tc.wantRel)
		}
		if sub != tc.wantSub {
			t.Errorf("path=%q: subAction=%q, want %q", tc.path, sub, tc.wantSub)
		}
		if relID != tc.wantRelID {
			t.Errorf("path=%q: relatedID=%q, want %q", tc.path, relID, tc.wantRelID)
		}
	}
}

// ---------------------------------------------------------------------------
// Helpers for ServeHTTP tests
// ---------------------------------------------------------------------------

func newHandlerWithRM(rm *mockRelationManager) *RelationManagerHandler {
	h := &RelationManagerHandler{
		managers: map[string]RelationManager{rm.Name(): rm},
	}
	return h
}

// ---------------------------------------------------------------------------
// TestRelationManagerHandler_GET_List
// ---------------------------------------------------------------------------

func TestRelationManagerHandler_GET_List(t *testing.T) {
	rm := newMockRM("posts")
	rm.listItems = []any{"item1", "item2"}

	h := newHandlerWithRM(rm)
	r := httptest.NewRequest(http.MethodGet, "/42/relations/posts", nil)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("GET list: status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "posts") {
		t.Errorf("GET list: body does not contain 'posts': %s", body)
	}
}

// ---------------------------------------------------------------------------
// TestRelationManagerHandler_GET_Form
// ---------------------------------------------------------------------------

func TestRelationManagerHandler_GET_Form(t *testing.T) {
	rm := newMockRM("posts")
	rm.formComponent = templ.ComponentFunc(func(_ context.Context, w io.Writer) error {
		_, err := io.WriteString(w, "<form>create post</form>")
		return err
	})

	h := newHandlerWithRM(rm)
	r := httptest.NewRequest(http.MethodGet, "/42/relations/posts/form", nil)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("GET form: status = %d, want %d", w.Code, http.StatusOK)
	}
	if !strings.Contains(w.Body.String(), "create post") {
		t.Errorf("GET form: body = %q, want form HTML", w.Body.String())
	}
}

func TestRelationManagerHandler_GET_Form_Forbidden(t *testing.T) {
	rm := newMockRM("posts")
	rm.canCreate = false

	h := newHandlerWithRM(rm)
	r := httptest.NewRequest(http.MethodGet, "/42/relations/posts/form", nil)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, r)

	if w.Code != http.StatusForbidden {
		t.Errorf("GET form forbidden: status = %d, want %d", w.Code, http.StatusForbidden)
	}
}

// ---------------------------------------------------------------------------
// TestRelationManagerHandler_POST_Create
// ---------------------------------------------------------------------------

func TestRelationManagerHandler_POST_Create(t *testing.T) {
	rm := newMockRM("posts")

	h := newHandlerWithRM(rm)
	r := httptest.NewRequest(http.MethodPost, "/42/relations/posts", nil)
	r.Header.Set("Referer", "/admin/users/42/edit")
	w := httptest.NewRecorder()

	h.ServeHTTP(w, r)

	if !rm.createCalled {
		t.Error("POST create: CreateRelated was not called")
	}
	if w.Code != http.StatusSeeOther {
		t.Errorf("POST create: status = %d, want %d", w.Code, http.StatusSeeOther)
	}
}

func TestRelationManagerHandler_POST_Create_Forbidden(t *testing.T) {
	rm := newMockRM("posts")
	rm.canCreate = false

	h := newHandlerWithRM(rm)
	r := httptest.NewRequest(http.MethodPost, "/42/relations/posts", nil)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, r)

	if w.Code != http.StatusForbidden {
		t.Errorf("POST create forbidden: status = %d, want %d", w.Code, http.StatusForbidden)
	}
}

// ---------------------------------------------------------------------------
// TestRelationManagerHandler_POST_Attach
// ---------------------------------------------------------------------------

func TestRelationManagerHandler_POST_Attach(t *testing.T) {
	rm := newMockRM("posts")

	h := newHandlerWithRM(rm)
	body := url.Values{"related_id": {"7"}}.Encode()
	r := httptest.NewRequest(http.MethodPost, "/42/relations/posts/attach",
		strings.NewReader(body))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.Header.Set("Referer", "/admin/users/42/edit")
	w := httptest.NewRecorder()

	h.ServeHTTP(w, r)

	if !rm.attachCalled {
		t.Error("POST attach: AttachRelated was not called")
	}
	if w.Code != http.StatusSeeOther {
		t.Errorf("POST attach: status = %d, want %d", w.Code, http.StatusSeeOther)
	}
}

func TestRelationManagerHandler_POST_Attach_MissingID(t *testing.T) {
	rm := newMockRM("posts")

	h := newHandlerWithRM(rm)
	r := httptest.NewRequest(http.MethodPost, "/42/relations/posts/attach", nil)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("POST attach missing ID: status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

// ---------------------------------------------------------------------------
// TestRelationManagerHandler_DELETE_Detach
// ---------------------------------------------------------------------------

func TestRelationManagerHandler_DELETE_Detach(t *testing.T) {
	rm := newMockRM("posts")

	h := newHandlerWithRM(rm)
	r := httptest.NewRequest(http.MethodDelete, "/42/relations/posts/detach/7", nil)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, r)

	if !rm.detachCalled {
		t.Error("DELETE detach: DetachRelated was not called")
	}
	if w.Code != http.StatusNoContent {
		t.Errorf("DELETE detach: status = %d, want %d", w.Code, http.StatusNoContent)
	}
}

// ---------------------------------------------------------------------------
// TestRelationManagerHandler_DELETE_Delete
// ---------------------------------------------------------------------------

func TestRelationManagerHandler_DELETE_Delete(t *testing.T) {
	rm := newMockRM("posts")

	h := newHandlerWithRM(rm)
	r := httptest.NewRequest(http.MethodDelete, "/42/relations/posts/99", nil)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, r)

	if !rm.deleteCalled {
		t.Error("DELETE delete: DeleteRelated was not called")
	}
	if w.Code != http.StatusNoContent {
		t.Errorf("DELETE delete: status = %d, want %d", w.Code, http.StatusNoContent)
	}
}

// ---------------------------------------------------------------------------
// TestRelationManagerHandler_UnknownRelation
// ---------------------------------------------------------------------------

func TestRelationManagerHandler_UnknownRelation(t *testing.T) {
	rm := newMockRM("posts")

	h := newHandlerWithRM(rm)
	r := httptest.NewRequest(http.MethodGet, "/42/relations/comments", nil)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, r)

	if w.Code != http.StatusNotFound {
		t.Errorf("Unknown relation: status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestRelationManagerHandler_BadPath(t *testing.T) {
	rm := newMockRM("posts")

	h := newHandlerWithRM(rm)
	r := httptest.NewRequest(http.MethodGet, "/bad", nil)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, r)

	if w.Code != http.StatusNotFound {
		t.Errorf("Bad path: status = %d, want %d", w.Code, http.StatusNotFound)
	}
}
