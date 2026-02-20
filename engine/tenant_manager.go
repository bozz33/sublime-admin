package engine

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
)

// TenantManager handles automatic tenant database creation and management.
type TenantManager struct {
	dbDSN    string // Master database DSN (for tenant registry)
	dbDriver string
	logger   *slog.Logger
}

// NewTenantManager creates a new tenant manager.
func NewTenantManager(masterDSN, driver string, logger *slog.Logger) *TenantManager {
	if logger == nil {
		logger = slog.Default()
	}
	return &TenantManager{
		dbDSN:    masterDSN,
		dbDriver: driver,
		logger:   logger,
	}
}

// InitializeTenantRegistry creates the tenant registry tables if they don't exist.
func (tm *TenantManager) InitializeTenantRegistry(ctx context.Context) error {
	db, err := sql.Open(tm.dbDriver, tm.dbDSN)
	if err != nil {
		return fmt.Errorf("failed to open master database: %w", err)
	}
	defer db.Close()

	// Create tenants table
	_, err = db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS tenants (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			domain TEXT,
			subdomain TEXT NOT NULL,
			database_dsn TEXT NOT NULL,
			status TEXT DEFAULT 'active',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create tenants table: %w", err)
	}

	// Create tenant_meta table for additional metadata
	_, err = db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS tenant_meta (
			tenant_id TEXT,
			key TEXT,
			value TEXT,
			PRIMARY KEY (tenant_id, key),
			FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create tenant_meta table: %w", err)
	}

	tm.logger.Info("tenant registry initialized")
	return nil
}

