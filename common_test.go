package marctools

import (
	"testing"
)

var counttests = []struct {
	in  string
	out int64
}{
	{"./fixtures/authoritybibs.mrc", 9},
	{"./fixtures/deweybrowse.mrc", 1},
	{"./fixtures/heb.mrc", 1},
	{"./fixtures/journals.mrc", 10},
	{"./fixtures/testbug2.mrc", 1},
	{"./fixtures/weird_ids.mrc", 8},
}

func TestCount(t *testing.T) {
	for _, tt := range counttests {
		count := RecordCount(tt.in)
		if count != tt.out {
			t.Errorf("RecordCount(%s) => %d, want: %d", tt.in, count, tt.out)
		}
	}
}
