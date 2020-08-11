package sets

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestStringSet(t *testing.T) {
	s := NewStringSet()

	s.Add("test2")
	s.Add("test3")
	s.Add("test1")
	s.Add("test1")

	want := []string{"test1", "test2", "test3"}

	if diff := cmp.Diff(want, s.Elements()); diff != "" {
		t.Fatalf("StringSet failed:\n%s", diff)
	}
}
