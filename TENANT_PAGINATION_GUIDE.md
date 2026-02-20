# Tenant Management & Server-Side Pagination Guide

## 🎯 Overview

This guide covers two major improvements to `sublimego-core`:

1. **Automatic Tenant Database Creation** — No more manual database setup
2. **Server-Side Pagination** — Efficient pagination for large datasets

---

## 🏢 Tenant Management

### Problem Solved
Previously, creating a tenant required:
- Manual database creation
- Manual tenant registration
- Risk of configuration errors

### Solution: TenantManager

The `TenantManager` handles everything automatically:

```go
// 1. Create tenant manager
manager := engine.NewTenantManager(
    "./data/master.db", // Master database for tenant registry
    "sqlite3",
    slog.Default(),
)

// 2. Initialize registry (creates tables if needed)
if err := manager.InitializeTenantRegistry(ctx); err != nil {
    panic(err)
}

// 3. Create tenant (automatic database provisioning!)
tenant := &engine.Tenant{
    ID:        "acme",
    Name:      "Acme Corporation",
    Domain:    "acme.example.com", 
    Subdomain: "acme",
    Meta: map[string]any{
        "plan":      "premium",
        "max_users": 100,
    },
}

if err := manager.CreateTenant(ctx, tenant); err != nil {
    panic(err)
}
```

### What Happens Automatically

1. **Database Creation**: `./data/tenants/acme.db` created automatically
2. **Schema Initialization**: Basic tables created (`schema_migrations`, etc.)
3. **Registry Entry**: Tenant recorded in master database
4. **Metadata Storage**: Custom metadata stored for future use

### Database Resolver

Use the tenant database resolver for automatic lookup:

```go
resolver := engine.NewTenantDatabaseResolver(manager, "example.com")
router := engine.NewMultiPanelRouter(resolver).WithPanelFactory(...)
```

---

## 📄 Server-Side Pagination

### Problem Solved
Previously:
- All records loaded into memory
- UI pagination only (inefficient for large datasets)
- No search/sort optimization

### Solution: PaginatedResource

Extend your resources with `PaginatedResource`:

```go
type ProductResource struct {
    *engine.SimplePaginatedResource
    db *sql.DB
}

func NewProductResource(db *sql.DB) *ProductResource {
    r := &ProductResource{
        SimplePaginatedResource: engine.NewSimplePaginatedResource("products", "Product", "Products"),
        db: db,
    }
    
    r.WithPaginatedList(func(ctx context.Context, params engine.PaginationParams) (*engine.PaginatedResult, error) {
        // Build query with pagination
        query := "SELECT id, name, price, created_at FROM products"
        var args []any
        
        // Add search filter
        if params.Search != "" {
            query += " WHERE name LIKE ?"
            args = append(args, "%"+params.Search+"%")
        }
        
        // Add sorting
        if params.Sort != "" {
            query += fmt.Sprintf(" ORDER BY %s %s", params.Sort, params.Order)
        }
        
        // Get total count
        countQuery := "SELECT COUNT(*) FROM products"
        if params.Search != "" {
            countQuery += " WHERE name LIKE ?"
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
        
        return engine.NewPaginatedResult(items, total, params.Page, params.PerPage), nil
    })
    
    return r
}
```

### URL Parameters

The pagination automatically handles these URL parameters:

- `page` - Current page (default: 1)
- `per_page` - Items per page (default: 15, max: 100)
- `search` - Search query
- `sort` - Sort column
- `order` - Sort order ("asc" or "desc")

Example URLs:
```
/products?page=2&per_page=20&search=laptop&sort=price&order=desc
```

### Handler Integration

Use `PaginatedCRUDHandler` instead of regular CRUD handler:

```go
// Instead of:
handler := engine.NewCRUDHandler(resource)

// Use:
handler := engine.NewPaginatedCRUDHandler(resource)
```

---

## 🔄 Complete Example

```go
func main() {
    // 1. Initialize tenant manager
    manager := engine.NewTenantManager("./data/master.db", "sqlite3", slog.Default())
    if err := manager.InitializeTenantRegistry(context.Background()); err != nil {
        panic(err)
    }
    
    // 2. Create tenant resolver
    resolver := engine.NewTenantDatabaseResolver(manager, "example.com")
    
    // 3. Set up multi-panel router with pagination
    router := engine.NewMultiPanelRouter(resolver).WithPanelFactory(func(ctx context.Context, t *engine.Tenant) (*engine.Panel, error) {
        // Open tenant database
        db, err := sql.Open("sqlite3", t.Meta["database_dsn"].(string))
        if err != nil {
            return nil, err
        }
        
        // Create panel
        panel := engine.NewPanel(fmt.Sprintf("%s Admin", t.Name))
        panel.SetSession(engine.NewIsolatedSession(t.ID, nil))
        
        // Add paginated resources
        panel.RegisterResource(NewProductResource(db))
        panel.RegisterResource(NewUserResource(db))
        
        return panel, nil
    })
    
    // 4. Validate session isolation
    router.validateSessionIsolation()
    
    // 5. Start server
    log.Fatal(http.ListenAndServe(":8080", router))
}
```

