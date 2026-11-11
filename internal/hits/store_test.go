package hits

import (
	"testing"
	"time"
)

func TestPutAssignsStableIDs(t *testing.T) {
	store := New(time.Minute)
	a := store.Put([]Hit{{Title: "A", URL: "https://example.com/a", Snippet: "s"}})
	b := store.Put([]Hit{{Title: "A again", URL: "https://example.com/a", Snippet: "s2"}})
	if a[0].ID == "" || a[0].ID != b[0].ID {
		t.Fatalf("ids = %q %q", a[0].ID, b[0].ID)
	}
	got, ok := store.Get(a[0].ID)
	if !ok || got.URL != "https://example.com/a" {
		t.Fatalf("get = %+v ok=%v", got, ok)
	}
}
