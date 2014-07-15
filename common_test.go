package marctools

import (
	"bytes"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"io/ioutil"
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

var idlisttests = []struct {
	in  string
	out []string
}{
	{"./fixtures/deweybrowse.mrc", []string{"testdeweybrowse"}},
	{"./fixtures/testbug2.mrc", []string{"testbug2"}},
	{"./fixtures/heb.mrc", []string{"testbug1"}},
	// {"./fixtures/weird_ids.mrc", []string{`|||pipe|||"quote`,
	// 	`dot.dash-underscore__3.colon:18`,
	// 	`dot.dash-underscore__3.space`,
	// 	`dollar$ign/slashcombo`,
	// 	`all'kinds"of'quotes"`,
	// 	`<angle>brackets&ampersands`,
	// 	`hashes#coming@ya`,
	// 	`wehave?sand%stospare`}},
	{"./fixtures/journals.mrc", []string{"testsample1",
		"testsample2",
		"testsample3",
		"testsample4",
		"testsample5",
		"testsample6",
		"testsample7",
		"testsample8",
		"testsample9",
		"testsample10"}},
}

func TestIdList(t *testing.T) {
	for _, tt := range idlisttests {
		ids := IdList(tt.in)
		if len(ids) != len(tt.out) {
			t.Errorf("IdList(%s) => %+v, want: %+v", tt.in, ids, tt.out)
		}
		for i := 0; i < len(ids); i++ {
			if ids[i] != tt.out[i] {
				t.Errorf("List element mismatch in IdList(%s)[%d] => %+v, want: %+v",
					tt.in, i, ids[i], tt.out[i])
			}
		}
	}
}

var marcmaptests = []struct {
	in  string
	out string
}{
	{"./fixtures/deweybrowse.mrc", "testdeweybrowse\t0\t613\n"},
	{"./fixtures/journals.mrc",
		"testsample1\t0\t1571\n" +
			"testsample2\t1571\t1195\n" +
			"testsample3\t2766\t1057\n" +
			"testsample4\t3823\t1361\n" +
			"testsample5\t5184\t1707\n" +
			"testsample6\t6891\t1532\n" +
			"testsample7\t8423\t1426\n" +
			"testsample8\t9849\t1251\n" +
			"testsample9\t11100\t2173\n" +
			"testsample10\t13273\t1195\n"},
}

func TestMarcMap(t *testing.T) {
	var b bytes.Buffer
	for _, tt := range marcmaptests {
		b.Reset()
		MarcMap(tt.in, &b)
		output := b.String()
		if output != tt.out {
			t.Errorf("MarcMap(%s) => %s, want: %s", tt.in, output, tt.out)
		}
	}
}

func TestMarcMapSqlite(t *testing.T) {
	file, err := ioutil.TempFile("", "marctools-TestMarcMapSqlite-")
	if err != nil {
		t.Errorf("%s\n", err)
	}

	MarcMapSqlite("./fixtures/journals.mrc", file.Name())

	db, err := sql.Open("sqlite3", file.Name())
	if err != nil {
		t.Errorf("%s\n", err)
	}
	defer db.Close()

	var offset, length string
	err = db.QueryRow("SELECT offset, length FROM seekmap WHERE id=?", "testsample3").Scan(&offset, &length)
	switch {
	case err == sql.ErrNoRows:
		t.Errorf("%s\n", err)
	case err != nil:
		t.Errorf("%s\n", err)
	default:
		if offset != "2766" || length != "1057" {
			t.Errorf("(%s, %s), want: (2766, 1057)\n", offset, length)
		}
	}
}
