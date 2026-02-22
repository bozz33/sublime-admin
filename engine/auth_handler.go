package engine

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/a-h/templ"
	authpkg "github.com/bozz33/sublimeadmin/auth"
	authtemplates "github.com/bozz33/sublimeadmin/views/auth"
	"golang.org/x/crypto/bcrypt"
)

// UserRepository is the interface the framework needs to authenticate users.
// Implement it in your project using your own ORM or database layer.
type UserRepository interface {
	FindByEmail(ctx context.Context, email string) (FrameworkUser, error)
	Create(ctx context.Context, name, email, hashedPassword string) (FrameworkUser, error)
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	ExistsByEmailExcluding(ctx context.Context, email string, excludeID int) (bool, error)
	UpdateNameEmail(ctx context.Context, id int, name, email string) error
	UpdatePassword(ctx context.Context, id int, hashedPassword string) error
	GetByID(ctx context.Context, id int) (FrameworkUser, error)
}

// FrameworkUser is the minimal user data the framework needs.
type FrameworkUser interface {
	GetID() int
	GetName() string
	GetEmail() string
	GetPassword() string
}

// AuthHandler handles authentication routes.
type AuthHandler struct {
	authManager *authpkg.Manager
	users       UserRepository
}

// NewAuthHandler creates a new authentication handler.
func NewAuthHandler(authManager *authpkg.Manager, users UserRepository) *AuthHandler {
	return &AuthHandler{
		authManager: authManager,
		users:       users,
	}
}

// ServeHTTP implements http.Handler for routing.
func (h *AuthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/login":
		switch r.Method {
		case http.MethodGet:
			h.showLogin(w, r)
		case http.MethodPost:
			h.handleLogin(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	case "/register":
		switch r.Method {
		case http.MethodGet:
			h.showRegister(w, r)
		case http.MethodPost:
			h.handleRegister(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	case "/logout":
		h.handleLogout(w, r)
	default:
		http.NotFound(w, r)
	}
}

// showLogin displays the login page.
func (h *AuthHandler) showLogin(w http.ResponseWriter, r *http.Request) {
	if h.authManager.IsAuthenticatedFromRequest(r) {
		http.Redirect(w, r, "/admin", http.StatusFound)
		return
	}

	templ.Handler(authtemplates.LoginPage()).ServeHTTP(w, r)
}

// handleLogin handles login form submission.
func (h *AuthHandler) handleLogin(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form", http.StatusBadRequest)
		return
	}

	email := r.FormValue("email")
	password := r.FormValue("password")

	dbUser, err := h.users.FindByEmail(r.Context(), email)
	if err != nil {
		h.showLoginWithError(w, r, "Invalid email or password")
		return
	}

	if !h.verifyPassword(password, dbUser.GetPassword()) {
		h.showLoginWithError(w, r, "Invalid email or password")
		return
	}

	authUser := &authpkg.User{
		ID:    dbUser.GetID(),
		Name:  dbUser.GetName(),
		Email: dbUser.GetEmail(),
	}

	if err := h.authManager.LoginWithRequest(r, authUser); err != nil {
		http.Error(w, "Login failed", http.StatusInternalServerError)
		return
	}

	// Remember Me: extend session lifetime via a long-lived cookie
	if r.FormValue("remember_me") == "1" || r.FormValue("remember_me") == "on" {
		http.SetCookie(w, &http.Cookie{
			Name:     "_remember",
			Value:    "1",
			Path:     "/",
			MaxAge:   int((30 * 24 * time.Hour).Seconds()),
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})
	}

	intendedURL := h.getIntendedURL(r)
	if intendedURL == "" {
		intendedURL = "/admin"
	}
	http.Redirect(w, r, intendedURL, http.StatusFound)
}

// showRegister displays the registration page.
func (h *AuthHandler) showRegister(w http.ResponseWriter, r *http.Request) {
	if h.authManager.IsAuthenticatedFromRequest(r) {
		http.Redirect(w, r, "/admin", http.StatusFound)
		return
	}

	templ.Handler(authtemplates.RegisterPage()).ServeHTTP(w, r)
}

// handleRegister handles registration form submission.
func (h *AuthHandler) handleRegister(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form", http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	email := r.FormValue("email")
	password := r.FormValue("password")
	passwordConfirm := r.FormValue("password_confirmation")

	if password != passwordConfirm {
		h.showRegisterWithError(w, r, "Passwords do not match")
		return
	}

	if len(password) < 6 {
		h.showRegisterWithError(w, r, "Password must be at least 6 characters")
		return
	}

	exists, err := h.users.ExistsByEmail(r.Context(), email)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	if exists {
		h.showRegisterWithError(w, r, "Email already in use")
		return
	}

	hashedPassword := h.hashPassword(password)
	newUser, err := h.users.Create(r.Context(), name, email, hashedPassword)
	if err != nil {
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}
	authUser := &authpkg.User{
		ID:    newUser.GetID(),
		Name:  newUser.GetName(),
		Email: newUser.GetEmail(),
	}

	if err := h.authManager.LoginWithRequest(r, authUser); err != nil {
		http.Error(w, "Login failed", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin", http.StatusFound)
}

// handleLogout logs out the user.
func (h *AuthHandler) handleLogout(w http.ResponseWriter, r *http.Request) {
	if err := h.authManager.LogoutWithRequest(r); err != nil {
		http.Error(w, "Logout failed", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/login", http.StatusFound)
}

// Helpers

func (h *AuthHandler) showLoginWithError(w http.ResponseWriter, r *http.Request, message string) {
	templ.Handler(authtemplates.LoginPage()).ServeHTTP(w, r)
}

func (h *AuthHandler) showRegisterWithError(w http.ResponseWriter, r *http.Request, message string) {
	templ.Handler(authtemplates.RegisterPage()).ServeHTTP(w, r)
}

func (h *AuthHandler) getIntendedURL(r *http.Request) string {
	return ""
}

func (h *AuthHandler) verifyPassword(password, hash string) bool {
	normalizedHash := hash
	if strings.HasPrefix(hash, "$2y$") {
		normalizedHash = "$2a$" + hash[4:]
	}

	err := bcrypt.CompareHashAndPassword([]byte(normalizedHash), []byte(password))
	return err == nil
}

func (h *AuthHandler) hashPassword(password string) string {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return ""
	}
	return string(hashedBytes)
}
