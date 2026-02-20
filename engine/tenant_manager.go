package engine

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"
)

// ---------------------------------------------------------------------------
// DatabaseStyleType — (inspiré de go-saas DatabaseStyleType)
// ---------------------------------------------------------------------------

// DatabaseStyleType controls how tenant data is isolated.
type DatabaseStyleType int32

const (
	// DatabaseStyleSingle — all tenants share one database (use tenant_id column).
	DatabaseStyleSingle DatabaseStyleType = 1 << 0
	// DatabaseStylePerTenant — each tenant has its own database.
	DatabaseStylePerTenant DatabaseStyleType = 1 << 1
	// DatabaseStyleMulti — hybrid: some tenants share, some are isolated.
	DatabaseStyleMulti DatabaseStyleType = 1 << 2
)

// ---------------------------------------------------------------------------
// ConnStrGenerator — (inspiré de go-saas ConnStrGenerator)
// ---------------------------------------------------------------------------

// ConnStrGenerator generates a database connection string for a tenant.
type ConnStrGenerator interface {
	Gen(ctx context.Context, tenant TenantInfo) (string, error)
}

// DefaultConnStrGenerator generates DSNs using a format string.
// Example: NewConnStrGenerator("./data/tenants/%s.db")
type DefaultConnStrGenerator struct {
	format string
}

func NewConnStrGenerator(format string) *DefaultConnStrGenerator {
	return &DefaultConnStrGenerator{format: format}
}

func (d *DefaultConnStrGenerator) Gen(_ context.Context, tenant TenantInfo) (string, error) {
	return fmt.Sprintf(d.format, tenant.GetId()), nil
}

// ---------------------------------------------------------------------------
// MigrationHook — hook for running migrations on tenant DB creation
// ---------------------------------------------------------------------------

// MigrationHook is called after a tenant database is created or upgraded.
// Use it to run schema migrations or seed data.
type MigrationHook func(ctx context.Context, db *sql.DB, tenant TenantInfo) error

// ---------------------------------------------------------------------------
// TenantManagerConfig — configuration
// ---------------------------------------------------------------------------

// TenantManagerConfig holds configuration for TenantManager.
type TenantManagerConfig struct {
	MasterDSN     string
	Driver        string
	DatabaseStyle DatabaseStyleType
	ConnStrGen    ConnStrGenerator
	MigrationHook MigrationHook
	Logger        *slog.Logger
	// Connection pool settings (official Go database/sql patterns)
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

// ---------------------------------------------------------------------------
// TenantManager — gestionnaire principal
// ---------------------------------------------------------------------------

// TenantManager handles automatic tenant database creation and management.
// Inspired by go-saas's multi-tenancy patterns.
type TenantManager struct {
	cfg      TenantManagerConfig
	store    TenantStore
	masterDB *sql.DB
	mu       sync.Mutex
	logger   *slog.Logger
}

// NewTenantManager creates a TenantManager with a SQL master database.
func NewTenantManager(cfg TenantManagerConfig) (*TenantManager, error) {
	cfg = applyTenantManagerDefaults(cfg)

	masterDB, err := sql.Open(cfg.Driver, cfg.MasterDSN)
	if err != nil {
		return nil, fmt.Errorf("open master database: %w", err)
	}
	// Official Go connection pool settings
	masterDB.SetMaxOpenConns(cfg.MaxOpenConns)
	masterDB.SetMaxIdleConns(cfg.MaxIdleConns)
	masterDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	masterDB.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

	return &TenantManager{
		cfg:      cfg,
		masterDB: masterDB,
		logger:   cfg.Logger,
	}, nil
}

// NewTenantManagerWithStore creates a TenantManager with a custom TenantStore.
// Use this with MemoryTenantStore for testing or simple setups.
func NewTenantManagerWithStore(store TenantStore, cfg TenantManagerConfig) *TenantManager {
	cfg = applyTenantManagerDefaults(cfg)
	return &TenantManager{
		cfg:    cfg,
		store:  store,
		logger: cfg.Logger,
	}
}

func applyTenantManagerDefaults(cfg TenantManagerConfig) TenantManagerConfig {
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}
	if cfg.ConnStrGen == nil {
		cfg.ConnStrGen = NewConnStrGenerator("./data/tenants/%s.db")
	}
	if cfg.DatabaseStyle == 0 {
		cfg.DatabaseStyle = DatabaseStylePerTenant
	}
	if cfg.MaxOpenConns == 0 {
		cfg.MaxOpenConns = 25
	}
	if cfg.MaxIdleConns == 0 {
		cfg.MaxIdleConns = 5
	}
	if cfg.ConnMaxLifetime == 0 {
		cfg.ConnMaxLifetime = 5 * time.Minute
	}
	if cfg.ConnMaxIdleTime == 0 {
		cfg.ConnMaxIdleTime = 2 * time.Minute
	}
	return cfg
}

