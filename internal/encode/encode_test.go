package encode

import (
	"net/url"
	"testing"
	"time"
)

func TestQueryRejectsUnsupported(t *testing.T) {
	type bad struct {
		M map[string]string `query:"m"`
	}
	if err := Query(url.Values{}, &bad{M: map[string]string{"a": "b"}}); err == nil {
		t.Error("expected error for map query field")
	}
	if err := Query(url.Values{}, 42); err == nil {
		t.Error("expected error for non-struct")
	}
}

func TestQueryScalarsAndTime(t *testing.T) {
	type Embedded struct {
		Cursor string `query:"cursor"`
	}
	type q struct {
		Embedded
		N     int       `query:"n"`
		F     float64   `query:"f"`
		B     *bool     `query:"b"`
		T     time.Time `query:"t"`
		Skip  string    `query:"-"`
		Empty string    `query:"empty"`
	}
	values := url.Values{}
	if err := Query(values, q{Embedded: Embedded{"c"}, N: 3, F: 1.5, T: time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC), Skip: "x"}); err != nil {
		t.Fatal(err)
	}
	want := "cursor=c&f=1.5&n=3&t=2026-01-02T03%3A04%3A05Z"
	if got := values.Encode(); got != want {
		t.Errorf("query = %q, want %q", got, want)
	}
}
