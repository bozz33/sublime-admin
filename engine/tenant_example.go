package engine

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
)

// ExampleTenantSetup shows how to set up automatic tenant management.
func ExampleTenantSetup() {
	// 1. Create tenant manager with new TenantManagerConfig API
	manager, err := NewTenantManager(TenantManagerConfig{
		MasterDSN:     "./data/master.db",
		Driver:        "sqlite3",
		DatabaseStyle: DatabaseStylePerTenant,
		ConnStrGen:    NewConnStrGenerator("./data/tenants/%s.db"),
		Logger:        slog.Default(),
		// Optional: run migrations on each new tenant DB
		MigrationHook: func(ctx context.Context, db *sql.DB, tenant TenantInfo) error {
			_, err := db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS users (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				email TEXT NOT NULL UNIQUE,
				name TEXT NOT NULL,
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP
			)`)
			return err
		},
	})
	if err != nil {
		panic(err)
	}

	// 2. Initialize tenant registry (creates tables if they don't exist)
	ctx := context.Background()
	if err := manager.InitializeTenantRegistry(ctx); err != nil {
		panic(err)
	}

	// 3. Create a new tenant — database provisioned automatically
	cfg := &TenantConfig{
		ID:        "acme",
		Name:      "Acme Corporation",
		Domain:    "acme.example.com",
		Subdomain: "acme",
		Meta:      map[string]string{"plan": "premium"},
	}

	if err := manager.CreateTenant(ctx, cfg); err != nil {
		panic(err)
	}

	// 4. Use tenant database resolver (with built-in LRU cache)
	resolver := NewTenantDatabaseResolver(manager, "example.com")

	// 5. Set up multi-panel router
	router := NewMultiPanelRouter(resolver)

	// 5b. For each tenant, create a panel with tenant-specific config
	// In practice, panels are created per-request by the router's resolve logic.
	_ = router
	_ = ctx

	// Example: create a panel for a specific tenant
	panel := NewPanel(fmt.Sprintf("%s Admin", cfg.Name)).
		WithPath("/admin")

	_ = panel

	// Now you can use router as your HTTP handler
	// http.ListenAndServe(":8080", router)
}

// ExamplePaginatedResource shows how to implement server-side pagination.
type ExampleProductResource struct {
	*SimplePaginatedResource
	db *sql.DB
}

func NewExampleProductResource(db *sql.DB) *ExampleProductResource {
	r := &ExampleProductResource{
		SimplePaginatedResource: &SimplePaginatedResource{
			SimpleResource: NewSimpleResource("products", "Product", "Products"),
		},
		db: db,
	}
	r.WithPaginatedList(func(ctx context.Context, params PaginationParams) (*PaginatedResult, error) {
		// Build query with pagination
		query := "SELECT id, name, price, created_at FROM products"
		var args []any
		var conditions []string

		// Add search filter
		if params.Search != "" {
			conditions = append(conditions, "name LIKE ?")
			args = append(args, "%"+params.Search+"%")
		}

		if len(conditions) > 0 {
			query += " WHERE " + fmt.Sprintf("%s", conditions[0])
		}

		// Add sorting
		if params.Sort != "" {
			query += fmt.Sprintf(" ORDER BY %s %s", params.Sort, params.Order)
		} else {
			query += " ORDER BY created_at DESC"
		}

		// Get total count
		countQuery := "SELECT COUNT(*) FROM products"
		if len(conditions) > 0 {
			countQuery += " WHERE " + fmt.Sprintf("%s", conditions[0])
		}

		var total int
		if err := db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
			return nil, err
		}

		// Add pagination
		offset := (params.Page - 1) * params.PerPage
		query += " LIMIT ? OFFSET ?"
		args = append(args, params.PerPage, offset)

		// Execute query
		rows, err := db.QueryContext(ctx, query, args...)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		var items []any
		for rows.Next() {
			var product struct {
				ID        int
				Name      string
				Price     float64
				CreatedAt string
			}
			if err := rows.Scan(&product.ID, &product.Name, &product.Price, &product.CreatedAt); err != nil {
				return nil, err
			}
			items = append(items, &product)
		}

		return NewPaginatedResult(items, total, params.Page, params.PerPage), nil
	})
	return r
}

// UserResource is an example of a tenant-aware resource.
type UserResource struct {
	*SimplePaginatedResource
	tenant *Tenant
	db     *sql.DB
}

func NewUserResource(db *sql.DB) *UserResource {
	r := &UserResource{
		SimplePaginatedResource: NewSimplePaginatedResource("users", "User", "Users"),
		db:                      db,
	}
	r.WithPaginatedList(func(ctx context.Context, params PaginationParams) (*PaginatedResult, error) {
		// Tenant-aware query - automatically filters by tenant if needed
		query := "SELECT id, email, name, created_at FROM users"
		var args []any

		// Add search
		if params.Search != "" {
			query += " WHERE email LIKE ? OR name LIKE ?"
			args = append(args, "%"+params.Search+"%", "%"+params.Search+"%")
		}

		// Add sorting and pagination
		query += " ORDER BY created_at DESC LIMIT ? OFFSET ?"
		args = append(args, params.PerPage, (params.Page-1)*params.PerPage)

		// Execute with tenant context if needed
		rows, err := db.QueryContext(ctx, query, args...)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		var items []any
		for rows.Next() {
			var user struct {
				ID        int
				Email     string
				Name      string
				CreatedAt string
			}
			if err := rows.Scan(&user.ID, &user.Email, &user.Name, &user.CreatedAt); err != nil {
				return nil, err
			}
			items = append(items, &user)
		}

		// Get total count
		countQuery := "SELECT COUNT(*) FROM users"
		if params.Search != "" {
			countQuery += " WHERE email LIKE ? OR name LIKE ?"
		}
		var total int
		if err := db.QueryRowContext(ctx, countQuery, args[:len(args)-2]...).Scan(&total); err != nil {
			return nil, err
		}

		return NewPaginatedResult(items, total, params.Page, params.PerPage), nil
	})
	return r
}

// SetTenant implements TenantAware interface for automatic tenant injection.
func (r *UserResource) SetTenant(tenant *Tenant) {
	r.tenant = tenant
	// You can use tenant info in queries, e.g.:
	// - Add tenant_id filter to all queries
	// - Use tenant-specific database
	// - Apply tenant-specific business rules
}
