package engine

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// TenantInfo is the interface all tenant types must implement.
type TenantInfo interface {
	GetId() string
	GetName() string
	GetDomain() string
	GetSubdomain() string
	GetDatabaseDSN() string
	GetMeta(key string) (string, bool)
}

// TenantConfig holds full tenant configuration.
// Mirrors go-saas's TenantConfig.
type TenantConfig struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Domain      string            `json:"domain"`
	Subdomain   string            `json:"subdomain"`
	DatabaseDSN string            `json:"database_dsn"`
	Status      string            `json:"status"`
	Meta        map[string]string `json:"meta"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

func (tc *TenantConfig) GetId() string        { return tc.ID }
func (tc *TenantConfig) GetName() string      { return tc.Name }
func (tc *TenantConfig) GetDomain() string    { return tc.Domain }
func (tc *TenantConfig) GetSubdomain() string { return tc.Subdomain }
func (tc *TenantConfig) GetDatabaseDSN() string {
	if tc.DatabaseDSN != "" {
		return tc.DatabaseDSN
	}
	return fmt.Sprintf("./data/tenants/%s.db", tc.ID)
}
func (tc *TenantConfig) GetMeta(key string) (string, bool) {
	if tc.Meta == nil {
		return "", false
	}
	v, ok := tc.Meta[key]
	return v, ok
}

// ToTenant converts TenantConfig to the existing Tenant struct.
func (tc *TenantConfig) ToTenant() *Tenant {
	meta := make(map[string]any)
	meta["database_dsn"] = tc.GetDatabaseDSN()
	for k, v := range tc.Meta {
		meta[k] = v
	}
	return &Tenant{
		ID:        tc.ID,
		Name:      tc.Name,
		Domain:    tc.Domain,
		Subdomain: tc.Subdomain,
		Meta:      meta,
	}
}

// TenantStore is the interface for tenant persistence.
type TenantStore interface {
	GetByNameOrId(ctx context.Context, nameOrId string) (*TenantConfig, error)
	List(ctx context.Context) ([]*TenantConfig, error)
	Create(ctx context.Context, cfg *TenantConfig) error
	Update(ctx context.Context, cfg *TenantConfig) error
	Delete(ctx context.Context, id string) error
}

// MemoryTenantStore is an in-memory TenantStore.
type MemoryTenantStore struct {
	mu      sync.RWMutex
	tenants map[string]*TenantConfig // keyed by id, domain, subdomain
}

// NewMemoryTenantStore creates an in-memory store pre-populated with configs.
func NewMemoryTenantStore(configs []*TenantConfig) *MemoryTenantStore {
	m := &MemoryTenantStore{tenants: make(map[string]*TenantConfig)}
	for _, c := range configs {
		m.index(c)
	}
	return m
}

func (m *MemoryTenantStore) index(c *TenantConfig) {
	m.tenants[c.ID] = c
	if c.Domain != "" {
		m.tenants[c.Domain] = c
	}
	if c.Subdomain != "" {
		m.tenants[c.Subdomain] = c
	}
}

func (m *MemoryTenantStore) GetByNameOrId(_ context.Context, nameOrId string) (*TenantConfig, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if t, ok := m.tenants[nameOrId]; ok {
		return t, nil
	}
	return nil, fmt.Errorf("tenant not found: %s", nameOrId)
}

func (m *MemoryTenantStore) List(_ context.Context) ([]*TenantConfig, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	seen := make(map[string]bool)
	var result []*TenantConfig
	for _, t := range m.tenants {
		if !seen[t.ID] {
			seen[t.ID] = true
			result = append(result, t)
		}
	}
	return result, nil
}

func (m *MemoryTenantStore) Create(_ context.Context, cfg *TenantConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.index(cfg)
	return nil
}

func (m *MemoryTenantStore) Update(ctx context.Context, cfg *TenantConfig) error {
	return m.Create(ctx, cfg)
}

func (m *MemoryTenantStore) Delete(_ context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if t, ok := m.tenants[id]; ok {
		delete(m.tenants, t.ID)
		delete(m.tenants, t.Domain)
		delete(m.tenants, t.Subdomain)
	}
	return nil
}

// ---------------------------------------------------------------------------
// SQLTenantStore — SQL-backed TenantStore
// ---------------------------------------------------------------------------

// SQLTenantStore persists tenants in a SQL database.
type SQLTenantStore struct {
	db *sql.DB
}

// NewSQLTenantStore creates a SQL-backed tenant store.
func NewSQLTenantStore(db *sql.DB) *SQLTenantStore {
	return &SQLTenantStore{db: db}
}

func (s *SQLTenantStore) GetByNameOrId(ctx context.Context, nameOrId string) (*TenantConfig, error) {
	var cfg TenantConfig
	var metaJSON string
	err := s.db.QueryRowContext(ctx, `
		SELECT id, name, domain, subdomain, database_dsn, status, meta, created_at, updated_at
		FROM tenants WHERE (id = ? OR domain = ? OR subdomain = ?) AND status = 'active'
	`, nameOrId, nameOrId, nameOrId).Scan(
		&cfg.ID, &cfg.Name, &cfg.Domain, &cfg.Subdomain,
		&cfg.DatabaseDSN, &cfg.Status, &metaJSON,
		&cfg.CreatedAt, &cfg.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("tenant not found: %s", nameOrId)
	}
	if err != nil {
		return nil, fmt.Errorf("query tenant: %w", err)
	}
	if metaJSON != "" {
		_ = json.Unmarshal([]byte(metaJSON), &cfg.Meta)
	}
	return &cfg, nil
}

func (s *SQLTenantStore) List(ctx context.Context) ([]*TenantConfig, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, name, domain, subdomain, database_dsn, status, meta, created_at, updated_at
		FROM tenants WHERE status = 'active' ORDER BY created_at
	`)
	if err != nil {
		return nil, fmt.Errorf("list tenants: %w", err)
	}
	defer rows.Close()

	var result []*TenantConfig
	for rows.Next() {
		var cfg TenantConfig
		var metaJSON string
		if err := rows.Scan(&cfg.ID, &cfg.Name, &cfg.Domain, &cfg.Subdomain,
			&cfg.DatabaseDSN, &cfg.Status, &metaJSON, &cfg.CreatedAt, &cfg.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan tenant: %w", err)
		}
		if metaJSON != "" {
			_ = json.Unmarshal([]byte(metaJSON), &cfg.Meta)
		}
		result = append(result, &cfg)
	}
	return result, nil
}

