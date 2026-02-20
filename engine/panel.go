package engine

import (
	"context"
	"net/http"
	"sort"
	"time"

	"github.com/a-h/templ"
	"github.com/alexedwards/scs/v2"
	"github.com/bozz33/sublimeadmin/auth"
	"github.com/bozz33/sublimeadmin/ui/layouts"
	"github.com/samber/lo"
)

// DatabaseClient is an interface for database operations.
// Users must implement this interface with their own database client.
type DatabaseClient interface {
	// FindUserByEmail finds a user by email for authentication.
	FindUserByEmail(ctx context.Context, email string) (*auth.User, string, error)
	// CreateUser creates a new user and returns the auth.User.
	CreateUser(ctx context.Context, name, email, hashedPassword string) (*auth.User, error)
	// UserExists checks if a user with the given email exists.
	UserExists(ctx context.Context, email string) (bool, error)
}

// DashboardProvider provides dashboard widgets.
type DashboardProvider interface {
	GetWidgets(ctx context.Context) []templ.Component
}

// AuthTemplates provides authentication page templates.
type AuthTemplates interface {
	LoginPage() templ.Component
	RegisterPage() templ.Component
}

// DashboardTemplates provides dashboard page templates.
type DashboardTemplates interface {
	Index(widgets []templ.Component) templ.Component
}

// Panel represents a complete dashboard.
type Panel struct {
	ID                 string
	Path               string
	BrandName          string
	DB                 DatabaseClient
	Resources          []Resource
	Pages              []Page
	AuthManager        *auth.Manager
	Session            *scs.SessionManager
	DashboardProvider  DashboardProvider
	AuthTemplates      AuthTemplates
	DashboardTemplates DashboardTemplates
}

// NewPanel initializes an empty Panel.
func NewPanel(id string) *Panel {
	return &Panel{
		ID:        id,
		BrandName: "SublimeGo",
		Resources: make([]Resource, 0),
		Pages:     make([]Page, 0),
	}
}

// Builder methods for fluent configuration.

func (p *Panel) SetPath(path string) *Panel {
	p.Path = path
	return p
}

func (p *Panel) SetDatabase(db DatabaseClient) *Panel {
	p.DB = db
	return p
}

func (p *Panel) SetBrandName(name string) *Panel {
	p.BrandName = name
	return p
}

func (p *Panel) SetAuthManager(authManager *auth.Manager) *Panel {
	p.AuthManager = authManager
	return p
}

func (p *Panel) SetSession(session *scs.SessionManager) *Panel {
	p.Session = session
	return p
}

func (p *Panel) SetDashboardProvider(provider DashboardProvider) *Panel {
	p.DashboardProvider = provider
	return p
}

func (p *Panel) SetAuthTemplates(templates AuthTemplates) *Panel {
	p.AuthTemplates = templates
	return p
}

func (p *Panel) SetDashboardTemplates(templates DashboardTemplates) *Panel {
	p.DashboardTemplates = templates
	return p
}

// AddResources adds a block of resources.
func (p *Panel) AddResources(rs ...Resource) *Panel {
	p.Resources = append(p.Resources, rs...)
	p.registerNavItems()
	return p
}

// AddPages adds custom pages to the panel.
func (p *Panel) AddPages(pages ...Page) *Panel {
	p.Pages = append(p.Pages, pages...)
	p.registerNavItems()
	return p
}

// navItem is a unified type for navigation items
type navItem struct {
	slug  string
	label string
	icon  string
	group string
	sort  int
}

