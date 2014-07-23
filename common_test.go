package marctools

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"github.com/miku/marc21"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
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

func TestMarcSplitDirectoryPrefix(t *testing.T) {
	dir, err := ioutil.TempDir("", "marctools-TestMarcSplitDirectoryPrefix-")
	if err != nil {
		t.Errorf("%s\n", err)
	}

	prefix := "prefix-"
	MarcSplitDirectoryPrefix("./fixtures/authoritybibs.mrc", 3, dir, prefix)

	expectedFiles := []string{filepath.Join(dir, fmt.Sprintf("%s%08d", prefix, 0)),
		filepath.Join(dir, fmt.Sprintf("%s%08d", prefix, 1)),
		filepath.Join(dir, fmt.Sprintf("%s%08d", prefix, 2))}

	for _, filename := range expectedFiles {
		if _, err := os.Stat(filename); os.IsNotExist(err) {
			t.Errorf("%s expected: %s\n", filename, err)
		}
		recordCount := RecordCount(filename)
		if recordCount != 3 {
			t.Errorf("unexpected record count in %s: %d, want: %d", filename, recordCount, 3)
		}
	}
}

var kvtests = []struct {
	in  string
	out map[string]string
}{
	{"key=value", map[string]string{"key": "value"}},
	{"k1=v1, k2=v2", map[string]string{"k1": "v1", "k2": "v2"}},
	{"k1 = v1,k2=   v2", map[string]string{"k1": "v1", "k2": "v2"}},
	{"k1=v1, k1=v2", map[string]string{"k1": "v2"}},
}

func TestKeyValueStringToMap(t *testing.T) {
	for _, tt := range kvtests {
		out, err := KeyValueStringToMap(tt.in)
		if err != nil {
			t.Errorf("KeyValueStringToMap(%s) err'd: %s", tt.in, err)
		}
		eq := reflect.DeepEqual(out, tt.out)
		if !eq {
			t.Errorf("KeyValueStringToMap(%s) => %+v, want: %+v", tt.in, out, tt.out)
		}
	}

	failures := []string{"key=",
		"keyvalue",
		"key=value=value",
		"key=value=value,=",
		"k1=v1,,, k1=v2"}
	for _, value := range failures {
		out, err := KeyValueStringToMap(value)
		if err == nil {
			t.Errorf("KeyValueStringToMap(%s) => %s, want: err!", value, out)
		}
	}
}

var stosettests = []struct {
	in  string
	out map[string]bool
}{
	{"key", map[string]bool{"key": true}},
	{"k1,k2", map[string]bool{"k1": true, "k2": true}},
	{"k1,  k2", map[string]bool{"k1": true, "k2": true}},
}

func TestStringToMapSet(t *testing.T) {
	for _, tt := range stosettests {
		out := StringToMapSet(tt.in)
		eq := reflect.DeepEqual(out, tt.out)
		if !eq {
			t.Errorf("StringToMapSet(%s) => %+v, want: %+v", tt.in, out, tt.out)
		}
	}
}

