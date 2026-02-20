package engine

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/alexedwards/scs/v2"
)

// ---------------------------------------------------------------------------
// Tenant — core type
// ---------------------------------------------------------------------------

// Tenant represents an isolated tenant in a multi-tenant setup.
type Tenant struct {
	// ID is the unique identifier (e.g. "acme", "globex").
	ID string
	// Domain is the full hostname that maps to this tenant (e.g. "acme.example.com").
	Domain string
	// Subdomain is the subdomain prefix (e.g. "acme" for "acme.example.com").
	Subdomain string
	// Name is the display name of the tenant.
	Name string
	// Meta holds arbitrary tenant-specific key/value data (DB DSN, plan, etc.).
	Meta map[string]any
}

// ---------------------------------------------------------------------------
// TenantResolver — extracts tenant from request
// ---------------------------------------------------------------------------

// TenantResolver resolves the current tenant from an HTTP request.
type TenantResolver interface {
	Resolve(r *http.Request) (*Tenant, bool)
}

// SubdomainResolver resolves tenants by subdomain (e.g. {tenant}.example.com).
type SubdomainResolver struct {
	mu      sync.RWMutex
	tenants map[string]*Tenant
	base    string
}

// NewSubdomainResolver creates a resolver that maps subdomains to tenants.
func NewSubdomainResolver(baseDomain string) *SubdomainResolver {
	return &SubdomainResolver{
		tenants: make(map[string]*Tenant),
		base:    baseDomain,
	}
}

// Register adds a tenant to the resolver.
func (r *SubdomainResolver) Register(t *Tenant) *SubdomainResolver {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tenants[t.Subdomain] = t
	if t.Domain != "" {
		r.tenants[t.Domain] = t
	}
	return r
}

