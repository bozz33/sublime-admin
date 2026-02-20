# Guide complet — sublime-admin (sublimego-core)

> Documentation exhaustive basée sur le code source réel du package `github.com/bozz33/sublimeadmin`.

---

## Table des matières

1. [Installation](#1-installation)
2. [Architecture générale](#2-architecture-générale)
3. [Panel — le point d'entrée](#3-panel--le-point-dentrée)
4. [Resource — CRUD automatique](#4-resource--crud-automatique)
5. [Form — formulaires](#5-form--formulaires)
6. [Table — tableaux de données](#6-table--tableaux-de-données)
7. [Actions](#7-actions)
8. [Auth — authentification](#8-auth--authentification)
9. [Middleware](#9-middleware)
10. [Flash messages](#10-flash-messages)
11. [Export CSV / Excel](#11-export-csv--excel)
12. [Pages custom](#12-pages-custom)
13. [Exemple complet end-to-end](#13-exemple-complet-end-to-end)

---

## 1. Installation

```bash
go get github.com/bozz33/sublimeadmin
```

---

## 2. Architecture générale

```
sublime-admin
├── engine/     → Panel, Resource, Page (cœur du framework)
├── form/       → Champs de formulaire + layouts
├── table/      → Colonnes, filtres, actions de table
├── actions/    → Actions sur les lignes (Edit, Delete, View, custom)
├── auth/       → User, Manager, sessions
├── middleware/ → RequireAuth, RequireRole, CORS, Rate limit, Recovery...
├── flash/      → Messages flash (success, error, warning, info)
├── export/     → CSV et Excel
├── apperrors/  → Erreurs HTTP structurées
└── validation/ → go-playground/validator wrapper
```

Flux : Panel → Resources → `panel.Router()` → `http.Handler` complet avec toutes les routes CRUD.

---

## 3. Panel — le point d'entrée

### Création

```go
session := scs.New()
session.Lifetime = 24 * time.Hour
authManager := auth.NewManager(session)

panel := engine.NewPanel("admin").
    SetPath("/admin").
    SetBrandName("Mon App").
    SetDatabase(myDB).
    SetAuthManager(authManager).
    SetSession(session).
    SetDashboardProvider(myDashboard).
    SetAuthTemplates(myAuthViews).
    SetDashboardTemplates(myDashViews)
```

### Interface DatabaseClient (obligatoire)

```go
type DatabaseClient interface {
    FindUserByEmail(ctx context.Context, email string) (*auth.User, string, error)
    CreateUser(ctx context.Context, name, email, hashedPassword string) (*auth.User, error)
    UserExists(ctx context.Context, email string) (bool, error)
}
```

### Enregistrement et démarrage

```go
panel.AddResources(
    &ProductResource{db: myDB},
    &UserResource{db: myDB},
)
panel.AddPages(&SettingsPage{})

http.ListenAndServe(":8080", panel.Router())
```

### Routes générées automatiquement

| Route | Description |
|---|---|
| `GET /login` | Page de connexion |
| `POST /login` | Traitement connexion |
| `GET /register` | Inscription |
| `GET /logout` | Déconnexion |
| `GET /` | Dashboard |
| `GET /{slug}` | Liste |
| `GET /{slug}/create` | Formulaire création |
| `POST /{slug}` | Création |
| `GET /{slug}/{id}/edit` | Formulaire édition |
| `PUT /{slug}/{id}` | Mise à jour |
| `DELETE /{slug}/{id}` | Suppression |

---

## 4. Resource — CRUD automatique

### Interface complète

```go
type Resource interface {
    Slug() string
    Label() string
    PluralLabel() string
    Icon() string
    Group() string          // groupe nav ("Catalogue"), vide = racine
    Sort() int              // ordre dans la nav

    Badge(ctx context.Context) string       // badge compteur "12"
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

### Exemple — ProductResource

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
            table.Text("Name").Label("Nom").Sortable().Searchable(),
            table.Text("Price").Label("Prix").Sortable(),
            table.Badge("Status").Label("Statut").Colors(map[string]string{
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
        form.NewSection("Informations").Schema(
            form.NewGrid(2).Schema(
                form.Text("Name").Label("Nom").Required(),
                form.Number("Price").Label("Prix").Required(),
            ),
            form.Textarea("Description").Label("Description").Rows(4),
            form.Toggle("Active").Label("Actif").Default(true),
        ),
    )
    if item != nil {
        f.Bind(item)
    }
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

## 5. Form — formulaires

### Champs texte

```go
form.Text("name").Label("Nom").Required().Placeholder("ex: iPhone 15").Default("valeur")
form.Email("email").Label("Email").Required()
form.Password("password").Label("Mot de passe").Required()
form.Number("price").Label("Prix").Default(0)
form.Textarea("description").Label("Description").Rows(5).Required()
```

### Sélection

```go
form.Select("status").
    Label("Statut").
    SetOptions(map[string]string{
        "active":   "Actif",
        "inactive": "Inactif",
    }).
    Default("active").
    Required()
```

### Booléens

```go
form.Checkbox("terms").Label("J'accepte les CGU").Default(false)

form.Toggle("active").
    Label("Actif").
    Labels("Oui", "Non").
    Default(true)
```

### Fichier

```go
form.FileUpload("avatar").
    Label("Photo").
    Accept("image/jpeg,image/png").
    MaxSize(5 * 1024 * 1024).
    Multiple()
```

### Éditeurs riches

```go
form.RichEditor("content").
    Label("Contenu").
    WithToolbar("bold", "italic", "link", "heading", "list", "image", "code").
    WithMaxLength(10000).
    Required()

form.MarkdownEditor("body").
    Label("Corps").
    Rows(20).
    Required()
```

### Tags

```go
form.Tags("tags").
    Label("Tags").
    WithSuggestions("go", "web", "api").
    WithMaxTags(10).
    WithSeparator(",").
    Default([]string{"go"})
```

### Clé-valeur

```go
form.KeyValue("metadata").
    Label("Métadonnées").
    WithLabels("Clé", "Valeur").
    WithMaxPairs(20).
    AddButtonLabel("Ajouter")
```

### Color picker

```go
form.ColorPicker("color").
    Label("Couleur").
    WithSwatches("#ef4444", "#22c55e", "#3b82f6").
    Default("#3b82f6")
```

### Slider

```go
form.Slider("discount").
    Label("Remise").
    Range(0, 100).
    WithStep(5).
    WithUnit("%").
    Default(0)
```

### Repeater

```go
form.Repeater("addresses", subFields...).
    Label("Adresses").
    Min(1).Max(5).
    AddButtonLabel("Ajouter une adresse")
```

### Layouts

```go
// Section
form.NewSection("Titre").
    Desc("Description").
    Collapsible().
    Schema(field1, field2)

// Grid colonnes
form.NewGrid(2).Schema(field1, field2)
form.NewGrid(3).Schema(field1, field2, field3)

// Tabs
form.NewTabs().
    AddTab("Général", field1, field2).
    AddTab("Sécurité", field3, field4)
```

### Formulaire complet avec layouts imbriqués

```go
f := form.New().SetSchema(
    form.NewSection("Identité").Schema(
        form.NewGrid(2).Schema(
            form.Text("first_name").Label("Prénom").Required(),
            form.Text("last_name").Label("Nom").Required(),
        ),
        form.Email("email").Label("Email").Required(),
    ),
    form.NewSection("Options").Collapsible().Schema(
        form.NewTabs().
            AddTab("Accès",
                form.Select("role").Label("Rôle").SetOptions(roles),
                form.Toggle("active").Label("Actif").Default(true),
            ).
            AddTab("Apparence",
                form.ColorPicker("color").Label("Couleur"),
                form.Slider("font_size").Label("Taille").Range(12, 24).WithUnit("px"),
            ),
    ),
)
f.Bind(existingUser) // pré-remplissage en édition
```

---

## 6. Table — tableaux de données

### Colonnes disponibles

```go
table.Text("Name").Label("Nom").Sortable().Searchable().Copyable()

table.Badge("Status").Label("Statut").Sortable().Colors(map[string]string{
    "active":  "success",
    "pending": "warning",
    "banned":  "danger",
})

table.Image("Avatar").Label("Photo").Round()
```

### Configuration table

```go
t := table.New(items).
    WithColumns(col1, col2, col3).
    SetActions(actions.EditAction("/url"), actions.DeleteAction("/url")).
    WithFilters(filter1, filter2).
    Search(true).
    Paginate(true).
    SetBaseURL("/products")
```

---

## 7. Actions

### Prédéfinies

```go
actions.EditAction("/products")    // → /products/{id}/edit
actions.DeleteAction("/products")  // modal de confirmation
actions.ViewAction("/products")    // → /products/{id}
```

### Personnalisée

```go
actions.New("publish").
    SetLabel("Publier").
    SetIcon("check-circle").
    SetColor("success").
    SetUrl(func(item any) string {
        return fmt.Sprintf("/products/%s/publish", actions.GetItemID(item))
    })
```

### Avec modal de confirmation

```go
actions.New("archive").
    SetLabel("Archiver").
    SetIcon("archive").
    SetColor("warning").
    RequiresDialog("Archiver ce produit ?", "Cette action est réversible.").
    SetUrl(func(item any) string {
        return fmt.Sprintf("/products/%s/archive", actions.GetItemID(item))
    })
```

### Couleurs : `"primary"` `"secondary"` `"success"` `"danger"` `"warning"` `"info"` `"gray"`

### Interface Identifiable

```go
func (p *Product) GetID() int { return p.ID }
// → actions.GetItemID(product) retourne "42"
```

---

## 8. Auth — authentification

### Initialisation

```go
session := scs.New()
authManager := auth.NewManager(session)
```

### User

```go
user := auth.NewUser(1, "alice@example.com", "Alice")
user.AddRole(auth.RoleAdmin)
user.AddPermission("products.create")
user.SetMetadata("email_verified", true)
```

### Vérifications User

```go
user.IsAdmin()           user.IsSuperAdmin()
user.IsAuthenticated()   user.IsGuest()
user.Can("perm")         user.Cannot("perm")
user.HasAnyPermission("p1", "p2")
user.HasAllPermissions("p1", "p2")
user.HasRole("admin")    user.HasAnyRole("admin", "mod")
```

### Manager

```go
authManager.Login(ctx, user)
authManager.LoginWithRequest(r, user)
authManager.Logout(ctx)
authManager.LogoutWithRequest(r)

user, err := authManager.UserFromRequest(r)
authManager.IsAuthenticatedFromRequest(r)
authManager.Can(ctx, "products.create")
authManager.HasRole(ctx, "admin")
authManager.IsAdmin(ctx)

authManager.SetIntendedURLFromRequest(r)
url := authManager.IntendedURL(ctx, "/")
```

### Helpers dans les handlers

```go
user := auth.CurrentUser(r)
auth.IsAuthenticated(r)
auth.IsGuest(r)
auth.Can(r, "products.create")
auth.HasRole(r, "admin")
auth.IsAdmin(r)
```

### Constantes

```go
// Rôles
auth.RoleGuest / RoleUser / RoleModerator / RoleAdmin / RoleSuperAdmin

// Permissions
auth.PermissionUsersView / PermissionUsersCreate / PermissionUsersUpdate / PermissionUsersDelete
auth.PermissionAdminAccess
```

---

## 9. Middleware

### Auth

```go
middleware.RequireAuth(authManager)
middleware.RequireGuest(authManager, "/dashboard")
middleware.LoadUser(authManager)          // sans forcer l'auth
middleware.RequirePermission(authManager, "products.create")
middleware.RequirePermissions(authManager, "p1", "p2")  // toutes requises
middleware.RequireAnyPermission(authManager, "p1", "p2")
middleware.RequireRole(authManager, "admin")
middleware.RequireAdmin(authManager)
middleware.RequireSuperAdmin(authManager)
middleware.Verified(authManager, "/verify-email")
```

### Autres

```go
middleware.Logger(logger)
middleware.Recovery(errorHandler)
middleware.CORS(config)
middleware.RateLimit(100, time.Minute)
middleware.Flash(flashManager)
middleware.Session(sessionManager)
```

### Composition

```go
// Stack
stack := middleware.NewStack(
    middleware.Logger(logger),
    middleware.RequireAuth(authManager),
)
mux.Handle("/admin/", stack.Then(handler))

// Chain (un seul middleware composé)
protected := middleware.Chain(
    middleware.RequireAuth(authManager),
    middleware.RequireAdmin(authManager),
)
mux.Handle("/admin/", protected(handler))

// Apply (helper direct)
mux.Handle("/api/", middleware.Apply(handler,
    middleware.RequireAuth(authManager),
    middleware.RateLimit(60, time.Minute),
))

// Conditionnel
middleware.Conditional(
    func(r *http.Request) bool { return r.Header.Get("X-API-Key") != "" },
    middleware.RequireAuth(authManager),
)

// Ignorer des chemins
middleware.SkipPaths([]string{"/health", "/metrics"}, middleware.Logger(logger))

// Group de routes
group := middleware.NewGroup(middleware.RequireAuth(authManager))
group.HandleFunc("/admin/users", usersHandler, mux)
group.HandleFunc("/admin/settings", settingsHandler, mux)
```

---

## 10. Flash messages

```go
flashManager := flash.NewManager(session)

// Ajouter
flashManager.Success(ctx, "Produit créé")
flashManager.Error(ctx, "Erreur")
flashManager.Warning(ctx, "Stock faible")
flashManager.Info(ctx, "Info")
flashManager.SuccessWithTitle(ctx, "Succès", "Produit créé avec succès")

// Depuis une request
flashManager.SuccessFromRequest(r, "Sauvegardé !")
flash.Success(r, "Sauvegardé !")  // helper global (manager dans contexte)

// Lire
messages := flashManager.GetAndClear(ctx)
flashManager.Has(ctx)
flashManager.HasType(ctx, flash.TypeSuccess)
flashManager.GetByType(ctx, flash.TypeError)
flashManager.Count(ctx)
```

---

## 11. Export CSV / Excel

### API fluent

```go
// CSV
err := export.New(export.FormatCSV).
    SetHeaders([]string{"ID", "Nom", "Prix"}).
    AddRow([]string{"1", "iPhone", "999.00"}).
    AddRows(moreRows).
    Write(w)

// Excel
err := export.New(export.FormatExcel).
    SetHeaders([]string{"ID", "Nom", "Prix"}).
    AddRows(rows).
    Write(w)
```

### Depuis des structs (via tags `export`)

```go
type Product struct {
    ID    int     `export:"ID"`
    Name  string  `export:"Nom"`
    Price float64 `export:"Prix"`
    Secret string `export:"-"` // ignoré
}

products := []Product{{1, "iPhone", 999.0, "secret"}}
export.ExportStructsCSV(w, products)
export.ExportStructsExcel(w, products)
```

### Helpers rapides

```go
export.QuickExportCSV(w, headers, rows)
export.QuickExportExcel(w, headers, rows)
```

### Dans un handler HTTP

```go
func ExportHandler(w http.ResponseWriter, r *http.Request) {
    format := export.FormatCSV
    filename := export.GenerateFilename("products", format)

    w.Header().Set("Content-Type", export.GetContentType(format))
    w.Header().Set("Content-Disposition", "attachment; filename="+filename)

    export.ExportStructsCSV(w, products)
}
```

---

## 12. Pages custom

### Interface Page

```go
type Page interface {
    Slug() string
    Label() string
    Icon() string
    Group() string
    Sort() int
    Handle(w http.ResponseWriter, r *http.Request)
}
```

### Exemple

```go
type SettingsPage struct{}

func (p *SettingsPage) Slug() string  { return "settings" }
func (p *SettingsPage) Label() string { return "Paramètres" }
func (p *SettingsPage) Icon() string  { return "cog" }
func (p *SettingsPage) Group() string { return "" }
func (p *SettingsPage) Sort() int     { return 99 }

func (p *SettingsPage) Handle(w http.ResponseWriter, r *http.Request) {
    // Ta logique custom
    settingsView().Render(r.Context(), w)
}

// Enregistrement
panel.AddPages(&SettingsPage{})
```

---

## 13. Exemple complet end-to-end

```go
package main

import (
    "context"
    "net/http"
    "time"

    "github.com/alexedwards/scs/v2"
    "github.com/bozz33/sublimeadmin/actions"
    "github.com/bozz33/sublimeadmin/auth"
    "github.com/bozz33/sublimeadmin/engine"
    "github.com/bozz33/sublimeadmin/flash"
    "github.com/bozz33/sublimeadmin/form"
    "github.com/bozz33/sublimeadmin/table"
)

func main() {
    // 1. Session
    session := scs.New()
    session.Lifetime = 24 * time.Hour

    // 2. Auth
    authManager := auth.NewManager(session)

    // 3. Flash
    flashManager := flash.NewManager(session)
    _ = flashManager

    // 4. DB (implémente engine.DatabaseClient)
    db := &MyDatabase{}

    // 5. Panel
    panel := engine.NewPanel("admin").
        SetPath("/admin").
        SetBrandName("Mon Admin").
        SetDatabase(db).
        SetAuthManager(authManager).
        SetSession(session)

    // 6. Resources
    panel.AddResources(
        &ArticleResource{db: db},
        &UserResource{db: db},
    )

    // 7. Démarrage
    http.ListenAndServe(":8080", panel.Router())
}

// --- ArticleResource ---

type Article struct {
    ID      int
    Title   string
    Content string
    Status  string
    Views   int
}

type ArticleResource struct{ db *MyDatabase }

func (r *ArticleResource) Slug() string        { return "articles" }
func (r *ArticleResource) Label() string       { return "Article" }
func (r *ArticleResource) PluralLabel() string { return "Articles" }
func (r *ArticleResource) Icon() string        { return "file-text" }
func (r *ArticleResource) Group() string       { return "Contenu" }
func (r *ArticleResource) Sort() int           { return 1 }

func (r *ArticleResource) Badge(ctx context.Context) string {
    return "" // ex: nombre de brouillons
}
func (r *ArticleResource) BadgeColor(ctx context.Context) string { return "warning" }

func (r *ArticleResource) CanCreate(ctx context.Context) bool { return true }
func (r *ArticleResource) CanRead(ctx context.Context) bool   { return true }
func (r *ArticleResource) CanUpdate(ctx context.Context) bool { return true }
func (r *ArticleResource) CanDelete(ctx context.Context) bool {
    return auth.UserFromContext(ctx).IsAdmin()
}

func (r *ArticleResource) Table(ctx context.Context) templ.Component {
    t := table.New(nil).
        WithColumns(
            table.Text("Title").Label("Titre").Sortable().Searchable(),
            table.Badge("Status").Label("Statut").Colors(map[string]string{
                "published": "success",
                "draft":     "warning",
                "archived":  "gray",
            }),
            table.Text("Views").Label("Vues").Sortable(),
        ).
        SetActions(
            actions.EditAction("/articles"),
            actions.New("publish").
                SetLabel("Publier").
                SetIcon("check").
                SetColor("success").
                SetUrl(func(item any) string {
                    return "/articles/" + actions.GetItemID(item) + "/publish"
                }),
            actions.DeleteAction("/articles"),
        )
    return renderTable(ctx, t)
}

func (r *ArticleResource) Form(ctx context.Context, item any) templ.Component {
    f := form.New().SetSchema(
        form.NewSection("Contenu").Schema(
            form.Text("Title").Label("Titre").Required(),
            form.RichEditor("Content").Label("Contenu").Required(),
        ),
        form.NewSection("Publication").Collapsible().Schema(
            form.NewGrid(2).Schema(
                form.Select("Status").Label("Statut").SetOptions(map[string]string{
                    "draft":     "Brouillon",
                    "published": "Publié",
                    "archived":  "Archivé",
                }).Default("draft"),
                form.Tags("Tags").Label("Tags").WithSuggestions("go", "web", "tutorial"),
            ),
        ),
    )
    if item != nil {
        f.Bind(item)
    }
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
        Title:   req.FormValue("Title"),
        Content: req.FormValue("Content"),
        Status:  req.FormValue("Status"),
    })
}
func (r *ArticleResource) Update(ctx context.Context, id string, req *http.Request) error {
    req.ParseForm()
    return r.db.UpdateArticle(ctx, id, &Article{
        Title:   req.FormValue("Title"),
        Content: req.FormValue("Content"),
        Status:  req.FormValue("Status"),
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