var recordmaptests = []struct {
	record        string
	filterMap     map[string]bool
	includeLeader bool
	out           string
}{
	// test with includeLeader=false
	{`00613cam a2200229Ma 4500001001600000005001700016008004100033020001500074035002300089040002500112041001800137043001200155050002400167049000900191082001600200082001600216100003000232245002200262250002300284260004700307300002900354testdeweybrowse20110419140028.0110214s1992    it a     b    001 0 ita d  a8820737493  a(OCoLC)ocm30585539  aRBNcRBNdOCLCGdPVU1 aitaalathlat  ae-it---14aDG848.15b.V53 1992  aPVUM  a123.45 .I39  a123.46 .Q391 aPerson, Fake,d1668-1744.10aDewey browse test  aFictional edition.  aMorano :bCentro di Studi Vichiani,c1992.  a296 p. :bill. ;c24 cm.`,
		map[string]bool{},
		false,
		`{"001":"testdeweybrowse","005":"20110419140028.0","008":"110214s1992    it a     b    001 0 ita d","020":[{"a":["8820737493"],"ind1":" ","ind2":" "}],"035":[{"a":["(OCoLC)ocm30585539"],"ind1":" ","ind2":" "}],"040":[{"a":["RBN"],"c":["RBN"],"d":["OCLCG","PVU"],"ind1":" ","ind2":" "}],"041":[{"a":["ita","lat"],"h":["lat"],"ind1":"1","ind2":" "}],"043":[{"a":["e-it---"],"ind1":" ","ind2":" "}],"049":[{"a":["PVUM"],"ind1":" ","ind2":" "}],"050":[{"a":["DG848.15"],"b":[".V53 1992"],"ind1":"1","ind2":"4"}],"082":[{"a":["123.45 .I39"],"ind1":" ","ind2":" "},{"a":["123.46 .Q39"],"ind1":" ","ind2":" "}],"100":[{"a":["Person, Fake,"],"d":["1668-1744."],"ind1":"1","ind2":" "}],"245":[{"a":["Dewey browse test"],"ind1":"1","ind2":"0"}],"250":[{"a":["Fictional edition."],"ind1":" ","ind2":" "}],"260":[{"a":["Morano :"],"b":["Centro di Studi Vichiani,"],"c":["1992."],"ind1":" ","ind2":" "}],"300":[{"a":["296 p. :"],"b":["ill. ;"],"c":["24 cm."],"ind1":" ","ind2":" "}]}`},
	// test with includeLeader=true
	{`00613cam a2200229Ma 4500001001600000005001700016008004100033020001500074035002300089040002500112041001800137043001200155050002400167049000900191082001600200082001600216100003000232245002200262250002300284260004700307300002900354testdeweybrowse20110419140028.0110214s1992    it a     b    001 0 ita d  a8820737493  a(OCoLC)ocm30585539  aRBNcRBNdOCLCGdPVU1 aitaalathlat  ae-it---14aDG848.15b.V53 1992  aPVUM  a123.45 .I39  a123.46 .Q391 aPerson, Fake,d1668-1744.10aDewey browse test  aFictional edition.  aMorano :bCentro di Studi Vichiani,c1992.  a296 p. :bill. ;c24 cm.`,
		map[string]bool{},
		true,
		`{"001":"testdeweybrowse","005":"20110419140028.0","008":"110214s1992    it a     b    001 0 ita d","020":[{"a":["8820737493"],"ind1":" ","ind2":" "}],"035":[{"a":["(OCoLC)ocm30585539"],"ind1":" ","ind2":" "}],"040":[{"a":["RBN"],"c":["RBN"],"d":["OCLCG","PVU"],"ind1":" ","ind2":" "}],"041":[{"a":["ita","lat"],"h":["lat"],"ind1":"1","ind2":" "}],"043":[{"a":["e-it---"],"ind1":" ","ind2":" "}],"049":[{"a":["PVUM"],"ind1":" ","ind2":" "}],"050":[{"a":["DG848.15"],"b":[".V53 1992"],"ind1":"1","ind2":"4"}],"082":[{"a":["123.45 .I39"],"ind1":" ","ind2":" "},{"a":["123.46 .Q39"],"ind1":" ","ind2":" "}],"100":[{"a":["Person, Fake,"],"d":["1668-1744."],"ind1":"1","ind2":" "}],"245":[{"a":["Dewey browse test"],"ind1":"1","ind2":"0"}],"250":[{"a":["Fictional edition."],"ind1":" ","ind2":" "}],"260":[{"a":["Morano :"],"b":["Centro di Studi Vichiani,"],"c":["1992."],"ind1":" ","ind2":" "}],"300":[{"a":["296 p. :"],"b":["ill. ;"],"c":["24 cm."],"ind1":" ","ind2":" "}],"leader":{"ba":"229","cs":"a","ic":"2","impldef":"m Ma ","length":"613","lol":"4","losp":"5","raw":"00613cam a2200229Ma 4500","sfcl":"2","status":"c","type":"a"}}`},
	// test with includeLeader=false and a simple filter
	{`00613cam a2200229Ma 4500001001600000005001700016008004100033020001500074035002300089040002500112041001800137043001200155050002400167049000900191082001600200082001600216100003000232245002200262250002300284260004700307300002900354testdeweybrowse20110419140028.0110214s1992    it a     b    001 0 ita d  a8820737493  a(OCoLC)ocm30585539  aRBNcRBNdOCLCGdPVU1 aitaalathlat  ae-it---14aDG848.15b.V53 1992  aPVUM  a123.45 .I39  a123.46 .Q391 aPerson, Fake,d1668-1744.10aDewey browse test  aFictional edition.  aMorano :bCentro di Studi Vichiani,c1992.  a296 p. :bill. ;c24 cm.`,
		map[string]bool{"001": true, "005": true},
		false,
		`{"001":"testdeweybrowse","005":"20110419140028.0"}`},
}