// registerNavItems injects navigation items into the sidebar.
func (p *Panel) registerNavItems() {
	var allItems []navItem

	for _, r := range p.Resources {
		allItems = append(allItems, navItem{
			slug:  r.Slug(),
			label: r.PluralLabel(),
			icon:  r.Icon(),
			group: r.Group(),
			sort:  r.Sort(),
		})
	}

	for _, pg := range p.Pages {
		allItems = append(allItems, navItem{
			slug:  pg.Slug(),
			label: pg.Label(),
			icon:  pg.Icon(),
			group: pg.Group(),
			sort:  pg.Sort(),
		})
	}

	sort.Slice(allItems, func(i, j int) bool {
		return allItems[i].sort < allItems[j].sort
	})

	grouped := lo.GroupBy(allItems, func(item navItem) string {
		if item.group == "" {
			return "_root"
		}
		return item.group
	})

	var navGroups []layouts.NavGroup

	if rootItems, ok := grouped["_root"]; ok {
		items := lo.Map(rootItems, func(item navItem, _ int) layouts.NavItem {
			return layouts.NavItem{
				Slug:  item.slug,
				Label: item.label,
				Icon:  item.icon,
			}
		})
		navGroups = append(navGroups, layouts.NavGroup{
			Label: "",
			Items: items,
		})
	}

	for groupName, items := range grouped {
		if groupName == "_root" {
			continue
		}
		navItems := lo.Map(items, func(item navItem, _ int) layouts.NavItem {
			return layouts.NavItem{
				Slug:  item.slug,
				Label: item.label,
				Icon:  item.icon,
			}
		})
		navGroups = append(navGroups, layouts.NavGroup{
			Label: groupName,
			Items: navItems,
		})
	}

	layouts.SetNavGroups(navGroups)
}

// Router generates the standard HTTP Handler with automatic CRUD.
func (p *Panel) Router() http.Handler {
	mux := http.NewServeMux()

	fs := http.FileServer(http.Dir("ui/assets"))
	mux.Handle("/assets/", http.StripPrefix("/assets/", fs))

	if p.AuthManager != nil && p.AuthTemplates != nil {
		authHandler := NewAuthHandler(p.AuthManager, p.DB, p.AuthTemplates)
		mux.Handle("/login", RequireGuest(p.AuthManager)(authHandler))
		mux.Handle("/register", RequireGuest(p.AuthManager)(authHandler))
		mux.Handle("/logout", authHandler)
	}

	dashboardHandler := RequireAuth(p.AuthManager, p.DB)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var widgets []templ.Component
		if p.DashboardProvider != nil {
			widgets = p.DashboardProvider.GetWidgets(r.Context())
		}
		if p.DashboardTemplates != nil {
			p.DashboardTemplates.Index(widgets).Render(r.Context(), w)
		}
	}))
	mux.Handle("/", dashboardHandler)

	for _, res := range p.Resources {
		handler := NewCRUDHandler(res)
		slug := res.Slug()
		protectedHandler := RequireAuth(p.AuthManager, p.DB)(handler)
		mux.Handle("/"+slug+"/", protectedHandler)
		mux.Handle("/"+slug, protectedHandler)
	}

	for _, pg := range p.Pages {
		pageHandler := NewPageHandler(pg)
		slug := pg.Slug()
		protectedHandler := RequireAuth(p.AuthManager, p.DB)(pageHandler)
		mux.Handle("/"+slug, protectedHandler)
	}

	if p.Session != nil {
		return p.Session.LoadAndSave(mux)
	}

	return mux
}

// ---------------------------------------------------------------------------
// ServerOption — functional options for Server()
// ---------------------------------------------------------------------------

// ServerOption configures the http.Server returned by Panel.Server().
type ServerOption func(*http.Server)

// WithReadTimeout sets the server read timeout.
func WithReadTimeout(d time.Duration) ServerOption {
	return func(s *http.Server) { s.ReadTimeout = d }
}

// WithWriteTimeout sets the server write timeout.
func WithWriteTimeout(d time.Duration) ServerOption {
	return func(s *http.Server) { s.WriteTimeout = d }
}

// WithIdleTimeout sets the server idle timeout.
func WithIdleTimeout(d time.Duration) ServerOption {
	return func(s *http.Server) { s.IdleTimeout = d }
}

// Server creates an *http.Server with secure timeouts by default.
// Override any timeout with the provided ServerOption functions.
//
//	srv := panel.Server(":8080")
//	log.Fatal(srv.ListenAndServe())
func (p *Panel) Server(addr string, opts ...ServerOption) *http.Server {
	srv := &http.Server{
		Addr:         addr,
		Handler:      p.Router(),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
	for _, opt := range opts {
		opt(srv)
	}
	return srv
}

// ListenAndServe starts the panel on addr with secure timeouts.
// It blocks until the server stops.
//
//	log.Fatal(panel.ListenAndServe(":8080"))
func (p *Panel) ListenAndServe(addr string, opts ...ServerOption) error {
	return p.Server(addr, opts...).ListenAndServe()
}