// CreateTenant creates a new tenant with automatic database provisioning.
func (tm *TenantManager) CreateTenant(ctx context.Context, tenant *Tenant) error {
	// Generate database DSN if not provided
	if tenant.Meta == nil {
		tenant.Meta = make(map[string]any)
	}
	if dsn, ok := tenant.Meta["database_dsn"].(string); !ok || dsn == "" {
		// Auto-generate SQLite database path
		dbPath := fmt.Sprintf("./data/tenants/%s.db", tenant.ID)
		tenant.Meta["database_dsn"] = dbPath
	}

	db, err := sql.Open(tm.dbDriver, tm.dbDSN)
	if err != nil {
		return fmt.Errorf("failed to open master database: %w", err)
	}
	defer db.Close()

	// Create tenant database
	if err := tm.createTenantDatabase(ctx, tenant.Meta["database_dsn"].(string)); err != nil {
		return fmt.Errorf("failed to create tenant database: %w", err)
	}

	// Insert tenant record
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Insert tenant
	_, err = tx.ExecContext(ctx, `
		INSERT INTO tenants (id, name, domain, subdomain, database_dsn, status)
		VALUES (?, ?, ?, ?, ?, 'active')
	`, tenant.ID, tenant.Name, tenant.Domain, tenant.Subdomain, tenant.Meta["database_dsn"])
	if err != nil {
		return fmt.Errorf("failed to insert tenant: %w", err)
	}

	// Insert metadata
	for key, value := range tenant.Meta {
		if key == "database_dsn" {
			continue // Already stored in dedicated column
		}
		_, err = tx.ExecContext(ctx, `
			INSERT INTO tenant_meta (tenant_id, key, value) VALUES (?, ?, ?)
		`, tenant.ID, key, fmt.Sprintf("%v", value))
		if err != nil {
			return fmt.Errorf("failed to insert tenant meta %s: %w", key, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit tenant creation: %w", err)
	}

	tm.logger.Info("tenant created", "tenant_id", tenant.ID, "database", tenant.Meta["database_dsn"])
	return nil
}

// createTenantDatabase creates and initializes a tenant's database.
func (tm *TenantManager) createTenantDatabase(ctx context.Context, dsn string) error {
	// Create tenant database
	db, err := sql.Open(tm.dbDriver, dsn)
	if err != nil {
		return fmt.Errorf("failed to open tenant database: %w", err)
	}
	defer db.Close()

	// Run basic schema initialization
	schema := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY,
			applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		
		INSERT OR IGNORE INTO schema_migrations (version) VALUES ('initial');
	`

	_, err = db.ExecContext(ctx, schema)
	if err != nil {
		return fmt.Errorf("failed to initialize tenant schema: %w", err)
	}

	return nil
}

// GetTenant retrieves a tenant from the registry.
func (tm *TenantManager) GetTenant(ctx context.Context, tenantID string) (*Tenant, error) {
	db, err := sql.Open(tm.dbDriver, tm.dbDSN)
	if err != nil {
		return nil, fmt.Errorf("failed to open master database: %w", err)
	}
	defer db.Close()

	var tenant Tenant
	var databaseDSN string

	err = db.QueryRowContext(ctx, `
		SELECT id, name, domain, subdomain, database_dsn FROM tenants WHERE id = ? AND status = 'active'
	`, tenantID).Scan(&tenant.ID, &tenant.Name, &tenant.Domain, &tenant.Subdomain, &databaseDSN)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("tenant not found: %s", tenantID)
		}
		return nil, fmt.Errorf("failed to query tenant: %w", err)
	}

	// Load metadata
	tenant.Meta = make(map[string]any)
	tenant.Meta["database_dsn"] = databaseDSN

	rows, err := db.QueryContext(ctx, `
		SELECT key, value FROM tenant_meta WHERE tenant_id = ?
	`, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to query tenant meta: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return nil, fmt.Errorf("failed to scan tenant meta: %w", err)
		}
		tenant.Meta[key] = value
	}

	return &tenant, nil
}

// ListTenants returns all active tenants.
func (tm *TenantManager) ListTenants(ctx context.Context) ([]*Tenant, error) {
	db, err := sql.Open(tm.dbDriver, tm.dbDSN)
	if err != nil {
		return nil, fmt.Errorf("failed to open master database: %w", err)
	}
	defer db.Close()

	rows, err := db.QueryContext(ctx, `
		SELECT id, name, domain, subdomain, database_dsn FROM tenants WHERE status = 'active' ORDER BY created_at
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query tenants: %w", err)
	}
	defer rows.Close()

	var tenants []*Tenant
	for rows.Next() {
		var tenant Tenant
		var databaseDSN string

		if err := rows.Scan(&tenant.ID, &tenant.Name, &tenant.Domain, &tenant.Subdomain, &databaseDSN); err != nil {
			return nil, fmt.Errorf("failed to scan tenant: %w", err)
		}

		tenant.Meta = make(map[string]any)
		tenant.Meta["database_dsn"] = databaseDSN
		tenants = append(tenants, &tenant)
	}

	return tenants, nil
}

// DeleteTenant soft-deletes a tenant (keeps database for backup).
func (tm *TenantManager) DeleteTenant(ctx context.Context, tenantID string) error {
	db, err := sql.Open(tm.dbDriver, tm.dbDSN)
	if err != nil {
		return fmt.Errorf("failed to open master database: %w", err)
	}
	defer db.Close()

	_, err = db.ExecContext(ctx, `
		UPDATE tenants SET status = 'deleted', updated_at = CURRENT_TIMESTAMP WHERE id = ?
	`, tenantID)
	if err != nil {
		return fmt.Errorf("failed to delete tenant: %w", err)
	}

	tm.logger.Info("tenant deleted", "tenant_id", tenantID)
	return nil
}

// TenantDatabaseResolver resolves tenants from the database registry.
type TenantDatabaseResolver struct {
	manager    *TenantManager
	baseDomain string
}

// NewTenantDatabaseResolver creates a resolver that uses the tenant database.
func NewTenantDatabaseResolver(manager *TenantManager, baseDomain string) *TenantDatabaseResolver {
	return &TenantDatabaseResolver{
		manager:    manager,
		baseDomain: baseDomain,
	}
}

// Resolve extracts the subdomain and resolves the tenant from the database.
func (r *TenantDatabaseResolver) Resolve(req *http.Request) (*Tenant, bool) {
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

	tenant, err := r.manager.GetTenant(req.Context(), slug)
	if err != nil {
		return nil, false
	}

	return tenant, true
}
