package cache

import "testing"

func TestUserPageKey(t *testing.T) {
	if got, want := UserPageKey("budi123"), "page:budi123"; got != want {
		t.Fatalf("expected cache key %q, got %q", want, got)
	}
}
