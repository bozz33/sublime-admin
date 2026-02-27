package actions

import (
	"context"
	"testing"
)

// ---------------------------------------------------------------------------
// NewGroup tests
// ---------------------------------------------------------------------------

func TestNewGroup_Label(t *testing.T) {
	g := NewGroup("Actions")
	if g.Label != "Actions" {
		t.Errorf("expected Label='Actions', got '%s'", g.Label)
	}
}

func TestNewGroup_DefaultIcon(t *testing.T) {
	g := NewGroup("Actions")
	if g.Icon != "more_vert" {
		t.Errorf("expected default Icon='more_vert', got '%s'", g.Icon)
	}
}

func TestNewGroup_DefaultColor(t *testing.T) {
	g := NewGroup("Actions")
	if g.Color != "gray" {
		t.Errorf("expected default Color='gray', got '%s'", g.Color)
	}
}

func TestNewGroup_Items_empty(t *testing.T) {
	g := NewGroup("Actions")
	if len(g.Items()) != 0 {
		t.Errorf("expected 0 items by default, got %d", len(g.Items()))
	}
}

// ---------------------------------------------------------------------------
// Add tests
// ---------------------------------------------------------------------------

func TestGroup_Add_single(t *testing.T) {
	a := New("edit").SetLabel("Edit")
	g := NewGroup("Actions").Add(a)

	if len(g.Items()) != 1 {
		t.Fatalf("expected 1 item after Add(), got %d", len(g.Items()))
	}
	if g.Items()[0].Name != "edit" {
		t.Errorf("expected first item Name='edit', got '%s'", g.Items()[0].Name)
	}
}

func TestGroup_Add_multiple(t *testing.T) {
	a1 := New("edit").SetLabel("Edit")
	a2 := New("delete").SetLabel("Delete")
	a3 := New("view").SetLabel("View")

	g := NewGroup("Actions").Add(a1, a2, a3)

	if len(g.Items()) != 3 {
		t.Errorf("expected 3 items after Add(), got %d", len(g.Items()))
	}
}

func TestGroup_Add_chained(t *testing.T) {
	a1 := New("edit")
	a2 := New("delete")

	g := NewGroup("Actions").Add(a1).Add(a2)

	if len(g.Items()) != 2 {
		t.Errorf("expected 2 items after chained Add() calls, got %d", len(g.Items()))
	}
}

func TestGroup_Items_preserves_order(t *testing.T) {
	names := []string{"edit", "view", "delete", "export"}
	acts := make([]*Action, len(names))
	for i, n := range names {
		acts[i] = New(n)
	}

	g := NewGroup("Actions").Add(acts...)
	items := g.Items()

	for i, name := range names {
		if items[i].Name != name {
			t.Errorf("expected items[%d].Name='%s', got '%s'", i, name, items[i].Name)
		}
	}
}

// ---------------------------------------------------------------------------
// SetIcon and SetColor tests
// ---------------------------------------------------------------------------

func TestGroup_SetIcon(t *testing.T) {
	g := NewGroup("Actions").SetIcon("settings")
	if g.Icon != "settings" {
		t.Errorf("expected Icon='settings', got '%s'", g.Icon)
	}
}

func TestGroup_SetColor(t *testing.T) {
	g := NewGroup("Actions").SetColor("danger")
	if g.Color != "danger" {
		t.Errorf("expected Color='danger', got '%s'", g.Color)
	}
}

func TestGroup_SetIcon_SetColor_fluent(t *testing.T) {
	g := NewGroup("Options").SetIcon("more_horiz").SetColor("primary")

	if g.Icon != "more_horiz" {
		t.Errorf("expected Icon='more_horiz', got '%s'", g.Icon)
	}
	if g.Color != "primary" {
		t.Errorf("expected Color='primary', got '%s'", g.Color)
	}
}

// ---------------------------------------------------------------------------
// IsAuthorized tests
// ---------------------------------------------------------------------------

func TestGroup_IsAuthorized_no_func(t *testing.T) {
	g := NewGroup("Actions")
	ctx := context.Background()

	// No AuthorizeFunc set → always authorized
	if !g.IsAuthorized(ctx, nil) {
		t.Error("expected IsAuthorized()=true when no AuthorizeFunc is set")
	}
}

func TestGroup_IsAuthorized_func_returns_true(t *testing.T) {
	g := NewGroup("Actions").Authorize(func(ctx context.Context, item any) bool {
		return true
	})
	ctx := context.Background()

	if !g.IsAuthorized(ctx, nil) {
		t.Error("expected IsAuthorized()=true when AuthorizeFunc returns true")
	}
}

func TestGroup_IsAuthorized_func_returns_false(t *testing.T) {
	g := NewGroup("Actions").Authorize(func(ctx context.Context, item any) bool {
		return false
	})
	ctx := context.Background()

	if g.IsAuthorized(ctx, nil) {
		t.Error("expected IsAuthorized()=false when AuthorizeFunc returns false")
	}
}

func TestGroup_IsAuthorized_with_item(t *testing.T) {
	type User struct{ Role string }

	g := NewGroup("Admin Actions").Authorize(func(ctx context.Context, item any) bool {
		u, ok := item.(User)
		if !ok {
			return false
		}
		return u.Role == "admin"
	})
	ctx := context.Background()

	adminUser := User{Role: "admin"}
	regularUser := User{Role: "viewer"}

	if !g.IsAuthorized(ctx, adminUser) {
		t.Error("expected IsAuthorized()=true for admin user")
	}
	if g.IsAuthorized(ctx, regularUser) {
		t.Error("expected IsAuthorized()=false for regular user")
	}
}

// ---------------------------------------------------------------------------
// MoreActionsGroup tests
// ---------------------------------------------------------------------------

func TestMoreActionsGroup_label(t *testing.T) {
	g := MoreActionsGroup()
	if g.Label != "More" {
		t.Errorf("expected Label='More', got '%s'", g.Label)
	}
}

func TestMoreActionsGroup_icon(t *testing.T) {
	g := MoreActionsGroup()
	if g.Icon != "more_vert" {
		t.Errorf("expected Icon='more_vert', got '%s'", g.Icon)
	}
}

func TestMoreActionsGroup_empty_items(t *testing.T) {
	g := MoreActionsGroup()
	if len(g.Items()) != 0 {
		t.Errorf("expected 0 items when called with no args, got %d", len(g.Items()))
	}
}

func TestMoreActionsGroup_with_actions(t *testing.T) {
	restore := RestoreAction("/items")
	forceDelete := ForceDeleteAction("/items")

	g := MoreActionsGroup(restore, forceDelete)

	if len(g.Items()) != 2 {
		t.Fatalf("expected 2 items, got %d", len(g.Items()))
	}
	if g.Items()[0].Name != "restore" {
		t.Errorf("expected first item Name='restore', got '%s'", g.Items()[0].Name)
	}
	if g.Items()[1].Name != "force-delete" {
		t.Errorf("expected second item Name='force-delete', got '%s'", g.Items()[1].Name)
	}
}

func TestMoreActionsGroup_is_authorized_no_func(t *testing.T) {
	g := MoreActionsGroup()
	ctx := context.Background()

	if !g.IsAuthorized(ctx, nil) {
		t.Error("expected IsAuthorized()=true by default for MoreActionsGroup")
	}
}
