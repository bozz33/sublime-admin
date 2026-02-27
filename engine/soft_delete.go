package engine

import "context"

// SoftDeletable is an optional interface for resources that support soft deletion.
// When implemented, the CRUDHandler will call SoftDelete instead of Delete,
// and expose additional routes for Restore and ForceDelete.
//
// Routes added when SoftDeletable is implemented:
//
//	POST /{slug}/{id}/restore      → Restore
//	DELETE /{slug}/{id}/force      → ForceDelete
//	DELETE /{slug}/{id}            → SoftDelete (replaces hard delete)
type SoftDeletable interface {
	// SoftDelete marks the item as deleted without removing it from the database.
	SoftDelete(ctx context.Context, id string) error

	// Restore un-deletes a soft-deleted item.
	Restore(ctx context.Context, id string) error

	// ForceDelete permanently removes an item (even if already soft-deleted).
	ForceDelete(ctx context.Context, id string) error

	// IsDeleted returns true if the given item is currently soft-deleted.
	IsDeleted(item any) bool
}

// TrashedMode controls which records are included in list queries.
type TrashedMode string

const (
	// TrashedModeActive returns only non-deleted records (default).
	TrashedModeActive TrashedMode = "active"

	// TrashedModeOnly returns only soft-deleted records.
	TrashedModeOnly TrashedMode = "only"

	// TrashedModeAll returns all records regardless of deletion state.
	TrashedModeAll TrashedMode = "all"
)

// TrashedModeFromString converts the "trashed" filter value to a TrashedMode.
// Returns TrashedModeActive for unknown values.
func TrashedModeFromString(s string) TrashedMode {
	switch TrashedMode(s) {
	case TrashedModeOnly:
		return TrashedModeOnly
	case TrashedModeAll:
		return TrashedModeAll
	default:
		return TrashedModeActive
	}
}
