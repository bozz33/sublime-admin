package engine

import (
	"net/http"
	"strings"

	"github.com/bozz33/sublimeadmin/auth"
	"github.com/bozz33/sublimeadmin/ui/layouts"
)

// AuthMiddleware provides authentication middleware.

// RequireAuth returns a middleware that requires authentication.
// Unauthenticated requests are redirected to the panel's login page.
func RequireAuth(authManager *auth.Manager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if authManager == nil {
				next.ServeHTTP(w, r)
				return
			}

			if !authManager.IsAuthenticatedFromRequest(r) {
				cfg := layouts.GetPanelConfigFromContext(r.Context())
				loginURL := strings.TrimRight(cfg.Path, "/") + "/login"
				if cfg.Path == "" || cfg.Path == "/" {
					loginURL = "/login"
				}
				http.Redirect(w, r, loginURL, http.StatusFound)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireGuest returns a middleware that requires the user to be a guest.
// Authenticated users are redirected to the panel's dashboard.
func RequireGuest(authManager *auth.Manager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if authManager == nil {
				next.ServeHTTP(w, r)
				return
			}

			if authManager.IsAuthenticatedFromRequest(r) {
				cfg := layouts.GetPanelConfigFromContext(r.Context())
				dashURL := cfg.Path
				if dashURL == "" {
					dashURL = "/"
				}
				http.Redirect(w, r, dashURL, http.StatusFound)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
