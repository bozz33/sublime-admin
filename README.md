# Sublime Admin

A modern Go framework for building admin panels, inspired by Laravel Filament.

**This is the core framework package without a project structure.** For a complete starter project with examples, see [SublimeGo](https://github.com/bozz33/SublimeGo).

## Which One Should I Use?

- **sublime-admin** (`github.com/bozz33/sublimeadmin`) - Core framework library only, use this as a dependency in your existing Go project
- **SublimeGo** (`github.com/bozz33/sublimego`) - Complete starter project with examples, database setup, and project structure

## Installation

```bash
go get github.com/bozz33/sublimeadmin@v1.0.0
```

## Features

- **Resource System**: Full CRUD with automatic generation
- **Form Builder**: Fluent form builder with validation
- **Table Builder**: Interactive tables with sorting, filters, and pagination
- **Actions**: Customizable actions with confirmation modals
- **Widgets**: Stats cards and charts (ApexCharts)
- **Navigation**: Advanced navigation with groups and badges
- **Authentication**: Built-in auth with bcrypt and sessions
- **Middleware**: Rate limiting, CORS, logging, recovery
- **Validation**: Extensible validation with custom rules
- **Export**: CSV and Excel export

## Quick Start

```go
package main

import (
    "github.com/bozz33/sublimeadmin/engine"
    "github.com/bozz33/sublimeadmin/form"
    "github.com/bozz33/sublimeadmin/table"
)

func main() {
    // Create your admin panel
    panel := engine.NewPanel("admin").
        SetPath("/admin").
        SetBrandName("My App")
    
    // Register resources
    panel.AddResources(
        &ProductResource{},
        &UserResource{},
    )
    
    // Start server
    http.ListenAndServe(":8080", panel.Router())
}
```

## Packages

| Package | Description |
|---------|-------------|
| `engine` | Core panel and resource management |
| `form` | Form builder with fields and validation |
| `table` | Table builder with columns and filters |
| `actions` | Action system with confirmation dialogs |
| `auth` | Authentication and authorization |
| `middleware` | HTTP middlewares (auth, CORS, rate limit, etc.) |
| `validation` | Input validation with custom rules |
| `widget` | Dashboard widgets (stats, charts) |
| `flash` | Flash messages |
| `errors` | Error handling |
| `export` | CSV/Excel export |
| `ui` | UI components and layouts |

## Documentation

For complete documentation, examples, and guides, see the [SublimeGo Starter Project](https://github.com/bozz33/SublimeGo).

## Relationship with SublimeGo

- **sublime-admin** is the core framework library (this repository)
- **SublimeGo** is a complete starter project that uses sublime-admin
- Use **SublimeGo** if you want to start a new admin panel project
- Use **sublime-admin** if you want to add admin panel functionality to an existing Go project

## License

MIT License
