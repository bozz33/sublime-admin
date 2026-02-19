# Architecture

This document describes the structure, patterns, and design decisions of the sublime-admin framework.

---

## Guiding Principles

1. **Idiomatic Go** — stdlib first; every dependency must be justified.
2. **Small interfaces** — each contract exposes the minimum required (Interface Segregation Principle).
3. **No magic** — everything is explicit, traceable, and testable.
4. **Single binary** — `//go:embed` for all assets; no external files required at runtime.
5. **Flat structure** — packages at the root level, no unnecessary `pkg/` wrapper.

---

## Package Layout

```
sublimego-core/
 actions/          # Row actions (edit, delete, custom, bulk)
 auth/             # Authentication, sessions, roles, permissions
 engine/           # Framework core: Panel, CRUD handlers, relations
 apperrors/        # Structured errors (package apperrors)
 export/           # CSV / Excel export
 flash/            # Session-based flash messages
 form/             # Form builder: fields, layouts, validation
 middleware/       # HTTP middlewares (auth, CORS, recovery, throttle, rate limit)
 table/            # Table builder: columns, filters, actions
 validation/       # Validation (go-playground/validator)
 go.mod
```

---

## Layer Diagram

```
                     HTTP Request

                         ↓

                Middlewares (middleware/)
   RequireAuth  RequireRole  CORS  Recovery
   RateLimit  Session  Flash

                         ↓

            engine.Panel  (engine/panel.go)
   Routing  Auth  Session  Resources  Pages

                    ↙           ↘
    Resource (CRUD)         Custom Page
    (engine/contract.go)    (engine/page)

                         ↓

                    Builders
   form.Form  table.Table  actions.Action

                         ↓

             Templ Templates  (views/ + ui/)

                         ↓

                  Your Database Layer
```

---

## Key Modules

### `engine/` — Framework Core

| File | Role |
|------|------|
| `panel.go` | Panel configuration, route registration, middleware injection |
| `contract.go` | Interfaces: `Resource`, `ResourceCRUD`, `ResourceViews`, `ResourcePermissions`, `ResourceNavigation` |

### `form/` — Form Builder

```go
// Standard fields
form.Text("name").Label("Name").Required()
form.Email("email").Label("Email")
form.Password("password").Label("Password")
form.Number("price").Label("Price")
form.Textarea("bio").Label("Bio").Rows(5)
form.Select("status").Label("Status").SetOptions(...)
form.Checkbox("active").Label("Active")
form.Toggle("featured").Label("Featured")
form.FileUpload("avatar").Label("Avatar").Accept("image/*")

// Advanced fields
form.RichEditor("content").Label("Content").WithToolbar("bold", "italic", "link")
form.MarkdownEditor("notes").Label("Notes").Rows(10)
form.Tags("tags").Label("Tags").WithSuggestions("go", "web", "admin")
form.KeyValue("meta").Label("Metadata").WithLabels("Key", "Value")
form.ColorPicker("color").Label("Color").WithSwatches("#ef4444", "#22c55e")
form.Slider("discount").Label("Discount").Range(0, 100).WithUnit("%")
form.Repeater("items", subFields...).Label("Items")

// Layouts
form.NewSection("General").Schema(...)
form.NewGrid(2).Schema(...)
form.NewTabs().AddTab("General", ...).AddTab("Advanced", ...)
```

### `table/` — Table Builder

```go
table.New(data).
    WithColumns(
        table.Text("name").Label("Name").Sortable().Searchable(),
        table.Badge("status").Label("Status").Colors(map[string]string{
            "active": "success", "inactive": "danger",
        }),
        table.Image("avatar").Label("Avatar").Round(),
    ).
    SetActions(
        actions.EditAction("/url"),
        actions.DeleteAction("/url"),
    ).
    Search(true).
    Paginate(true)
```

### `auth/` — Authentication

| File | Role |
|------|------|
| `manager.go` | `Manager` — Login, Logout, session, permissions, roles |
| `user.go` | `User` struct — permissions, roles, metadata, helpers |

### `middleware/` — HTTP Middlewares

| Middleware | Role |
|---|---|
| `RequireAuth` | Redirect unauthenticated users to login |
| `RequireGuest` | Redirect authenticated users away |
| `RequireRole` / `RequirePermission` | Role/permission gate |
| `LoadUser` | Load user into context without forcing auth |
| `Recovery` | Panic → 500 with structured error |
| `Logger` | Structured HTTP request logging |
| `CORS` | Cross-Origin Resource Sharing |
| `RateLimit` | Per-IP request throttling |
| `Session` | Session middleware (wraps SCS) |
| `Flash` | Flash message injection |

### `flash/` — Flash Messages

```go
flashManager := flash.NewManager(session)
flashManager.Success(ctx, "Saved!")
flashManager.Error(ctx, "Failed")
flashManager.Warning(ctx, "Low stock")
flashManager.Info(ctx, "Update available")

messages := flashManager.GetAndClear(ctx)
```

### `export/` — CSV / Excel Export

```go
export.New(export.FormatCSV).
    SetHeaders([]string{"ID", "Name"}).
    AddRows(rows).
    Write(w)

export.ExportStructsExcel(w, myStructSlice)
```

### `actions/` — Row Actions

```go
actions.EditAction("/products")
actions.DeleteAction("/products")
actions.ViewAction("/products")
actions.New("publish").SetLabel("Publish").SetColor("success").SetUrl(...)
```

---

## Error Handling

The `apperrors/` package provides structured, HTTP-aware errors:

```go
import "github.com/bozz33/sublimeadmin/apperrors"

err := apperrors.NotFound("user not found")
err := apperrors.Forbidden("access denied")
err := apperrors.Validation("email", "invalid format")

apperrors.Handle(w, r, err)
```

---

## Configuration

When used via the SublimeGo starter, configuration is loaded by `appconfig/loader.go` via Viper.
Priority order: environment variables > `config.yaml` > defaults.

```yaml
server:
  host: "0.0.0.0"
  port: 8080

database:
  driver: "sqlite"
  url: "file:app.db?cache=shared&_fk=1"

engine:
  base_path: "/admin"
  brand_name: "My Admin"
  items_per_page: 25

auth:
  session_lifetime: 24h
  cookie_name: "sublimego_session"
  bcrypt_cost: 12
```

---

## Design Patterns

### Builder Pattern
Used throughout `form`, `table`, and `actions` for a fluent, readable API.

### Interface Segregation
`Resource` is composed of small focused interfaces (`ResourceMeta`, `ResourceCRUD`,
`ResourceViews`, `ResourcePermissions`, `ResourceNavigation`) — each implementable independently.

### Dependency Injection
`Panel` accepts its dependencies (DB, Auth, Session) via `Set*` setters.

### Explicit Registration
Resources and pages are **always registered manually** via `panel.AddResources(...)` and
`panel.AddPages(...)`. There is no auto-discovery in `sublimego-core`.

> **Note:** The [SublimeGo](https://github.com/bozz33/sublimego) starter project adds an optional
> `go generate` layer that auto-generates the registration call, but the underlying mechanism
> is still the same explicit `AddResources`.

---

## Testing

```bash
go test ./...           # All tests
go test -cover ./...    # With coverage report
go test -race ./...     # Race condition detection
go test ./form/... -v   # Specific package, verbose output
```

---

## Further Reading

- [GUIDE_EN.md](GUIDE_EN.md) — Complete usage guide
- [PANEL_CONFIG.md](PANEL_CONFIG.md) — Full configuration reference
