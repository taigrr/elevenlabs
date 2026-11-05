package stream

import (
	"context"
	"errors"
	"testing"
)

func TestPaginateStopsEarlyAndOnError(t *testing.T) {
	calls := 0
	fetch := func(_ context.Context, cursor string) (Page[int], error) {
		calls++
		switch cursor {
		case "":
			return Page[int]{Items: []int{1, 2}, Next: "c1", HasMore: true}, nil
		case "c1":
			return Page[int]{Items: []int{3}, Next: "c2", HasMore: true}, nil
		default:
			return Page[int]{}, errors.New("boom")
		}
	}
	var got []int
	var gotErr error
	for v, err := range Paginate(context.Background(), "", fetch) {
		if err != nil {
			gotErr = err
			break
		}
		got = append(got, v)
	}
	if len(got) != 3 || gotErr == nil || calls != 3 {
		t.Errorf("got=%v err=%v calls=%d", got, gotErr, calls)
	}

	calls = 0
	for v := range Paginate(context.Background(), "", fetch) {
		if v == 2 {
			break
		}
	}
	if calls != 1 {
		t.Errorf("early break fetched %d pages", calls)
	}

	same := func(context.Context, string) (Page[int], error) {
		calls++
		return Page[int]{Items: []int{1}, Next: "same", HasMore: true}, nil
	}
	calls = 0
	for range Paginate(context.Background(), "same", same) {
	}
	if calls != 1 {
		t.Errorf("repeated cursor should stop, fetched %d", calls)
	}
}
