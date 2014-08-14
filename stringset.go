package marctools

// StringSet is map disguised as set
type StringSet struct {
	set map[string]struct{}
}

// NewStringSet returns an empty set
func NewStringSet() *StringSet {
	return &StringSet{set: make(map[string]struct{})}
}

// Add adds a string to a set, returns true if added, false it it already existed (noop)
func (set *StringSet) Add(s string) bool {
	_, found := set.set[s]
	set.set[s] = struct{}{}
	return !found // False if it existed already
}

// Contains returns true if given string is in the set, false otherwise
func (set *StringSet) Contains(s string) bool {
	_, found := set.set[s]
	return found
}

// Size returns current number of elements in the set
func (set *StringSet) Size() int {
	return len(set.set)
}
