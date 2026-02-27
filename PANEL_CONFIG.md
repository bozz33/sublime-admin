# Panel Configuration

**Full configuration reference for SublimeGo panels.**

This guide covers all available configuration options for customizing your admin panel.

---

## Table of Contents

1. [Basic Configuration](#basic-configuration)
2. [Authentication](#authentication)
3. [Branding](#branding)
4. [Navigation](#navigation)
5. [Middleware](#middleware)
6. [Notifications](#notifications)
7. [Multi-tenancy](#multi-tenancy)
8. [Performance](#performance)
9. [Complete Example](#complete-example)

---

## Basic Configuration

### Creating a Panel

```go
package main

import (
    "github.com/bozz33/sublimeadmin/engine"
    "github.com/bozz33/sublimeadmin/auth"
    "github.com/bozz33/sublimeadmin/middleware"
)

func main() {
    panel := engine.NewPanel("admin")
    
    // Basic configuration
    panel.
        WithPath("/admin").
        WithDatabase(db).
        WithBrandName("My App").
        WithLogo("/logo.png").
        WithFavicon("/favicon.ico")
    
    // Start server
    http.ListenAndServe(":8080", panel.Router())
}
```

### Required Configuration

| Method | Description | Required |
|--------|-------------|----------|
| `WithPath(path)` | Base path for admin panel | Yes |
| `WithDatabase(db)` | Database connection | Yes |
| `WithUsers(repo)` | User repository | Yes |

---

## Authentication

### Basic Auth Setup

```go
panel.
    WithAuthManager(auth.NewManager(db)).
    WithSession(&scs.SessionStore{
        Lifetime: 24 * time.Hour,
        Cookie: scs.Cookie{
            Name:     "admin_session",
            SameSite: http.SameSiteLaxMode,
            Secure:   true,
            HttpOnly: true,
        },
    })
```

### Authentication Features

```go
panel.
    // Registration
    EnableRegistration().
    
    // Email verification
    EnableEmailVerification().
    
    // Password reset
    EnablePasswordReset().
    
    // User profile
    EnableProfile().
    
    // Custom auth manager
    WithAuthManager(&CustomAuthManager{
        db:     db,
        mailer: mailer,
    })
```

### Custom Auth Manager

```go
type CustomAuthManager struct {
    db     *ent.Client
    mailer mailer.Mailer
}

func (m *CustomAuthManager) Authenticate(email, password string) (*auth.User, error) {
    user, err := m.db.User.Query().
        Where(user.Email(email)).
        Only(context.Background())
    if err != nil {
        return nil, err
    }
    
    if !bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)) {
        return nil, auth.ErrInvalidCredentials
    }
    
    return &auth.User{
        ID:    user.ID,
        Email: user.Email,
        Name:  user.Name,
        Role:  user.Role,
    }, nil
}

func (m *CustomAuthManager) SendPasswordReset(email string) error {
    token := generateResetToken()
    
    // Save token
    m.db.User.Update().
        Where(user.Email(email)).
        SetResetToken(token).
        SetResetExpires(time.Now().Add(1 * time.Hour)).
        Exec(context.Background())
    
    // Send email
    return m.mailer.Send(mailer.Message{
        To:      []string{email},
        Subject: "Password Reset",
        Body:    fmt.Sprintf("Reset link: /reset?token=%s", token),
    })
}
```

---

## Branding

### Visual Customization

```go
panel.
    WithBrandName("My Application").
    WithLogo("/assets/logo.png").
    WithFavicon("/assets/favicon.ico").
    WithPrimaryColor("#3B82F6").
    WithCustomColor("accent", "#10B981").
    EnableDarkMode()
```

### Color Configuration

```go
panel.
    // Primary colors
    WithPrimaryColor("#3B82F6").      // Blue
    WithCustomColor("secondary", "#6B7280").  // Gray
    WithCustomColor("success", "#10B981").    // Green
    WithCustomColor("warning", "#F59E0B").    // Yellow
    WithCustomColor("danger", "#EF4444").      // Red
    WithCustomColor("info", "#3B82F6").       // Blue
    
    // Theme colors
    WithCustomColor("sidebar", "#1F2937").    // Dark sidebar
    WithCustomColor("topbar", "#FFFFFF").     // Light topbar
    WithCustomColor("card", "#FFFFFF").       // Card background
```

### Dynamic Color Palettes

```go
import "github.com/bozz33/sublimeadmin/color"

palette := color.NewPalette().
    AddPrimary("#3B82F6").
    AddSecondary("#6B7280").
    AddSuccess("#10B981").
    AddWarning("#F59E0B").
    AddDanger("#EF4444").
    GenerateCSSVars()

panel.WithColorPalette(palette)
```

---

## Navigation

### Manual Navigation Items

```go
panel.
    AddNavItem(engine.NavItem{
        Label: "Dashboard",
        Icon:  "dashboard",
        URL:   "/admin",
    }).
    AddNavItem(engine.NavItem{
        Label: "Users",
        Icon:  "people",
        URL:   "/admin/users",
    }).
    AddNavItem(engine.NavItem{
        Label: "Settings",
        Icon:  "settings",
        URL:   "/admin/settings",
    })
```

### Navigation Groups

```go
panel.
    AddNavGroup(engine.NavigationGroup{
        Label: "Management",
        Icon:  "admin_panel_settings",
        Items: []engine.NavItem{
            {Label: "Users", Icon: "people", URL: "/admin/users"},
            {Label: "Roles", Icon: "admin_panel_settings", URL: "/admin/roles"},
            {Label: "Permissions", Icon: "security", URL: "/admin/permissions"},
        },
    }).
    AddNavGroup(engine.NavigationGroup{
        Label: "Content",
        Icon:  "article",
        Items: []engine.NavItem{
            {Label: "Posts", Icon: "description", URL: "/admin/posts"},
            {Label: "Pages", Icon: "article", URL: "/admin/pages"},
            {Label: "Media", Icon: "image", URL: "/admin/media"},
        },
    })
```

### Dynamic Navigation

```go
panel.WithNavigationHook(func(ctx context.Context) []engine.NavItem {
    user := auth.UserFromContext(ctx)
    items := []engine.NavItem{
        {Label: "Dashboard", Icon: "dashboard", URL: "/admin"},
    }
    
    // Add admin-only items
    if user.Role == "admin" {
        items = append(items,
            engine.NavItem{Label: "Settings", Icon: "settings", URL: "/admin/settings"},
            engine.NavItem{Label: "Logs", Icon: "list_alt", URL: "/admin/logs"},
        )
    }
    
    return items
})
```

### Resource Navigation

Resources are automatically added to navigation. Customize with:

```go
type UserResource struct {
    engine.BaseResource
}

func (r *UserResource) NavigationBadge(ctx context.Context) string {
    count, _ := r.db.User.Query().Count(ctx)
    return fmt.Sprintf("%d", count)
}

func (r *UserResource) NavigationBadgeColor(ctx context.Context) string {
    return "blue"
}

func (r *UserResource) NavigationIcon() string {
    return "people"
}

func (r *UserResource) NavigationGroup() string {
    return "Management"
}
```

---

## Middleware

### Built-in Middleware

```go
panel.
    // Authentication middleware
    WithMiddleware(middleware.RequireAuth).
    
    // Role-based access
    WithMiddleware(middleware.RequireRole("admin")).
    
    // Permission-based access
    WithMiddleware(middleware.RequirePermission("users.view")).
    
    // CORS
    WithMiddleware(middleware.CORS(middleware.CORSOptions{
        AllowedOrigins: []string{"https://example.com"},
        AllowedMethods: []string{"GET", "POST", "PUT", "DELETE"},
        AllowedHeaders: []string{"Content-Type", "Authorization"},
    })).
    
    // CSRF protection
    WithMiddleware(middleware.CSRF(middleware.CSRFOptions{
        Secret:        "your-secret-key",
        SecureCookie:  true,
        HTTPOnly:      true,
    })).
    
    // Rate limiting
    WithMiddleware(middleware.RateLimit(100, time.Minute)).
    
    // Recovery
    WithMiddleware(middleware.Recovery).
    
    // Request logging
    WithMiddleware(middleware.Logger).
    
    // Security headers
    WithMiddleware(middleware.SecureHeaders)
```

### Custom Middleware

```go
func CustomMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Custom logic
        ctx := context.WithValue(r.Context(), "custom_key", "custom_value")
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

panel.WithMiddleware(CustomMiddleware)
```

### Conditional Middleware

```go
panel.WithMiddleware(func(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if strings.HasPrefix(r.URL.Path, "/admin/api/") {
            // API-specific middleware
            middleware.RequireAuth(next).ServeHTTP(w, r)
        } else {
            next.ServeHTTP(w, r)
        }
    })
})
```

---

## Notifications

### Basic Notification Setup

```go
panel.
    EnableNotifications().
    WithNotificationStore(&notifications.DatabaseStore{
        DB: db,
    })
```

### Notification Configuration

```go
panel.
    // Enable notifications
    EnableNotifications().
    
    // Custom store
    WithNotificationStore(&CustomNotificationStore{
        redis: redisClient,
    }).
    
    // Notification handlers
    WithNotificationHandler(&notifications.Handler{
        Store: store,
        Prefix: "/admin/notifications",
    })
```

### Custom Notification Store

```go
type CustomNotificationStore struct {
    redis *redis.Client
}

func (s *CustomNotificationStore) Send(ctx context.Context, n *notifications.Notification) error {
    data, _ := json.Marshal(n)
    return s.redis.LPush(ctx, fmt.Sprintf("notifications:%s", n.UserID), data).Err()
}

func (s *CustomNotificationStore) GetUnread(ctx context.Context, userID string) ([]*notifications.Notification, error) {
    data, err := s.redis.LRange(ctx, fmt.Sprintf("notifications:%s", userID), 0, -1).Result()
    if err != nil {
        return nil, err
    }
    
    var notifs []*notifications.Notification
    for _, d := range data {
        var n notifications.Notification
        if err := json.Unmarshal([]byte(d), &n); err == nil {
            notifs = append(notifs, &n)
        }
    }
    return notifs, nil
}
```

---

## Multi-tenancy

### Subdomain-based Tenancy

```go
resolver := engine.NewSubdomainResolver("example.com").
    Register(&engine.Tenant{
        ID:       "acme",
        Subdomain: "acme",
        Name:     "Acme Corp",
        Database: "acme_db",
    }).
    Register(&engine.Tenant{
        ID:       "techco",
        Subdomain: "techco",
        Name:     "Tech Company",
        Database: "techco_db",
    })

router := engine.NewMultiPanelRouter(resolver).
    RegisterPanel("acme", acmePanel).
    RegisterPanel("techco", techcoPanel)

panel.WithTenantResolver(resolver)
```

### Path-based Tenancy

```go
resolver := engine.NewPathResolver("/app").
    Register(&engine.Tenant{
        ID:   "tenant1",
        Path: "tenant1",
        Name: "Tenant 1",
    }).
    Register(&engine.Tenant{
        ID:   "tenant2",
        Path: "tenant2",
        Name: "Tenant 2",
    })
```

### Custom Tenant Resolver

```go
type CustomTenantResolver struct {
    tenants map[string]*engine.Tenant
}

func (r *CustomTenantResolver) Resolve(req *http.Request) (*engine.Tenant, error) {
    // Custom resolution logic
    header := req.Header.Get("X-Tenant-ID")
    if tenant, exists := r.tenants[header]; exists {
        return tenant, nil
    }
    return nil, engine.ErrTenantNotFound
}

func (r *CustomTenantResolver) Register(tenant *engine.Tenant) error {
    r.tenants[tenant.ID] = tenant
    return nil
}
```

### Tenant-aware Resources

```go
type TenantResource struct {
    engine.BaseResource
}

func (r *TenantResource) List(ctx context.Context) ([]any, error) {
    tenant := engine.TenantFromContext(ctx)
    
    // Query tenant-specific data
    users, err := r.db.User.Query().
        Where(user.TenantID(tenant.ID)).
        All(ctx)
    if err != nil {
        return nil, err
    }
    
    result := make([]any, len(users))
    for i, u := range users {
        result[i] = u
    }
    return result, nil
}
```

---

## Performance

### Caching

```go
panel.
    // Enable response caching
    WithCache(&cache.MemoryCache{
        TTL: 5 * time.Minute,
    }).
    
    // Static asset caching
    WithStaticCacheControl(middleware.CacheControl{
        MaxAge: 30 * 24 * time.Hour, // 30 days
    })
```

### Gzip Compression

```go
panel.WithMiddleware(middleware.Gzip(middleware.GzipOptions{
    Level: 6,
    Types: []string{
        "text/html",
        "text/css",
        "application/javascript",
        "application/json",
    },
}))
```

### Database Connection Pool

```go
db, err := ent.Open("postgres", dsn, 
    ent.Log(log.Stdlib),
    ent.Driver(&sql.DB{
        SetMaxOpenConns:    25,
        SetMaxIdleConns:    25,
        SetConnMaxLifetime: time.Hour,
    }),
)
```

### Performance Monitoring

```go
panel.
    WithPerformanceMiddleware().
    WithMetrics(&engine.Metrics{
        Enabled: true,
        Path:    "/admin/metrics",
    })
```

---

## Complete Example

```go
package main

import (
    "context"
    "net/http"
    "time"
    
    "entgo.io/ent/ent"
    "github.com/bozz33/sublimeadmin/engine"
    "github.com/bozz33/sublimeadmin/auth"
    "github.com/bozz33/sublimeadmin/middleware"
    "github.com/bozz33/sublimeadmin/notifications"
    "github.com/bozz33/sublimeadmin/cache"
)

func main() {
    // Database connection
    db, err := ent.Open("postgres", "postgres://user:pass@localhost/db")
    if err != nil {
        panic(err)
    }
    defer db.Close()
    
    // Create panel
    panel := engine.NewPanel("admin").
        WithPath("/admin").
        WithDatabase(db).
        WithBrandName("My Application").
        WithLogo("/assets/logo.png").
        WithFavicon("/assets/favicon.ico").
        WithPrimaryColor("#3B82F6").
        EnableDarkMode()
    
    // Authentication
    panel.
        WithAuthManager(auth.NewManager(db)).
        WithSession(&scs.SessionStore{
            Lifetime: 24 * time.Hour,
            Cookie: scs.Cookie{
                Name:     "admin_session",
                SameSite: http.SameSiteLaxMode,
                Secure:   true,
                HttpOnly: true,
            },
        }).
        EnableRegistration().
        EnableEmailVerification().
        EnablePasswordReset().
        EnableProfile()
    
    // Middleware
    panel.
        WithMiddleware(middleware.Recovery).
        WithMiddleware(middleware.Logger).
        WithMiddleware(middleware.SecureHeaders).
        WithMiddleware(middleware.CORS(middleware.CORSOptions{
            AllowedOrigins: []string{"https://example.com"},
            AllowedMethods: []string{"GET", "POST", "PUT", "DELETE"},
            AllowedHeaders: []string{"Content-Type", "Authorization"},
        })).
        WithMiddleware(middleware.CSRF(middleware.CSRFOptions{
            Secret:        "your-secret-key",
            SecureCookie:  true,
            HTTPOnly:      true,
        })).
        WithMiddleware(middleware.RateLimit(100, time.Minute)).
        WithMiddleware(middleware.Gzip(middleware.GzipOptions{
            Level: 6,
            Types: []string{
                "text/html",
                "text/css",
                "application/javascript",
                "application/json",
            },
        }))
    
    // Notifications
    panel.
        EnableNotifications().
        WithNotificationStore(&notifications.DatabaseStore{
            DB: db,
        })
    
    // Navigation
    panel.
        AddNavGroup(engine.NavigationGroup{
            Label: "Management",
            Icon:  "admin_panel_settings",
            Items: []engine.NavItem{
                {Label: "Users", Icon: "people", URL: "/admin/users"},
                {Label: "Roles", Icon: "admin_panel_settings", URL: "/admin/roles"},
            },
        }).
        AddNavGroup(engine.NavigationGroup{
            Label: "Content",
            Icon:  "article",
            Items: []engine.NavItem{
                {Label: "Posts", Icon: "description", URL: "/admin/posts"},
                {Label: "Pages", Icon: "article", URL: "/admin/pages"},
            },
        })
    
    // Resources
    panel.AddResources(
        NewUserResource(db),
        NewPostResource(db),
        NewPageResource(db),
    )
    
    // Custom pages
    panel.AddPages(
        NewDashboardPage(),
        NewSettingsPage(),
    )
    
    // Performance
    panel.
        WithCache(&cache.MemoryCache{
            TTL: 5 * time.Minute,
        }).
        WithStaticCacheControl(middleware.CacheControl{
            MaxAge: 30 * 24 * time.Hour,
        })
    
    // Start server
    server := &http.Server{
        Addr:         ":8080",
        Handler:      panel.Router(),
        ReadTimeout:  15 * time.Second,
        WriteTimeout: 15 * time.Second,
        IdleTimeout:  60 * time.Second,
    }
    
    log.Printf("Server starting on %s", server.Addr)
    if err := server.ListenAndServe(); err != nil {
        log.Fatal(err)
    }
}
```

---

## Environment Variables

Create a `.env` file for configuration:

```bash
# Database
DB_HOST=localhost
DB_PORT=5432
DB_NAME=sublimego
DB_USER=sublimego
DB_PASSWORD=password

# Panel
PANEL_PATH=/admin
PANEL_BRAND_NAME=My Application
PANEL_PRIMARY_COLOR=#3B82F6

# Authentication
SESSION_SECRET=your-secret-key
JWT_SECRET=your-jwt-secret

# Email
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=your-email@gmail.com
SMTP_PASSWORD=your-app-password

# Redis (for notifications/sessions)
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=

# Multi-tenancy
TENANT_MODE=subdomain
DOMAIN=example.com
```

Load with Viper:

```go
import "github.com/spf13/viper"

viper.SetConfigFile(".env")
viper.ReadInConfig()

db, err := ent.Open("postgres", fmt.Sprintf(
    "postgres://%s:%s@%s:%s/%s",
    viper.GetString("DB_USER"),
    viper.GetString("DB_PASSWORD"),
    viper.GetString("DB_HOST"),
    viper.GetString("DB_PORT"),
    viper.GetString("DB_NAME"),
))
```

---

## Next Steps

1. **Environment Setup**: Configure environment variables
2. **Database Setup**: Run migrations and seed data
3. **Authentication**: Configure auth providers
4. **Resources**: Create your first resources
5. **Testing**: Write tests for your configuration
6. **Deployment**: Configure for production

For more examples, see the [examples directory](https://github.com/bozz33/sublime-admin/tree/main/examples).
