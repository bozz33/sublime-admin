# Panel Colors Configuration Guide

> Complete guide for configuring all semantic colors in SublimeGo panels — primary, danger, success, warning, info, and secondary.

---

## Overview

Le template SublimeGo utilise **6 couleurs sémantiques** définies dans le CSS :

| Couleur | Utilisation | Défaut |
|---------|-------------|--------|
| **primary** | Boutons principaux, liens actifs, focus | Green (#22c55e) |
| **danger** | Erreurs, suppressions, alertes critiques | Red (#ef4444) |
| **success** | Confirmations, validations, succès | Emerald (#10b981) |
| **warning** | Avertissements, actions à risque | Amber (#f59e0b) |
| **info** | Informations, aide, tooltips | Blue (#3b82f6) |
| **secondary** | Boutons secondaires, textes neutres | Slate (#64748b) |

Chaque couleur génère automatiquement **11 nuances** (50, 100, 200...950) pour une intégration parfaite avec Tailwind CSS.

---

## Configuration Simple

### 1. Couleur Primaire Uniquement

```go
panel := engine.NewPanel("admin").
    WithCustomColor("#3b82f6")  // Change primary to blue
```

### 2. Couleurs Sémantiques Individuelles

```go
panel := engine.NewPanel("admin").
    WithSemanticColors(map[string]string{
        "primary":   "#3b82f6",  // Blue
        "danger":    "#ef4444",  // Red
        "success":   "#10b981",  // Emerald
        "warning":   "#f59e0b",  // Amber
        "info":      "#06b6d4",  // Cyan
        "secondary": "#64748b",  // Slate
    })
```

### 3. Schéma de Couleurs Complet

```go
import "github.com/bozz33/sublimego/color"

c := color.Color{}
panel := engine.NewPanel("admin").
    WithColors(&engine.ColorScheme{
        Primary:   c.Hex("#3b82f6"),
        Danger:    c.Hex("#ef4444"),
        Success:   c.Hex("#10b981"),
        Warning:   c.Hex("#f59e0b"),
        Info:      c.Hex("#06b6d4"),
        Secondary: c.Hex("#64748b"),
    })
```

---

## Où Sont Utilisées Ces Couleurs ?

### Primary (Couleur Principale)

**Dans le CSS :**
- `.btn-primary` — Boutons d'action principaux
- `.badge-primary` — Badges de statut actif
- `.alert-primary` — Alertes informatives
- `.progress-primary` — Barres de progression
- `.bg-gradient-primary` — Arrière-plans dégradés
- `.text-gradient-primary` — Textes avec dégradé
- `.stat-icon-primary` — Icônes de statistiques

**Dans les composants :**
```go
// Boutons
actions.New("save").Color("primary")

// Badges
table.Badge("status").Color("primary")

// Navigation active
// Automatiquement appliqué aux liens actifs
```

### Danger (Erreurs/Suppressions)

**Dans le CSS :**
- `.btn-danger` — Boutons de suppression
- `.badge-danger` — Badges d'erreur
- `.alert-danger` — Alertes d'erreur
- `.form-input-danger` — Champs en erreur
- `.form-helper-danger` — Messages d'erreur

**Dans les composants :**
```go
// Actions destructives
actions.New("delete").Color("danger")

// Notifications
notifications.Danger("Error").WithBody("Failed to save")

// Badges
table.Badge("status").Colors(map[string]string{
    "rejected": "danger",
})
```

### Success (Confirmations)

**Dans le CSS :**
- `.btn-success` — Boutons de confirmation
- `.badge-success` — Badges de succès
- `.alert-success` — Alertes de succès
- `.progress-success` — Barres de progression réussies
- `.form-input-success` — Champs validés

**Dans les composants :**
```go
// Notifications
notifications.Success("Saved").WithBody("Changes saved successfully")

// Badges
table.Badge("status").Colors(map[string]string{
    "approved": "success",
    "active":   "success",
})
```

### Warning (Avertissements)

**Dans le CSS :**
- `.btn-warning` — Boutons d'avertissement
- `.badge-warning` — Badges d'avertissement
- `.alert-warning` — Alertes d'avertissement
- `.progress-warning` — Barres de progression en attente

**Dans les composants :**
```go
// Notifications
notifications.Warning("Warning").WithBody("Unsaved changes")

// Badges
table.Badge("status").Colors(map[string]string{
    "pending":  "warning",
    "draft":    "warning",
})
```

### Info (Informations)

**Dans le CSS :**
- `.btn-info` — Boutons informatifs
- `.badge-info` — Badges informatifs
- `.alert-info` — Alertes informatives
- `.progress-info` — Barres de progression informatives

**Dans les composants :**
```go
// Notifications
notifications.Info("Info").WithBody("New version available")

// Badges
table.Badge("type").Colors(map[string]string{
    "info": "info",
})
```

### Secondary (Neutre)

**Dans le CSS :**
- `.btn-secondary` — Boutons secondaires/annulation
- `.badge-secondary` — Badges neutres
- `.bg-gradient-secondary` — Arrière-plans neutres

**Dans les composants :**
```go
// Boutons d'annulation
actions.New("cancel").Color("secondary")
```

---

## Exemples de Thèmes Prédéfinis

### Thème Bleu (Corporate)

```go
panel.WithSemanticColors(map[string]string{
    "primary":   "#3b82f6",  // Blue
    "danger":    "#ef4444",  // Red
    "success":   "#10b981",  // Emerald
    "warning":   "#f59e0b",  // Amber
    "info":      "#06b6d4",  // Cyan
    "secondary": "#64748b",  // Slate
})
```

### Thème Violet (Creative)

```go
panel.WithSemanticColors(map[string]string{
    "primary":   "#8b5cf6",  // Purple
    "danger":    "#ef4444",  // Red
    "success":   "#10b981",  // Emerald
    "warning":   "#f59e0b",  // Amber
    "info":      "#3b82f6",  // Blue
    "secondary": "#6b7280",  // Gray
})
```

### Thème Orange (Énergique)

```go
panel.WithSemanticColors(map[string]string{
    "primary":   "#f97316",  // Orange
    "danger":    "#dc2626",  // Red
    "success":   "#16a34a",  // Green
    "warning":   "#eab308",  // Yellow
    "info":      "#0ea5e9",  // Sky
    "secondary": "#71717a",  // Zinc
})
```

### Thème Rose (Moderne)

```go
panel.WithSemanticColors(map[string]string{
    "primary":   "#ec4899",  // Pink
    "danger":    "#ef4444",  // Red
    "success":   "#10b981",  // Emerald
    "warning":   "#f59e0b",  // Amber
    "info":      "#8b5cf6",  // Purple
    "secondary": "#6b7280",  // Gray
})
```

---

## Génération Automatique des CSS Variables

Le panel génère automatiquement les CSS variables pour toutes les couleurs :

```css
@theme {
  /* Primary */
  --color-primary-50: #f0fdf4;
  --color-primary-100: #dcfce7;
  --color-primary-200: #bbf7d0;
  --color-primary-300: #86efac;
  --color-primary-400: #4ade80;
  --color-primary-500: #22c55e;
  --color-primary-600: #16a34a;
  --color-primary-700: #15803d;
  --color-primary-800: #166534;
  --color-primary-900: #14532d;
  --color-primary-950: #052e16;
  
  /* Danger */
  --color-danger-50: #fef2f2;
  --color-danger-100: #fee2e2;
  /* ... */
  
  /* Success, Warning, Info, Secondary */
  /* ... */
}
```

Ces variables sont utilisées automatiquement par Tailwind CSS :
- `bg-primary-500` → `background-color: var(--color-primary-500)`
- `text-danger-600` → `color: var(--color-danger-600)`
- `border-success-300` → `border-color: var(--color-success-300)`

---

## Utilisation dans les Templates Personnalisés

```html
<!-- Utiliser les couleurs sémantiques dans vos templates -->
<button class="bg-primary-600 hover:bg-primary-700 text-white">
    Save
</button>

<div class="border-l-4 border-danger-500 bg-danger-50 p-4">
    <p class="text-danger-700">Error message</p>
</div>

<span class="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-success-100 text-success-700">
    Active
</span>
```

---

## Exemple Complet

```go
package main

import (
    "github.com/bozz33/sublimego/color"
    "github.com/bozz33/sublimego/engine"
)

func main() {
    // 1. Créer un schéma de couleurs personnalisé
    c := color.Color{}
    customColors := &engine.ColorScheme{
        Primary:   c.Hex("#3b82f6"),  // Blue
        Danger:    c.Hex("#ef4444"),  // Red
        Success:   c.Hex("#10b981"),  // Emerald
        Warning:   c.Hex("#f59e0b"),  // Amber
        Info:      c.Hex("#06b6d4"),  // Cyan
        Secondary: c.Hex("#64748b"),  // Slate
    }
    
    // 2. Configurer le panel
    panel := engine.NewPanel("admin").
        SetPath("/admin").
        SetBrandName("My App").
        WithColors(customColors).  // Appliquer le schéma
        WithDarkMode(true)
    
    // 3. Les couleurs sont maintenant utilisées partout :
    //    - Boutons primary/danger/success/warning/info/secondary
    //    - Badges avec les mêmes couleurs
    //    - Alertes et notifications
    //    - Barres de progression
    //    - États de formulaires
    //    - Gradients et icônes
    
    // 4. Le CSS est généré automatiquement au boot
    router := panel.Router()
    
    // 5. Servir l'application
    http.ListenAndServe(":8080", router)
}
```

---

## Best Practices

1. **Cohérence** — Utilisez les couleurs sémantiques de manière cohérente
2. **Contraste** — Vérifiez le contraste WCAG pour l'accessibilité
3. **Dark mode** — Testez vos couleurs en mode sombre
4. **Limitation** — Ne surchargez pas avec trop de couleurs personnalisées
5. **Documentation** — Documentez votre schéma de couleurs pour l'équipe

---

## Différences avec Filament

| Aspect | Filament PHP | SublimeGo |
|--------|-------------|-----------|
| Couleurs sémantiques | 6 (primary, danger, success, warning, info, gray) | 6 (primary, danger, success, warning, info, secondary) |
| Configuration | `FilamentColor::register()` | `panel.WithColors()` ou `panel.WithSemanticColors()` |
| Génération de nuances | ✅ Auto (50-950) | ✅ Auto (50-950) |
| CSS variables | ✅ | ✅ |
| Support hex/rgb | ✅ | ✅ |

SublimeGo suit exactement le même modèle que Filament pour une expérience familière.
