# sublime-admin — File Map

> Quick reference for every file in the framework. One line per file.

---

## Core Engine (`engine/`)

| File | Role |
|------|------|
| `engine/panel.go` | `Panel` struct, fluent builder (`Set*`), `Router()`, `Server()`, `ListenAndServe()` with secure timeouts |
| `engine/contract.go` | Core interfaces: `Resource`, `Column`, `Row`, `Field`, `FormState`, context helpers |
| `engine/base_resource.go` | `BaseResource` — default no-op implementations for all `Resource` methods |
| `engine/crud_handler.go` | HTTP handler for List / Create / Edit / Delete / Show actions |
| `engine/auth_handler.go` | Login / Register / Logout HTTP handler |
| `engine/auth_middleware.go` | `RequireAuth`, `RequireGuest` middleware wrappers |
| `engine/csrf_middleware.go` | CSRF token middleware integrated with the panel |
| `engine/page.go` | `Page` interface and `BasePage` for custom panel pages |
| `engine/page_handler.go` | HTTP handler for custom pages |
| `engine/identifiable.go` | `Identifiable` interface — `GetID()` helper for records |
| `engine/relations.go` | `Relation`, `RelationBuilder` (BelongsTo/HasOne/HasMany/ManyToMany), `RelationManager`, `RelationManagerHandler` |
| `engine/tenant.go` | `Tenant`, `SubdomainResolver`, `PathResolver`, `DynamicSubdomainResolver` (TTL cache), `MultiPanelRouter` (factory + cache), `NewIsolatedSession`, `TenantMiddleware`, `TenantAware` |
| `engine/perf.go` | `EnablePprof`, `PprofAuthMiddleware`, `ETagMiddleware`, `SecurityHeadersMiddleware` |
| `engine/export_import_handler.go` | `ExportHandler` (CSV/Excel), `ImportHandler` (CSV/Excel/JSON), `ResourceExportable`, `ResourceImportable` |
| `engine/password_reset_handler.go` | `/forgot-password` + `/reset-password` handler — in-memory token store, mailer integration |
| `engine/profile_handler.go` | `/profile` handler — update name/email, change password, no ORM required |
| `engine/doc.go` | Package documentation |

---

## Authentication (`auth/`)

| File | Role |
|------|------|
| `auth/manager.go` | `Manager` — session-based login/logout, user loading, `IsAuthenticated` |
| `auth/user.go` | `User` struct, permissions, roles, `ToMap`/`FromMap` for session storage |
| `auth/mfa.go` | TOTP-based MFA — `GenerateSecret`, `Validate`, `GenerateRecoveryCodes`, recovery code helpers |
| `auth/doc.go` | Package documentation |

---

## Form Builder (`form/`)

| File | Role |
|------|------|
| `form/contract.go` | `Component`, `Field`, `Layout` interfaces |
| `form/fields.go` | All field types: Text, Textarea, Select, Checkbox, Toggle, FileUpload, Repeater, RichEditor, Markdown, Tags, KeyValue, ColorPicker, Slider |
| `form/layout.go` | Layout components: `Section`, `Grid`, `Tabs` |

---

## Table Builder (`table/`)

| File | Role |
|------|------|
| `table/columns.go` | Column types: Text, Badge, Image, Boolean, Date (with relative time) |
| `table/bulk_actions.go` | `BulkAction`, `BulkDelete`, `BulkExport` |
| `table/table.go` | `Table` struct, pagination, filters, `WithBulkActions` |

---

## Middleware (`middleware/`)

| File | Role |
|------|------|
| `middleware/middleware.go` | `Middleware` type, `Stack`, `Chain`, `Conditional`, `SkipPaths`, `Group` |
| `middleware/auth.go` | `RequireAuth`, `RequirePermission`, `RequireRole`, `LoadUser`, `Auth`, `Verified` |
| `middleware/csrf.go` | `CSRFManager`, `CSRF()` middleware, `GetCSRFToken`, `CSRFField` |
| `middleware/cors.go` | CORS middleware with configurable origins/methods/headers |
| `middleware/flash.go` | Flash message middleware — injects `flash.Manager` into context |
| `middleware/logger.go` | HTTP request logger middleware |
| `middleware/recovery.go` | Panic recovery middleware |
| `middleware/ratelimit.go` | Token-bucket rate limiting per IP |
| `middleware/session.go` | Session loading middleware helper |

