# sublime-admin

**The core Go framework library for building admin panels.**
Use this package as a dependency in your existing Go project.
Production-ready framework for building sophisticated admin interfaces in Go.

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

### Forms (22 field types)
- **Basic**: Text, Email, Password, Number, Textarea, Select, Checkbox, Toggle, DatePicker, FileUpload, Hidden
- **Advanced**: RichEditor, MarkdownEditor, Tags, KeyValue, ColorPicker, Slider, Radio, CheckboxList, Repeater
- **Layouts**: Section, Grid, Tabs, Wizard, Callout, Fieldset, Flex, Split
- **Validation**: Live validation with Datastar SSE, custom rules, error handling

### Tables (13 column types + 4 inline edit)
- **Display columns**: Text, Badge, Image, Boolean, Date, Avatar, Icon, Color, Tags
- **Inline edit**: TextInput, Select, Toggle, Checkbox (with Datastar PATCH)
- **Features**: Sorting, search, pagination, filters (6 types), bulk actions
- **Advanced**: Summaries (Sum, Average, Min, Max, Count), Grouping, Column manager, Export/Import

### Actions
- **Basic actions**: Edit, Delete, View, Create, Export, Import, Restore, ForceDelete
- **Modal actions**: Confirmation dialogs, form modals, slide-over panels
- **Features**: Lifecycle hooks (Before/After/OnSuccess/OnFailure), rate limiting, authorization
- **Advanced**: Action groups, bulk actions, custom handlers

### Infolist (12 entry types)
- **Entries**: Text, Badge, Boolean, Date, Image, Color, Icon, List, Link, Code, KeyValue, Repeatable
- **Layout**: Sections with columns (1-3), descriptions, icons
- **Features**: Rich formatting, syntax highlighting, responsive design

### Enum Package
- **Interfaces**: HasLabel, HasColor, HasIcon, HasDescription, HasGroup
- **Helpers**: Options, Labels, Colors, Icons, BadgeColor, RadioOptions, CheckboxOptions
- **Features**: Generic helpers, grouped options, type-safe conversions

### Authentication & Security
- **Auth**: Session-based, bcrypt passwords, role/permission management
- **MFA/TOTP**: RFC 6238 compliant, recovery codes, QR provisioning
- **Middleware**: Auth, CORS, CSRF, Rate limiting, Recovery, Security headers
- **Features**: Login throttling, secure sessions, multi-factor support

### Notifications
- **Stores**: In-memory (dev), DatabaseStore (production)
- **Delivery**: SSE streaming, Datastar badge updates, real-time
- **Features**: Per-user channels, unread counts, mark as read

### Advanced Architecture
- **Multi-tenancy**: Subdomain/Path resolvers, tenant-aware routing
- **Relations**: BelongsTo, HasOne, HasMany, ManyToMany with UI
- **Plugins**: Boot interface, registry system
- **Jobs**: Background queue with SQLite persistence
- **Logger**: Structured logging (slog), rotation, request tracking
- **Mailer**: SMTP, LogMailer, HTML templates

### UI & Theming
- **Components**: 32+ atomic components, 6 layouts
- **Features**: Dark mode, responsive, Material Icons, Tailwind CSS
- **Reactivity**: Datastar SSE (11KB), replaces HTMX+Alpine.js
- **Widgets**: Stats, Charts, Grid, Timeline, Progress, Table, List

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
| `engine` | Core panel, CRUD handlers, multi-tenancy, relations (13 interfaces) |
| `form` | Form builder (22 fields), layouts, validation, live SSE validation |
| `table` | Table builder (13 columns + 4 inline), filters, summaries, grouping |
| `actions` | Actions (8 built-in), modals, lifecycle hooks, rate limiting |
| `infolist` | Read-only detail views (12 entry types) |
| `enum` | Generic enum helpers (10 helpers, type-safe) |
| `auth` | Authentication, sessions, roles, permissions, MFA/TOTP |
| `middleware` | HTTP middlewares (auth, CORS, CSRF, recovery, rate limit) |
| `notifications` | Notifications (memory + database stores), SSE streaming |
| `color` | Dynamic color palettes, CSS variables, Tailwind integration |
| `widget` | Dashboard widgets (Stats, Charts, Grid, Timeline, Progress, Table, List) |
| `search` | Global search with scoring, registry, QuickSearch interface |
| `export` | CSV / Excel export with struct tags |
| `importer` | CSV import with validation |
| `jobs` | Background job queue with SQLite persistence |
| `validation` | Input validation (go-playground/validator + custom) |
| `flash` | Session-based flash messages |
| `apperrors` | Structured errors with HTTP handlers |
| `logger` | Structured logging (slog) with rotation |
| `mailer` | SMTP + LogMailer with HTML templates |
| `plugin` | Plugin system with Boot interface |
| `datastar` | SSE SDK for Go (11KB, replaces HTMX+Alpine.js) |
| `ui` | 32+ Templ UI components and 6 layouts |
| `views` | Generic views (forms, tables, modals, widgets) |
| `cmd/sublimego` | CLI (make:resource, make:page, make:widget, make:enum, make:action) |

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
        form.Select("status").Label("Status").Options(map[string]string{
            "draft":     "Draft",
            "published": "Published",
        }),
        form.Tags("tags").Label("Tags"),
        form.ColorPicker("color").Label("Color"),
        form.MarkdownEditor("specs").Label("Specifications"),
        form.Slider("price").Label("Price").Min(0).Max(1000),
    )
    return views.GenericForm(f)
}