---

## 🎯 Benefits

### Tenant Management
- ✅ **Zero Manual Setup** — Databases created automatically
- ✅ **Consistent Schema** — Every tenant gets the same base schema
- ✅ **Central Registry** — All tenants tracked in master DB
- ✅ **Metadata Support** — Store custom tenant settings
- ✅ **Soft Delete** — Tenants can be deactivated without data loss

### Server-Side Pagination
- ✅ **Memory Efficient** — Only load needed records
- ✅ **Fast Queries** — Database-level LIMIT/OFFSET
- ✅ **Search Integration** — Full-text search in pagination
- ✅ **Sorting Support** — Client-controlled sorting
- ✅ **URL Preservation** — Pagination state maintained across actions

---

## 🔧 Migration Guide

### From Manual Tenants

**Before:**
```go
// Manual database creation
os.MkdirAll("./data/tenants", 0755)
db, _ := sql.Open("sqlite3", "./data/tenants/acme.db")
db.Exec(`CREATE TABLE users...`)

// Manual tenant registration
resolver := engine.NewSubdomainResolver("example.com")
resolver.Register(&engine.Tenant{
    ID: "acme",
    Subdomain: "acme",
    // ...
})
```

**After:**
```go
// Automatic everything
manager := engine.NewTenantManager("./data/master.db", "sqlite3", nil)
manager.InitializeTenantRegistry(ctx)
manager.CreateTenant(ctx, &engine.Tenant{
    ID: "acme",
    Subdomain: "acme",
    // ...
})
resolver := engine.NewTenantDatabaseResolver(manager, "example.com")
```

### From Client-Side Pagination

**Before:**
```go
func (r *ProductResource) List(ctx context.Context) ([]any, error) {
    // Loads ALL products into memory
    return r.db.QueryContext(ctx, "SELECT * FROM products")
}
```

**After:**
```go
func (r *ProductResource) ListPaginated(ctx context.Context, params engine.PaginationParams) (*engine.PaginatedResult, error) {
    // Loads only needed products
    offset := (params.Page - 1) * params.PerPage
    query := "SELECT * FROM products LIMIT ? OFFSET ?"
    // ... pagination logic
}
```

---

## 📚 API Reference

### TenantManager

- `NewTenantManager(dsn, driver, logger)` — Create manager
- `InitializeTenantRegistry(ctx)` — Create registry tables
- `CreateTenant(ctx, tenant)` — Create tenant with database
- `GetTenant(ctx, id)` — Retrieve tenant
- `ListTenants(ctx)` — List all tenants
- `DeleteTenant(ctx, id)` — Soft-delete tenant

### PaginatedResource

- `NewSimplePaginatedResource(slug, label, plural)` — Create resource
- `WithPaginatedList(fn)` — Set pagination function
- `ParsePaginationParams(r)` — Extract pagination from request

### PaginationParams

- `Page` — Current page (1-based)
- `PerPage` — Items per page (1-100)
- `Search` — Search query
- `Sort` — Sort column
- `Order` — "asc" or "desc"

### PaginatedResult

- `Items` — Current page items
- `Total` — Total items count
- `Page` — Current page number
- `PerPage` — Items per page
- `TotalPages` — Total pages
- `HasNext` — Has next page
- `HasPrev` — Has previous page

---

## 🚀 Next Steps

1. **Update Your Resources** — Convert to `PaginatedResource`
2. **Replace Tenant Setup** — Use `TenantManager` instead of manual registration
3. **Update Handlers** — Use `PaginatedCRUDHandler`
4. **Test Pagination** — Verify URL parameters work correctly
5. **Monitor Performance** — Check memory usage with large datasets

---

## 🐛 Troubleshooting

### Tenant Creation Fails
- Check master database DSN is accessible
- Verify write permissions to `./data/tenants/`
- Check SQLite driver is imported

### Pagination Not Working
- Ensure resource implements `PaginatedResource`
- Use `PaginatedCRUDHandler` instead of `CRUDHandler`
- Check URL parameters are being parsed

### Session Isolation Warning
- Use `NewIsolatedSession()` for each tenant
- Call `validateSessionIsolation()` at startup
- Ensure each tenant has unique session cookie name
