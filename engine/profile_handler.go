package engine

import (
	"net/http"

	"github.com/a-h/templ"
	authpkg "github.com/bozz33/sublimeadmin/auth"
	authtemplates "github.com/bozz33/sublimeadmin/views/auth"
)

// ProfileHandler handles GET/POST /profile.
type ProfileHandler struct {
	authManager *authpkg.Manager
	users       UserRepository
}

// NewProfileHandler creates a new profile handler.
func NewProfileHandler(authManager *authpkg.Manager, users UserRepository) *ProfileHandler {
	return &ProfileHandler{authManager: authManager, users: users}
}

func (h *ProfileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.showProfile(w, r)
	case http.MethodPost:
		action := r.FormValue("_action")
		switch action {
		case "change_password":
			h.handleChangePassword(w, r)
		default:
			h.handleUpdateProfile(w, r)
		}
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *ProfileHandler) showProfile(w http.ResponseWriter, r *http.Request) {
	u := h.currentUser(r)
	templ.Handler(authtemplates.ProfilePage(u, "", "")).ServeHTTP(w, r)
}

func (h *ProfileHandler) handleUpdateProfile(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form", http.StatusBadRequest)
		return
	}

	u := h.currentUser(r)
	name := r.FormValue("name")
	email := r.FormValue("email")

	if name == "" || email == "" {
		templ.Handler(authtemplates.ProfilePage(u, "Name and email are required.", "")).ServeHTTP(w, r)
		return
	}

	// Check email uniqueness (skip if unchanged)
	if email != u.Email {
		exists, err := h.users.ExistsByEmailExcluding(r.Context(), email, u.ID)
		if err != nil {
			templ.Handler(authtemplates.ProfilePage(u, "Database error.", "")).ServeHTTP(w, r)
			return
		}
		if exists {
			templ.Handler(authtemplates.ProfilePage(u, "This email is already in use.", "")).ServeHTTP(w, r)
			return
		}
	}

	err := h.users.UpdateNameEmail(r.Context(), u.ID, name, email)
	if err != nil {
		templ.Handler(authtemplates.ProfilePage(u, "Failed to update profile.", "")).ServeHTTP(w, r)
		return
	}

	// Refresh session user
	u.Name = name
	u.Email = email
	_ = h.authManager.UpdateUserFromRequest(r, u)

	templ.Handler(authtemplates.ProfilePage(u, "", "Profile updated successfully.")).ServeHTTP(w, r)
}

func (h *ProfileHandler) handleChangePassword(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form", http.StatusBadRequest)
		return
	}

	u := h.currentUser(r)
	current := r.FormValue("current_password")
	newPwd := r.FormValue("new_password")
	confirm := r.FormValue("new_password_confirmation")

	if newPwd != confirm {
		templ.Handler(authtemplates.ProfilePage(u, "New passwords do not match.", "")).ServeHTTP(w, r)
		return
	}
	if len(newPwd) < 8 {
		templ.Handler(authtemplates.ProfilePage(u, "Password must be at least 8 characters.", "")).ServeHTTP(w, r)
		return
	}

	// Load current hash from DB
	dbUser, err := h.users.GetByID(r.Context(), u.ID)
	if err != nil {
		templ.Handler(authtemplates.ProfilePage(u, "User not found.", "")).ServeHTTP(w, r)
		return
	}

	ah := &AuthHandler{}
	if !ah.verifyPassword(current, dbUser.GetPassword()) {
		templ.Handler(authtemplates.ProfilePage(u, "Current password is incorrect.", "")).ServeHTTP(w, r)
		return
	}

	newHash := ah.hashPassword(newPwd)
	err = h.users.UpdatePassword(r.Context(), u.ID, newHash)
	if err != nil {
		templ.Handler(authtemplates.ProfilePage(u, "Failed to update password.", "")).ServeHTTP(w, r)
		return
	}

	templ.Handler(authtemplates.ProfilePage(u, "", "Password changed successfully.")).ServeHTTP(w, r)
}

// currentUser returns the authenticated user from context, falling back to a
// minimal user built from the session ID.
func (h *ProfileHandler) currentUser(r *http.Request) *authpkg.User {
	u := authpkg.UserFromContext(r.Context())
	if u != nil && u.IsAuthenticated() {
		return u
	}
	id := h.authManager.UserIDFromRequest(r)
	return &authpkg.User{ID: id}
}
