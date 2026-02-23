package table

import (
	"reflect"
)

// Eval[T] represents a value that can be either static or dynamically computed
// from the cell value and the full record. This is the Go-idiomatic equivalent
// of Filament's EvaluatesClosures pattern (->color(fn($state) => ...)).
//
// Usage:
//
//	// Static value
//	col.WithColorEval(Static[string]("green"))
//
//	// Dynamic — depends on the cell value
//	col.WithColorEval(Dynamic[string](func(value string, _ any) string {
//	    if value == "active" { return "green" }
//	    return "red"
//	}))
//
//	// Dynamic — depends on the full record
//	col.WithColorEval(Dynamic[string](func(_ string, record any) string {
//	    if u, ok := record.(*User); ok && u.IsAdmin { return "purple" }
//	    return "gray"
//	}))
type Eval[T any] struct {
	static  T
	dynamic func(value string, record any) T
	isDyn   bool
}

// Static returns an Eval that always resolves to the given value.
func Static[T any](v T) Eval[T] {
	return Eval[T]{static: v}
}

// Dynamic returns an Eval that calls fn(value, record) at render time.
func Dynamic[T any](fn func(value string, record any) T) Eval[T] {
	return Eval[T]{dynamic: fn, isDyn: true}
}

// Resolve returns the evaluated value given the cell value and the full record.
func (e Eval[T]) Resolve(value string, record any) T {
	if e.isDyn && e.dynamic != nil {
		return e.dynamic(value, record)
	}
	return e.static
}

// IsSet returns true if the Eval has been explicitly set (static or dynamic).
func (e Eval[T]) IsSet() bool {
	if e.isDyn {
		return true
	}
	return !reflect.DeepEqual(e.static, reflect.Zero(reflect.TypeOf(&e.static).Elem()).Interface())
}
