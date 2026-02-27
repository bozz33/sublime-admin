package actions

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// NewModal constructor tests
// ---------------------------------------------------------------------------

func TestNewModal_defaults(t *testing.T) {
	m := NewModal("archive")

	if m.Action.Name != "archive" {
		t.Errorf("expected Name='archive', got '%s'", m.Action.Name)
	}
	if m.Size != ModalSizeMD {
		t.Errorf("expected Size=ModalSizeMD, got '%s'", m.Size)
	}
	if !m.CloseOnEsc {
		t.Error("expected CloseOnEsc=true by default")
	}
	if !m.CloseOnBack {
		t.Error("expected CloseOnBack=true by default")
	}
	if m.FormMethod != "POST" {
		t.Errorf("expected FormMethod='POST', got '%s'", m.FormMethod)
	}
}

func TestNewModal_WithSize(t *testing.T) {
	m := NewModal("export").WithSize(ModalSizeLG)
	if m.Size != ModalSizeLG {
		t.Errorf("expected Size=ModalSizeLG, got '%s'", m.Size)
	}
}

func TestNewModal_SlideOver(t *testing.T) {
	m := NewModal("details").SlideOver()
	if !m.IsSlideOver {
		t.Error("expected IsSlideOver=true after SlideOver()")
	}
}

func TestNewModal_WithDatabaseTransaction(t *testing.T) {
	m := NewModal("publish").WithDatabaseTransaction()
	if !m.DatabaseTransaction {
		t.Error("expected DatabaseTransaction=true after WithDatabaseTransaction()")
	}
}

func TestNewModal_WithHTMXLoad(t *testing.T) {
	m := NewModal("preview").WithHTMXLoad("/items/123/preview")
	if m.HTMXLoadURL != "/items/123/preview" {
		t.Errorf("expected HTMXLoadURL='/items/123/preview', got '%s'", m.HTMXLoadURL)
	}
}

func TestNewModal_WithForm(t *testing.T) {
	m := NewModal("rename").WithForm("/items/123/rename",
		ModalField{Name: "name", Label: "Name", Type: "text", Required: true},
	)
	if m.FormAction != "/items/123/rename" {
		t.Errorf("expected FormAction='/items/123/rename', got '%s'", m.FormAction)
	}
	if len(m.FormFields) != 1 {
		t.Fatalf("expected 1 form field, got %d", len(m.FormFields))
	}
	if m.FormFields[0].Name != "name" {
		t.Errorf("expected field Name='name', got '%s'", m.FormFields[0].Name)
	}
}

// ---------------------------------------------------------------------------
// Fluent setters (return *ModalAction for chaining)
// ---------------------------------------------------------------------------

func TestModalAction_SetLabel(t *testing.T) {
	m := NewModal("archive").SetLabel("Archive Item")
	if m.Action.Label != "Archive Item" {
		t.Errorf("expected Label='Archive Item', got '%s'", m.Action.Label)
	}
}

func TestModalAction_SetIcon(t *testing.T) {
	m := NewModal("archive").SetIcon("archive")
	if m.Action.Icon != "archive" {
		t.Errorf("expected Icon='archive', got '%s'", m.Action.Icon)
	}
}

func TestModalAction_SetColor(t *testing.T) {
	m := NewModal("delete").SetColor(ColorDanger)
	if m.Action.Color != ColorDanger {
		t.Errorf("expected Color='danger', got '%s'", m.Action.Color)
	}
}

func TestModalAction_WithSuccessMessage(t *testing.T) {
	m := NewModal("publish").WithSuccessMessage("Published successfully.")
	if m.Action.SuccessMessage != "Published successfully." {
		t.Errorf("expected SuccessMessage='Published successfully.', got '%s'", m.Action.SuccessMessage)
	}
}

// ---------------------------------------------------------------------------
// ServeHTTP GET — renders modal HTML fragment
// ---------------------------------------------------------------------------

