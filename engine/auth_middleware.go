package engine

import (
"net/http"

"github.com/bozz33/sublimeadmin/auth"
)

// AuthMiddleware provides authentication middleware.

// RequireAuth returns a middleware that requires authentication.
func RequireAuth(authManager *auth.Manager, db DatabaseClient) func(http.Handler) http.Handler {
return func(next http.Handler) http.Handler {
return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
if authManager == nil {
next.ServeHTTP(w, r)
return
}

if !authManager.IsAuthenticatedFromRequest(r) {
http.Redirect(w, r, "/login", http.StatusFound)
return
}

next.ServeHTTP(w, r)
})
}
}

// RequireGuest returns a middleware that requires the user to be a guest.
func RequireGuest(authManager *auth.Manager) func(http.Handler) http.Handler {
return func(next http.Handler) http.Handler {
return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
if authManager == nil {
next.ServeHTTP(w, r)
return
}

if authManager.IsAuthenticatedFromRequest(r) {
http.Redirect(w, r, "/admin", http.StatusFound)
return
}

next.ServeHTTP(w, r)
})
}
}