func (s *SQLTenantStore) Create(ctx context.Context, cfg *TenantConfig) error {
	metaJSON, _ := json.Marshal(cfg.Meta)
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO tenants (id, name, domain, subdomain, database_dsn, status, meta)
		VALUES (?, ?, ?, ?, ?, 'active', ?)
	`, cfg.ID, cfg.Name, cfg.Domain, cfg.Subdomain, cfg.DatabaseDSN, string(metaJSON))
	if err != nil {
		return fmt.Errorf("insert tenant: %w", err)
	}
	return nil
}

func (s *SQLTenantStore) Update(ctx context.Context, cfg *TenantConfig) error {
	metaJSON, _ := json.Marshal(cfg.Meta)
	_, err := s.db.ExecContext(ctx, `
		UPDATE tenants SET name=?, domain=?, subdomain=?, database_dsn=?, meta=?, updated_at=CURRENT_TIMESTAMP
		WHERE id=?
	`, cfg.Name, cfg.Domain, cfg.Subdomain, cfg.DatabaseDSN, string(metaJSON), cfg.ID)
	if err != nil {
		return fmt.Errorf("update tenant: %w", err)
	}
	return nil
}

func (s *SQLTenantStore) Delete(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE tenants SET status='deleted', updated_at=CURRENT_TIMESTAMP WHERE id=?
	`, id)
	if err != nil {
		return fmt.Errorf("delete tenant: %w", err)
	}
	return nil
}