func TestModalAction_ServeHTTP_GET_renders_html(t *testing.T) {
	m := NewModal("archive").SetLabel("Archive").WithForm("/items/1/archive",
		ModalField{Name: "reason", Label: "Reason", Type: "textarea"},
	)

	req := httptest.NewRequest(http.MethodGet, "/items/1/archive", nil)
	rw := httptest.NewRecorder()
	m.ServeHTTP(rw, req)

	if rw.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rw.Code)
	}
	ct := rw.Header().Get("Content-Type")
	if !strings.HasPrefix(ct, "text/html") {
		t.Errorf("expected text/html Content-Type, got '%s'", ct)
	}
	body := rw.Body.String()
	if !strings.Contains(body, "modal-fragment") {
		t.Error("expected 'modal-fragment' id in rendered HTML")
	}
	if !strings.Contains(body, "Archive") {
		t.Error("expected modal label 'Archive' in rendered HTML")
	}
}

func TestModalAction_ServeHTTP_GET_confirmation_dialog(t *testing.T) {
	m := ConfirmAction("delete", "Delete this item?", "This action cannot be undone.")
	m.SetColor(ColorDanger)

	req := httptest.NewRequest(http.MethodGet, "/items/1", nil)
	rw := httptest.NewRecorder()
	m.ServeHTTP(rw, req)

	if rw.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rw.Code)
	}
	body := rw.Body.String()
	if !strings.Contains(body, "Delete this item?") {
		t.Error("expected modal title in body")
	}
	if !strings.Contains(body, "This action cannot be undone.") {
		t.Error("expected description in body")
	}
}

// ---------------------------------------------------------------------------
// ServeHTTP POST — redirects after processing
// ---------------------------------------------------------------------------

func TestModalAction_ServeHTTP_POST_redirects(t *testing.T) {
	m := NewModal("archive").WithForm("/items/1/archive")

	form := url.Values{}
	form.Set("reason", "no longer needed")

	req := httptest.NewRequest(http.MethodPost, "/items/1/archive", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rw := httptest.NewRecorder()
	m.ServeHTTP(rw, req)

	if rw.Code != http.StatusSeeOther {
		t.Errorf("expected 303 redirect after POST, got %d", rw.Code)
	}
	location := rw.Header().Get("Location")
	if location != "/items/1/archive" {
		t.Errorf("expected redirect to '/items/1/archive', got '%s'", location)
	}
}

func TestModalAction_ServeHTTP_POST_fallback_referer(t *testing.T) {
	m := NewModal("archive")
	// No FormAction → should redirect to Referer
	req := httptest.NewRequest(http.MethodPost, "/modal", strings.NewReader(""))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Referer", "/items")
	rw := httptest.NewRecorder()
	m.ServeHTTP(rw, req)

	if rw.Code != http.StatusSeeOther {
		t.Errorf("expected 303, got %d", rw.Code)
	}
	if rw.Header().Get("Location") != "/items" {
		t.Errorf("expected redirect to '/items' (Referer), got '%s'", rw.Header().Get("Location"))
	}
}

func TestModalAction_ServeHTTP_POST_fallback_root(t *testing.T) {
	m := NewModal("archive")
	// No FormAction, no Referer → redirect to "/"
	req := httptest.NewRequest(http.MethodPost, "/modal", strings.NewReader(""))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rw := httptest.NewRecorder()
	m.ServeHTTP(rw, req)

	if rw.Code != http.StatusSeeOther {
		t.Errorf("expected 303, got %d", rw.Code)
	}
	if rw.Header().Get("Location") != "/" {
		t.Errorf("expected redirect to '/', got '%s'", rw.Header().Get("Location"))
	}
}

// ---------------------------------------------------------------------------
// ServeHTTP — method not allowed
// ---------------------------------------------------------------------------

func TestModalAction_ServeHTTP_PUT_method_not_allowed(t *testing.T) {
	m := NewModal("test")
	req := httptest.NewRequest(http.MethodPut, "/modal", nil)
	rw := httptest.NewRecorder()
	m.ServeHTTP(rw, req)

	if rw.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405 for PUT, got %d", rw.Code)
	}
}

