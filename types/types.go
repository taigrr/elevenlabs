package types

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// Clearable is an optional update field that can either set a value or
// explicitly clear it. Leave a *Clearable field nil to keep the current
// value. It encodes as the value itself or as the string "clear".
type Clearable[T any] struct {
	value T
	clear bool
}

// SetTo returns a Clearable that sets the field to v.
func SetTo[T any](v T) *Clearable[T] { return &Clearable[T]{value: v} }

// ClearField returns a Clearable that removes the field's current value.
func ClearField[T any]() *Clearable[T] { return &Clearable[T]{clear: true} }

// Value returns the value to set and whether the field is being set rather than cleared.
func (c Clearable[T]) Value() (T, bool) { return c.value, !c.clear }

// IsClear reports whether the field is being cleared.
func (c Clearable[T]) IsClear() bool { return c.clear }

// MarshalJSON implements json.Marshaler.
func (c Clearable[T]) MarshalJSON() ([]byte, error) {
	if c.clear {
		return []byte(`"clear"`), nil
	}
	return json.Marshal(c.value)
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *Clearable[T]) UnmarshalJSON(b []byte) error {
	if bytes.Equal(bytes.TrimSpace(b), []byte(`"clear"`)) {
		*c = Clearable[T]{clear: true}
		return nil
	}
	*c = Clearable[T]{}
	return json.Unmarshal(b, &c.value)
}

// LimitOrUnlimited is a numeric limit that may be "unlimited".
type LimitOrUnlimited struct {
	// Limit is the numeric limit when Unlimited is false.
	Limit float64
	// Unlimited is true when there is no limit.
	Unlimited bool
}

// MarshalJSON implements json.Marshaler.
func (l LimitOrUnlimited) MarshalJSON() ([]byte, error) {
	if l.Unlimited {
		return []byte(`"unlimited"`), nil
	}
	return json.Marshal(l.Limit)
}

// UnmarshalJSON implements json.Unmarshaler.
func (l *LimitOrUnlimited) UnmarshalJSON(b []byte) error {
	var s string
	if json.Unmarshal(b, &s) == nil {
		if s != "unlimited" {
			return fmt.Errorf("elevenlabs: unexpected limit %q", s)
		}
		*l = LimitOrUnlimited{Unlimited: true}
		return nil
	}
	*l = LimitOrUnlimited{}
	return json.Unmarshal(b, &l.Limit)
}

// Ptr returns a pointer to v. It is a convenience for populating optional
// request fields, e.g. Seed: types.Ptr(42).
func Ptr[T any](v T) *T { return &v }
