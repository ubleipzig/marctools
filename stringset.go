package marctools

// poor mans string set
type StringSet struct {
	set map[string]struct{}
}

func NewStringSet() *StringSet {
	return &StringSet{set: make(map[string]struct{})}
}

func (set *StringSet) Add(s string) bool {
	_, found := set.set[s]
	set.set[s] = struct{}{}
	return !found // False if it existed already
}

func (set *StringSet) Contains(s string) bool {
	_, found := set.set[s]
	return found
}

func (set *StringSet) Size() int {
	return len(set.set)
}
