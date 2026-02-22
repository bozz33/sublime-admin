# sublime-admin

**The core Go framework library for building admin panels.**
Use this package as a dependency in your existing Go project.

[![Go Reference](https://pkg.go.dev/badge/github.com/bozz33/sublimeadmin.svg)](https://pkg.go.dev/github.com/bozz33/sublimeadmin)
[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?logo=go)](https://go.dev)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

---

## Two Repositories

sublimeadmin is split into two complementary repositories:

| | [sublimeadmin](https://github.com/bozz33/sublimeadmin) | [sublime-admin](https://github.com/bozz33/sublime-admin) |
|---|---|---|
| **Type** | Complete starter project | Core framework library |
| **Use when** | Starting a new admin panel project | Adding admin panel to an existing Go project |
| **Includes** | DB setup, Ent schemas, CLI, examples, views | Framework packages only |
| **Import** | `github.com/bozz33/sublimeadmin` | `github.com/bozz33/sublimeadmin` |

> **Not sure which to pick?** Start with **sublimeadmin**  it includes everything out of the box.

---

## Installation

```bash
go get github.com/bozz33/sublimeadmin@latest
```

---

## Features

### Forms
- Fields: Text, Email, Password, Number, Textarea, Select, Checkbox, Toggle, DatePicker, FileUpload
- Advanced fields: **RichEditor**, **MarkdownEditor**, **TagsInput**, **KeyValue**, **ColorPicker**, **Slider**
- Layouts: Section, Grid, Tabs, **Wizard/Steps**, Callout, Repeater

### Tables
- Columns: Text, Badge, Boolean, Date, Image
- Sorting, search, pagination, filters, bulk actions
- **Summaries** (sum, average, min, max, count) in the table footer
- **Grouping**  group rows by column value with optional collapse
- Export CSV/Excel, Import CSV/Excel/JSON

### Authentication & Security
- Bcrypt password hashing + secure sessions
- Middleware: RequireAuth, RequireRole, RequirePermission, login throttling
- **MFA/TOTP**  two-factor authentication (RFC 6238) + recovery codes
- Role and permission management

### Notifications
- In-memory store (development)
- **DatabaseStore**  persistent backend via any Ent client
- SSE (Server-Sent Events) for real-time delivery
- **Broadcaster**  per-user SSE fan-out with 30s heartbeat

### Advanced Architecture
- **Multi-tenancy**  `SubdomainResolver`, `PathResolver`, `MultiPanelRouter`, `TenantAware`
- **Render Hooks**  10 named UI injection points
- **Plugin system**  `Plugin` interface with `Boot()` and thread-safe registry
- **Nested Resources**  `RelationManager` (BelongsTo, HasMany, ManyToMany)
- Background jobs with SQLite persistence
- Structured logger (`log/slog`) with log rotation
- Built-in mailer

---

## Quick Start

```go
package main

import (
    "net/http"

    "github.com/bozz33/sublimeadmin/engine"
    "github.com/bozz33/sublimeadmin/form"
    "github.com/bozz33/sublimeadmin/table"
)

func main() {
    panel := engine.NewPanel("admin").
        WithPath("/admin").
        WithBrandName("My App").
        WithDatabase(db)

    panel.AddResources(
        NewProductResource(db),
        NewUserResource(db),
    )

    http.ListenAndServe(":8080", panel.Router())
}
```

---

## Package Reference

| Package | Description |
|---------|-------------|
| `engine` | Core panel, CRUD handlers, multi-tenancy, relations |
| `form` | Form builder  fields, layouts, validation |
| `table` | Table builder  columns, filters, summaries, grouping |
| `actions` | Row actions (edit, delete, custom, bulk) |
| `auth` | Authentication, sessions, roles, permissions, MFA/TOTP |
| `middleware` | HTTP middlewares (auth, CORS, CSRF, recovery, throttle) |
| `notifications` | Notifications  memory store, DatabaseStore, SSE Broadcaster |
| `hooks` | Render Hooks  named UI injection points |
| `plugin` | Plugin system |
| `validation` | Input validation (go-playground/validator + gorilla/schema) |
| `widget` | Dashboard widgets (stats cards, ApexCharts) |
| `flash` | Session-based flash messages |
| `apperrors` | Structured errors  AppError, Handler, typed constructors |
| `export` | CSV / Excel export |
| `ui` | Templ UI components and layouts |

---

## Resource Example

```go
type ProductResource struct {
    engine.BaseResource
    db *ent.Client
}

func (r *ProductResource) Slug() string        { return "products" }
func (r *ProductResource) Label() string       { return "Product" }
func (r *ProductResource) PluralLabel() string { return "Products" }
func (r *ProductResource) Icon() string        { return "package" }

func (r *ProductResource) Form(ctx context.Context, item any) templ.Component {
    f := form.New().SetSchema(
        form.Text("name").Label("Name").Required(),
        form.RichEditor("description").Label("Description"),
        form.Select("status").Label("Status").SetOptions(map[string]string{
            "draft":     "Draft",
            "published": "Published",
        }),
        form.Tags("tags").Label("Tags"),
        form.ColorPicker("color").Label("Color"),
    )
    return views.GenericForm(f)
}

func (r *ProductResource) Table(ctx context.Context) templ.Component {
    t := table.New(nil).
        WithColumns(
            table.Text("name").WithLabel("Name").Sortable().Searchable(),
            table.Badge("status").WithLabel("Status"),
            table.Date("created_at").WithLabel("Created").Sortable(),
        ).
        WithSummaries(
            table.NewSummary("price", table.SummarySum).WithLabel("Total").WithFormat("$%.2f"),
        ).
        WithGroups(
            table.GroupBy("status").WithLabel("By status").Collapsible(),
        )
    return views.GenericTable(t)
}
```

---

## MFA Example

```go
mfa := auth.NewMFA(auth.DefaultMFAConfig())

// Generate a secret for a user
secret, _ := mfa.GenerateSecret()

// Provisioning URI for QR code
uri := mfa.ProvisioningURI(secret, user.Email)

// Validate a submitted TOTP code
if !mfa.Validate(secret, code) {
    // invalid
}

// Recovery codes
codes, _ := mfa.GenerateRecoveryCodes()
auth.SetRecoveryCodes(user, codes)
auth.UseRecoveryCode(user, submittedCode) // consumes the code
```

---

## Render Hooks Example

```go
import "github.com/bozz33/sublimeadmin/hooks"

hooks.Register(hooks.AfterContent, func(ctx context.Context) templ.Component {
    return myBannerComponent()
})
```

---

## Multi-Tenancy Example

```go
resolver := engine.NewSubdomainResolver("example.com").
    Register(&engine.Tenant{ID: "acme", Subdomain: "acme", Name: "Acme Corp"})

router := engine.NewMultiPanelRouter(resolver).
    RegisterPanel("acme", acmePanel)

http.ListenAndServe(":8080", engine.TenantMiddleware(resolver, true)(router))
```

---

## Documentation

Full documentation, guides, and examples are available in the
[sublimeadmin starter project](https://github.com/bozz33/sublimeadmin):

- [ARCHITECTURE.md](https://github.com/bozz33/sublimeadmin/blob/main/ARCHITECTURE.md)
- [RESOURCES_GUIDE.md](https://github.com/bozz33/sublimeadmin/blob/main/RESOURCES_GUIDE.md)
- [PANEL_CONFIG.md](https://github.com/bozz33/sublimeadmin/blob/main/PANEL_CONFIG.md)
- [TEMPLATING.md](https://github.com/bozz33/sublimeadmin/blob/main/TEMPLATING.md)
- [CONTRIBUTING.md](https://github.com/bozz33/sublimeadmin/blob/main/CONTRIBUTING.md)

---

## Contributing

Contributions are welcome  please read
[CONTRIBUTING.md](https://github.com/bozz33/sublimeadmin/blob/main/CONTRIBUTING.md)
for guidelines.

---

## License

MIT  see [LICENSE](LICENSE).

---

*Inspired by [Laravel Filament](https://filamentphp.com/)  Built with [Ent](https://entgo.io/) and [Templ](https://templ.guide/)*
