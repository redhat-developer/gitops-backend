package sets

import "sort"

// StringSet is a set of strings with simple functionality for getting a sorted
// set of elements.
type StringSet map[string]bool

// NewStringSet creates a new empty StringSet.
func NewStringSet() StringSet {
	return StringSet{}
}

// Add adds a string to the set.
func (s StringSet) Add(n string) {
	s[n] = true
}

// Elements returns a sorted list of elements in the set.
func (s StringSet) Elements() []string {
	e := []string{}
	for k := range s {
		e = append(e, k)
	}
	sort.Strings(e)
	return e
}
