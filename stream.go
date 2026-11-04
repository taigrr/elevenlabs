package elevenlabs

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"iter"
)

// Stream decodes a newline-delimited JSON response one message at a time.
// It must be closed when no longer needed.
//
//	for stream.Next() {
//		chunk := stream.Current()
//	}
//	if err := stream.Err(); err != nil { ... }
type Stream[T any] struct {
	body    io.ReadCloser
	reader  *bufio.Reader
	current T
	err     error
	done    bool
}

func newStream[T any](body io.ReadCloser) *Stream[T] {
	return &Stream[T]{body: body, reader: bufio.NewReader(body)}
}

// Next advances to the next message, returning false at the end of the
// stream or on error.
func (s *Stream[T]) Next() bool {
	for !s.done {
		line, err := s.reader.ReadBytes('\n')
		if errors.Is(err, io.EOF) {
			s.done = true
		} else if err != nil {
			s.done, s.err = true, fmt.Errorf("elevenlabs: read stream: %w", err)
			return false
		}
		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		var msg T
		if err := json.Unmarshal(line, &msg); err != nil {
			s.done, s.err = true, fmt.Errorf("elevenlabs: decode stream message: %w", err)
			return false
		}
		s.current = msg
		return true
	}
	return false
}

// Current returns the message read by the last successful call to Next.
func (s *Stream[T]) Current() T { return s.current }

// Err returns the first error encountered while reading the stream.
func (s *Stream[T]) Err() error { return s.err }

// Close releases the underlying connection.
func (s *Stream[T]) Close() error { return s.body.Close() }

// All returns an iterator over the remaining messages. The stream is closed
// when iteration finishes or stops early.
func (s *Stream[T]) All() iter.Seq2[T, error] {
	return func(yield func(T, error) bool) {
		defer s.Close()
		for s.Next() {
			if !yield(s.current, nil) {
				return
			}
		}
		if s.err != nil {
			var zero T
			yield(zero, s.err)
		}
	}
}

// page is one page of a cursor-paginated listing.
type page[T any] struct {
	items   []T
	next    string
	hasMore bool
}

// paginate iterates over every item of a cursor-paginated endpoint, calling
// fetch with the cursor of each successive page (empty for the first page
// unless start is set). Iteration stops at the first error, which is yielded.
func paginate[T any](ctx context.Context, start string, fetch func(ctx context.Context, cursor string) (page[T], error)) iter.Seq2[T, error] {
	return func(yield func(T, error) bool) {
		cursor := start
		for {
			p, err := fetch(ctx, cursor)
			if err != nil {
				var zero T
				yield(zero, err)
				return
			}
			for _, item := range p.items {
				if !yield(item, nil) {
					return
				}
			}
			if !p.hasMore || p.next == "" || p.next == cursor {
				return
			}
			cursor = p.next
		}
	}
}