// InitializeTenantRegistry creates the tenant registry tables if they don't exist.
func (tm *TenantManager) InitializeTenantRegistry(ctx context.Context) error {
	if tm.masterDB == nil {
		return fmt.Errorf("no master database configured; use NewTenantManager")
	}
	_, err := tm.masterDB.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS tenants (
			id          TEXT PRIMARY KEY,
			name        TEXT NOT NULL,
			domain      TEXT DEFAULT '',
			subdomain   TEXT NOT NULL DEFAULT '',
			database_dsn TEXT NOT NULL,
			status      TEXT DEFAULT 'active',
			meta        TEXT DEFAULT '{}',
			created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at  DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("create tenants table: %w", err)
	}
	tm.store = NewSQLTenantStore(tm.masterDB)
	tm.logger.Info("tenant registry initialized")
	return nil
}

// CreateTenant creates a new tenant with automatic database provisioning.
func (tm *TenantManager) CreateTenant(ctx context.Context, cfg *TenantConfig) error {
	if cfg.Meta == nil {
		cfg.Meta = make(map[string]string)
	}

	// Auto-generate DSN via ConnStrGenerator (go-saas pattern)
	if cfg.DatabaseDSN == "" {
		dsn, err := tm.cfg.ConnStrGen.Gen(ctx, cfg)
		if err != nil {
			return fmt.Errorf("generate DSN: %w", err)
		}
		cfg.DatabaseDSN = dsn
	}

	// Provision per-tenant database
	if tm.cfg.DatabaseStyle == DatabaseStylePerTenant || tm.cfg.DatabaseStyle == DatabaseStyleMulti {
		if err := tm.provisionTenantDatabase(ctx, cfg); err != nil {
			return fmt.Errorf("provision tenant database: %w", err)
		}
	}

	if err := tm.store.Create(ctx, cfg); err != nil {
		return fmt.Errorf("persist tenant: %w", err)
	}

	tm.logger.Info("tenant created", "id", cfg.ID, "dsn", cfg.DatabaseDSN)
	return nil
}

// provisionTenantDatabase creates and initializes a tenant's database.
func (tm *TenantManager) provisionTenantDatabase(ctx context.Context, cfg *TenantConfig) error {
	db, err := sql.Open(tm.cfg.Driver, cfg.DatabaseDSN)
	if err != nil {
		return fmt.Errorf("open tenant database: %w", err)
	}
	defer db.Close()

	// Apply connection pool settings to tenant DB too
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(3)
	db.SetConnMaxLifetime(tm.cfg.ConnMaxLifetime)
	db.SetConnMaxIdleTime(tm.cfg.ConnMaxIdleTime)

	// Base schema: migration tracking table
	_, err = db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version    TEXT PRIMARY KEY,
			applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	// Run user-provided migration hook (go-saas Seed/Migrate pattern)
	if tm.cfg.MigrationHook != nil {
		if err := tm.cfg.MigrationHook(ctx, db, cfg); err != nil {
			return fmt.Errorf("migration hook: %w", err)
		}
	}

	tm.logger.Info("tenant database provisioned", "id", cfg.ID, "dsn", cfg.DatabaseDSN)
	return nil
}

// GetTenant retrieves a tenant config by ID, domain, or subdomain.
func (tm *TenantManager) GetTenant(ctx context.Context, nameOrId string) (*TenantConfig, error) {
	return tm.store.GetByNameOrId(ctx, nameOrId)
}

// ListTenants returns all active tenants.
func (tm *TenantManager) ListTenants(ctx context.Context) ([]*TenantConfig, error) {
	return tm.store.List(ctx)
}

// UpdateTenant updates an existing tenant's configuration.
func (tm *TenantManager) UpdateTenant(ctx context.Context, cfg *TenantConfig) error {
	if err := tm.store.Update(ctx, cfg); err != nil {
		return fmt.Errorf("update tenant: %w", err)
	}
	tm.logger.Info("tenant updated", "id", cfg.ID)
	return nil
}