func (r *ProductResource) Table(ctx context.Context) templ.Component {
    t := table.New(nil).
        WithColumns(
            table.Text("name").WithLabel("Name").Sortable().Searchable(),
            table.Badge("status").WithLabel("Status").Colors(map[string]string{
                "draft": "gray", "published": "green",
            }),
            table.Date("created_at").WithLabel("Created").Sortable(),
            table.Tags("tags").WithLabel("Tags"),
            table.ViewColumn("view").WithLabel("View").URL(func(item any) string {
                return fmt.Sprintf("/admin/products/%d", item.(*Product).ID)
            }),
            table.TextInputColumn("price").WithLabel("Price").PatchURL(func(item any) string {
                return fmt.Sprintf("/admin/products/%d", item.(*Product).ID)
            }),
        ).
        WithFilters(
            table.SelectFilter("status").Label("Status").Options([]table.FilterOption{
                {Value: "draft", Label: "Draft"},
                {Value: "published", Label: "Published"},
            }),
            table.DateFilter("created_at").Label("Created"),
        ).
        WithSummaries(
            table.NewSummary("price", table.SummarySum).WithLabel("Total").WithFormat("$%.2f"),
        ).
        WithGroups(
            table.GroupBy("status").WithLabel("By status").Collapsible(),
        ).
        EnableColumnManager()
    return views.GenericTable(t)
}

func (r *ProductResource) View(ctx context.Context, item any) templ.Component {
    p := item.(*Product)
    il := infolist.New().
        AddSection("Basic Info", 2,
            infolist.TextEntry("name", "Name", p.Name),
            infolist.BadgeEntry("status", "Status", p.Status, "green"),
        ).
        AddSection("Details", 1,
            infolist.CodeEntry("specs", "Specifications", p.Specs, "markdown"),
            infolist.KeyValueEntry("metadata", "Metadata", 
                infolist.KeyValuePair{Key: "Created", Value: p.CreatedAt.Format("2006-01-02")},
                infolist.KeyValuePair{Key: "Updated", Value: p.UpdatedAt.Format("2006-01-02")},
            ),
        )
    return views.GenericInfolist(il)
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

### Framework Documentation
- **[ARCHITECTURE.md](ARCHITECTURE.md)** - Complete architecture overview
- **[AUDIT_FEVRIER_2026.md](AUDIT_FEVRIER_2026.md)** - Full audit report (95% Filament parity)
- **[CONTRIBUTING.md](CONTRIBUTING.md)** - Contribution guidelines

### Getting Started
```bash
# Install CLI
go install github.com/bozz33/sublimeadmin/cmd/sublimego@latest

# Create a new resource
sublimego make:resource Product

# Create an enum
sublimego make:enum Status

# Create an action
sublimego make:action BulkExport
```

### Examples & Guides
- **Resource Examples** - Complete CRUD resources with relations
- **Widget Examples** - Dashboard widgets and charts
- **Multi-tenancy** - Tenant-aware applications
- **Plugin Development** - Extending the framework

### Key Features Showcase
- **Live Validation** - Real-time form validation with Datastar SSE
- **Inline Editing** - Edit table cells without page reload
- **Relation Management** - BelongsTo, HasMany, ManyToMany with UI
- **Column Manager** - Toggle column visibility
- **Modal Actions** - Confirmation dialogs and form modals
- **Real-time Notifications** - SSE streaming with Datastar

---

## Contributing

Contributions are welcome! Please read [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

### How to Contribute
1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

---

## License

MIT - see [LICENSE](LICENSE).

---

## Status

**SublimeGo is production-ready**.
The framework is actively maintained and used in production environments.

- ✅ **22 form fields** with live validation
- ✅ **13 table columns** with inline editing
- ✅ **12 infolist entries** for detail views
- ✅ **8 built-in actions** with modal support
- ✅ **Multi-tenancy** support
- ✅ **Real-time notifications** with SSE
- ✅ **Complete CLI** for code generation
- ✅ **34 test files** with good coverage


