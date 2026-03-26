package usecase

import (
	"errors"
	"testing"
	"time"
)

func TestClampPageSize(t *testing.T) {
	if clampPageSize(0) != 20 {
		t.Fatalf("default")
	}
	if clampPageSize(100) != 50 {
		t.Fatalf("max")
	}
	if clampPageSize(15) != 15 {
		t.Fatalf("mid")
	}
}

func TestDedupeStrings(t *testing.T) {
	got := dedupeStrings([]string{"a", "", "a", "b"})
	if len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Fatalf("got %v", got)
	}
}

func TestDecodeNextCursor(t *testing.T) {
	u := "11111111-1111-1111-1111-111111111111"
	raw := "2020-05-01T12:00:00Z|" + u
	ct, cid, err := decodeNextCursor(&raw)
	if err != nil {
		t.Fatal(err)
	}
	if cid == nil || *cid != u {
		t.Fatalf("id %v", cid)
	}
	if ct == nil || !ct.Equal(time.Date(2020, 5, 1, 12, 0, 0, 0, time.UTC)) {
		t.Fatalf("time %v", ct)
	}
	bad := "nope"
	_, _, err = decodeNextCursor(&bad)
	if !errors.Is(err, ErrInvalidArgument) {
		t.Fatalf("want invalid, got %v", err)
	}
}

func TestEncodeDecodeNextCursor_roundTrip(t *testing.T) {
	t0 := time.Date(2020, 5, 1, 12, 0, 0, 0, time.UTC)
	id := "11111111-1111-1111-1111-111111111111"
	s := encodeNextCursor(t0, id)
	ct, cid, err := decodeNextCursor(s)
	if err != nil {
		t.Fatal(err)
	}
	if *cid != id {
		t.Fatalf("id")
	}
	if !ct.UTC().Equal(t0) {
		t.Fatalf("time got %v want %v", ct.UTC(), t0)
	}
}
