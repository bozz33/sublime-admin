package form

import "context"

// formErrKey is the unexported context key for FormErrors.
type formErrKey struct{}

// FormErrors maps field names to error messages.
// The special key "_error" holds a general (non-field-specific) message.
//
// Example — return from a resource's Create/Update:
//
//	return form.FormErrors{
//	    "email":    "Email already taken.",
//	    "password": "Must be at least 8 characters.",
//	}
type FormErrors map[string]string

// Error implements the error interface so FormErrors can be returned directly.
func (fe FormErrors) Error() string {
	if msg, ok := fe["_error"]; ok {
		return msg
	}
	return "validation failed"
}

// FieldErrors implements ValidationErrors.
func (fe FormErrors) FieldErrors() map[string]string {
	return map[string]string(fe)
}

// ValidationErrors is an optional interface for errors that carry
// per-field validation messages. When CRUDHandler.Store or Update
// receives an error implementing this interface, it re-renders the
// form with inline field errors instead of returning HTTP 500.
type ValidationErrors interface {
	error
	FieldErrors() map[string]string
}

// WithFormErrors returns a context carrying the given FormErrors.
// Called automatically by CRUDHandler — no need to call manually.
func WithFormErrors(ctx context.Context, errs FormErrors) context.Context {
	return context.WithValue(ctx, formErrKey{}, errs)
}

// GetFormErrors retrieves FormErrors from context.
// Returns nil if no errors are in context (i.e. fresh form).
func GetFormErrors(ctx context.Context) FormErrors {
	if e, ok := ctx.Value(formErrKey{}).(FormErrors); ok {
		return e
	}
	return nil
}
