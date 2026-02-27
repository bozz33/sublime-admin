# Templating Guide

**Working with Templ templates in SublimeGo.**

This guide covers how to use Templ templates effectively in SublimeGo, including built-in components, custom templates, and best practices.

---

## Table of Contents

1. [Introduction to Templ](#introduction-to-templ)
2. [Built-in Components](#built-in-components)
3. [Layouts](#layouts)
4. [Custom Components](#custom-components)
5. [Datastar Integration](#datastar-integration)
6. [Form Templates](#form-templates)
7. [Table Templates](#table-templates)
8. [Best Practices](#best-practices)
9. [Examples](#examples)

---

## Introduction to Templ

Templ is a Go templating language that compiles to Go code. In SublimeGo, we use Templ for all UI components.

### Basic Templ Syntax

```templ
// Basic template
templ hello(name string) {
    <div>Hello, { name }</div>
}

// Conditional rendering
templ greeting(user *User) {
    { if user != nil }
        <div>Hello, { user.Name }</div>
    { else }
        <div>Hello, Guest</div>
    { endif }
}

// Loop
templ userList(users []*User) {
    <ul>
        { for _, user := range users }
            <li>{ user.Name }</li>
        { endfor }
    </ul>
}

// CSS classes
templ button(text string, class string) {
    <button class={ class }>{ text }</button>
}

// Attributes
templ link(url, text string) {
    <a href={ url }>{ text }</a>
}
```

### Using Templates in Go

```go
// Import generated templates
import "github.com/bozz33/sublimeadmin/views"

// Render template
func handler(w http.ResponseWriter, r *http.Request) {
    component := views.Hello("World")
    component.Render(r.Context(), w)
}
```

---

## Built-in Components

SublimeGo provides 32+ atomic components in `ui/atoms/` and 6 layouts in `ui/layouts/`.

### Atomic Components

#### Button
```templ
// Basic button
<button class="btn btn-primary">Click me</button>

// Using the component
@atoms.Button("Click me", atoms.ButtonProps{
    Class: "btn-primary",
    Icon:  "add",
})
```

#### Badge
```templ
@atoms.Badge("New", atoms.BadgeProps{
    Color: "green",
    Size:  "sm",
})
```

#### Modal
```templ
@atoms.Modal("modal-id", atoms.ModalProps{
    Title: "Confirm Action",
    Size:  "md",
}) {
    <p>Are you sure you want to proceed?</p>
}
```

#### Alert
```templ
@atoms.Alert("Success message", atoms.AlertProps{
    Type:  "success",
    Icon:  "check_circle",
})
```

#### Card
```templ
@atoms.Card(atoms.CardProps{
    Title: "Card Title",
    Class: "shadow-lg",
}) {
    <p>Card content goes here</p>
}
```

### Form Components

#### Input Field
```templ
@atoms.InputField("name", atoms.InputProps{
    Label:    "Name",
    Type:     "text",
    Required: true,
    Class:    "form-input",
})
```

#### Select Field
```templ
@atoms.SelectField("role", atoms.SelectProps{
    Label:    "Role",
    Options: []atoms.SelectOption{
        {Value: "user", Label: "User"},
        {Value: "admin", Label: "Admin"},
    },
    Required: true,
})
```

#### Checkbox
```templ
@atoms.CheckboxField("active", atoms.CheckboxProps{
    Label: "Active User",
    Checked: user.Active,
})
```

### Table Components

#### Table
```templ
@atoms.Table(atoms.TableProps{
    Headers: []string{"Name", "Email", "Role"},
    Class:  "table-striped",
}) {
    { for _, user := range users }
        <tr>
            <td>{ user.Name }</td>
            <td>{ user.Email }</td>
            <td>
                @atoms.Badge(user.Role, atoms.BadgeProps{
                    Color: getRoleColor(user.Role),
                })
            </td>
        </tr>
    { endfor }
}
```

---

## Layouts

SublimeGo provides 6 main layouts in `ui/layouts/`.

### Base Layout
```templ
// Base layout with sidebar and topbar
@layouts.Base("Dashboard", layouts.BaseProps{
    SidebarOpen: true,
    DarkMode:    false,
}) {
    <div class="p-6">
        <h1>Dashboard Content</h1>
    </div>
}
```

### Auth Layout
```templ
@layouts.Auth("Login", layouts.AuthProps{
    ShowLogo:    true,
    Background:  "bg-gradient-to-br from-blue-500 to-purple-600",
}) {
    <div class="w-full max-w-md">
        @atoms.Card(atoms.CardProps{
            Title: "Sign In",
            Class: "shadow-xl",
        }) {
            // Login form
        }
    </div>
}
```

### Sidebar Layout
```templ
@layouts.Sidebar("App Name", layouts.SidebarProps{
    Items: []layouts.SidebarItem{
        {Label: "Dashboard", Icon: "dashboard", URL: "/"},
        {Label: "Users", Icon: "people", URL: "/users"},
    },
    CurrentPath: "/users",
}) {
    <main class="flex-1 p-6">
        { children }
    </main>
}
```

### Topbar Layout
```templ
@layouts.Topbar("App Name", layouts.TopbarProps{
    User: &layouts.User{
        Name:  "John Doe",
        Email: "john@example.com",
        Avatar: "/avatar.jpg",
    },
    Notifications: []layouts.Notification{
        {Title: "New user", Body: "John joined", Time: "2m ago"},
    },
}) {
    <div class="p-6">
        { children }
    </div>
}
```

### Flash Layout
```templ
@layouts.Flash(layouts.FlashProps{
    Messages: []layouts.FlashMessage{
        {Type: "success", Text: "Operation completed"},
        {Type: "error", Text: "Something went wrong"},
    },
}) {
    <div class="container mx-auto">
        { children }
    </div>
}
```

### Config Layout
```templ
@layouts.Config("Settings", layouts.ConfigProps{
    Breadcrumbs: []layouts.Breadcrumb{
        {Label: "Home", URL: "/"},
        {Label: "Settings", URL: "/settings"},
    },
}) {
    <div class="space-y-6">
        { children }
    </div>
}
```

---

## Custom Components

### Creating Custom Components

Create a new template file `components/custom.templ`:

```templ
package components

// Custom card with image
templ ImageCard(title, description, image string) {
    <div class="bg-white rounded-lg shadow-md overflow-hidden">
        <img src={ image } alt={ title } class="w-full h-48 object-cover">
        <div class="p-4">
            <h3 class="text-lg font-semibold">{ title }</h3>
            <p class="text-gray-600 mt-2">{ description }</p>
        </div>
    </div>
}

// Status indicator
templ StatusIndicator(status string) {
    <div class="flex items-center space-x-2">
        { switch status {
        case "active" }
            <div class="w-2 h-2 bg-green-500 rounded-full"></div>
            <span class="text-green-600">Active</span>
        case "inactive" }
            <div class="w-2 h-2 bg-gray-500 rounded-full"></div>
            <span class="text-gray-600">Inactive</span>
        { default }
            <div class="w-2 h-2 bg-yellow-500 rounded-full"></div>
            <span class="text-yellow-600">Pending</span>
        { endswitch }
    </div>
}

// Progress bar
templ ProgressBar(current, max int, label string) {
    <div class="w-full">
        <div class="flex justify-between text-sm text-gray-600 mb-1">
            <span>{ label }</span>
            <span>{ current }/{ max }</span>
        </div>
        <div class="w-full bg-gray-200 rounded-full h-2">
            <div class="bg-blue-600 h-2 rounded-full" style={ fmt.Sprintf("width: %.1f%%", float64(current)/float64(max)*100) }></div>
        </div>
    </div>
}
```

### Using Custom Components

```go
// Import generated components
import "github.com/bozz33/sublimeadmin/components"

func handler(w http.ResponseWriter, r *http.Request) {
    component := components.ImageCard(
        "Beautiful Sunset",
        "A stunning sunset over the mountains",
        "/sunset.jpg",
    )
    component.Render(r.Context(), w)
}
```

### Component Composition

```templ
// User profile card
templ UserProfileCard(user *User) {
    @atoms.Card(atoms.CardProps{
        Title: user.Name,
        Class: "shadow-lg",
    }) {
        <div class="flex items-center space-x-4 mb-4">
            <img src={ user.Avatar } alt={ user.Name } class="w-16 h-16 rounded-full">
            <div>
                <h3 class="text-lg font-semibold">{ user.Name }</h3>
                <p class="text-gray-600">{ user.Email }</p>
            </div>
        </div>
        
        <div class="space-y-2">
            @StatusIndicator(user.Status)
            @ProgressBar(user.CompletedTasks, user.TotalTasks, "Tasks Progress")
        </div>
        
        <div class="mt-4 flex space-x-2">
            @atoms.Button("Edit", atoms.ButtonProps{
                Class: "btn-secondary",
                Icon:  "edit",
            })
            @atoms.Button("Delete", atoms.ButtonProps{
                Class: "btn-danger",
                Icon:  "delete",
            })
        </div>
    }
}
```

---

## Datastar Integration

SublimeGo uses Datastar for reactivity instead of HTMX + Alpine.js.

### Datastar Attributes

```templ
// Signal binding
<div data-signal-{ signalName }={ value }></div>

// Event handlers
<button @click={ "increment()" }>Click me</button>

// Conditional rendering
<div @show={ "showModal" }">Modal content</div>

// Loop rendering
<div @for={ "user in users" }>
    <span>{ user.name }</span>
</div>
```

### Live Form Validation

```templ
templ TextInputField(name, label string, value string) {
    <div class="form-group">
        <label for={ name }>{ label }</label>
        <input 
            type="text" 
            id={ name } 
            name={ name } 
            value={ value }
            @input={ "validateField('" + name + "', $event.target.value)" }
            data-signal-{ name + "Error" }={ "" }
            class={ "form-input " + classes }
        />
        <div 
            id={ name + "-error" } 
            @text={ name + "Error" }
            class="text-red-500 text-sm mt-1"
        ></div>
    </div>
}
```

### Real-time Updates

```templ
templ NotificationBadge(count int) {
    <div class="relative">
        <button @click={ "toggleNotifications()" }>
            <span class="material-icons">notifications</span>
        </button>
        { if count > 0 }
            <span 
                @text={ "notifCount" }
                class="absolute -top-1 -right-1 bg-red-500 text-white text-xs rounded-full w-5 h-5 flex items-center justify-center"
            >
                { count }
            </span>
        { endif }
    </div>
}
```

### SSE Integration

```go
// Server-side event handler
func (h *NotificationHandler) HandleBadgeStream(w http.ResponseWriter, r *http.Request) {
    sse := datastar.NewSSE(w)
    
    // Send initial count
    count := h.store.GetUnreadCount(userID)
    sse.MergeSignals(map[string]any{
        "notifCount": count,
    })
    
    // Listen for updates
    ch := h.store.Subscribe(userID)
    defer h.store.Unsubscribe(userID, ch)
    
    for {
        select {
        case <-r.Context().Done():
            return
        case notification := <-ch:
            count := h.store.GetUnreadCount(userID)
            sse.MergeSignals(map[string]any{
                "notifCount": count,
            })
        }
    }
}
```

---

## Form Templates

### Form Layout Template

```templ
// Form with sections
templ FormLayout(title string, sections []FormSection) {
    @layouts.Config(title, layouts.ConfigProps{}) {
        <form method="POST" class="space-y-6">
            { for _, section := range sections }
                <div class="bg-white rounded-lg shadow p-6">
                    <h2 class="text-lg font-semibold mb-4">{ section.Title }</h2>
                    { if section.Description != "" }
                        <p class="text-gray-600 mb-4">{ section.Description }</p>
                    { endif }
                    
                    <div class={ fmt.Sprintf("grid grid-cols-%d gap-4", section.Columns) }>
                        { for _, field := range section.Fields }
                            { field }
                        { endfor }
                    </div>
                </div>
            { endfor }
            
            <div class="flex justify-end space-x-2">
                @atoms.Button("Cancel", atoms.ButtonProps{
                    Type:  "button",
                    Class: "btn-secondary",
                    Attrs: map[string]string{"@click": "history.back()"},
                })
                @atoms.Button("Save", atoms.ButtonProps{
                    Type:  "submit",
                    Class: "btn-primary",
                })
            </div>
        </form>
    }
}
```

### Field Templates

```templ
// Text input with validation
templ TextField(name, label, value string, required bool, error string) {
    <div class="form-group">
        <label for={ name } class="form-label">
            { label }
            { if required }
                <span class="text-red-500">*</span>
            { endif }
        </label>
        <input 
            type="text" 
            id={ name } 
            name={ name } 
            value={ value }
            required={ required }
            class={ "form-input " + classes }
            @input={ "validateField('" + name + "', $event.target.value)" }
        />
        { if error != "" }
            <div class="text-red-500 text-sm mt-1">{ error }</div>
        { endif }
    </div>
}

// Select field
templ SelectField(name, label string, options []SelectOption, value string, required bool) {
    <div class="form-group">
        <label for={ name } class="form-label">
            { label }
            { if required }
                <span class="text-red-500">*</span>
            { endif }
        </label>
        <select 
            id={ name } 
            name={ name } 
            required={ required }
            class="form-select"
        >
            <option value="">Select...</option>
            { for _, option := range options }
                <option 
                    value={ option.Value } 
                    { if option.Value == value }selected{ endif }
                >
                    { option.Label }
                </option>
            { endfor }
        </select>
    </div>
}
```

---

## Table Templates

### Table Template

```templ
templ DataTable(columns []TableColumn, rows []TableRow, actions []TableAction) {
    <div class="bg-white shadow rounded-lg overflow-hidden">
        <div class="overflow-x-auto">
            <table class="min-w-full divide-y divide-gray-200">
                <thead class="bg-gray-50">
                    <tr>
                        { for _, column := range columns }
                            <th class={ "px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider " + column.Class }>
                                { column.Label }
                                { if column.Sortable }
                                    <button @click={ "sort('" + column.Key + "')" } class="ml-1">
                                        <span class="material-icons text-sm">sort</span>
                                    </button>
                                { endif }
                            </th>
                        { endfor }
                        { if len(actions) > 0 }
                            <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                                Actions
                            </th>
                        { endif }
                    </tr>
                </thead>
                <tbody class="bg-white divide-y divide-gray-200">
                    { for _, row := range rows }
                        <tr class="hover:bg-gray-50">
                            { for _, column := range columns }
                                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                                    { column.Render(row.Data[column.Key]) }
                                </td>
                            { endfor }
                            { if len(actions) > 0 }
                                <td class="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                                    <div class="flex justify-end space-x-2">
                                        { for _, action := range actions }
                                            @atoms.Button(action.Label, atoms.ButtonProps{
                                                Class: action.Class,
                                                Icon:  action.Icon,
                                                Attrs: map[string]string{
                                                    "@click": action.ClickHandler,
                                                },
                                            })
                                        { endfor }
                                    </div>
                                </td>
                            { endif }
                        </tr>
                    { endfor }
                </tbody>
            </table>
        </div>
    </div>
}
```

### Pagination Template

```templ
templ Pagination(currentPage, totalPages, totalItems int, perPage int) {
    <div class="bg-white px-4 py-3 flex items-center justify-between border-t border-gray-200 sm:px-6">
        <div class="flex-1 flex justify-between sm:hidden">
            @atoms.Button("Previous", atoms.ButtonProps{
                Class: "btn-secondary",
                Disabled: currentPage == 1,
                Attrs: map[string]string{"@click": "goToPage(" + fmt.Sprintf("%d", currentPage-1) + ")"},
            })
            @atoms.Button("Next", atoms.ButtonProps{
                Class: "btn-secondary",
                Disabled: currentPage == totalPages,
                Attrs: map[string]string{"@click": "goToPage(" + fmt.Sprintf("%d", currentPage+1) + ")"},
            })
        </div>
        <div class="hidden sm:flex-1 sm:flex sm:items-center sm:justify-between">
            <div>
                <p class="text-sm text-gray-700">
                    Showing
                    <span class="font-medium">{ (currentPage-1)*perPage + 1 }</span>
                    to
                    <span class="font-medium">{ min(currentPage*perPage, totalItems) }</span>
                    of
                    <span class="font-medium">{ totalItems }</span>
                    results
                </p>
            </div>
            <div>
                <nav class="relative z-0 inline-flex rounded-md shadow-sm -space-x-px">
                    { for i := 1; i <= totalPages; i++ }
                        <button 
                            @click={ "goToPage(" + fmt.Sprintf("%d", i) + ")" }
                            class={ "relative inline-flex items-center px-4 py-2 border text-sm font-medium " + 
                                (i == currentPage ? "z-10 bg-blue-50 border-blue-500 text-blue-600" : "bg-white border-gray-300 text-gray-500 hover:bg-gray-50") }
                        >
                            { i }
                        </button>
                    { endfor }
                </nav>
            </div>
        </div>
    </div>
}
```

---

## Best Practices

### 1. Component Organization

```
ui/
├── atoms/          # Small, reusable components
│   ├── button.templ
│   ├── input.templ
│   └── badge.templ
├── molecules/      # Combinations of atoms
│   ├── form-field.templ
│   └── table-row.templ
├── organisms/      # Complex UI sections
│   ├── data-table.templ
│   └── sidebar.templ
└── templates/      # Page templates
    ├── dashboard.templ
    └── user-profile.templ
```

### 2. Naming Conventions

```templ
// Use PascalCase for component names
templ UserProfileCard(user *User) { }

// Use camelCase for variables
templ DataTable(data []TableData) { }

// Use kebab-case for CSS classes
<div class="user-profile-card"></div>
```

### 3. Props Structure

```go
// Define props as structs for better type safety
type ButtonProps struct {
    Label  string
    Class  string
    Icon   string
    Type   string
    Attrs  map[string]string
}

templ Button(props ButtonProps) {
    <button 
        type={ props.Type }
        class={ props.Class }
        { for key, value := range props.Attrs }
            { key }={ value }
        { endfor }
    >
        { if props.Icon != "" }
            <span class="material-icons">{ props.Icon }</span>
        { endif }
        { props.Label }
    </button>
}
```

### 4. Error Handling

```templ
templ SafeRender(content string) {
    { if content != "" }
        <div>{ content }</div>
    { else }
        <div class="text-gray-500">No content available</div>
    { endif }
}

templ ErrorBoundary(err error) {
    { if err != nil }
        <div class="bg-red-50 border border-red-200 rounded-md p-4">
            <div class="flex">
                <div class="flex-shrink-0">
                    <span class="material-icons text-red-400">error</span>
                </div>
                <div class="ml-3">
                    <h3 class="text-sm font-medium text-red-800">Error</h3>
                    <div class="mt-2 text-sm text-red-700">
                        { err.Error() }
                    </div>
                </div>
            </div>
        </div>
    { endif }
}
```

### 5. Performance Optimization

```templ
// Use fragments for conditional rendering
templ ConditionalContent(showDetails bool, content string) {
    { if showDetails }
        <div class="details">{ content }</div>
    { endif }
}

// Avoid unnecessary computations
templ ExpensiveList(items []Item) {
    { range _, item := range items }
        <div>{ item.Name }</div>
    { endrange }
}
```

---

## Examples

### Complete Dashboard Page

```templ
templ Dashboard(user *User, stats *DashboardStats, recentActivity []Activity) {
    @layouts.Base("Dashboard", layouts.BaseProps{
        SidebarOpen: true,
        DarkMode:    user.Preferences.DarkMode,
    }) {
        <div class="p-6">
            <!-- Header -->
            <div class="mb-8">
                <h1 class="text-2xl font-bold text-gray-900">Dashboard</h1>
                <p class="text-gray-600">Welcome back, { user.Name }!</p>
            </div>
            
            <!-- Stats Grid -->
            <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
                @atoms.StatCard("Total Users", stats.TotalUsers, "people", "blue")
                @atoms.StatCard("Revenue", fmt.Sprintf("$%.2f", stats.Revenue), "attach_money", "green")
                @atoms.StatCard("Orders", stats.TotalOrders, "shopping_cart", "purple")
                @atoms.StatCard("Conversion", fmt.Sprintf("%.1f%%", stats.ConversionRate), "trending_up", "orange")
            </div>
            
            <!-- Charts -->
            <div class="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-8">
                @atoms.Chart("revenue-chart", "Revenue Trend", stats.RevenueData)
                @atoms.Chart("user-growth", "User Growth", stats.UserGrowthData)
            </div>
            
            <!-- Recent Activity -->
            <div class="bg-white shadow rounded-lg">
                <div class="px-4 py-5 sm:px-6">
                    <h3 class="text-lg leading-6 font-medium text-gray-900">Recent Activity</h3>
                </div>
                <div class="border-t border-gray-200">
                    { for _, activity := range recentActivity }
                        <div class="px-4 py-4 sm:px-6 hover:bg-gray-50">
                            <div class="flex items-center space-x-3">
                                <img src={ activity.User.Avatar } alt={ activity.User.Name } class="w-8 h-8 rounded-full">
                                <div class="flex-1">
                                    <p class="text-sm text-gray-900">
                                        <span class="font-medium">{ activity.User.Name }</span>
                                        { activity.Action }
                                    </p>
                                    <p class="text-xs text-gray-500">{ activity.TimeAgo }</p>
                                </div>
                            </div>
                        </div>
                    { endfor }
                </div>
            </div>
        </div>
    }
}
```

### User Profile Page

```templ
templ UserProfile(user *User, permissions []Permission) {
    @layouts.Config("User Profile", layouts.ConfigProps{
        Breadcrumbs: []layouts.Breadcrumb{
            {Label: "Home", URL: "/"},
            {Label: "Users", URL: "/users"},
            {Label: user.Name, URL: ""},
        },
    }) {
        <div class="max-w-4xl mx-auto">
            <!-- Profile Header -->
            <div class="bg-white shadow rounded-lg p-6 mb-6">
                <div class="flex items-center space-x-6">
                    <img src={ user.Avatar } alt={ user.Name } class="w-24 h-24 rounded-full">
                    <div class="flex-1">
                        <h1 class="text-2xl font-bold text-gray-900">{ user.Name }</h1>
                        <p class="text-gray-600">{ user.Email }</p>
                        @StatusIndicator(user.Status)
                    </div>
                    <div class="flex space-x-2">
                        @atoms.Button("Edit Profile", atoms.ButtonProps{
                            Class: "btn-primary",
                            Icon:  "edit",
                        })
                        @atoms.Button("Settings", atoms.ButtonProps{
                            Class: "btn-secondary",
                            Icon:  "settings",
                        })
                    </div>
                </div>
            </div>
            
            <!-- Details Tabs -->
            @atoms.Tabs("profile-tabs", []atoms.Tab{
                {Label: "Overview", Icon: "info"},
                {Label: "Activity", Icon: "history"},
                {Label: "Permissions", Icon: "security"},
            }) {
                <!-- Overview Tab -->
                <div class="space-y-6">
                    <div class="bg-white shadow rounded-lg p-6">
                        <h2 class="text-lg font-semibold mb-4">Information</h2>
                        <dl class="grid grid-cols-1 gap-x-4 gap-y-6 sm:grid-cols-2">
                            <div>
                                <dt class="text-sm font-medium text-gray-500">Full Name</dt>
                                <dd class="mt-1 text-sm text-gray-900">{ user.Name }</dd>
                            </div>
                            <div>
                                <dt class="text-sm font-medium text-gray-500">Email</dt>
                                <dd class="mt-1 text-sm text-gray-900">{ user.Email }</dd>
                            </div>
                            <div>
                                <dt class="text-sm font-medium text-gray-500">Role</dt>
                                <dd class="mt-1 text-sm text-gray-900">
                                    @atoms.Badge(user.Role, atoms.BadgeProps{
                                        Color: getRoleColor(user.Role),
                                    })
                                </dd>
                            </div>
                            <div>
                                <dt class="text-sm font-medium text-gray-500">Joined</dt>
                                <dd class="mt-1 text-sm text-gray-900">{ user.CreatedAt.Format("January 2, 2006") }</dd>
                            </div>
                        </dl>
                    </div>
                    
                    <div class="bg-white shadow rounded-lg p-6">
                        <h2 class="text-lg font-semibold mb-4">Statistics</h2>
                        <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
                            @ProgressBar(user.CompletedTasks, user.TotalTasks, "Tasks Completed")
                            @ProgressBar(user.ProjectsCompleted, user.TotalProjects, "Projects Completed")
                            @ProgressBar(user.HoursLogged, user.TargetHours, "Hours Logged")
                        </div>
                    </div>
                </div>
                
                <!-- Activity Tab -->
                <div>
                    @ActivityTimeline(user.Activity)
                </div>
                
                <!-- Permissions Tab -->
                <div>
                    <div class="bg-white shadow rounded-lg">
                        <div class="px-4 py-5 sm:px-6">
                            <h3 class="text-lg leading-6 font-medium text-gray-900">Permissions</h3>
                        </div>
                        <div class="border-t border-gray-200">
                            { for _, permission := range permissions }
                                <div class="px-4 py-4 sm:px-6">
                                    <div class="flex items-center justify-between">
                                        <div>
                                            <h4 class="text-sm font-medium text-gray-900">{ permission.Name }</h4>
                                            <p class="text-sm text-gray-500">{ permission.Description }</p>
                                        </div>
                                        <div>
                                            { if permission.Granted }
                                                @atoms.Badge("Granted", atoms.BadgeProps{
                                                    Color: "green",
                                                })
                                            { else }
                                                @atoms.Badge("Denied", atoms.BadgeProps{
                                                    Color: "red",
                                                })
                                            { endif }
                                        </div>
                                    </div>
                                </div>
                            { endfor }
                        </div>
                    </div>
                </div>
            }
        </div>
    }
}
```

---

## Next Steps

1. **Learn Templ Basics**: Read the [Templ documentation](https://templ.guide/)
2. **Study Built-in Components**: Explore `ui/atoms/` and `ui/layouts/`
3. **Practice Component Composition**: Build complex UI from simple components
4. **Master Datastar**: Learn SSE and reactive patterns
5. **Optimize Performance**: Use fragments and avoid unnecessary computations

For more examples, see the [examples directory](https://github.com/bozz33/sublime-admin/tree/main/examples).