---

## Features

| Package | File | Role |
|---------|------|------|
| `flash/` | `flash/flash.go` | Flash messages: Success, Error, Warning, Info — session-backed |
| `export/` | `export/exporter.go` | CSV/Excel export — `FromStructs`, `Write`, `GenerateFilename` |
| `importer/` | `importer/importer.go` | CSV/Excel/JSON import — `ImportFromFile`, `MapToStruct`, hooks |
| `notifications/` | `notifications/notification.go` | In-memory notification store, fluent builders, SSE `MarshalSSE` |
| `notifications/` | `notifications/broadcast.go` | `Broadcaster` — SSE fan-out per user, heartbeat, `BroadcastAll` |
| `notifications/` | `notifications/handler.go` | HTTP handler: list, unread, stream (SSE), mark-read, `NotificationStore` interface |
| `hooks/` | `hooks/hooks.go` | UI render hooks — 10 positions, `Register`, `Render`, `Has`, `Clear` |
| `infolist/` | `infolist/infolist.go` | Read-only detail view — Text, Badge, Boolean, Date, Image, Color entries |
| `search/` | `search/search.go` | Global fuzzy search — `Searchable` interface, `GlobalSearch`, `CalculateScore`, `HighlightMatch` |
| `plugin/` | `plugin/plugin.go` | Plugin registry — `Register`, `Boot`, `Get`, `All` |
| `mailer/` | `mailer/mailer.go` | `Mailer` interface, `SMTPMailer`, `LogMailer`, `NoopMailer` |
| `widget/` | `widget/` | Dashboard widgets: StatsCard, Chart, Badge, etc. |
| `jobs/` | `jobs/` | Background jobs with persistence |
| `validation/` | `validation/` | Input validation helpers |
| `logger/` | `logger/` | Structured logging (`log/slog` wrapper) |
| `apperrors/` | `apperrors/` | Typed application errors |
| `actions/` | `actions/action.go` | Resource actions: Edit, Delete, View, Create, Export, Import, Restore, ForceDelete |
| `registry/` | `registry/` | Resource registry for auto-discovery |

---

## UI (`ui/`)

| Directory | Role |
|-----------|------|
| `ui/layouts/` | Base layouts: sidebar, topbar, nav groups |
| `ui/components/` | Reusable components: table, form, action buttons, alerts |
| `ui/assets/` | Static assets (CSS, JS) served at `/assets/` |

---

## Multi-Tenancy Quick Reference

```go
// Static tenants (known at startup)
resolver := engine.NewSubdomainResolver("example.com")
resolver.Register(&engine.Tenant{ID: "acme", Subdomain: "acme", Name: "Acme Corp"})

// Dynamic tenants (from DB, with TTL cache)
resolver := engine.NewDynamicSubdomainResolver("example.com",
    func(ctx context.Context, slug string) (*engine.Tenant, error) {
        return db.FindTenantBySlug(ctx, slug)
    },
    5*time.Minute,
)

// Router with factory + cache
router := engine.NewMultiPanelRouter(resolver).
    WithPanelCache(10 * time.Minute).
    WithPanelFactory(func(ctx context.Context, t *engine.Tenant) (*engine.Panel, error) {
        return buildTenantPanel(t), nil
    })

// Session isolation (prevents cross-tenant cookie collisions)
session := engine.NewIsolatedSession(tenant.ID, myStore)
panel.SetSession(session)

// Secure server startup
log.Fatal(panel.ListenAndServe(":8080"))
```

---

## Security Checklist

- [ ] Use `engine.NewIsolatedSession(slug, store)` for each tenant panel
- [ ] Call `multiRouter.validateSessionIsolation()` at startup
- [ ] Use `panel.ListenAndServe()` (not `http.ListenAndServe`) — includes timeouts
- [ ] Add `engine.SecurityHeadersMiddleware` to your handler chain
- [ ] Protect `/debug/pprof/` with `engine.PprofAuthMiddleware(token)`
- [ ] Enable `middleware.NewCSRFManager().CSRF` on all POST forms
- [ ] Use `auth.NewMFA(nil)` for TOTP-based 2FA on sensitive panels
