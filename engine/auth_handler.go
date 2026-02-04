package engine

import (
"net/http"
"strings"

"github.com/a-h/templ"
"github.com/bozz33/sublime-admin/auth"
"golang.org/x/crypto/bcrypt"
)

// AuthHandler handles authentication routes.
type AuthHandler struct {
authManager *auth.Manager
db          DatabaseClient
templates   AuthTemplates
}

// NewAuthHandler creates a new authentication handler.
func NewAuthHandler(authManager *auth.Manager, db DatabaseClient, templates AuthTemplates) *AuthHandler {
return &AuthHandler{
authManager: authManager,
db:          db,
templates:   templates,
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

func (h *AuthHandler) showLogin(w http.ResponseWriter, r *http.Request) {
if h.authManager.IsAuthenticatedFromRequest(r) {
http.Redirect(w, r, "/admin", http.StatusFound)
return
}
templ.Handler(h.templates.LoginPage()).ServeHTTP(w, r)
}

func (h *AuthHandler) handleLogin(w http.ResponseWriter, r *http.Request) {
if err := r.ParseForm(); err != nil {
http.Error(w, "Invalid form", http.StatusBadRequest)
return
}

email := r.FormValue("email")
password := r.FormValue("password")

user, hashedPassword, err := h.db.FindUserByEmail(r.Context(), email)
if err != nil {
h.showLoginWithError(w, r, "Invalid email or password")
return
}

if !h.verifyPassword(password, hashedPassword) {
h.showLoginWithError(w, r, "Invalid email or password")
return
}

if err := h.authManager.LoginWithRequest(r, user); err != nil {
http.Error(w, "Login failed", http.StatusInternalServerError)
return
}

http.Redirect(w, r, "/admin", http.StatusFound)
}

func (h *AuthHandler) showRegister(w http.ResponseWriter, r *http.Request) {
if h.authManager.IsAuthenticatedFromRequest(r) {
http.Redirect(w, r, "/admin", http.StatusFound)
return
}
templ.Handler(h.templates.RegisterPage()).ServeHTTP(w, r)
}

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

exists, err := h.db.UserExists(r.Context(), email)
if err != nil {
http.Error(w, "Database error", http.StatusInternalServerError)
return
}
if exists {
h.showRegisterWithError(w, r, "Email already in use")
return
}

hashedPassword := h.hashPassword(password)
user, err := h.db.CreateUser(r.Context(), name, email, hashedPassword)
if err != nil {
http.Error(w, "Failed to create user", http.StatusInternalServerError)
return
}

if err := h.authManager.LoginWithRequest(r, user); err != nil {
http.Error(w, "Login failed", http.StatusInternalServerError)
return
}

http.Redirect(w, r, "/admin", http.StatusFound)
}

func (h *AuthHandler) handleLogout(w http.ResponseWriter, r *http.Request) {
if err := h.authManager.LogoutWithRequest(r); err != nil {
http.Error(w, "Logout failed", http.StatusInternalServerError)
return
}
http.Redirect(w, r, "/login", http.StatusFound)
}

func (h *AuthHandler) showLoginWithError(w http.ResponseWriter, r *http.Request, message string) {
templ.Handler(h.templates.LoginPage()).ServeHTTP(w, r)
}

func (h *AuthHandler) showRegisterWithError(w http.ResponseWriter, r *http.Request, message string) {
templ.Handler(h.templates.RegisterPage()).ServeHTTP(w, r)
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