// DeleteTenant soft-deletes a tenant (keeps database for backup).
func (tm *TenantManager) DeleteTenant(ctx context.Context, id string) error {
	if err := tm.store.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete tenant: %w", err)
	}
	tm.logger.Info("tenant deleted", "id", id)
	return nil
}

// OpenTenantDB opens a connection to a tenant's database with proper pool settings.
func (tm *TenantManager) OpenTenantDB(ctx context.Context, nameOrId string) (*sql.DB, error) {
	cfg, err := tm.store.GetByNameOrId(ctx, nameOrId)
	if err != nil {
		return nil, err
	}
	db, err := sql.Open(tm.cfg.Driver, cfg.DatabaseDSN)
	if err != nil {
		return nil, fmt.Errorf("open tenant db: %w", err)
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(3)
	db.SetConnMaxLifetime(tm.cfg.ConnMaxLifetime)
	db.SetConnMaxIdleTime(tm.cfg.ConnMaxIdleTime)
	return db, nil
}

// Store returns the underlying TenantStore.
func (tm *TenantManager) Store() TenantStore { return tm.store }

// ---------------------------------------------------------------------------
// TenantResolveContrib — résolveurs chaînés (inspiré de go-saas contrib pattern)
// ---------------------------------------------------------------------------

// TenantResolveContrib is a single resolver in a chain.
// Inspired by go-saas's ContextContrib / contrib pattern.
type TenantResolveContrib interface {
	Name() string
	Resolve(r *http.Request) (string, bool)
}

// ChainedTenantResolver tries multiple resolvers in order.
// Inspired by go-saas's DefaultTenantResolver which chains multiple contribs.
type ChainedTenantResolver struct {
	store   TenantStore
	contribs []TenantResolveContrib
	cache   *tenantLRUCache
}

// NewChainedTenantResolver creates a resolver that tries contribs in order.
func NewChainedTenantResolver(store TenantStore, ttl time.Duration, contribs ...TenantResolveContrib) *ChainedTenantResolver {
	return &ChainedTenantResolver{
		store:    store,
		contribs: contribs,
		cache:    newTenantLRUCache(128, ttl),
	}
}

// Resolve tries each contrib in order, returns first match.
func (r *ChainedTenantResolver) Resolve(req *http.Request) (*Tenant, bool) {
	for _, contrib := range r.contribs {
		if slug, ok := contrib.Resolve(req); ok && slug != "" {
			// Check LRU cache first
			if t, hit := r.cache.get(slug); hit {
				return t, true
			}
			cfg, err := r.store.GetByNameOrId(req.Context(), slug)
			if err != nil {
				continue
			}
			t := cfg.ToTenant()
			r.cache.set(slug, t)
			return t, true
		}
	}
	return nil, false
}

// ---------------------------------------------------------------------------
// Built-in contrib resolvers (inspiré de go-saas)
// ---------------------------------------------------------------------------

// DomainContrib resolves tenant from the full hostname.
type DomainContrib struct{ BaseDomain string }

func NewDomainContrib(baseDomain string) *DomainContrib { return &DomainContrib{BaseDomain: baseDomain} }
func (c *DomainContrib) Name() string                   { return "domain" }
func (c *DomainContrib) Resolve(r *http.Request) (string, bool) {
	host := stripPort(r.Host)
	if c.BaseDomain != "" && strings.HasSuffix(host, "."+c.BaseDomain) {
		return strings.TrimSuffix(host, "."+c.BaseDomain), true
	}
	if host != "" && host != c.BaseDomain {
		return host, true
	}
	return "", false
}

// HeaderContrib resolves tenant from a request header (e.g. X-Tenant-ID).
type HeaderContrib struct{ Header string }

func NewHeaderContrib(header string) *HeaderContrib { return &HeaderContrib{Header: header} }
func (c *HeaderContrib) Name() string               { return "header" }
func (c *HeaderContrib) Resolve(r *http.Request) (string, bool) {
	v := r.Header.Get(c.Header)
	return v, v != ""
}

// QueryContrib resolves tenant from a query parameter (e.g. ?tenant=acme).
type QueryContrib struct{ Param string }

func NewQueryContrib(param string) *QueryContrib { return &QueryContrib{Param: param} }
func (c *QueryContrib) Name() string             { return "query" }
func (c *QueryContrib) Resolve(r *http.Request) (string, bool) {
	v := r.URL.Query().Get(c.Param)
	return v, v != ""
}

// CookieContrib resolves tenant from a cookie.
type CookieContrib struct{ CookieName string }

func NewCookieContrib(name string) *CookieContrib { return &CookieContrib{CookieName: name} }
func (c *CookieContrib) Name() string             { return "cookie" }
func (c *CookieContrib) Resolve(r *http.Request) (string, bool) {
	cookie, err := r.Cookie(c.CookieName)
	if err != nil {
		return "", false
	}
	return cookie.Value, cookie.Value != ""
}

// PathContrib resolves tenant from the first URL path segment (e.g. /acme/admin).
type PathContrib struct{}

func NewPathContrib() *PathContrib              { return &PathContrib{} }
func (c *PathContrib) Name() string             { return "path" }
func (c *PathContrib) Resolve(r *http.Request) (string, bool) {
	path := strings.TrimPrefix(r.URL.Path, "/")
	seg := strings.SplitN(path, "/", 2)[0]
	return seg, seg != ""
}

// ---------------------------------------------------------------------------
// TenantDatabaseResolver — backward-compatible resolver using TenantManager
// ---------------------------------------------------------------------------

// TenantDatabaseResolver resolves tenants from the database registry.
// Kept for backward compatibility; prefer ChainedTenantResolver for new code.
type TenantDatabaseResolver struct {
	manager    *TenantManager
	baseDomain string
	cache      *tenantLRUCache
}

// NewTenantDatabaseResolver creates a resolver backed by TenantManager.
func NewTenantDatabaseResolver(manager *TenantManager, baseDomain string) *TenantDatabaseResolver {
	return &TenantDatabaseResolver{
		manager:    manager,
		baseDomain: baseDomain,
		cache:      newTenantLRUCache(128, 5*time.Minute),
	}
}

func (r *TenantDatabaseResolver) Resolve(req *http.Request) (*Tenant, bool) {
	host := stripPort(req.Host)
	var slug string
	if strings.HasSuffix(host, "."+r.baseDomain) {
		slug = strings.TrimSuffix(host, "."+r.baseDomain)
	} else {
		slug = host
	}
	if slug == "" {
		return nil, false
	}
	if t, hit := r.cache.get(slug); hit {
		return t, true
	}
	cfg, err := r.manager.GetTenant(req.Context(), slug)
	if err != nil {
		return nil, false
	}
	t := cfg.ToTenant()
	r.cache.set(slug, t)
	return t, true
}

// ---------------------------------------------------------------------------
// tenantLRUCache — LRU cache with TTL (inspiré de go-saas Cache[K,V])
// ---------------------------------------------------------------------------

type tenantCacheEntry struct {
	tenant *Tenant
	at     time.Time
}

type tenantLRUCache struct {
	mu      sync.Mutex
	entries map[string]*tenantCacheEntry
	order   []string
	cap     int
	ttl     time.Duration
}

func newTenantLRUCache(capacity int, ttl time.Duration) *tenantLRUCache {
	if capacity <= 0 {
		capacity = 128
	}
	if ttl <= 0 {
		ttl = 5 * time.Minute
	}
	return &tenantLRUCache{
		entries: make(map[string]*tenantCacheEntry),
		cap:     capacity,
		ttl:     ttl,
	}
}

func (c *tenantLRUCache) get(key string) (*Tenant, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	e, ok := c.entries[key]
	if !ok {
		return nil, false
	}
	if time.Since(e.at) > c.ttl {
		delete(c.entries, key)
		return nil, false
	}
	// Move to front (LRU)
	c.moveToFront(key)
	return e.tenant, true
}

func (c *tenantLRUCache) set(key string, t *Tenant) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.entries) >= c.cap {
		// Evict oldest
		if len(c.order) > 0 {
			oldest := c.order[0]
			c.order = c.order[1:]
			delete(c.entries, oldest)
		}
	}
	c.entries[key] = &tenantCacheEntry{tenant: t, at: time.Now()}
	c.order = append(c.order, key)
}

func (c *tenantLRUCache) delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.entries, key)
}

func (c *tenantLRUCache) flush() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries = make(map[string]*tenantCacheEntry)
	c.order = nil
}

func (c *tenantLRUCache) moveToFront(key string) {
	for i, k := range c.order {
		if k == key {
			c.order = append(c.order[:i], c.order[i+1:]...)
			c.order = append(c.order, key)
			return
		}
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func stripPort(host string) string {
	if idx := strings.LastIndex(host, ":"); idx != -1 {
		return host[:idx]
	}
	return host
}
