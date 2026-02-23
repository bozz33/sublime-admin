# Analyse approfondie : Filament + Livewire — Architecture complète

> Analyse du code source complet des dossiers `filament/` et `livewire/`
> pour comprendre comment appliquer les mêmes patterns à SublimeGo.

---

## TABLE DES MATIÈRES

1. [Architecture globale Filament](#1-architecture-globale)
2. [Le socle : support/ (Component, ViewComponent, Concerns)](#2-support)
3. [Tables : colonnes, rendu, filtres, actions, pagination](#3-tables)
4. [Forms : champs, layouts, validation](#4-forms)
5. [Actions : modals, lifecycle, confirmation](#5-actions)
6. [Widgets : Stats, Charts](#6-widgets)
7. [Panel & Resources : le pipeline complet](#7-panel)
8. [Livewire : le moteur de réactivité](#8-livewire)
9. [Synthèse : Filament vs SublimeGo](#9-synthese)
10. [Plan de correction complet](#10-plan)

---

## 1. Architecture globale Filament {#1-architecture-globale}

Filament est composé de **10 packages** indépendants mais interconnectés :

```
filament/
├── support/       ← SOCLE : Component, ViewComponent, EvaluatesClosures, Concerns (34 traits)
├── schemas/       ← Schema engine (layout, state, components)
├── tables/        ← Table engine (columns, filters, bulk actions, pagination)
├── forms/         ← Form engine (fields, validation, layouts)
├── actions/       ← Action engine (modals, lifecycle, confirmation)
├── widgets/       ← Widget engine (stats, charts)
├── infolists/     ← Read-only detail views
├── notifications/ ← Toast notifications
├── query-builder/ ← Advanced query builder
└── filament/      ← Panel (Resources, Pages, Navigation, Auth)
```

### Principe fondamental : **TOUT est un ViewComponent qui sait se rendre**

```
Component (abstract)
  ├── EvaluatesClosures   ← évalue les closures avec injection de dépendances
  ├── Configurable        ← setUp() appelé à la construction
  ├── Macroable           ← extensible dynamiquement
  └── Conditionable       ← when(), unless()

ViewComponent extends Component implements Htmlable
  ├── $view              ← nom de la vue Blade
  ├── toHtml()           ← si HasEmbeddedView → toEmbeddedHtml(), sinon render()
  ├── render()           ← retourne view($this->getView(), [...])
  └── getExtraViewData() ← données supplémentaires pour la vue
```

**Clé** : `toHtml()` vérifie si le composant implémente `HasEmbeddedView`. Si oui, il appelle
`toEmbeddedHtml()` qui génère le HTML directement en PHP (pas de vue Blade). C'est le cas
de `TextColumn`, `ImageColumn`, `IconColumn` — pour la performance.

---

## 2. Le socle : support/ {#2-support}

### 2.1 EvaluatesClosures — Le cœur de la configuration dynamique

```php
trait EvaluatesClosures {
    public function evaluate(mixed $value, array $namedInjections = []): mixed {
        if (! $value instanceof Closure) return $value;  // valeur statique
        // Résout les dépendances par nom et type, puis appelle la closure
        $dependencies = [];
        foreach ((new ReflectionFunction($value))->getParameters() as $param) {
            $dependencies[] = $this->resolveClosureDependencyForEvaluation($param, ...);
        }
        return $value(...$dependencies);
    }
}
```

**Pourquoi c'est important** : TOUTE propriété dans Filament peut être soit une valeur statique,
soit une closure. Exemple :
```php
TextColumn::make('status')
    ->color(fn (string $state) => match ($state) {
        'active' => 'success',    // ← la couleur dépend de la VALEUR
        'inactive' => 'danger',
    })
```

La closure reçoit automatiquement `$state`, `$record`, `$column`, `$livewire` par injection.

### 2.2 Les 34 Concerns de support/

Chaque Concern est un **trait réutilisable** qui ajoute une capacité :

| Concern | Méthodes | Utilisé par |
|---------|----------|-------------|
| `HasColor` | `color()`, `getColor($state)` | Column, Action, Stat, Badge |
| `HasIcon` | `icon()`, `getIcon($state)` | Column, Action, Stat |
| `HasBadge` | `badge()`, `getBadge()` | Action, NavigationItem |
| `HasAlignment` | `alignment()`, `getAlignment()` | Column, Schema |
| `HasWidth` | `width()`, `getWidth()` | Column |
| `HasPlaceholder` | `placeholder()` | Column, Field |
| `HasExtraAttributes` | `extraAttributes()` | Tout |
| `CanGrow` | `grow()`, `canGrow()` | Column |
| `CanSpanColumns` | `columnSpan()`, `columnStart()` | Column, Field, Widget |
| `CanWrap` | `wrap()`, `canWrap()` | Column, TextColumn |
| `HasWeight` | `weight()`, `getWeight()` | TextColumn |
| `HasFontFamily` | `fontFamily()` | TextColumn |
| `HasLineClamp` | `lineClamp()` | TextColumn |
| `CanBeCopied` | `copyable()`, `isCopyable()` | TextColumn |
| `HasCellState` | `getState()`, `getStateFromRecord()` | Column |
| `HasVerticalAlignment` | `verticalAlignment()` | Column |

### 2.3 HasCellState — Comment Filament extrait la valeur

```php
trait HasCellState {
    public function getState(): mixed {
        $state = ($this->getStateUsing !== null)
            ? $this->evaluate($this->getStateUsing)    // ← custom accessor
            : $this->getStateFromRecord();              // ← auto via Eloquent
        if (blank($state)) $state = $this->getDefaultState();
        return $state;
    }

    public function getStateFromRecord(): mixed {
        $record = $this->getRecord();
        // Gère les relations (user.name), les nested attributes, etc.
        return data_get($record, $this->getName());
    }
}
```

**Différence avec SublimeGo** : Filament passe le **record complet** à la colonne, qui extrait
elle-même la valeur. SublimeGo convertit tout en `[]string` avant le rendu → perte d'info.

---

## 3. Tables : le système complet {#3-tables}

### 3.1 Table.php — Configuration via traits (30 Concerns)

```php
class Table extends ViewComponent {
    use HasColumns;           // columns(), getVisibleColumns()
    use HasFilters;           // filters(), getFilters(), filtersLayout()
    use HasBulkActions;       // bulkActions(), isSelectionEnabled()
    use HasActions;           // getAction(), cacheAction()
    use HasHeaderActions;     // headerActions()
    use HasToolbarActions;    // toolbarActions()
    use CanPaginateRecords;   // paginated(), paginationPageOptions()
    use CanSearchRecords;     // searchable(), isSearchable()
    use CanSortRecords;       // sortable(), getSortColumn()
    use CanGroupRecords;      // grouping(), getGrouping()
    use CanReorderRecords;    // reorderable(), isReordering()
    use HasEmptyState;        // emptyStateHeading(), emptyStateIcon()
    use HasColumnManager;     // hasColumnManager(), toggleableColumns()
    use HasContent;           // content(), getContent()
    use HasHeader;            // header(), getHeader()
    use HasRecordUrl;         // recordUrl(), getRecordUrl()
    use HasRecordAction;      // recordAction(), getRecordAction()
    // ... 30 traits au total
}
```

### 3.2 Columns — Hiérarchie et types

```
Column extends ViewComponent
  ├── 30 Concerns (BelongsToTable, CanBeSearchable, CanBeSortable, HasColor, HasIcon, ...)
  ├── record($record)         ← reçoit le record courant
  ├── renderInLayout()        ← rendu dans un content layout
  └── getExtraViewData()      ← données pour la vue

TextColumn extends Column implements HasEmbeddedView
  ├── badge(), isBadge()      ← mode badge (pill coloré)
  ├── CanFormatState          ← date(), money(), numeric(), limit()
  ├── HasColor                ← color(fn($state) => ...)
  ├── HasDescription          ← description('above'|'below')
  ├── HasIcon                 ← icon(), iconPosition()
  ├── HasWeight               ← weight('bold')
  ├── CanBeCopied             ← copyable()
  ├── CanWrap                 ← wrap()
  └── toEmbeddedHtml()        ← GÉNÈRE LE HTML DIRECTEMENT (550 lignes !)

ImageColumn extends Column implements HasEmbeddedView
  ├── circular(), square()
  ├── stacked(), overlap(), ring()
  ├── limit(), limitedRemainingText()
  ├── disk(), visibility()
  └── toEmbeddedHtml()        ← <img> avec fallback, stacking, etc.

IconColumn extends Column implements HasEmbeddedView
  ├── boolean()               ← mode check/cross
  ├── HasColor, HasIcon
  └── toEmbeddedHtml()

BadgeColumn extends TextColumn (DEPRECATED → use TextColumn->badge())
BooleanColumn extends IconColumn (DEPRECATED → use IconColumn->boolean())

SelectColumn, CheckboxColumn, TextInputColumn, ToggleColumn
  → colonnes ÉDITABLES inline (pas juste affichage)
```

### 3.3 Le rendu des cellules — La ligne clé

```php
// index.blade.php ligne 1934-1996
@foreach ($columns as $column)
    @php
        $column->record($record);       // 1. Donne le record à la colonne
        $column->rowLoop($loop);        // 2. Donne le loop
        $column->recordKey($recordKey); // 3. Donne la clé
    @endphp

    <td {{ $column->getExtraCellAttributeBag()->class([...]) }}>
        <{{ $wrapperTag }} ...>
            {{ $column }}               // 4. LA COLONNE SE REND ELLE-MÊME
        </{{ $wrapperTag }}>            //    via toHtml() → toEmbeddedHtml()
    </td>
@endforeach
```

**La table ne sait RIEN du type de colonne.** Elle boucle et dit "rends-toi".

### 3.4 Filters — Objets configurables

```php
// Utilisation dans une Resource :
Table::make()
    ->filters([
        SelectFilter::make('status')
            ->options(['active' => 'Active', 'inactive' => 'Inactive']),
        TernaryFilter::make('verified')
            ->label('Email verified'),
        Filter::make('created_at')
            ->form([DatePicker::make('from'), DatePicker::make('until')])
            ->query(fn (Builder $query, array $data) => $query->when(...)),
    ])
```

Chaque filtre est un objet avec `getSchemaComponents()` qui retourne les champs du formulaire.

### 3.5 Pagination — Configurable via la Table

```php
Table::make()
    ->paginated([10, 25, 50, 100, 'all'])   // options de per_page
    ->defaultPaginationPageOption(25)
    ->extremePaginationLinks()                // first/last page links
    ->paginationMode(PaginationMode::Simple)  // simple ou cursor
```

### 3.6 BulkActions & HeaderActions

```php
Table::make()
    ->headerActions([
        CreateAction::make(),
        ExportAction::make(),
    ])
    ->bulkActions([
        BulkActionGroup::make([
            DeleteBulkAction::make(),
            ExportBulkAction::make(),
        ]),
    ])
```

Les actions sont des **objets Action** avec modal, confirmation, lifecycle hooks.

---

## 4. Forms : champs et layouts {#4-forms}

### 4.1 Hiérarchie

```
Component (schemas/)
  └── Field extends Component
        ├── HasName, HasLabel
        ├── CanBeValidated, CanBeMarkedAsRequired
        ├── HasHelperText, HasHint
        └── $view = 'filament-forms::components.text-input'

TextInput extends Field
  ├── email(), password(), tel(), url(), numeric()
  ├── mask(), prefix(), suffix()
  ├── autocomplete(), autofocus()
  ├── readOnly(), disabled()
  └── $view = 'filament-forms::components.text-input'

Select extends Field
  ├── options(), searchable(), preload()
  ├── relationship()          ← charge depuis Eloquent
  ├── createOptionForm()      ← formulaire de création inline
  ├── editOptionForm()        ← formulaire d'édition inline
  └── native()                ← select HTML natif vs custom

FileUpload extends BaseFileUpload
  ├── image(), avatar()
  ├── disk(), directory(), visibility()
  ├── maxSize(), acceptedFileTypes()
  └── multiple(), reorderable()
```

### 4.2 Pattern commun : chaque champ a sa vue Blade

```php
class TextInput extends Field {
    protected string $view = 'filament-forms::components.text-input';
}
```

La vue Blade reçoit `$field` et se rend en utilisant les getters :
`$field->getLabel()`, `$field->isRequired()`, `$field->getPrefix()`, etc.

---

## 5. Actions : le système complet {#5-actions}

### 5.1 Action.php — 49 Concerns !

```php
class Action extends ViewComponent implements Arrayable {
    use CanBeDisabled, CanBeHidden;
    use CanOpenModal;           // modal(), slideOver(), modalWidth()
    use CanOpenUrl;             // url(), openUrlInNewTab()
    use CanRedirect;            // redirect(), redirectRoute()
    use CanRequireConfirmation; // requiresConfirmation()
    use CanNotify;              // successNotification(), failureNotification()
    use HasAction;              // action(Closure)
    use HasLifecycleHooks;      // before(), after()
    use HasSchema;              // form([...fields...])
    use HasLabel, HasName;
    use HasColor, HasIcon, HasBadge;
    use HasSize;                // size('sm'|'md'|'lg')
    use InteractsWithRecord;    // record(), getRecord()
    use CanBeRateLimited;       // rateLimited()
    use CanUseDatabaseTransactions; // databaseTransaction()
    // ... 49 traits
}
```

### 5.2 Lifecycle hooks

```php
Action::make('approve')
    ->before(function () { /* validation custom */ })
    ->action(function (Model $record) {
        $record->update(['status' => 'approved']);
    })
    ->after(function () { /* notification */ })
    ->requiresConfirmation()
    ->modalHeading('Approve this record?')
    ->modalDescription('This action cannot be undone.')
    ->color('success')
    ->icon('heroicon-o-check')
```

### 5.3 Modals

Les actions peuvent ouvrir des modals avec :
- Un formulaire (`->form([...fields...])`)
- Un infolist (`->infolist([...entries...])`)
- Du contenu custom (`->modalContent(view('...'))`)
- Des boutons footer custom (`->modalFooterActions([...])`)
- Slide-over (`->slideOver()`)

---

## 6. Widgets {#6-widgets}

### 6.1 Widget extends Livewire\Component

```php
abstract class Widget extends LivewireComponent {
    use CanBeLazy;
    protected string $view;
    protected int|string|array $columnSpan = 1;
    public function render(): View { return view($this->view, $this->getViewData()); }
}
```

### 6.2 StatsOverviewWidget

```php
class StatsOverviewWidget extends Widget {
    protected function getStats(): array {
        return [
            Stat::make('Total Revenue', '$192.1k')
                ->description('32k increase')
                ->descriptionIcon('heroicon-m-arrow-trending-up')
                ->descriptionColor('success')
                ->chart([7, 2, 10, 3, 15, 4, 17])
                ->color('success'),
        ];
    }
}
```

`Stat` est un `Component` avec `HasColor`, `HasDescription`, `HasLabel` + `chart()`, `icon()`.

### 6.3 ChartWidget

```php
abstract class ChartWidget extends Widget {
    abstract protected function getType(): string;  // 'line', 'bar', 'pie', etc.
    protected function getData(): array { return []; }
    protected function getOptions(): array|RawJs|null { return null; }
}
```

---

## 7. Panel & Resources {#7-panel}

### 7.1 Panel.php — Configuration via 48 Concerns

```php
class Panel extends Component {
    use HasAuth, HasAvatars, HasBrandLogo, HasColors;
    use HasComponents, HasDarkMode, HasNavigation;
    use HasRoutes, HasMiddleware, HasPlugins;
    use HasSidebar, HasTopbar, HasTheme;
    // ... 48 traits
}
```

### 7.2 Resource.php — Le point d'entrée CRUD

```php
abstract class Resource {
    protected static ?string $model = null;

    public static function form(Schema $schema): Schema { return $schema; }
    public static function table(Table $table): Table { return $table; }
    public static function infolist(Schema $schema): Schema { return $schema; }

    // Auto-découverte du modèle depuis le nom de la classe
    public static function getModel(): string {
        return static::$model ?? str(class_basename(static::class))
            ->beforeLast('Resource')
            ->prepend(app()->getNamespace() . 'Models\\');
    }
}
```

### 7.3 ListRecords — La page de liste

```php
class ListRecords extends Page implements HasTable {
    use InteractsWithTable;

    // URL state automatique via Livewire #[Url]
    public ?array $tableFilters = null;
    public ?string $tableSearch = '';
    public ?string $tableSort = null;

    protected function makeTable(): Table {
        return $this->makeBaseTable()
            ->query(fn () => $this->getTableQuery())
            ->recordUrl(fn (Model $record) => $this->getResourceUrl('edit', ['record' => $record]))
            ->recordAction(fn (Model $record) => /* view or edit */);
        // Puis : static::getResource()::configureTable($table);
    }
}
```

**Pipeline complet** :
1. `ListRecords::makeTable()` crée la Table
2. `Resource::configureTable($table)` appelle `Resource::table($table)` (défini par le dev)
3. Le dev configure colonnes, filtres, actions dans `table()`
4. Livewire rend la vue `filament-tables::index`
5. La vue boucle sur `$columns`, chaque colonne se rend via `{{ $column }}`

---

## 8. Livewire : le moteur {#8-livewire}

### 8.1 Component — Le composant réactif

```php
abstract class Component {
    use HandlesEvents;       // dispatch(), on()
    use HandlesRedirects;    // redirect()
    use HandlesValidation;   // validate(), rules()
    use HandlesAttributes;   // #[Locked], #[Url], #[Computed]
    use HandlesStreaming;     // stream()
}
```

### 8.2 HandleComponents — Mount + Render + Dehydrate

```php
public function mount($name, $params = [], $key = null) {
    $component = app('livewire')->new($name);
    trigger('mount', $component, $params, $key, $parent);
    $html = $this->render($component, '<div></div>');
    trigger('dehydrate', $component, $context);
    $snapshot = $this->snapshot($component, $context);
    // Insère wire:snapshot et wire:effects dans le HTML
    return $html;
}
```

### 8.3 Équivalent dans SublimeGo

SublimeGo utilise **HTMX + Alpine.js** au lieu de Livewire :
- **HTMX** = requêtes partielles (hx-get, hx-post, hx-swap)
- **Alpine.js** = réactivité côté client (x-data, x-show, x-on)
- **SSE** = notifications en temps réel

C'est un bon choix. Pas besoin de copier Livewire. Mais il faut copier le **pattern de rendu** :
chaque composant sait se rendre lui-même.

---

## 9. Synthèse : Filament vs SublimeGo {#9-synthese}

### 9.1 Ce que Filament fait bien (et qu'on doit copier)

| Pattern Filament | SublimeGo actuel | Ce qu'il faut faire |
|-----------------|-------------------|---------------------|
| **Chaque colonne se rend elle-même** (`{{ $column }}`) | `renderCell()` centralisé | Ajouter `Render()` à `table.Column` |
| **Le record complet est passé à la colonne** | `Row.Cells []string` | Passer le record + la valeur |
| **Closures évaluées dynamiquement** (`color(fn($state) => ...)`) | Valeurs statiques | Ajouter `func(value string) string` callbacks |
| **Concerns/Traits réutilisables** (HasColor, HasIcon, HasDescription) | Champs dupliqués dans chaque struct | Interfaces Go + embedding |
| **Actions = objets configurables** avec modal, lifecycle | Juste des URLs | Enrichir `actions.Action` |
| **Filtres = objets avec formulaire** | `FilterDef{Key, Label, Type, Options}` | Enrichir `table.Filter` |
| **Pagination configurable** (options, mode, scroll) | Basique | Enrichir `Pagination` |

### 9.2 Ce que SublimeGo a déjà (et qu'il faut garder)

- ✅ `table/` package avec `TextColumn`, `BadgeColumn`, `ImageColumn`, `BooleanColumn`, `DateColumn`
- ✅ API fluent (`Text("name").WithLabel("Nom").Sortable()`)
- ✅ `Column` interface (`Key()`, `Label()`, `Type()`, `Value()`, `IsSortable()`)
- ✅ `actions/` package avec `Action` struct
- ✅ `widget/` package avec Stats et Charts
- ✅ HTMX + Alpine.js (pas besoin de Livewire)
- ✅ Templates Templ (équivalent de Blade)

### 9.3 Le problème fondamental

**Le CRUD (`engine/`) n'utilise PAS le package `table/` !**

```
engine/contract.go:
  type Column struct { Key, Label, Type string }  ← struct PLATE
  type Row struct { Cells []string }               ← valeurs STRING

table/columns.go:
  type TextColumn struct { ... }                   ← objet RICHE
  type BadgeColumn struct { ColorMap map[string]string }
  // JAMAIS utilisé par le CRUD !
```

---

## 10. Plan de correction complet {#10-plan}

### Phase 1 : Ajouter `Render()` à l'interface `table.Column`

```go
// table/table.go
type Column interface {
    Key() string
    Label() string
    Type() string
    IsSortable() bool
    IsSearchable() bool
    IsCopyable() bool
    Value(item any) string
    Render(value string, record any) templ.Component  // NOUVEAU
}
```

Chaque type de colonne implémente `Render()` avec son propre template Templ.

### Phase 2 : Créer les templates Templ pour chaque type de colonne

```
table/
  ├── columns.go          ← types existants + Render()
  ├── column_render.templ ← templates Templ pour chaque type
  │     ├── TextCellView(value, prefix, suffix string)
  │     ├── BadgeCellView(value, color string)
  │     ├── ImageCellView(url string, circular bool)
  │     ├── BooleanCellView(value bool)
  │     ├── DateCellView(value string)
  │     └── AvatarCellView(name, initials, bgColor string)
  └── ...
```

### Phase 3 : Migrer `engine/` de `engine.Column` vers `table.Column`

- `TableState.Columns` : `[]engine.Column` → `[]table.Column`
- `BaseResource.SetTableColumns()` : accepte `...table.Column`
- `BuildTableState()` : utilise `col.Value(item)` au lieu de `getColumnValue()`
- Garder `engine.Column` comme alias de compatibilité (wrapper vers TextColumn)

### Phase 4 : Modifier `list.templ` pour le rendu par composant

```templ
// AVANT (centralisé) :
for j, cell := range row.Cells {
    @renderCell(state.Columns, j, cell)  // gros switch
}

// APRÈS (chaque colonne se rend) :
for j, cell := range row.Cells {
    @state.Columns[j].Render(cell, nil)  // la colonne se rend elle-même
}
```

### Phase 5 : Enrichir les colonnes

**TextColumn** :
- `Badge()` → active le rendu badge (comme Filament)
- `Description(field string)` → sous-texte
- `Icon(icon string)` → icône avant/après
- `Color(fn func(string) string)` → couleur dynamique
- `Weight(w string)` → bold, semibold
- `Copyable()` → bouton copier
- `Prefix(s string)`, `Suffix(s string)` → préfixe/suffixe
- `Limit(n int)` → tronquer le texte

**BadgeColumn** :
- `Colors(map[string]string)` → couleurs par valeur
- `Icon(fn func(string) string)` → icône par valeur

**ImageColumn** :
- `Circular()` → rond
- `Square()` → carré
- `Size(w, h int)` → dimensions custom

**BooleanColumn** :
- `TrueIcon(icon string)`, `FalseIcon(icon string)`
- `TrueColor(color string)`, `FalseColor(color string)`

**DateColumn** :
- `DateFormat(format string)` ← existe déjà
- `ShowRelative()` ← existe déjà
- `Since()` → "il y a 2 heures"

**NOUVEAUX** :
- `IconColumn` → icône avec couleur conditionnelle
- `AvatarColumn` → cercle coloré avec initiales + nom
- `ColorColumn` → pastille de couleur

### Phase 6 : Enrichir les filtres

```go
// Actuellement :
type FilterDef struct { Key, Label, Type string; Options []FilterOption }

// Cible :
type SelectFilter struct { ... }  // dropdown
type BooleanFilter struct { ... } // toggle
type DateFilter struct { ... }    // date range
type CustomFilter struct { ... }  // formulaire custom
```

### Phase 7 : Enrichir les actions

```go
// Actuellement :
type Action struct { Label, Icon, Color, URL string }

// Cible :
type Action struct {
    // ... existant
    RequiresConfirmation bool
    ModalHeading         string
    ModalDescription     string
    BeforeFunc           func(ctx context.Context) error
    AfterFunc            func(ctx context.Context) error
}
```

### Phase 8 : Page header fidèle au template

```
┌─────────────────────────────────────────────────┐
│ Dashboard > Users                    [+ Ajouter] │
│ Manage your users                                │
├─────────────────────────────────────────────────┤
│ [Search...] [Filters ▾] [Columns ▾]            │
├─────────────────────────────────────────────────┤
│ Table...                                         │
├─────────────────────────────────────────────────┤
│ Showing 1-10 of 100  [< 1 2 3 ... 10 >]        │
└─────────────────────────────────────────────────┘
```

### Ordre d'implémentation

1. **Phase 1-2** : `Render()` + templates Templ (fondation)
2. **Phase 3-4** : Migration engine → table.Column + list.templ
3. **Phase 5** : Enrichir les colonnes (badge, avatar, icon, etc.)
4. **Phase 8** : Page header fidèle
5. **Phase 6** : Filtres enrichis
6. **Phase 7** : Actions enrichies

---

## ANNEXE A : Infolists (read-only detail views) {#annexe-infolists}

> Package `filament/infolists/` — 103 items

Les infolists sont le **miroir read-only des forms**. Même pattern, même hiérarchie.

### Hiérarchie

```
Entry extends Component (schemas/)
  ├── HasName, HasLabel, HasPlaceholder
  ├── HasAlignment, HasHelperText, HasHint, HasTooltip
  ├── CanOpenUrl
  └── $viewIdentifier = 'entry'

TextEntry extends Entry implements HasEmbeddedView
  ├── badge(), isBadge()           ← identique à TextColumn
  ├── CanFormatState               ← date(), money(), numeric(), limit()
  ├── HasColor, HasIcon, HasIconColor
  ├── HasWeight, HasFontFamily, HasLineClamp
  ├── CanBeCopied, CanWrap
  ├── bulleted(), prose(), listWithLineBreaks()
  └── toEmbeddedHtml()             ← génère le HTML directement

ImageEntry extends Entry implements HasEmbeddedView
  ├── circular(), square()
  ├── stacked(), overlap(), ring(), limit()
  ├── disk(), visibility(), defaultImageUrl()
  └── toEmbeddedHtml()

IconEntry extends Entry implements HasEmbeddedView
  ├── boolean()                    ← mode check/cross
  ├── trueIcon(), falseIcon(), trueColor(), falseColor()
  ├── HasColor, HasIcon
  └── toEmbeddedHtml()

ColorEntry extends Entry implements HasEmbeddedView
  → affiche une pastille de couleur

CodeEntry extends Entry implements HasEmbeddedView
  → affiche du code avec syntax highlighting

KeyValueEntry extends Entry
  → affiche des paires clé-valeur

RepeatableEntry extends Entry
  → affiche une liste de sous-entrées (relation hasMany)
```

### Pattern clé : Symétrie Form ↔ Infolist

| Form | Infolist | Table |
|------|----------|-------|
| `TextInput::make('name')` | `TextEntry::make('name')` | `TextColumn::make('name')` |
| `Select::make('status')` | `TextEntry::make('status')->badge()` | `TextColumn::make('status')->badge()` |
| `FileUpload::make('avatar')` | `ImageEntry::make('avatar')` | `ImageColumn::make('avatar')` |
| `Toggle::make('active')` | `IconEntry::make('active')->boolean()` | `IconColumn::make('active')->boolean()` |

**Leçon pour SublimeGo** : Notre package `infolist/` (à créer) devrait suivre exactement
le même pattern que `table/columns.go` mais pour l'affichage read-only en détail.

---

## ANNEXE B : Notifications {#annexe-notifications}

> Package `filament/notifications/` — 97 items

### Notification extends ViewComponent implements HasEmbeddedView

```php
class Notification extends ViewComponent implements Arrayable, HasEmbeddedView {
    use HasTitle, HasBody, HasIcon, HasIconColor, HasColor;
    use HasStatus;      // success(), danger(), warning(), info()
    use HasDuration;    // duration(5000), persistent()
    use HasActions;     // actions([Action::make('undo')->...])
    use HasDate;        // date()
    use HasId;          // id()
    use CanBeInline;    // inline()
}
```

### 3 modes de livraison

```php
// 1. Flash notification (session)
Notification::make()
    ->title('Saved successfully')
    ->success()
    ->send();                    // → session flash

// 2. Broadcast (WebSocket)
Notification::make()
    ->title('New order')
    ->body('Order #1234')
    ->broadcast($user);          // → Laravel Broadcasting

// 3. Database (persistent)
Notification::make()
    ->title('New message')
    ->sendToDatabase($user);     // → notifications table
```

### Sérialisation

```php
public function toArray(): array {
    return [
        'id' => $this->getId(),
        'actions' => array_map(fn ($a) => $a->toArray(), $this->getActions()),
        'body' => $this->getBody(),
        'color' => $this->getColor(),
        'duration' => $this->getDuration(),
        'icon' => $icon,
        'iconColor' => $this->getIconColor(),
        'status' => $this->getStatus(),
        'title' => $this->getTitle(),
    ];
}
```

**Leçon pour SublimeGo** : Notre `notifications/` package a déjà un builder pattern similaire
avec SSE. La sérialisation `toArray()` est utile pour le stockage en DB.

---

## ANNEXE C : Query Builder {#annexe-query-builder}

> Package `filament/query-builder/` — 90 items

### Constraint — Le composant de base

```php
class Constraint extends Component {
    use HasLabel, HasName, HasOperators, HasIcon;
    // Chaque contrainte a un attribut, un label, et des opérateurs
}
```

### 6 types de contraintes

| Type | Description | Opérateurs |
|------|-------------|------------|
| `TextConstraint` | Texte libre | contains, starts_with, ends_with, equals |
| `NumberConstraint` | Numérique | equals, gt, gte, lt, lte, between |
| `DateConstraint` | Date/heure | before, after, between, is_today, is_this_week |
| `BooleanConstraint` | Vrai/faux | is_true, is_false |
| `SelectConstraint` | Choix parmi options | is, is_not |
| `RelationshipConstraint` | Relations Eloquent | has, doesnt_have, count |

**Leçon pour SublimeGo** : Ce package est avancé et pas prioritaire. Mais le pattern
`Constraint` avec `Operators` est intéressant pour un futur système de filtres avancés.

---

## ANNEXE D : Schemas (layout engine) — Complet {#annexe-schemas}

> Package `filament/schemas/` — 228 items

### Component — La base de TOUT layout

```php
class Component extends ViewComponent {
    // 30 Concerns :
    use BelongsToContainer, BelongsToModel;
    use CanBeConcealed, CanBeDisabled, CanBeHidden;
    use CanBeGridContainer, CanBeLiberatedFromContainerGrid;
    use CanBeRepeated, CanPartiallyRender, CanPoll;
    use CanGrow, CanOrderColumns, CanSpanColumns;
    use Cloneable;
    use HasActions, HasChildComponents;
    use HasColumns, HasGap;
    use HasEntryWrapper, HasFieldWrapper;
    use HasExtraAttributes;
    use HasHeadings, HasId, HasInlineLabel, HasKey;
    use HasMaxWidth, HasMeta;
    use HasState, HasStateBindingModifiers;
}
```

### Closure dependency injection

```php
// Component résout automatiquement les paramètres des closures :
protected function resolveDefaultClosureDependencyForEvaluationByName(string $parameterName): array {
    return match ($parameterName) {
        'context', 'operation' => [$this->getContainer()->getOperation()],
        'get' => [$this->makeGetUtility()],      // Get utility pour lire d'autres champs
        'livewire' => [$this->getLivewire()],
        'model' => [$this->getModel()],
        'rawState' => [$this->getRawState()],
        'record' => [$this->getRecord()],
        'set' => [$this->makeSetUtility()],       // Set utility pour modifier d'autres champs
        'state' => [$this->getState()],
        default => parent::...,
    };
}
```

### Layout components

| Component | Description | Concerns |
|-----------|-------------|----------|
| `Section` | Carte avec heading, description, icône | CanBeCollapsed, HasDescription, HasHeading, HasIcon, HasHeaderActions, HasFooterActions |
| `Tabs` | Onglets | CanPersistTab, HasChildComponents |
| `Wizard` | Étapes séquentielles | HasChildComponents |
| `Grid` | Grille responsive | HasColumns |
| `Group` | Groupement invisible | HasChildComponents |
| `Flex` | Flexbox | HasChildComponents |
| `Fieldset` | Fieldset HTML | HasChildComponents |
| `Form` | Formulaire | HasChildComponents |
| `Html` | HTML brut | - |
| `Text` | Texte | HasColor, HasWeight |
| `Icon` | Icône | HasColor, HasIcon |
| `Image` | Image | - |

### HasState — 27 688 octets, le plus gros Concern

Gère tout le cycle de vie de l'état d'un composant :
- `statePath()`, `getStatePath()` — chemin dans le state tree
- `getState()`, `getRawState()` — lecture
- `state()`, `default()` — écriture
- `afterStateHydrated()`, `afterStateUpdated()` — hooks
- `dehydrateState()`, `hydrateState()` — sérialisation
- `mutateDehydratedState()`, `mutateRelationshipDataBeforeSave()` — transformation

---

## ANNEXE E : Panel & Auth — Complet {#annexe-panel}

> Package `filament/filament/` — 1745 items

### Panel — 48 Concerns

```php
class Panel extends Component {
    use HasAuth;           // login(), registration(), passwordReset(), emailVerification()
    use HasAvatars;        // defaultAvatarProvider()
    use HasBrandLogo;      // brandLogo(), brandLogoHeight()
    use HasBrandName;      // brandName()
    use HasColors;         // colors(['primary' => '#...'])
    use HasComponents;     // resources(), pages(), widgets(), livewireComponents()
    use HasDarkMode;       // darkMode()
    use HasFont;           // font('Inter')
    use HasIcons;          // icons([...])
    use HasMiddleware;     // middleware([...]), authMiddleware([...])
    use HasNavigation;     // navigation(), navigationItems(), navigationGroups()
    use HasPlugins;        // plugins([...])
    use HasRoutes;         // routes(), path()
    use HasSidebar;        // sidebarCollapsibleOnDesktop(), sidebarFullyCollapsibleOnDesktop()
    use HasTheme;          // viteTheme(), theme()
    use HasTopbar;         // topbar()
    use HasTopNavigation;  // topNavigation()
    // ... 48 au total
}
```

### Auth Pages

| Page | Description | Fonctionnalités |
|------|-------------|-----------------|
| `Login` | Connexion | Rate limiting, MFA, remember me, form schema |
| `Register` | Inscription | Form schema, validation |
| `EditProfile` | Profil | Form schema, password change |
| `PasswordReset/RequestPasswordReset` | Mot de passe oublié | Email, rate limiting |
| `PasswordReset/ResetPassword` | Reset | Token validation |
| `EmailVerification/EmailVerificationPrompt` | Vérification email | Resend, rate limiting |

### MultiFactor Auth

```
MultiFactor/
  ├── App/          ← TOTP (Google Authenticator)
  ├── Email/        ← Code par email
  ├── Contracts/    ← MultiFactorAuthenticationProvider interface
  └── Pages/        ← MultiFactorChallenge page
```

### Navigation

```php
NavigationItem extends Component {
    // label, icon, activeIcon, badge, badgeColor, url, sort, group
    // childItems, isActive, isHidden, shouldOpenUrlInNewTab
}

NavigationGroup extends Component {
    // label, icon, items, isCollapsed, isCollapsible
    // extraSidebarAttributes, extraTopbarAttributes
}

NavigationManager {
    // Gère la construction et le cache de la navigation
}
```

### Dashboard

```php
class Dashboard extends Page {
    protected static string $routePath = '/';
    protected static ?int $navigationSort = -2;

    public function content(Schema $schema): Schema {
        return $schema->components([
            $this->getFiltersFormContentComponent(),
            $this->getWidgetsContentComponent(),  // Grid de widgets
        ]);
    }
}
```

### Http

```
Http/
  ├── Controllers/
  │     ├── Auth/     ← EmailVerificationController, LogoutController
  │     └── AssetController  ← Sert les assets compilés
  └── Middleware/
        ├── Authenticate
        ├── DisableBladeIconComponents
        ├── DispatchServingFilamentEvent
        ├── IdentifyTenant
        ├── MirrorConfigToSubpackages
        └── SetUpPanel
```

---

## ANNEXE F : Livewire — Complet {#annexe-livewire}

> Package `livewire/livewire/` — 267 items

### Component — 11 traits

```php
abstract class Component {
    use HandlesEvents;                    // dispatch(), on(), listeners
    use HandlesRedirects;                 // redirect(), redirectRoute()
    use HandlesStreaming;                 // stream()
    use HandlesAttributes;                // #[Locked], #[Url], #[Computed], #[Renderless]
    use HandlesValidation;                // validate(), rules(), messages()
    use HandlesFormObjects;               // Form objects
    use HandlesJsEvaluation;              // js()
    use HandlesReleaseTokens;             // Token-based concurrency
    use HandlesPageComponents;            // Full-page components
    use HandlesDisablingBackButtonCache;  // Disable browser back cache
    use InteractsWithProperties;          // Property access
}
```

### HandleComponents — Le cœur du lifecycle

```
mount($name, $params, $key)
  1. new($name)                    ← instancie le composant
  2. trigger('mount', ...)         ← boot → initialize → mount → booted
  3. render($component)            ← rend la vue Blade
  4. trigger('dehydrate', ...)     ← sérialise l'état
  5. snapshot($component)          ← crée le snapshot (data + memo + checksum)
  6. insertAttributes(wire:snapshot, wire:effects)

update($snapshot, $updates, $calls)
  1. fromSnapshot($snapshot)       ← vérifie checksum, hydrate
  2. trigger('hydrate', ...)       ← boot → initialize → hydrate → booted
  3. updateProperties(...)         ← applique les updates (updating* → updated*)
  4. callMethods(...)              ← appelle les méthodes côté serveur
  5. render($component)            ← re-rend la vue
  6. trigger('dehydrate', ...)     ← re-sérialise
  7. snapshot($component)          ← nouveau snapshot
```

### 52 Features (hooks système)

| Feature | Description |
|---------|-------------|
| `SupportLifecycleHooks` | boot, mount, hydrate, dehydrate, rendering, rendered, updating*, updated* |
| `SupportPagination` | Pagination auto avec URL sync |
| `SupportEvents` | dispatch(), listeners, event bus |
| `SupportValidation` | validate(), rules(), messages(), withValidator() |
| `SupportComputed` | #[Computed] properties avec cache |
| `SupportAttributes` | #[Locked], #[Url], #[Renderless] |
| `SupportQueryString` | Sync automatique URL ↔ propriétés |
| `SupportFileUploads` | Upload avec preview, validation, storage |
| `SupportFormObjects` | Form objects pour grouper la logique |
| `SupportModels` | Hydration/dehydration automatique des modèles |
| `SupportNavigate` | SPA-like navigation (wire:navigate) |
| `SupportLazyLoading` | Lazy loading de composants |
| `SupportStreaming` | Server-sent streaming |
| `SupportRedirects` | Redirections côté serveur |
| `SupportTesting` | Helpers de test (assertSee, assertSet, etc.) |
| `SupportReactiveProps` | Props réactives entre composants |
| `SupportIsolating` | Isolation de composants (pas de re-render parent) |
| `SupportEntangle` | Sync bidirectionnelle JS ↔ PHP |
| `SupportPolling` | wire:poll pour refresh périodique |
| `SupportNestingComponents` | Composants imbriqués |

### Lifecycle Hooks en détail

```
MOUNT (premier rendu) :
  boot() → bootTraits() → initialize() → mount($params) → booted()

HYDRATE (requête suivante) :
  boot() → bootTraits() → initialize() → hydrate() → hydrateXx() → booted()

UPDATE (propriété modifiée) :
  updating($path, $value) → updatingXx($value) → [set value] → updated($path) → updatedXx($value)

RENDER :
  rendering($view, $data) → [render blade] → rendered($view, $html)

DEHYDRATE (après render) :
  dehydrate() → dehydrateXx($value)
```

### Property Synthesizers

Livewire sérialise/désérialise les propriétés PHP automatiquement :
- `CarbonSynth` — Carbon dates
- `CollectionSynth` — Laravel Collections
- `StringableSynth` — Stringable objects
- `EnumSynth` — PHP Enums
- `StdClassSynth` — stdClass objects
- `ArraySynth` — Arrays
- `IntSynth`, `FloatSynth` — Primitives

**Leçon pour SublimeGo** : Pas besoin de copier ce système. HTMX + Alpine.js gère
la réactivité côté client. Mais le pattern lifecycle (hooks before/after) est utile
pour nos Actions et notre CRUD handler.

---

## ANNEXE G : Inventaire complet des fichiers analysés {#annexe-inventaire}

### Filament (10 dossiers, ~4584 items)

| Dossier | Items | Fichiers clés analysés |
|---------|-------|----------------------|
| `support/` | 497 | `ViewComponent.php`, `Component.php`, `EvaluatesClosures.php`, `HasCellState.php`, 34 Concerns |
| `tables/` | 327 | `Table.php`, `Column.php`, `TextColumn.php`, `ImageColumn.php`, `IconColumn.php`, `HasColumns.php`, `HasFilters.php`, `HasBulkActions.php`, `CanPaginateRecords.php`, `index.blade.php` (2226 lignes) |
| `forms/` | 370 | `Field.php`, `TextInput.php`, `Select.php`, 48 Concerns |
| `actions/` | 1050 | `Action.php` (864 lignes, 49 Concerns), `CanOpenModal.php`, `HasLifecycleHooks.php`, `HasAction.php`, `HasSchema.php` |
| `widgets/` | 77 | `Widget.php`, `StatsOverviewWidget.php`, `ChartWidget.php`, `Stat.php` |
| `schemas/` | 228 | `Schema.php`, `Component.php` (30 Concerns), `Section.php`, 40 Concerns, `HasState.php` (27KB) |
| `infolists/` | 103 | `Entry.php`, `TextEntry.php`, `ImageEntry.php`, `IconEntry.php`, `ColorEntry.php`, 9 Concerns |
| `notifications/` | 97 | `Notification.php` (405 lignes), 10 Concerns, `HasStatus.php`, `HasActions.php` |
| `query-builder/` | 90 | `Constraint.php`, 6 types de contraintes, 4 Concerns |
| `filament/` | 1745 | `Panel.php` (48 Concerns), `Resource.php`, `ListRecords.php`, `CreateRecord.php`, `EditRecord.php`, `Login.php`, `Dashboard.php`, `NavigationItem.php`, `NavigationGroup.php`, `Page.php` |

### Livewire (1 dossier, 267 items)

| Sous-dossier | Fichiers clés analysés |
|-------------|----------------------|
| `src/` | `Component.php` (11 traits), `HandleComponents.php` (565 lignes — mount, update, render, hydrate, dehydrate, snapshot) |
| `Features/` | 52 features : `SupportLifecycleHooks`, `SupportPagination`, `SupportEvents`, `SupportValidation`, `SupportComputed`, `SupportAttributes` |
| `Mechanisms/` | `RenderComponent.php`, `HandleComponents/`, `ComponentRegistry.php`, `DataStore.php`, `ExtendBlade/` |
