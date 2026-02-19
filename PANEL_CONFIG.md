# Panel Configuration

Complete reference for configuring a sublime-admin panel.

---

## Configuration File

When used via the [SublimeGo](https://github.com/bozz33/sublimego) starter, configuration is
managed via `config/config.yaml` and loaded by `appconfig/loader.go` using Viper.

Priority order (highest to lowest):

1. Environment variables (prefixed with `SUBLIMEGO_`)
2. `config/config.yaml`
3. Built-in defaults

When using `sublimego-core` directly (without the starter), pass configuration values
programmatically via the `Panel` builder methods.

---

## Full Configuration Reference (SublimeGo starter)

```yaml
#  Server 
server:
  host: "0.0.0.0"          # Bind address
  port: 8080               # HTTP port
  read_timeout: 15s        # Max time to read request
  write_timeout: 15s       # Max time to write response
  idle_timeout: 60s        # Keep-alive timeout

#  Database 
database:
  driver: "sqlite"         # sqlite | postgres | mysql
  url: "file:app.db?cache=shared&_fk=1"
  max_open_conns: 25
  max_idle_conns: 5
  conn_max_lifetime: 5m

#  Engine (Panel) 
engine:
  base_path: "/admin"      # URL prefix for the admin panel
  brand_name: "My Admin"   # Displayed in the sidebar header
  brand_logo: ""           # Optional path to a logo image
  items_per_page: 25       # Default pagination size
  date_format: "2006-01-02"
  datetime_format: "2006-01-02 15:04"

#  Authentication 
auth:
  session_lifetime: 24h
  cookie_name: "sublimego_session"
  cookie_secure: true      # Set false for local HTTP development
  cookie_same_site: "lax"  # strict | lax | none
  bcrypt_cost: 12          # 10-14 recommended
  login_throttle: 5        # Max failed attempts before lockout
  lockout_duration: 15m

#  Logger 
logger:
  level: "info"            # debug | info | warn | error
  format: "json"           # json | text
  output: "stdout"         # stdout | stderr | /path/to/file.log
  max_size_mb: 100         # Log rotation: max file size
  max_backups: 3           # Log rotation: number of backups
  max_age_days: 28         # Log rotation: max age

#  Jobs 
jobs:
  workers: 5               # Concurrent job workers
  poll_interval: 5s        # How often to poll for new jobs
  retention: 72h           # How long to keep completed jobs

#  Mailer 
mailer:
  host: "smtp.example.com"
  port: 587
  username: ""
  password: ""             # Use SUBLIMEGO_MAILER_PASSWORD env var
  from: "admin@example.com"
  from_name: "My Admin"
  tls: true
```

---

## Environment Variables

Every config key maps to an environment variable using the `SUBLIMEGO_` prefix and `_` as separator:

| Config key | Environment variable |
|------------|---------------------|
| `server.port` | `SUBLIMEGO_SERVER_PORT` |
| `database.url` | `SUBLIMEGO_DATABASE_URL` |
| `auth.bcrypt_cost` | `SUBLIMEGO_AUTH_BCRYPT_COST` |
| `mailer.password` | `SUBLIMEGO_MAILER_PASSWORD` |
| `engine.brand_name` | `SUBLIMEGO_ENGINE_BRAND_NAME` |
| `engine.base_path` | `SUBLIMEGO_ENGINE_BASE_PATH` |

---

## Panel Setup in Code

This is the **primary API** when using `sublimego-core` directly.

```go
package main

import (
    "net/http"
    "time"

    "github.com/alexedwards/scs/v2"
    "github.com/bozz33/sublimeadmin/auth"
    "github.com/bozz33/sublimeadmin/engine"
    "github.com/bozz33/sublimeadmin/flash"
)

func main() {
    // Session manager
    session := scs.New()
    session.Lifetime = 24 * time.Hour
    session.Cookie.Name = "my_session"
    session.Cookie.Secure = true
    session.Cookie.SameSite = http.SameSiteLaxMode

    // Auth manager
    authManager := auth.NewManager(session)

    // Flash manager
    flashManager := flash.NewManager(session)
    _ = flashManager

    // Panel
    panel := engine.NewPanel("admin").
        SetPath("/admin").
        SetBrandName("My Application").
        SetDatabase(db).           // implements engine.DatabaseClient
        SetAuthManager(authManager).
        SetSession(session).
        SetDashboardProvider(myDashboard).    // optional
        SetAuthTemplates(myAuthViews).        // optional
        SetDashboardTemplates(myDashViews)    // optional

    // Register resources — ALWAYS MANUAL in sublimego-core
    panel.AddResources(
        NewProductResource(db),
        NewUserResource(db),
        NewOrderResource(db),
    )

    // Register custom pages
    panel.AddPages(
        NewSettingsPage(),
        NewReportsPage(),
    )

    http.ListenAndServe(":8080", panel.Router())
}
```

### Panel Builder Methods

| Method | Description |
|--------|-------------|
| `SetPath(string)` | URL prefix for all admin routes (default: `/admin`) |
| `SetBrandName(string)` | Name shown in the sidebar header |
| `SetDatabase(DatabaseClient)` | DB adapter for user auth (required) |
| `SetAuthManager(*auth.Manager)` | Auth manager (required) |
| `SetSession(*scs.SessionManager)` | SCS session manager (required) |
| `SetDashboardProvider(...)` | Custom dashboard Templ component |
| `SetAuthTemplates(...)` | Custom login/register Templ templates |
| `SetDashboardTemplates(...)` | Custom dashboard layout templates |
| `AddResources(...Resource)` | Register CRUD resources |
| `AddPages(...Page)` | Register custom pages |
| `Router()` | Build and return the `http.Handler` |

---

## Required DatabaseClient Interface

```go
type DatabaseClient interface {
    FindUserByEmail(ctx context.Context, email string) (*auth.User, string, error)
    CreateUser(ctx context.Context, name, email, hashedPassword string) (*auth.User, error)
    UserExists(ctx context.Context, email string) (bool, error)
}
```

Example implementation:

```go
type MyDB struct{ db *sql.DB }

func (d *MyDB) FindUserByEmail(ctx context.Context, email string) (*auth.User, string, error) {
    var id int
    var name, hashedPwd string
    err := d.db.QueryRowContext(ctx,
        "SELECT id, name, password FROM users WHERE email = ?", email,
    ).Scan(&id, &name, &hashedPwd)
    if err != nil {
        return nil, "", err
    }
    return auth.NewUser(id, email, name), hashedPwd, nil
}

func (d *MyDB) CreateUser(ctx context.Context, name, email, hashedPwd string) (*auth.User, error) {
    res, err := d.db.ExecContext(ctx,
        "INSERT INTO users (name, email, password) VALUES (?, ?, ?)", name, email, hashedPwd,
    )
    if err != nil {
        return nil, err
    }
    id, _ := res.LastInsertId()
    return auth.NewUser(int(id), email, name), nil
}

func (d *MyDB) UserExists(ctx context.Context, email string) (bool, error) {
    var count int
    err := d.db.QueryRowContext(ctx,
        "SELECT COUNT(*) FROM users WHERE email = ?", email,
    ).Scan(&count)
    return count > 0, err
}
```

---

## Auto-generated Routes

`panel.Router()` registers the following routes automatically:

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/login` | Login page |
| `POST` | `/login` | Login handler |
| `GET` | `/register` | Registration page |
| `POST` | `/register` | Registration handler |
| `GET` | `/logout` | Logout |
| `GET` | `/` | Dashboard |
| `GET` | `/{slug}` | Resource list |
| `GET` | `/{slug}/create` | Create form |
| `POST` | `/{slug}` | Create |
| `GET` | `/{slug}/{id}/edit` | Edit form |
| `PUT` | `/{slug}/{id}` | Update |
| `DELETE` | `/{slug}/{id}` | Delete |

---

## Session Configuration

```go
session := scs.New()
session.Lifetime = 24 * time.Hour       // session duration
session.IdleTimeout = 2 * time.Hour     // inactivity timeout
session.Cookie.Name = "my_session"
session.Cookie.Secure = true            // HTTPS only (set false for local dev)
session.Cookie.HttpOnly = true
session.Cookie.SameSite = http.SameSiteLaxMode

// Use a persistent store (default is in-memory)
// session.Store = mysqlstore.New(db, 1*time.Minute)
// session.Store = redisstore.New(pool)
```

---

## Auth Manager Configuration

```go
authManager := auth.NewManager(session)

// Login
err := authManager.Login(ctx, user)
err := authManager.LoginWithRequest(r, user)

// Logout
err := authManager.Logout(ctx)

// Post-login redirect
authManager.SetIntendedURLFromRequest(r)
redirectURL := authManager.IntendedURL(ctx, "/dashboard")

// Checks
authManager.IsAuthenticatedFromRequest(r)
authManager.Can(ctx, "products.create")
authManager.HasRole(ctx, "admin")
authManager.IsAdmin(ctx)
```

---

## Middleware Configuration

Apply middlewares to your own routes outside the panel:

```go
mux := http.NewServeMux()

// Panel handles /admin/*
mux.Handle("/admin/", panel.Router())

// Your own protected routes
protected := middleware.Chain(
    middleware.Session(session),
    middleware.RequireAuth(authManager),
)
mux.Handle("/api/", protected(myAPIHandler))

// Public routes with rate limiting
mux.Handle("/api/public/", middleware.Apply(publicHandler,
    middleware.RateLimit(100, time.Minute),
))

http.ListenAndServe(":8080", mux)
```

---

## Further Reading

- [GUIDE_EN.md](GUIDE_EN.md) — Complete usage guide (all packages)
- [ARCHITECTURE.md](ARCHITECTURE.md) — Package layout and design patterns