func TestRecordToMap(t *testing.T) {
	for _, tt := range recordmaptests {
		reader := strings.NewReader(tt.record)
		record, err := marc21.ReadRecord(reader)
		if err != nil {
			t.Error(err)
		}
		result := RecordToMap(record, &tt.filterMap, tt.includeLeader)
		if result == nil {
			t.Error("RecordToMap should not return nil")
		}
		b, err := json.Marshal(result)
		if err != nil {
			t.Error("RecordToMap should return something JSON-serializable")
		}
		if string(b) != tt.out {
			t.Errorf("RecordToMap(%s, %+v, %v) => %+v, want: %+v", tt.record, tt.filterMap, tt.includeLeader, string(b), tt.out)
		}
	}
}

var recordtotsvtests = []struct {
	record              string
	tags                []string
	fillNA              string
	separator           string
	skipIncompleteLines bool
	out                 string
}{
	{`00613cam a2200229Ma 4500001001600000005001700016008004100033020001500074035002300089040002500112041001800137043001200155050002400167049000900191082001600200082001600216100003000232245002200262250002300284260004700307300002900354testdeweybrowse20110419140028.0110214s1992    it a     b    001 0 ita d  a8820737493  a(OCoLC)ocm30585539  aRBNcRBNdOCLCGdPVU1 aitaalathlat  ae-it---14aDG848.15b.V53 1992  aPVUM  a123.45 .I39  a123.46 .Q391 aPerson, Fake,d1668-1744.10aDewey browse test  aFictional edition.  aMorano :bCentro di Studi Vichiani,c1992.  a296 p. :bill. ;c24 cm.`,
		[]string{"001"},
		"<NULL>",
		"",
		true,
		"testdeweybrowse\n",
	},
	{`00613cam a2200229Ma 4500001001600000005001700016008004100033020001500074035002300089040002500112041001800137043001200155050002400167049000900191082001600200082001600216100003000232245002200262250002300284260004700307300002900354testdeweybrowse20110419140028.0110214s1992    it a     b    001 0 ita d  a8820737493  a(OCoLC)ocm30585539  aRBNcRBNdOCLCGdPVU1 aitaalathlat  ae-it---14aDG848.15b.V53 1992  aPVUM  a123.45 .I39  a123.46 .Q391 aPerson, Fake,d1668-1744.10aDewey browse test  aFictional edition.  aMorano :bCentro di Studi Vichiani,c1992.  a296 p. :bill. ;c24 cm.`,
		[]string{"001", "020.a", "082.a"},
		"<NULL>",
		"",
		true,
		"testdeweybrowse\t8820737493\t123.45 .I39\n",
	},
	{`00613cam a2200229Ma 4500001001600000005001700016008004100033020001500074035002300089040002500112041001800137043001200155050002400167049000900191082001600200082001600216100003000232245002200262250002300284260004700307300002900354testdeweybrowse20110419140028.0110214s1992    it a     b    001 0 ita d  a8820737493  a(OCoLC)ocm30585539  aRBNcRBNdOCLCGdPVU1 aitaalathlat  ae-it---14aDG848.15b.V53 1992  aPVUM  a123.45 .I39  a123.46 .Q391 aPerson, Fake,d1668-1744.10aDewey browse test  aFictional edition.  aMorano :bCentro di Studi Vichiani,c1992.  a296 p. :bill. ;c24 cm.`,
		[]string{"001", "020.a", "082.a"},
		"<NULL>",
		"|",
		true,
		"testdeweybrowse\t8820737493\t123.45 .I39|123.46 .Q39\n",
	},
	{`00613cam a2200229Ma 4500001001600000005001700016008004100033020001500074035002300089040002500112041001800137043001200155050002400167049000900191082001600200082001600216100003000232245002200262250002300284260004700307300002900354testdeweybrowse20110419140028.0110214s1992    it a     b    001 0 ita d  a8820737493  a(OCoLC)ocm30585539  aRBNcRBNdOCLCGdPVU1 aitaalathlat  ae-it---14aDG848.15b.V53 1992  aPVUM  a123.45 .I39  a123.46 .Q391 aPerson, Fake,d1668-1744.10aDewey browse test  aFictional edition.  aMorano :bCentro di Studi Vichiani,c1992.  a296 p. :bill. ;c24 cm.`,
		[]string{"999"},
		"<NULL>",
		"|",
		false,
		"<NULL>\n",
	},
	{`00613cam a2200229Ma 4500001001600000005001700016008004100033020001500074035002300089040002500112041001800137043001200155050002400167049000900191082001600200082001600216100003000232245002200262250002300284260004700307300002900354testdeweybrowse20110419140028.0110214s1992    it a     b    001 0 ita d  a8820737493  a(OCoLC)ocm30585539  aRBNcRBNdOCLCGdPVU1 aitaalathlat  ae-it---14aDG848.15b.V53 1992  aPVUM  a123.45 .I39  a123.46 .Q391 aPerson, Fake,d1668-1744.10aDewey browse test  aFictional edition.  aMorano :bCentro di Studi Vichiani,c1992.  a296 p. :bill. ;c24 cm.`,
		[]string{"999"},
		"<NULL>",
		"|",
		true,
		"",
	},
	{`00613cam a2200229Ma 4500001001600000005001700016008004100033020001500074035002300089040002500112041001800137043001200155050002400167049000900191082001600200082001600216100003000232245002200262250002300284260004700307300002900354testdeweybrowse20110419140028.0110214s1992    it a     b    001 0 ita d  a8820737493  a(OCoLC)ocm30585539  aRBNcRBNdOCLCGdPVU1 aitaalathlat  ae-it---14aDG848.15b.V53 1992  aPVUM  a123.45 .I39  a123.46 .Q391 aPerson, Fake,d1668-1744.10aDewey browse test  aFictional edition.  aMorano :bCentro di Studi Vichiani,c1992.  a296 p. :bill. ;c24 cm.`,
		[]string{"001", "@Length"},
		"<NULL>",
		"",
		false,
		"testdeweybrowse\t613\n",
	},
}

func TestRecordToTSV(t *testing.T) {
	for _, tt := range recordtotsvtests {
		reader := strings.NewReader(tt.record)
		record, err := marc21.ReadRecord(reader)
		if err != nil {
			t.Error(err)
		}
		result := RecordToTSV(record, &tt.tags, &tt.fillNA, &tt.separator, &tt.skipIncompleteLines)
		if result == nil {
			t.Error("RecordToTSV should not return nil")
		}
		if *result != tt.out {
			t.Errorf("RecordToTSV(%s, %v, %s, %s, %t) => %+v, want: %+v", record, tt.tags, tt.fillNA, tt.separator, tt.skipIncompleteLines, *result, tt.out)
		}
	}
}
