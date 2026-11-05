package stream

import (
	"context"
	"iter"
)

// Page is one page of a cursor-paginated listing.
type Page[T any] struct {
	Items   []T
	Next    string
	HasMore bool
}

// Paginate iterates over every item of a cursor-paginated endpoint, calling
// fetch with the cursor of each successive page (empty for the first page
// unless start is set). Iteration stops at the first error, which is yielded.
func Paginate[T any](ctx context.Context, start string, fetch func(ctx context.Context, cursor string) (Page[T], error)) iter.Seq2[T, error] {
	return func(yield func(T, error) bool) {
		cursor := start
		for {
			p, err := fetch(ctx, cursor)
			if err != nil {
				var zero T
				yield(zero, err)
				return
			}
			for _, item := range p.Items {
				if !yield(item, nil) {
					return
				}
			}
			if !p.HasMore || p.Next == "" || p.Next == cursor {
				return
			}
			cursor = p.Next
		}
	}
}