// ---------------------------------------------------------------------------
// ConfirmAction and DeleteModalAction helpers
// ---------------------------------------------------------------------------

func TestConfirmAction_defaults(t *testing.T) {
	m := ConfirmAction("publish", "Publish?", "This will go live.")

	if m.Action.Name != "publish" {
		t.Errorf("expected Name='publish', got '%s'", m.Action.Name)
	}
	if !m.Action.RequiresConfirmation {
		t.Error("expected RequiresConfirmation=true")
	}
	if m.Action.ModalTitle != "Publish?" {
		t.Errorf("expected ModalTitle='Publish?', got '%s'", m.Action.ModalTitle)
	}
	if m.Action.ConfirmLabel != "Confirm" {
		t.Errorf("expected ConfirmLabel='Confirm', got '%s'", m.Action.ConfirmLabel)
	}
	if m.Action.CancelLabel != "Cancel" {
		t.Errorf("expected CancelLabel='Cancel', got '%s'", m.Action.CancelLabel)
	}
}

func TestDeleteModalAction_configuration(t *testing.T) {
	m := DeleteModalAction("/items")

	if m.Action.Name != "delete" {
		t.Errorf("expected Name='delete', got '%s'", m.Action.Name)
	}
	if m.Action.Color != ColorDanger {
		t.Errorf("expected Color='danger', got '%s'", m.Action.Color)
	}
	if m.Action.Method != "DELETE" {
		t.Errorf("expected Method='DELETE', got '%s'", m.Action.Method)
	}
	if m.Action.SuccessMessage == "" {
		t.Error("expected non-empty SuccessMessage")
	}
}

// ---------------------------------------------------------------------------
// Color constants
// ---------------------------------------------------------------------------

func TestColorConstants_values(t *testing.T) {
	cases := []struct {
		constant Color
		expected string
	}{
		{ColorPrimary, "primary"},
		{ColorSecondary, "secondary"},
		{ColorDanger, "danger"},
		{ColorWarning, "warning"},
		{ColorSuccess, "success"},
		{ColorInfo, "info"},
		{ColorGray, "gray"},
	}
	for _, c := range cases {
		if c.constant != c.expected {
			t.Errorf("expected Color constant '%s', got '%s'", c.expected, c.constant)
		}
	}
}

func TestColorConstants_usable_as_string(t *testing.T) {
	// Color is a type alias for string; must be directly comparable.
	var col Color = ColorDanger
	if col != "danger" {
		t.Errorf("Color type must be string-compatible, got '%s'", col)
	}
}

// ---------------------------------------------------------------------------
// modalSizeClass helper
// ---------------------------------------------------------------------------

func TestModalSizeClass(t *testing.T) {
	cases := []struct {
		size     ModalSize
		expected string
	}{
		{ModalSizeSM, "max-w-sm"},
		{ModalSizeMD, "max-w-lg"},
		{ModalSizeLG, "max-w-2xl"},
		{ModalSizeXL, "max-w-4xl"},
		{ModalSizeFull, "max-w-full"},
	}
	for _, c := range cases {
		got := modalSizeClass(c.size)
		if got != c.expected {
			t.Errorf("modalSizeClass(%s): expected '%s', got '%s'", c.size, c.expected, got)
		}
	}
}

// ---------------------------------------------------------------------------
// htmlEscape helper
// ---------------------------------------------------------------------------

func TestHtmlEscape(t *testing.T) {
	raw := `<script>alert("xss")</script>`
	escaped := htmlEscape(raw)

	if strings.Contains(escaped, "<script>") {
		t.Error("expected < to be escaped")
	}
	if strings.Contains(escaped, `"`) {
		t.Error("expected \" to be escaped")
	}
}
