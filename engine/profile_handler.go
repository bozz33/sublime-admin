package engine

import (
	"context"
	"fmt"
	"net/http"

	"github.com/bozz33/sublimeadmin/auth"
)

// ProfileUpdatable extends DatabaseClient with profile management capability.
// Implement this on your DB client to enable the ProfileHandler.
type ProfileUpdatable interface {
	DatabaseClient
	// UpdateUserProfile updates the name and email of the user identified by userID.
	UpdateUserProfile(ctx context.Context, userID int, name, email string) error
	// EmailTakenByOther returns true if the email is already used by a different user.
	EmailTakenByOther(ctx context.Context, userID int, email string) (bool, error)
	// GetHashedPassword returns the stored bcrypt hash for the user identified by userID.
	GetHashedPassword(ctx context.Context, userID int) (string, error)
	// UpdateUserPassword sets a new hashed password for the user identified by userID.
	UpdateUserPassword(ctx context.Context, userID int, hashedPassword string) error
}

// ProfileTemplates provides the templates needed by ProfileHandler.
// If nil, a minimal built-in HTML form is used.
type ProfileTemplates interface {
	ProfilePage(user *auth.User, errMsg, successMsg string) interface {
		Render(ctx context.Context, w interface{ Write([]byte) (int, error) }) error
	}
}

// ProfileHandler handles GET/POST /profile.
// It depends only on the ProfileUpdatable interface — no ORM required.
type ProfileHandler struct {
	authManager *auth.Manager
	db          ProfileUpdatable
	templates   ProfileTemplates
}

// NewProfileHandler creates a new profile handler.
// Pass nil for templates to use the built-in minimal HTML form.
func NewProfileHandler(authManager *auth.Manager, db ProfileUpdatable, templates ProfileTemplates) *ProfileHandler {
	return &ProfileHandler{authManager: authManager, db: db, templates: templates}
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

func (h *ProfileHandler) currentUser(r *http.Request) *auth.User {
	u := auth.UserFromContext(r.Context())
	if u != nil && u.IsAuthenticated() {
		return u
	}
	id := h.authManager.UserIDFromRequest(r)
	return &auth.User{ID: id}
}

func (h *ProfileHandler) showProfile(w http.ResponseWriter, r *http.Request) {
	u := h.currentUser(r)
	h.renderProfile(w, r, u, "", "")
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
		h.renderProfile(w, r, u, "Name and email are required.", "")
		return
	}

	if email != u.Email {
		taken, err := h.db.EmailTakenByOther(r.Context(), u.ID, email)
		if err != nil {
			h.renderProfile(w, r, u, "Database error.", "")
			return
		}
		if taken {
			h.renderProfile(w, r, u, "This email is already in use.", "")
			return
		}
	}

	if err := h.db.UpdateUserProfile(r.Context(), u.ID, name, email); err != nil {
		h.renderProfile(w, r, u, "Failed to update profile.", "")
		return
	}

	u.Name = name
	u.Email = email
	_ = h.authManager.UpdateUserFromRequest(r, u)

	h.renderProfile(w, r, u, "", "Profile updated successfully.")
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
		h.renderProfile(w, r, u, "New passwords do not match.", "")
		return
	}
	if len(newPwd) < 8 {
		h.renderProfile(w, r, u, "Password must be at least 8 characters.", "")
		return
	}

	currentHash, err := h.db.GetHashedPassword(r.Context(), u.ID)
	if err != nil {
		h.renderProfile(w, r, u, "User not found.", "")
		return
	}

	ah := &AuthHandler{}
	if !ah.verifyPassword(current, currentHash) {
		h.renderProfile(w, r, u, "Current password is incorrect.", "")
		return
	}

	newHash := ah.hashPassword(newPwd)
	if err := h.db.UpdateUserPassword(r.Context(), u.ID, newHash); err != nil {
		h.renderProfile(w, r, u, "Failed to update password.", "")
		return
	}

	h.renderProfile(w, r, u, "", "Password changed successfully.")
}

func (h *ProfileHandler) renderProfile(w http.ResponseWriter, r *http.Request, u *auth.User, errMsg, successMsg string) {
	if h.templates != nil {
		comp := h.templates.ProfilePage(u, errMsg, successMsg)
		if comp != nil {
			_ = comp.Render(r.Context(), w)
			return
		}
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = fmt.Fprint(w, profileHTML(u, errMsg, successMsg))
}

// ---------------------------------------------------------------------------
// Minimal HTML fallback
// ---------------------------------------------------------------------------

func profileHTML(u *auth.User, errMsg, successMsg string) string {
	var msg string
	if errMsg != "" {
		msg = `<p style="color:red">` + errMsg + `</p>`
	}
	if successMsg != "" {
		msg = `<p style="color:green">` + successMsg + `</p>`
	}
	name := ""
	email := ""
	if u != nil {
		name = u.Name
		email = u.Email
	}
	return `<!DOCTYPE html><html><body>
<h2>Profile</h2>` + msg + `
<form method="POST">
  <label>Name<br><input type="text" name="name" value="` + name + `" required /></label><br><br>
  <label>Email<br><input type="email" name="email" value="` + email + `" required /></label><br><br>
  <button type="submit">Update Profile</button>
</form>
<hr>
<h3>Change Password</h3>
<form method="POST">
  <input type="hidden" name="_action" value="change_password" />
  <label>Current Password<br><input type="password" name="current_password" required /></label><br><br>
  <label>New Password<br><input type="password" name="new_password" required minlength="8" /></label><br><br>
  <label>Confirm New Password<br><input type="password" name="new_password_confirmation" required minlength="8" /></label><br><br>
  <button type="submit">Change Password</button>
</form>
</body></html>`
}
