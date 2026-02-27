# Resources Guide

**Building resources step by step with SublimeGo.**

This guide walks you through creating complete CRUD resources with forms, tables, actions, and advanced features.

---

## Table of Contents

1. [Basic Resource](#basic-resource)
2. [Forms](#forms)
3. [Tables](#tables)
4. [Actions](#actions)
5. [Relations](#relations)
6. [Infolist](#infolist)
7. [Advanced Features](#advanced-features)
8. [Complete Example](#complete-example)

---

## Basic Resource

A resource represents a data model in your admin panel. Here's the minimal structure:

```go
package main

import (
    "context"
    "fmt"

    "github.com/bozz33/sublimeadmin/engine"
    "github.com/a-h/templ"
)

type UserResource struct {
    engine.BaseResource
    db *ent.Client
}

// Required interfaces
func (r *UserResource) Slug() string        { return "users" }
func (r *UserResource) Label() string       { return "User" }
func (r *UserResource) PluralLabel() string { return "Users" }
func (r *UserResource) Icon() string        { return "person" }

// Optional interfaces
func (r *UserResource) CanCreate(ctx context.Context) bool  { return true }
func (r *UserResource) CanRead(ctx context.Context) bool   { return true }
func (r *UserResource) CanUpdate(ctx context.Context) bool  { return true }
func (r *UserResource) CanDelete(ctx context.Context) bool  { return true }
```

---

## Forms

### Basic Form

```go
func (r *UserResource) Form(ctx context.Context, item any) templ.Component {
    user := item.(*ent.User)
    
    f := form.New().SetSchema(
        form.Text("name").Label("Name").Required(),
        form.Email("email").Label("Email").Required(),
        form.Password("password").Label("Password").Required(),
        form.Select("role").Label("Role").Options(map[string]string{
            "user":  "User",
            "admin": "Admin",
        }),
    )
    
    // Set initial values for edit form
    if user != nil {
        f.SetValues(map[string]any{
            "name":  user.Name,
            "email": user.Email,
            "role":  user.Role,
        })
    }
    
    return views.GenericForm(f)
}
```

### Advanced Form Features

```go
f := form.New().SetSchema(
    // Basic fields
    form.Text("name").Label("Name").Required().Placeholder("Enter name"),
    form.Textarea("bio").Label("Bio").Rows(4),
    
    // Advanced fields
    form.RichEditor("content").Label("Content").Toolbar([]string{
        "bold", "italic", "link", "heading", "list", "image",
    }),
    form.MarkdownEditor("specs").Label("Specifications"),
    form.Tags("skills").Label("Skills").Suggestions([]string{
        "Go", "JavaScript", "Python", "Docker",
    }),
    form.ColorPicker("theme").Label("Theme Color").Swatches([]string{
        "#3B82F6", "#10B981", "#F59E0B", "#EF4444",
    }),
    form.Slider("experience").Label("Experience").Min(0).Max(10),
    
    // Layout components
    form.Section("Personal Info", "Basic user information",
        form.Grid(2,
            form.Text("first_name").Label("First Name"),
            form.Text("last_name").Label("Last Name"),
        ),
    ),
    
    form.Tabs(
        form.Tab("Account", "Account settings",
            form.Email("email").Label("Email"),
            form.Password("password").Label("Password"),
        ),
        form.Tab("Profile", "Profile information",
            form.Text("phone").Label("Phone"),
            form.Select("country").Label("Country"),
        ),
    ),
)
```

### Live Validation

```go
form.Text("email").Label("Email").
    Required().
    WithLiveValidation("/users/validate-field"),
```

The validation endpoint is automatically handled by `CRUDHandler.ValidateField()`.

---

## Tables

### Basic Table

```go
func (r *UserResource) Table(ctx context.Context) templ.Component {
    t := table.New(nil).
        WithColumns(
            table.Text("name").WithLabel("Name").Sortable().Searchable(),
            table.Email("email").WithLabel("Email"),
            table.Badge("role").WithLabel("Role").Colors(map[string]string{
                "admin": "red",
                "user":  "green",
            }),
            table.Date("created_at").WithLabel("Created").Sortable(),
        )
    
    return views.GenericTable(t)
}
```

### Advanced Table Features

```go
t := table.New(nil).
    WithColumns(
        // Display columns
        table.Text("name").WithLabel("Name").
            Sortable().Searchable().
            WithIcon("person").
            WithColorFunc(func(value string, item any) string {
                if item.(*ent.User).Active {
                    return "green"
                }
                return "gray"
            }),
        
        table.Avatar("avatar").WithLabel("Avatar").
            WithInitials(func(item any) string {
                user := item.(*ent.User)
                return string(user.Name[0]) + string(user.Name[1])
            }),
        
        // Inline edit columns
        table.TextInputColumn("email").WithLabel("Email").
            PatchURL(func(item any) string {
                return fmt.Sprintf("/users/%d", item.(*ent.User).ID)
            }),
        
        table.SelectColumn("role").WithLabel("Role").
            Options([]table.SelectColOption{
                {Value: "user", Label: "User"},
                {Value: "admin", Label: "Admin"},
            }).
            PatchURL(func(item any) string {
                return fmt.Sprintf("/users/%d", item.(*ent.User).ID)
            }),
        
        // View column
        table.ViewColumn("view").WithLabel("View").
            URL(func(item any) string {
                return fmt.Sprintf("/admin/users/%d", item.(*ent.User).ID)
            }),
    ).
    WithFilters(
        table.SelectFilter("role").Label("Role").Options([]table.FilterOption{
            {Value: "user", Label: "User"},
            {Value: "admin", Label: "Admin"},
        }),
        table.DateFilter("created_at").Label("Created"),
        table.TextFilter("search").Label("Search"),
    ).
    WithSummaries(
        table.NewSummary("salary", table.SummaryAverage).WithLabel("Avg Salary"),
        table.NewSummary("projects", table.SummaryCount).WithLabel("Total Projects"),
    ).
    WithGroups(
        table.GroupBy("department").WithLabel("Department").
            Collapsible().CollapsedByDefault(),
    ).
    EnableColumnManager().
    WithStriped().
    WithDeferred()
```

---

## Actions

### Basic Actions

```go
func (r *UserResource) Actions(ctx context.Context) []*actions.Action {
    return []*actions.Action{
        actions.EditAction("/admin/users"),
        actions.DeleteAction("/admin/users"),
        actions.ViewAction("/admin/users"),
    }
}
```

### Custom Actions

```go
func (r *UserResource) Actions(ctx context.Context) []*actions.Action {
    return []*actions.Action{
        // Simple action
        actions.New("activate").
            SetLabel("Activate").
            SetIcon("check").
            SetColor(actions.Success).
            Execute(func(ctx context.Context, item any) error {
                user := item.(*ent.User)
                return r.db.User.UpdateOne(user).
                    SetActive(true).
                    Exec(ctx)
            }),
        
        // Modal action with form
        actions.NewModal("send-email").
            SetLabel("Send Email").
            SetIcon("mail").
            SetForm(func(ctx context.Context, item any) templ.Component {
                return EmailForm(item.(*ent.User))
            }).
            Execute(func(ctx context.Context, item any, r *http.Request) error {
                user := item.(*ent.User)
                subject := r.FormValue("subject")
                body := r.FormValue("body")
                return r.mailer.Send(mailer.Message{
                    To:      []string{user.Email},
                    Subject: subject,
                    Body:    body,
                })
            }),
        
        // Bulk action
        actions.New("bulk-activate").
            SetLabel("Activate Selected").
            SetIcon("check_circle").
            SetBulk(true).
            Execute(func(ctx context.Context, items []any) error {
                for _, item := range items {
                    user := item.(*ent.User)
                    if err := r.db.User.UpdateOne(user).
                        SetActive(true).
                        Exec(ctx); err != nil {
                        return err
                    }
                }
                return nil
            }),
    }
}
```

### Action Lifecycle Hooks

```go
action := actions.New("delete").
    BeforeFunc(func(ctx context.Context, item any) error {
        // Log before deletion
        log.Printf("Deleting user: %v", item)
        return nil
    }).
    AfterFunc(func(ctx context.Context, item any) error {
        // Send notification
        r.notifications.Send(ctx, &notifications.Notification{
            Title: "User Deleted",
            Body:  fmt.Sprintf("User %v was deleted", item),
            Level: notifications.Info,
        })
        return nil
    }).
    OnSuccess(func(ctx context.Context, item any) {
        // Flash message
        flash.Set(ctx, "success", "User deleted successfully")
    }).
    OnFailure(func(ctx context.Context, item any, err error) {
        // Error handling
        flash.Set(ctx, "error", "Failed to delete user")
    })
```

---

## Relations

### Defining Relations

```go
func (r *UserResource) Relations() []engine.Relation {
    return []engine.Relation{
        engine.BelongsTo("company", "Company", "company_id", "id").
            SetDisplayField("name").
            SetEager(true),
        
        engine.HasMany("posts", "Post", "user_id", "id").
            SetDisplayField("title").
            SetEager(true),
        
        engine.ManyToMany("roles", "Role", "user_roles", "user_id", "role_id").
            SetDisplayField("name"),
    }
}
```

### Relation Manager

```go
func (r *UserResource) RelationManagers() []engine.RelationManager {
    return []engine.RelationManager{
        &PostRelationManager{
            db:     r.db,
            parent: r,
        },
    }
}

type PostRelationManager struct {
    engine.BaseRelationManager
    db     *ent.Client
    parent *UserResource
}

func (rm *PostRelationManager) RelationName() string { return "posts" }

func (rm *PostRelationManager) Form(ctx context.Context, parentID string) templ.Component {
    f := form.New().SetSchema(
        form.Text("title").Label("Title").Required(),
        form.RichEditor("content").Label("Content"),
        form.Select("status").Label("Status").Options(map[string]string{
            "draft":     "Draft",
            "published": "Published",
        }),
    )
    return views.RelationForm(f, parentID)
}
```

---

## Infolist

### Basic Infolist

```go
func (r *UserResource) View(ctx context.Context, item any) templ.Component {
    user := item.(*ent.User)
    
    il := infolist.New().
        AddSection("Basic Info", 2,
            infolist.TextEntry("name", "Name", user.Name),
            infolist.BadgeEntry("role", "Role", user.Role, "blue"),
            infolist.BooleanEntry("active", "Active", user.Active),
        ).
        AddSection("Contact", 1,
            infolist.TextEntry("email", "Email", user.Email),
            infolist.TextEntry("phone", "Phone", user.Phone),
        )
    
    return views.GenericInfolist(il)
}
```

### Advanced Infolist

```go
il := infolist.New().
    AddSection("Profile", 2,
        infolist.ImageEntry("avatar", "Avatar", user.AvatarURL),
        infolist.TextEntry("name", "Name", user.Name),
        infolist.BadgeEntry("status", "Status", user.Status, "green"),
        infolist.DateEntry("created", "Created", user.CreatedAt, "2006-01-02"),
    ).
    AddSection("Bio", 1,
        infolist.CodeEntry("bio", "Biography", user.Bio, "markdown").
            WithLanguage("markdown"),
    ).
    AddSection("Skills", 1,
        infolist.TagsEntry("skills", "Skills", user.Skills).
            WithColorMap(map[string]string{
                "Go":       "blue",
                "Python":   "green",
                "Docker":   "cyan",
            }),
    ).
    AddSection("Activity", 3,
        infolist.KeyValueEntry("stats", "Statistics",
            infolist.KeyValuePair{Key: "Posts", Value: fmt.Sprintf("%d", user.PostCount)},
            infolist.KeyValuePair{Key: "Comments", Value: fmt.Sprintf("%d", user.CommentCount)},
            infolist.KeyValuePair{Key: "Likes", Value: fmt.Sprintf("%d", user.LikeCount)},
        ),
    )
```

---

## Advanced Features

### Search

```go
func (r *UserResource) Search(ctx context.Context, query string) ([]any, error) {
    users, err := r.db.User.Query().
        Where(
            user.Or(
                user.NameContains(query),
                user.EmailContains(query),
            ),
        ).
        Limit(20).
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

### Export

```go
func (r *UserResource) Export(ctx context.Context, format string) ([]byte, error) {
    users, err := r.db.User.Query().All(ctx)
    if err != nil {
        return nil, err
    }
    
    exp := export.New(format)
    exp.SetHeaders([]string{"ID", "Name", "Email", "Role", "Created"})
    
    for _, u := range users {
        exp.AddRow([]string{
            fmt.Sprintf("%d", u.ID),
            u.Name,
            u.Email,
            u.Role,
            u.CreatedAt.Format("2006-01-02"),
        })
    }
    
    return exp.Write()
}
```

### Validation

```go
func (r *UserResource) ValidateField(ctx context.Context, field, value string) error {
    switch field {
    case "email":
        if !strings.Contains(value, "@") {
            return validation.NewError(field, "Invalid email format")
        }
        // Check uniqueness
        exists, err := r.db.User.Query().
            Where(user.Email(value)).
            Exist(ctx)
        if err != nil {
            return err
        }
        if exists {
            return validation.NewError(field, "Email already exists")
        }
    case "name":
        if len(value) < 2 {
            return validation.NewError(field, "Name must be at least 2 characters")
        }
    }
    return nil
}
```

### Hooks

```go
func (r *UserResource) BeforeCreate(ctx context.Context, req *http.Request) error {
    // Hash password
    password := req.FormValue("password")
    hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        return err
    }
    req.Form.Set("password", string(hashed))
    return nil
}

func (r *UserResource) AfterCreate(ctx context.Context, item any) error {
    user := item.(*ent.User)
    
    // Send welcome email
    go func() {
        r.mailer.Send(mailer.Message{
            To:      []string{user.Email},
            Subject: "Welcome to our platform",
            Body:    fmt.Sprintf("Welcome %s!", user.Name),
        })
    }()
    
    // Create notification
    r.notifications.Send(ctx, &notifications.Notification{
        Title: "New User",
        Body:  fmt.Sprintf("%s joined the platform", user.Name),
        Level: notifications.Info,
    })
    
    return nil
}
```

---

## Complete Example

```go
package main

import (
    "context"
    "fmt"
    "net/http"
    
    "entgo.io/ent/ent"
    "github.com/bozz33/sublimeadmin/actions"
    "github.com/bozz33/sublimeadmin/engine"
    "github.com/bozz33/sublimeadmin/form"
    "github.com/bozz33/sublimeadmin/infolist"
    "github.com/bozz33/sublimeadmin/table"
    "github.com/bozz33/sublimeadmin/views"
    "github.com/a-h/templ"
)

type UserResource struct {
    engine.BaseResource
    db *ent.Client
}

// Required interfaces
func (r *UserResource) Slug() string        { return "users" }
func (r *UserResource) Label() string       { return "User" }
func (r *UserResource) PluralLabel() string { return "Users" }
func (r *UserResource) Icon() string        { return "person" }

// CRUD operations
func (r *UserResource) List(ctx context.Context) ([]any, error) {
    users, err := r.db.User.Query().All(ctx)
    if err != nil {
        return nil, err
    }
    result := make([]any, len(users))
    for i, u := range users {
        result[i] = u
    }
    return result, nil
}

func (r *UserResource) Get(ctx context.Context, id string) (any, error) {
    userID, err := strconv.Atoi(id)
    if err != nil {
        return nil, err
    }
    return r.db.User.Get(ctx, userID)
}

func (r *UserResource) Create(ctx context.Context, req *http.Request) error {
    name := req.FormValue("name")
    email := req.FormValue("email")
    password := req.FormValue("password")
    role := req.FormValue("role")
    
    _, err := r.db.User.Create().
        SetName(name).
        SetEmail(email).
        SetPassword(password).
        SetRole(role).
        Save(ctx)
    return err
}

func (r *UserResource) Update(ctx context.Context, id string, req *http.Request) error {
    userID, err := strconv.Atoi(id)
    if err != nil {
        return err
    }
    
    update := r.db.User.UpdateOneID(userID)
    
    if name := req.FormValue("name"); name != "" {
        update.SetName(name)
    }
    if email := req.FormValue("email"); email != "" {
        update.SetEmail(email)
    }
    if role := req.FormValue("role"); role != "" {
        update.SetRole(role)
    }
    
    return update.Exec(ctx)
}

func (r *UserResource) Delete(ctx context.Context, id string) error {
    userID, err := strconv.Atoi(id)
    if err != nil {
        return err
    }
    return r.db.User.DeleteOneID(userID).Exec(ctx)
}

// Views
func (r *UserResource) Form(ctx context.Context, item any) templ.Component {
    user := item.(*ent.User)
    
    f := form.New().SetSchema(
        form.Section("Basic Info", "User information",
            form.Grid(2,
                form.Text("name").Label("Name").Required(),
                form.Email("email").Label("Email").Required(),
            ),
            form.Select("role").Label("Role").Options(map[string]string{
                "user":  "User",
                "admin": "Admin",
            }),
        ),
        form.Section("Profile", "Additional details",
            form.Textarea("bio").Label("Bio").Rows(4),
            form.Tags("skills").Label("Skills").Suggestions([]string{
                "Go", "JavaScript", "Python", "Docker",
            }),
        ),
    )
    
    if user != nil {
        f.SetValues(map[string]any{
            "name":  user.Name,
            "email": user.Email,
            "role":  user.Role,
            "bio":   user.Bio,
            "skills": user.Skills,
        })
    }
    
    return views.GenericForm(f)
}

func (r *UserResource) Table(ctx context.Context) templ.Component {
    t := table.New(nil).
        WithColumns(
            table.Avatar("name").WithLabel("Name").
                WithInitials(func(item any) string {
                    user := item.(*ent.User)
                    return string(user.Name[0]) + string(user.Name[1])
                }),
            table.Text("email").WithLabel("Email").Searchable(),
            table.Badge("role").WithLabel("Role").Colors(map[string]string{
                "admin": "red",
                "user":  "green",
            }),
            table.Tags("skills").WithLabel("Skills"),
            table.Date("created_at").WithLabel("Created").Sortable(),
            table.ViewColumn("view").WithLabel("View").URL(func(item any) string {
                return fmt.Sprintf("/admin/users/%d", item.(*ent.User).ID)
            }),
        ).
        WithFilters(
            table.SelectFilter("role").Label("Role").Options([]table.FilterOption{
                {Value: "user", Label: "User"},
                {Value: "admin", Label: "Admin"},
            }),
            table.TextFilter("search").Label("Search"),
        ).
        EnableColumnManager()
    
    return views.GenericTable(t)
}

func (r *UserResource) View(ctx context.Context, item any) templ.Component {
    user := item.(*ent.User)
    
    il := infolist.New().
        AddSection("Profile", 2,
            infolist.ImageEntry("avatar", "Avatar", user.AvatarURL),
            infolist.TextEntry("name", "Name", user.Name),
            infolist.BadgeEntry("role", "Role", user.Role, "blue"),
            infolist.DateEntry("created", "Created", user.CreatedAt, "2006-01-02"),
        ).
        AddSection("Details", 1,
            infolist.TextEntry("email", "Email", user.Email),
            infolist.CodeEntry("bio", "Biography", user.Bio, "markdown"),
            infolist.TagsEntry("skills", "Skills", user.Skills),
        )
    
    return views.GenericInfolist(il)
}

// Actions
func (r *UserResource) Actions(ctx context.Context) []*actions.Action {
    return []*actions.Action{
        actions.EditAction("/admin/users"),
        actions.DeleteAction("/admin/users"),
        actions.ViewAction("/admin/users"),
        actions.New("activate").
            SetLabel("Activate").
            SetIcon("check").
            SetColor(actions.Success).
            Execute(func(ctx context.Context, item any) error {
                user := item.(*ent.User)
                return r.db.User.UpdateOne(user).
                    SetActive(true).
                    Exec(ctx)
            }),
    }
}

// Optional interfaces
func (r *UserResource) Search(ctx context.Context, query string) ([]any, error) {
    users, err := r.db.User.Query().
        Where(
            user.Or(
                user.NameContains(query),
                user.EmailContains(query),
            ),
        ).
        Limit(20).
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

func (r *UserResource) ValidateField(ctx context.Context, field, value string) error {
    switch field {
    case "email":
        if !strings.Contains(value, "@") {
            return fmt.Errorf("invalid email format")
        }
    case "name":
        if len(value) < 2 {
            return fmt.Errorf("name must be at least 2 characters")
        }
    }
    return nil
}
```

---

## Next Steps

1. **CLI Generation**: Use `sublimego make:resource User` to generate boilerplate
2. **Relations**: Add related resources with `RelationManager`
3. **Permissions**: Implement role-based access control
4. **Testing**: Write unit tests for your resources
5. **Deployment**: Configure for production with proper middleware

For more examples, see the [examples directory](https://github.com/bozz33/sublime-admin/tree/main/examples).
