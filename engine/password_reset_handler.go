package engine

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/bozz33/sublimeadmin/mailer"
)

// PasswordResettable extends DatabaseClient with password reset capability.
// Implement this on your DB client to enable the PasswordResetHandler.
type PasswordResettable interface {
	DatabaseClient
	// UserExistsByEmail checks if a user with the given email exists.
	UserExistsByEmail(ctx interface{ Deadline() (time.Time, bool); Done() <-chan struct{}; Err() error; Value(any) any }, email string) (bool, error)
	// UpdatePassword sets a new hashed password for the user identified by email.
	UpdatePassword(ctx interface{ Deadline() (time.Time, bool); Done() <-chan struct{}; Err() error; Value(any) any }, email, hashedPassword string) error
}

// resetToken holds a password reset token with expiry.
type resetToken struct {
	email     string
	expiresAt time.Time
}

// resetStore is an in-memory token store (replace with DB for production).
var resetStore = struct {
	mu     sync.Mutex
	tokens map[string]resetToken
}{tokens: make(map[string]resetToken)}

// PasswordResetTemplates provides the templates needed by PasswordResetHandler.
type PasswordResetTemplates interface {
	ForgotPasswordPage(errMsg, successMsg string) interface {
		Render(ctx interface{ Deadline() (time.Time, bool); Done() <-chan struct{}; Err() error; Value(any) any }, w interface{ Write([]byte) (int, error) }) error
	}
	ResetPasswordPage(token, email, errMsg string) interface {
		Render(ctx interface{ Deadline() (time.Time, bool); Done() <-chan struct{}; Err() error; Value(any) any }, w interface{ Write([]byte) (int, error) }) error
	}
}

// PasswordResetHandler handles /forgot-password and /reset-password.
// It depends only on the DatabaseClient interface (no ORM required).
type PasswordResetHandler struct {
	db      PasswordResettable
	mailer  mailer.Mailer
	baseURL string
}

// NewPasswordResetHandler creates a new password reset handler.
// Pass a &mailer.LogMailer{} for development or mailer.NewSMTPMailer(cfg) for production.
func NewPasswordResetHandler(db PasswordResettable, m mailer.Mailer, baseURL string) *PasswordResetHandler {
	if m == nil {
		m = &mailer.LogMailer{}
	}
	return &PasswordResetHandler{db: db, mailer: m, baseURL: baseURL}
}

func (h *PasswordResetHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/forgot-password":
		switch r.Method {
		case http.MethodGet:
			h.showForgotForm(w, r)
		case http.MethodPost:
			h.handleForgotPassword(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	case "/reset-password":
		switch r.Method {
		case http.MethodGet:
			h.showResetForm(w, r)
		case http.MethodPost:
			h.handleResetPassword(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	default:
		http.NotFound(w, r)
	}
}

func (h *PasswordResetHandler) showForgotForm(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = fmt.Fprint(w, forgotPasswordHTML("", ""))
}

func (h *PasswordResetHandler) showResetForm(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	email := r.URL.Query().Get("email")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = fmt.Fprint(w, resetPasswordHTML(token, email, ""))
}

func (h *PasswordResetHandler) handleForgotPassword(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form", http.StatusBadRequest)
		return
	}
	email := r.FormValue("email")
	if email == "" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = fmt.Fprint(w, forgotPasswordHTML("Email is required.", ""))
		return
	}

	exists, _ := h.db.UserExistsByEmail(r.Context(), email)
	if exists {
		token := generateResetToken()
		resetStore.mu.Lock()
		resetStore.tokens[token] = resetToken{
			email:     email,
			expiresAt: time.Now().Add(1 * time.Hour),
		}
		resetStore.mu.Unlock()

		resetLink := fmt.Sprintf("%s/reset-password?token=%s&email=%s", h.baseURL, token, email)
		_ = h.mailer.Send(mailer.Message{
			To:      []string{email},
			Subject: "Reset your password",
			Body:    fmt.Sprintf("Click the link below to reset your password (valid 1 hour):\n\n%s\n", resetLink),
		})
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = fmt.Fprint(w, forgotPasswordHTML("", "If that email exists, a reset link has been sent."))
}

func (h *PasswordResetHandler) handleResetPassword(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form", http.StatusBadRequest)
		return
	}

	token := r.FormValue("token")
	email := r.FormValue("email")
	password := r.FormValue("password")
	confirm := r.FormValue("password_confirmation")

	showErr := func(msg string) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = fmt.Fprint(w, resetPasswordHTML(token, email, msg))
	}

	if password != confirm {
		showErr("Passwords do not match.")
		return
	}
	if len(password) < 8 {
		showErr("Password must be at least 8 characters.")
		return
	}

	resetStore.mu.Lock()
	entry, ok := resetStore.tokens[token]
	if ok {
		delete(resetStore.tokens, token)
	}
	resetStore.mu.Unlock()

	if !ok || entry.email != email || time.Now().After(entry.expiresAt) {
		showErr("This reset link is invalid or has expired.")
		return
	}

	ah := &AuthHandler{}
	newHash := ah.hashPassword(password)
	if err := h.db.UpdatePassword(r.Context(), email, newHash); err != nil {
		showErr("Failed to reset password. Please try again.")
		return
	}

	http.Redirect(w, r, "/login?reset=1", http.StatusFound)
}

// generateResetToken returns a cryptographically random 32-byte hex token.
func generateResetToken() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// ---------------------------------------------------------------------------
// Minimal HTML fallbacks (used when no template system is wired up)
// ---------------------------------------------------------------------------

func forgotPasswordHTML(errMsg, successMsg string) string {
	var msg string
	if errMsg != "" {
		msg = `<p style="color:red">` + errMsg + `</p>`
	}
	if successMsg != "" {
		msg = `<p style="color:green">` + successMsg + `</p>`
	}
	return `<!DOCTYPE html><html><body>
<h2>Forgot Password</h2>` + msg + `
<form method="POST">
  <label>Email<br><input type="email" name="email" required /></label><br><br>
  <button type="submit">Send Reset Link</button>
</form>
<p><a href="/login">Back to login</a></p>
</body></html>`
}

func resetPasswordHTML(token, email, errMsg string) string {
	var msg string
	if errMsg != "" {
		msg = `<p style="color:red">` + errMsg + `</p>`
	}
	return `<!DOCTYPE html><html><body>
<h2>Reset Password</h2>` + msg + `
<form method="POST">
  <input type="hidden" name="token" value="` + token + `" />
  <input type="hidden" name="email" value="` + email + `" />
  <label>New Password<br><input type="password" name="password" required minlength="8" /></label><br><br>
  <label>Confirm Password<br><input type="password" name="password_confirmation" required minlength="8" /></label><br><br>
  <button type="submit">Reset Password</button>
</form>
</body></html>`
}
