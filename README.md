# Sublime Admin

A modern Go framework for building admin panels, inspired by Laravel Filament.

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

See the [SublimeGo Starter](https://github.com/bozz33/SublimeGo) for a complete example project.

## License

MIT License
