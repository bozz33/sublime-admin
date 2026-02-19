# Complete Guide — sublime-admin (sublimego-core)

> Exhaustive documentation based on the actual source code of `github.com/bozz33/sublimeadmin`.

---

## Table of Contents

1. [Installation](#1-installation)
2. [General Architecture](#2-general-architecture)
3. [Panel — Entry Point](#3-panel--entry-point)
4. [Resource Registration — Manual or Automatic?](#4-resource-registration--manual-or-automatic)
5. [Resource — Automatic CRUD](#5-resource--automatic-crud)
6. [Form — Forms](#6-form--forms)
7. [Table — Data Tables](#7-table--data-tables)
8. [Actions](#8-actions)
9. [Auth — Authentication](#9-auth--authentication)
10. [Middleware](#10-middleware)
11. [Flash Messages](#11-flash-messages)
12. [Export CSV / Excel](#12-export-csv--excel)
13. [Custom Pages](#13-custom-pages)
14. [Complete End-to-End Example](#14-complete-end-to-end-example)

---

## 1. Installation

```bash
go get github.com/bozz33/sublimeadmin
```

Key dependencies: `alexedwards/scs/v2`, `a-h/templ`, `samber/lo`, `xuri/excelize/v2`.

---

## 2. General Architecture

```
sublime-admin
├── engine/     → Panel, Resource, Page (framework core)
├── form/       → Form fields + layouts
├── table/      → Columns, filters, table actions
├── actions/    → Row actions (Edit, Delete, View, custom)
├── auth/       → User, Manager, sessions
├── middleware/ → RequireAuth, RequireRole, CORS, Rate limit, Recovery...
├── flash/      → Flash messages (success, error, warning, info)
├── export/     → CSV and Excel
├── apperrors/  → Structured HTTP errors
└── validation/ → go-playground/validator wrapper
```

**Flow:** Create a `Panel` → implement `Resource` interfaces → **manually register** them → `panel.Router()` returns a complete `http.Handler`.

---

## 3. Panel — Entry Point

```go
session := scs.New()
session.Lifetime = 24 * time.Hour
authManager := auth.NewManager(session)

panel := engine.NewPanel("admin").
    SetPath("/admin").
    SetBrandName("My App").
    SetDatabase(myDB).
    SetAuthManager(authManager).
    SetSession(session).
    SetDashboardProvider(myDashboard).
    SetAuthTemplates(myAuthViews).
    SetDashboardTemplates(myDashViews)
```

### Required DatabaseClient Interface

```go
type DatabaseClient interface {
    FindUserByEmail(ctx context.Context, email string) (*auth.User, string, error)
    CreateUser(ctx context.Context, name, email, hashedPassword string) (*auth.User, error)
    UserExists(ctx context.Context, email string) (bool, error)
}
```

### Auto-generated Routes

| Route | Description |
|---|---|
| `GET /login` | Login page |
| `POST /login` | Login handler |
| `GET /register` | Registration page |
| `GET /logout` | Logout |
| `GET /` | Dashboard |
| `GET /{slug}` | Resource list |
| `GET /{slug}/create` | Create form |
| `POST /{slug}` | Create |
| `GET /{slug}/{id}/edit` | Edit form |
| `PUT /{slug}/{id}` | Update |
| `DELETE /{slug}/{id}` | Delete |

---

## 4. Resource Registration — Manual or Automatic?

### Short answer: **Manual in sublimego-core**

Registration is always **explicit**:

```go
panel.AddResources(
    &ProductResource{db: myDB},
    &UserResource{db: myDB},
    &OrderResource{db: myDB},
)
panel.AddPages(&SettingsPage{})

http.ListenAndServe(":8080", panel.Router())
```

There is **no auto-discovery, no reflection scanning, no magic** in `sublimego-core`. You are in full control.

### What about the SublimeGo starter project?

The [SublimeGo](https://github.com/bozz33/sublimego) starter adds an **optional code generation layer**:

```go
// generate.go (SublimeGo only)
//go:generate go run ./cmd/scanner -dir=. -output=internal/registry/provider_gen.go
```

Running `go generate` scans your project for types implementing `engine.Resource` and generates a `provider_gen.go` that calls `panel.AddResources(...)` for you. But the underlying mechanism is still the same manual call — just generated.

### Comparison

| Feature | sublimego-core | SublimeGo starter |
|---|---|---|
| Resource registration | Manual (`AddResources`) | Manual + optional `go generate` |
| Auto-discovery | ❌ No | ✅ Via code generation |
| Code generation tool | ❌ No | ✅ `cmd/scanner` |

---

## 5. Resource — Automatic CRUD

### Full Interface

```go
type Resource interface {
    Slug() string           // URL identifier: "products"
    Label() string          // singular: "Product"
    PluralLabel() string    // plural: "Products"
    Icon() string           // icon: "package"
    Group() string          // nav group: "Catalogue" (empty = root)
    Sort() int              // nav order: 10

    Badge(ctx context.Context) string       // counter badge: "12"
    BadgeColor(ctx context.Context) string  // "danger", "warning"...

    Table(ctx context.Context) templ.Component
    Form(ctx context.Context, item any) templ.Component

    CanCreate(ctx context.Context) bool
    CanRead(ctx context.Context) bool
    CanUpdate(ctx context.Context) bool
    CanDelete(ctx context.Context) bool

    List(ctx context.Context) ([]any, error)
    Get(ctx context.Context, id string) (any, error)
    Create(ctx context.Context, r *http.Request) error
    Update(ctx context.Context, id string, r *http.Request) error
    Delete(ctx context.Context, id string) error
    BulkDelete(ctx context.Context, ids []string) error
}
```

### Example — ProductResource

```go
type ProductResource struct{ db ProductStore }

func (r *ProductResource) Slug() string        { return "products" }
func (r *ProductResource) Label() string       { return "Product" }
func (r *ProductResource) PluralLabel() string { return "Products" }
func (r *ProductResource) Icon() string        { return "package" }
func (r *ProductResource) Group() string       { return "Catalogue" }
func (r *ProductResource) Sort() int           { return 10 }
func (r *ProductResource) Badge(ctx context.Context) string      { return "" }
func (r *ProductResource) BadgeColor(ctx context.Context) string { return "" }

func (r *ProductResource) CanCreate(ctx context.Context) bool {
    return auth.UserFromContext(ctx).Can("products.create")
}
func (r *ProductResource) CanRead(ctx context.Context) bool   { return true }
func (r *ProductResource) CanUpdate(ctx context.Context) bool {
    return auth.UserFromContext(ctx).Can("products.update")
}
func (r *ProductResource) CanDelete(ctx context.Context) bool {
    return auth.UserFromContext(ctx).IsAdmin()
}

func (r *ProductResource) Table(ctx context.Context) templ.Component {
    t := table.New(nil).
        WithColumns(
            table.Text("Name").Label("Name").Sortable().Searchable(),
            table.Text("Price").Label("Price").Sortable(),
            table.Badge("Status").Label("Status").Colors(map[string]string{
                "active": "success", "inactive": "danger",
            }),
        ).
        SetActions(
            actions.EditAction("/products"),
            actions.DeleteAction("/products"),
        )
    return renderTable(ctx, t)
}

func (r *ProductResource) Form(ctx context.Context, item any) templ.Component {
    f := form.New().SetSchema(
        form.NewSection("Information").Schema(
            form.NewGrid(2).Schema(
                form.Text("Name").Label("Name").Required(),
                form.Number("Price").Label("Price").Required(),
            ),
            form.Textarea("Description").Label("Description").Rows(4),
            form.Toggle("Active").Label("Active").Default(true),
        ),
    )
    if item != nil { f.Bind(item) }
    return renderForm(ctx, f)
}

func (r *ProductResource) List(ctx context.Context) ([]any, error) {
    products, err := r.db.All(ctx)
    result := make([]any, len(products))
    for i, p := range products { result[i] = p }
    return result, err
}
func (r *ProductResource) Get(ctx context.Context, id string) (any, error) {
    return r.db.FindByID(ctx, id)
}
func (r *ProductResource) Create(ctx context.Context, req *http.Request) error {
    req.ParseForm()
    price, _ := strconv.ParseFloat(req.FormValue("Price"), 64)
    return r.db.Create(ctx, &Product{Name: req.FormValue("Name"), Price: price})
}
func (r *ProductResource) Update(ctx context.Context, id string, req *http.Request) error {
    req.ParseForm()
    price, _ := strconv.ParseFloat(req.FormValue("Price"), 64)
    return r.db.Update(ctx, id, &Product{Name: req.FormValue("Name"), Price: price})
}
func (r *ProductResource) Delete(ctx context.Context, id string) error {
    return r.db.Delete(ctx, id)
}
func (r *ProductResource) BulkDelete(ctx context.Context, ids []string) error {
    for _, id := range ids {
        if err := r.db.Delete(ctx, id); err != nil { return err }
    }
    return nil
}
```

---

## 6. Form — Forms

### Text Fields

```go
form.Text("name").Label("Name").Required().Placeholder("e.g. iPhone 15").Default("value")
form.Email("email").Label("Email").Required()
form.Password("password").Label("Password").Required()
form.Number("price").Label("Price").Default(0)
form.Textarea("description").Label("Description").Rows(5)
```

### Select

```go
form.Select("status").
    Label("Status").
    SetOptions(map[string]string{"active": "Active", "inactive": "Inactive"}).
    Default("active").Required()
```

### Booleans

```go
form.Checkbox("terms").Label("I accept the Terms").Default(false)
form.Toggle("active").Label("Active").Labels("Yes", "No").Default(true)
```

### File Upload

```go
form.FileUpload("avatar").
    Label("Profile Picture").
    Accept("image/jpeg,image/png").
    MaxSize(5 * 1024 * 1024).
    Multiple()
```

### Rich Editors

```go
form.RichEditor("content").
    Label("Content").
    WithToolbar("bold", "italic", "link", "heading", "list", "image", "code").
    WithMaxLength(10000).Required()

form.MarkdownEditor("body").Label("Body").Rows(20).Required()
```

### Tags

```go
form.Tags("tags").
    Label("Tags").
    WithSuggestions("go", "web", "api").
    WithMaxTags(10).WithSeparator(",").
    Default([]string{"go"})
```

### Key-Value

```go
form.KeyValue("metadata").
    Label("Metadata").
    WithLabels("Key", "Value").
    WithMaxPairs(20).
    AddButtonLabel("Add entry")
```

### Color Picker

```go
form.ColorPicker("color").
    Label("Color").
    WithSwatches("#ef4444", "#22c55e", "#3b82f6").
    Default("#3b82f6")
```

### Slider

```go
form.Slider("discount").
    Label("Discount").Range(0, 100).WithStep(5).WithUnit("%").Default(0)
```

### Repeater

```go
form.Repeater("addresses", subFields...).
    Label("Addresses").Min(1).Max(5).AddButtonLabel("Add address")
```

### Layouts

```go
// Section
form.NewSection("Title").Desc("Description").Collapsible().Schema(field1, field2)

// Grid
form.NewGrid(2).Schema(field1, field2)
form.NewGrid(3).Schema(field1, field2, field3)

// Tabs
form.NewTabs().
    AddTab("General", field1, field2).
    AddTab("Security", field3, field4)
```

### Full Form with Nested Layouts

```go
f := form.New().SetSchema(
    form.NewSection("Identity").Schema(
        form.NewGrid(2).Schema(
            form.Text("first_name").Label("First Name").Required(),
            form.Text("last_name").Label("Last Name").Required(),
        ),
        form.Email("email").Label("Email").Required(),
    ),
    form.NewSection("Options").Collapsible().Schema(
        form.NewTabs().
            AddTab("Access",
                form.Select("role").Label("Role").SetOptions(roles),
                form.Toggle("active").Label("Active").Default(true),
            ).
            AddTab("Appearance",
                form.ColorPicker("color").Label("Color"),
                form.Slider("font_size").Label("Font Size").Range(12, 24).WithUnit("px"),
            ),
    ),
)
f.Bind(existingUser) // pre-fill in edit mode
```

---

## 7. Table — Data Tables

```go
table.Text("Name").Label("Name").Sortable().Searchable().Copyable()
table.Badge("Status").Label("Status").Colors(map[string]string{
    "active": "success", "pending": "warning", "banned": "danger",
})
table.Image("Avatar").Label("Photo").Round()

t := table.New(items).
    WithColumns(col1, col2, col3).
    SetActions(actions.EditAction("/url"), actions.DeleteAction("/url")).
    WithFilters(filter1).
    Search(true).Paginate(true).
    SetBaseURL("/products")
```

---

## 8. Actions

```go
// Predefined
actions.EditAction("/products")
actions.DeleteAction("/products")
actions.ViewAction("/products")

// Custom
actions.New("publish").
    SetLabel("Publish").SetIcon("check-circle").SetColor("success").
    SetUrl(func(item any) string {
        return "/products/" + actions.GetItemID(item) + "/publish"
    })

// With confirmation modal
actions.New("archive").
    SetLabel("Archive").SetIcon("archive").SetColor("warning").
    RequiresDialog("Archive this product?", "This action can be undone.").
    SetUrl(func(item any) string {
        return "/products/" + actions.GetItemID(item) + "/archive"
    })
```

Colors: `"primary"` `"secondary"` `"success"` `"danger"` `"warning"` `"info"` `"gray"`

```go
// Identifiable interface for GetItemID
func (p *Product) GetID() int { return p.ID }
```

---

## 9. Auth — Authentication

```go
session := scs.New()
authManager := auth.NewManager(session)

user := auth.NewUser(1, "alice@example.com", "Alice")
user.AddRole(auth.RoleAdmin)
user.AddPermission("products.create")
user.SetMetadata("email_verified", true)

// User checks
user.IsAdmin() / user.IsSuperAdmin() / user.IsAuthenticated() / user.IsGuest()
user.Can("perm") / user.Cannot("perm")
user.HasAnyPermission("p1", "p2") / user.HasAllPermissions("p1", "p2")
user.HasRole("admin") / user.HasAnyRole("admin", "moderator")

// Manager
authManager.Login(ctx, user)
authManager.Logout(ctx)
user, err := authManager.UserFromRequest(r)
authManager.IsAuthenticatedFromRequest(r)
authManager.Can(ctx, "products.create")
authManager.IsAdmin(ctx)
authManager.SetIntendedURLFromRequest(r)
url := authManager.IntendedURL(ctx, "/")

// Context helpers in handlers
auth.CurrentUser(r) / auth.IsAuthenticated(r) / auth.IsAdmin(r)
auth.Can(r, "products.create") / auth.HasRole(r, "admin")
```

Constants: `auth.RoleGuest / RoleUser / RoleModerator / RoleAdmin / RoleSuperAdmin`
`auth.PermissionUsersView / PermissionUsersCreate / PermissionAdminAccess`

---

## 10. Middleware

```go
// Auth
middleware.RequireAuth(authManager)
middleware.RequireGuest(authManager, "/dashboard")
middleware.LoadUser(authManager)
middleware.RequirePermission(authManager, "products.create")
middleware.RequirePermissions(authManager, "p1", "p2")
middleware.RequireAnyPermission(authManager, "p1", "p2")
middleware.RequireRole(authManager, "admin")
middleware.RequireAdmin(authManager)
middleware.RequireSuperAdmin(authManager)
middleware.Verified(authManager, "/verify-email")

// Other
middleware.Logger(logger)
middleware.Recovery(errorHandler)
middleware.CORS(config)
middleware.RateLimit(100, time.Minute)
middleware.Flash(flashManager)
middleware.Session(sessionManager)

// Composition
stack := middleware.NewStack(middleware.Logger(logger), middleware.RequireAuth(authManager))
mux.Handle("/admin/", stack.Then(handler))

protected := middleware.Chain(middleware.RequireAuth(authManager), middleware.RequireAdmin(authManager))
mux.Handle("/admin/", protected(handler))

middleware.Apply(handler, middleware.RequireAuth(authManager), middleware.RateLimit(60, time.Minute))
middleware.SkipPaths([]string{"/health"}, middleware.Logger(logger))
middleware.Conditional(func(r *http.Request) bool { return true }, middleware.RequireAuth(authManager))

group := middleware.NewGroup(middleware.RequireAuth(authManager))
group.HandleFunc("/admin/users", usersHandler, mux)
```

---

## 11. Flash Messages

```go
flashManager := flash.NewManager(session)

// Add
flashManager.Success(ctx, "Created successfully")
flashManager.Error(ctx, "An error occurred")
flashManager.Warning(ctx, "Low stock")
flashManager.Info(ctx, "Update available")
flashManager.SuccessWithTitle(ctx, "Success", "Product created")

// From request
flashManager.SuccessFromRequest(r, "Saved!")
flash.Success(r, "Saved!")  // global helper (manager in context)

// Read
messages := flashManager.GetAndClear(ctx)
flashManager.Has(ctx) / flashManager.HasType(ctx, flash.TypeSuccess)
flashManager.Count(ctx) / flashManager.GetByType(ctx, flash.TypeError)
```

Types: `flash.TypeSuccess` `flash.TypeError` `flash.TypeWarning` `flash.TypeInfo`

---

## 12. Export CSV / Excel

```go
// Fluent API
export.New(export.FormatCSV).
    SetHeaders([]string{"ID", "Name", "Price"}).
    AddRow([]string{"1", "iPhone", "999.00"}).
    Write(w)

// From structs (export tags)
type Product struct {
    ID    int     `export:"ID"`
    Name  string  `export:"Name"`
    Price float64 `export:"Price"`
    Token string  `export:"-"` // ignored
}
export.ExportStructsCSV(w, products)
export.ExportStructsExcel(w, products)

// Quick helpers
export.QuickExportCSV(w, headers, rows)
export.QuickExportExcel(w, headers, rows)

// In HTTP handler
func ExportHandler(w http.ResponseWriter, r *http.Request) {
    format := export.FormatCSV
    w.Header().Set("Content-Type", export.GetContentType(format))
    w.Header().Set("Content-Disposition",
        "attachment; filename="+export.GenerateFilename("products", format))
    export.ExportStructsCSV(w, products)
}
```

---

## 13. Custom Pages

```go
type Page interface {
    Slug() string
    Label() string
    Icon() string
    Group() string
    Sort() int
    Handle(w http.ResponseWriter, r *http.Request)
}

type SettingsPage struct{}

func (p *SettingsPage) Slug() string  { return "settings" }
func (p *SettingsPage) Label() string { return "Settings" }
func (p *SettingsPage) Icon() string  { return "cog" }
func (p *SettingsPage) Group() string { return "" }
func (p *SettingsPage) Sort() int     { return 99 }
func (p *SettingsPage) Handle(w http.ResponseWriter, r *http.Request) {
    settingsView().Render(r.Context(), w)
}

panel.AddPages(&SettingsPage{})
```

---

## 14. Complete End-to-End Example

```go
package main

import (
    "context"
    "net/http"
    "strconv"
    "time"

    "github.com/alexedwards/scs/v2"
    "github.com/a-h/templ"
    "github.com/bozz33/sublimeadmin/actions"
    "github.com/bozz33/sublimeadmin/auth"
    "github.com/bozz33/sublimeadmin/engine"
    "github.com/bozz33/sublimeadmin/form"
    "github.com/bozz33/sublimeadmin/table"
)

func main() {
    session := scs.New()
    session.Lifetime = 24 * time.Hour

    authManager := auth.NewManager(session)
    db := &MyDatabase{}

    panel := engine.NewPanel("admin").
        SetPath("/admin").
        SetBrandName("My Admin").
        SetDatabase(db).
        SetAuthManager(authManager).
        SetSession(session)

    // Registration is MANUAL — list every resource explicitly
    panel.AddResources(
        &ArticleResource{db: db},
        &UserResource{db: db},
    )
    panel.AddPages(&SettingsPage{})

    http.ListenAndServe(":8080", panel.Router())
}

type Article struct {
    ID      int
    Title   string
    Content string
    Status  string
}

type ArticleResource struct{ db *MyDatabase }

func (r *ArticleResource) Slug() string        { return "articles" }
func (r *ArticleResource) Label() string       { return "Article" }
func (r *ArticleResource) PluralLabel() string { return "Articles" }
func (r *ArticleResource) Icon() string        { return "file-text" }
func (r *ArticleResource) Group() string       { return "Content" }
func (r *ArticleResource) Sort() int           { return 1 }
func (r *ArticleResource) Badge(ctx context.Context) string      { return "" }
func (r *ArticleResource) BadgeColor(ctx context.Context) string { return "" }

func (r *ArticleResource) CanCreate(ctx context.Context) bool { return true }
func (r *ArticleResource) CanRead(ctx context.Context) bool   { return true }
func (r *ArticleResource) CanUpdate(ctx context.Context) bool { return true }
func (r *ArticleResource) CanDelete(ctx context.Context) bool {
    return auth.UserFromContext(ctx).IsAdmin()
}

func (r *ArticleResource) Table(ctx context.Context) templ.Component {
    t := table.New(nil).
        WithColumns(
            table.Text("Title").Label("Title").Sortable().Searchable(),
            table.Badge("Status").Label("Status").Colors(map[string]string{
                "published": "success",
                "draft":     "warning",
                "archived":  "gray",
            }),
        ).
        SetActions(
            actions.EditAction("/articles"),
            actions.New("publish").
                SetLabel("Publish").SetIcon("check").SetColor("success").
                SetUrl(func(item any) string {
                    return "/articles/" + actions.GetItemID(item) + "/publish"
                }),
            actions.DeleteAction("/articles"),
        )
    return renderTable(ctx, t)
}

func (r *ArticleResource) Form(ctx context.Context, item any) templ.Component {
    f := form.New().SetSchema(
        form.NewSection("Content").Schema(
            form.Text("Title").Label("Title").Required(),
            form.RichEditor("Content").Label("Content").Required(),
        ),
        form.NewSection("Publishing").Collapsible().Schema(
            form.NewGrid(2).Schema(
                form.Select("Status").Label("Status").SetOptions(map[string]string{
                    "draft": "Draft", "published": "Published", "archived": "Archived",
                }).Default("draft"),
                form.Tags("Tags").Label("Tags").WithSuggestions("go", "web", "tutorial"),
            ),
        ),
    )
    if item != nil { f.Bind(item) }
    return renderForm(ctx, f)
}

func (r *ArticleResource) List(ctx context.Context) ([]any, error) {
    articles, err := r.db.AllArticles(ctx)
    result := make([]any, len(articles))
    for i, a := range articles { result[i] = a }
    return result, err
}
func (r *ArticleResource) Get(ctx context.Context, id string) (any, error) {
    return r.db.FindArticle(ctx, id)
}
func (r *ArticleResource) Create(ctx context.Context, req *http.Request) error {
    req.ParseForm()
    return r.db.CreateArticle(ctx, &Article{
        Title: req.FormValue("Title"), Content: req.FormValue("Content"),
        Status: req.FormValue("Status"),
    })
}
func (r *ArticleResource) Update(ctx context.Context, id string, req *http.Request) error {
    req.ParseForm()
    return r.db.UpdateArticle(ctx, id, &Article{
        Title: req.FormValue("Title"), Content: req.FormValue("Content"),
        Status: req.FormValue("Status"),
    })
}
func (r *ArticleResource) Delete(ctx context.Context, id string) error {
    return r.db.DeleteArticle(ctx, id)
}
func (r *ArticleResource) BulkDelete(ctx context.Context, ids []string) error {
    for _, id := range ids {
        if err := r.db.DeleteArticle(ctx, id); err != nil { return err }
    }
    return nil
}
```