// Resolve extracts the subdomain from the host and looks up the tenant.
func (r *SubdomainResolver) Resolve(req *http.Request) (*Tenant, bool) {
	host := req.Host
	if idx := strings.LastIndex(host, ":"); idx != -1 {
		host = host[:idx]
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	if t, ok := r.tenants[host]; ok {
		return t, true
	}
	if strings.HasSuffix(host, "."+r.base) {
		sub := strings.TrimSuffix(host, "."+r.base)
		if t, ok := r.tenants[sub]; ok {
			return t, true
		}
	}
	return nil, false
}

// PathResolver resolves tenants from the URL path prefix (e.g. /acme/admin/...).
type PathResolver struct {
	mu      sync.RWMutex
	tenants map[string]*Tenant
}

// NewPathResolver creates a resolver that maps URL path prefixes to tenants.
func NewPathResolver() *PathResolver {
	return &PathResolver{tenants: make(map[string]*Tenant)}
}

// Register adds a tenant.
func (r *PathResolver) Register(t *Tenant) *PathResolver {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tenants[t.ID] = t
	return r
}

// Resolve extracts the first path segment and looks up the tenant.
func (r *PathResolver) Resolve(req *http.Request) (*Tenant, bool) {
	path := strings.TrimPrefix(req.URL.Path, "/")
	seg := strings.SplitN(path, "/", 2)[0]
	r.mu.RLock()
	defer r.mu.RUnlock()
	if t, ok := r.tenants[seg]; ok {
		return t, true
	}
	return nil, false
}

// ---------------------------------------------------------------------------
// Context helpers
// ---------------------------------------------------------------------------

const contextKeyTenant contextKey = "tenant"

// WithTenant injects a tenant into the context.
func WithTenant(ctx context.Context, t *Tenant) context.Context {
	return context.WithValue(ctx, contextKeyTenant, t)
}

// TenantFromContext retrieves the current tenant from context.
func TenantFromContext(ctx context.Context) *Tenant {
	if t, ok := ctx.Value(contextKeyTenant).(*Tenant); ok {
		return t
	}
	return nil
}

// ---------------------------------------------------------------------------
// TenantMiddleware — injects tenant into every request context
// ---------------------------------------------------------------------------

// TenantMiddleware resolves the tenant for each request and injects it into ctx.
func TenantMiddleware(resolver TenantResolver, requireTenant bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tenant, ok := resolver.Resolve(r)
			if !ok {
				if requireTenant {
					http.NotFound(w, r)
					return
				}
				next.ServeHTTP(w, r)
				return
			}
			ctx := WithTenant(r.Context(), tenant)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// ---------------------------------------------------------------------------
// MultiPanelRouter — routes requests to the correct Panel per tenant
// ---------------------------------------------------------------------------

// MultiPanelRouter dispatches HTTP requests to the correct Panel based on the resolved tenant.
type MultiPanelRouter struct {
	resolver      TenantResolver
	panels        map[string]*Panel
	mu            sync.RWMutex
	handlers      map[string]http.Handler
	cachedEntries map[string]*cachedPanelEntry
	fallback      http.Handler
	factory       PanelFactory
	cacheTTL      time.Duration
}

// NewMultiPanelRouter creates a router that dispatches to per-tenant panels.
func NewMultiPanelRouter(resolver TenantResolver) *MultiPanelRouter {
	return &MultiPanelRouter{
		resolver:      resolver,
		panels:        make(map[string]*Panel),
		handlers:      make(map[string]http.Handler),
		cachedEntries: make(map[string]*cachedPanelEntry),
		cacheTTL:      10 * time.Minute,
	}
}

// RegisterPanel associates a Panel with a tenant ID.
func (m *MultiPanelRouter) RegisterPanel(tenantID string, p *Panel) *MultiPanelRouter {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.panels[tenantID] = p
	return m
}

// SetFallback sets the handler used when no tenant matches.
func (m *MultiPanelRouter) SetFallback(h http.Handler) *MultiPanelRouter {
	m.fallback = h
	return m
}

// ServeHTTP resolves the tenant and dispatches to the correct panel handler.
func (m *MultiPanelRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	tenant, ok := m.resolver.Resolve(r)
	if !ok {
		if m.fallback != nil {
			m.fallback.ServeHTTP(w, r)
		} else {
			http.NotFound(w, r)
		}
		return
	}
	ctx := WithTenant(r.Context(), tenant)
	r = r.WithContext(ctx)
	h := m.getHandler(tenant.ID)
	if h == nil {
		http.Error(w, "Panel not configured for tenant: "+tenant.ID, http.StatusServiceUnavailable)
		return
	}
	h.ServeHTTP(w, r)
}

func (m *MultiPanelRouter) getHandler(tenantID string) http.Handler {
	// Fast path: check static handlers (no TTL — built once from RegisterPanel)
	m.mu.RLock()
	h, ok := m.handlers[tenantID]
	// Also check factory cache with TTL
	entry, hasEntry := m.cachedEntries[tenantID]
	m.mu.RUnlock()

	if ok {
		return h
	}
	if hasEntry && (m.cacheTTL == 0 || time.Since(entry.createdAt) < m.cacheTTL) {
		return entry.handler
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Double-check after acquiring write lock
	if h, ok = m.handlers[tenantID]; ok {
		return h
	}
	if entry, hasEntry = m.cachedEntries[tenantID]; hasEntry {
		if m.cacheTTL == 0 || time.Since(entry.createdAt) < m.cacheTTL {
			return entry.handler
		}
	}

	// Try static panel map first
	if p, ok := m.panels[tenantID]; ok {
		h = p.Router()
		m.handlers[tenantID] = h
		return h
	}

	// Fall back to factory
	if m.factory == nil {
		return nil
	}
	tenant := &Tenant{ID: tenantID}
	p, err := m.factory(context.Background(), tenant)
	if err != nil || p == nil {
		return nil
	}
	h = p.Router()
	m.cachedEntries[tenantID] = &cachedPanelEntry{handler: h, createdAt: time.Now()}
	return h
}

// ---------------------------------------------------------------------------
// TenantAware — optional interface for resources that need tenant isolation
// ---------------------------------------------------------------------------

// TenantAware is an optional interface for resources that scope their data to the current tenant.
type TenantAware interface {
	SetTenant(tenant *Tenant)
}

// TenantResourceMiddleware wraps a Panel's handler to call SetTenant on all TenantAware resources.
func TenantResourceMiddleware(p *Panel) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tenant := TenantFromContext(r.Context())
			if tenant != nil {
				for _, res := range p.Resources {
					if ta, ok := res.(TenantAware); ok {
						ta.SetTenant(tenant)
					}
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

// ---------------------------------------------------------------------------
// DynamicSubdomainResolver — resolves tenants from a DB/function with TTL cache
// ---------------------------------------------------------------------------

// cachedTenantEntry holds a tenant with its cache timestamp.
type cachedTenantEntry struct {
	tenant *Tenant
	at     time.Time
}

// DynamicSubdomainResolver resolves tenants by calling a user-supplied function
// (e.g. a DB query) and caches results for cacheTTL to avoid per-request DB hits.
type DynamicSubdomainResolver struct {
	baseDomain string
	resolveFn  func(ctx context.Context, slug string) (*Tenant, error)
	cacheTTL   time.Duration
	cache      sync.Map // slug -> *cachedTenantEntry
}

// NewDynamicSubdomainResolver creates a resolver that calls resolveFn to look up
// tenants by subdomain slug, caching results for cacheTTL.
//
//	resolver := engine.NewDynamicSubdomainResolver("example.com",
//	    func(ctx context.Context, slug string) (*engine.Tenant, error) {
//	        return db.FindTenantBySlug(ctx, slug)
//	    },
//	    5*time.Minute,
//	)
func NewDynamicSubdomainResolver(
	baseDomain string,
	resolveFn func(ctx context.Context, slug string) (*Tenant, error),
	cacheTTL time.Duration,
) *DynamicSubdomainResolver {
	if cacheTTL <= 0 {
		cacheTTL = 5 * time.Minute
	}
	return &DynamicSubdomainResolver{
		baseDomain: baseDomain,
		resolveFn:  resolveFn,
		cacheTTL:   cacheTTL,
	}
}

// Resolve extracts the subdomain from the request host and resolves the tenant.
func (r *DynamicSubdomainResolver) Resolve(req *http.Request) (*Tenant, bool) {
	host := req.Host
	if idx := strings.LastIndex(host, ":"); idx != -1 {
		host = host[:idx]
	}

	var slug string
	if strings.HasSuffix(host, "."+r.baseDomain) {
		slug = strings.TrimSuffix(host, "."+r.baseDomain)
	} else {
		slug = host
	}
	if slug == "" {
		return nil, false
	}

	if cached, ok := r.cache.Load(slug); ok {
		entry := cached.(*cachedTenantEntry)
		if time.Since(entry.at) < r.cacheTTL {
			if entry.tenant == nil {
				return nil, false
			}
			return entry.tenant, true
		}
	}

	tenant, err := r.resolveFn(req.Context(), slug)
	if err != nil {
		r.cache.Store(slug, &cachedTenantEntry{tenant: nil, at: time.Now()})
		return nil, false
	}
	r.cache.Store(slug, &cachedTenantEntry{tenant: tenant, at: time.Now()})
	return tenant, true
}

// InvalidateCache removes a specific slug from the cache (e.g. after tenant update).
func (r *DynamicSubdomainResolver) InvalidateCache(slug string) {
	r.cache.Delete(slug)
}

// InvalidateAll clears the entire tenant cache.
func (r *DynamicSubdomainResolver) InvalidateAll() {
	r.cache.Range(func(key, _ any) bool {
		r.cache.Delete(key)
		return true
	})
}

// ---------------------------------------------------------------------------
// NewIsolatedSession — session helper for multi-tenant isolation
// ---------------------------------------------------------------------------

// NewIsolatedSession creates a session manager with a tenant-specific cookie name
// to prevent session collisions between tenants sharing the same domain.
//
//	session := engine.NewIsolatedSession("acme", mysqlstore.New(tenantDB, time.Minute))
func NewIsolatedSession(slug string, store scs.Store) *scs.SessionManager {
	s := scs.New()
	s.Cookie.Name = "sa_" + slug
	s.Cookie.Secure = true
	s.Cookie.HttpOnly = true
	s.Cookie.SameSite = http.SameSiteLaxMode
	if store != nil {
		s.Store = store
	}
	return s
}

// ---------------------------------------------------------------------------
// MultiPanelRouter — WithPanelFactory + WithPanelCache + validateSessionIsolation
// ---------------------------------------------------------------------------

// PanelFactory is a function that builds a Panel for a given Tenant on demand.
type PanelFactory func(ctx context.Context, t *Tenant) (*Panel, error)

// cachedPanelEntry holds a built panel handler with its creation time.
type cachedPanelEntry struct {
	handler   http.Handler
	createdAt time.Time
}

// WithPanelFactory sets a factory function used to build panels on demand.
// When set, RegisterPanel is no longer required — panels are built lazily and
// cached for cacheTTL (set via WithPanelCache, default 10 minutes).
func (m *MultiPanelRouter) WithPanelFactory(factory PanelFactory) *MultiPanelRouter {
	m.factory = factory
	return m
}

// WithPanelCache sets the TTL for cached panel handlers built by the factory.
// Default is 10 minutes. Pass 0 to disable caching (rebuild on every request).
func (m *MultiPanelRouter) WithPanelCache(ttl time.Duration) *MultiPanelRouter {
	m.cacheTTL = ttl
	return m
}

// validateSessionIsolation panics at startup if two registered panels share
// the same session cookie name — which would cause cross-tenant session leaks.
func (m *MultiPanelRouter) validateSessionIsolation() {
	m.mu.RLock()
	defer m.mu.RUnlock()
	seen := make(map[string]string) // cookieName -> tenantID
	for tenantID, p := range m.panels {
		if p.Session == nil {
			continue
		}
		name := p.Session.Cookie.Name
		if name == "" {
			name = "session"
		}
		if existing, conflict := seen[name]; conflict {
			panic(fmt.Sprintf(
				"engine: session cookie collision — tenants %q and %q both use cookie %q; "+
					"use engine.NewIsolatedSession(slug, store) to give each tenant a unique cookie",
				existing, tenantID, name,
			))
		}
		seen[name] = tenantID
	}
}
