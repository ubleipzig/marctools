package marctools

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"

	// sqlite3 bindings used by marcmap command
	_ "github.com/mattn/go-sqlite3"
	// use marc22 since it can read XML, needed by marcxmltojson
	"github.com/miku/marc22"
)

// AppVersion is displayed by all command line tools
const AppVersion = "1.6.4"

// JsonConversionOptions specify parameters for the MARC to JSON conversion
type JSONConversionOptions struct {
	FilterMap     map[string]bool   // which tags to include
	MetaMap       map[string]string // meta information
	IncludeLeader bool
	PlainMode     bool // only dump the content
	IgnoreErrors  bool
	RecordKey     string
}

// Batchworker batches work of MARC records to JSON
func BatchWorker(in chan []*marc22.Record, out chan []byte, wg *sync.WaitGroup, options JSONConversionOptions) {
	defer wg.Done()
	for records := range in {
		for _, record := range records {
			recordMap := RecordMap(record, options.FilterMap, options.IncludeLeader)
			if options.PlainMode {
				b, err := json.Marshal(recordMap)
				if err != nil {
					if !options.IgnoreErrors {
						log.Fatal(err)
					}
					log.Println(err)
					continue
				}
				out <- b
			} else {
				m := map[string]interface{}{
					options.RecordKey: recordMap,
					"meta":            options.MetaMap,
				}
				b, err := json.Marshal(m)
				if err != nil {
					if !options.IgnoreErrors {
						log.Fatal(err)
					}
					log.Println(err)
					continue
				}
				out <- b
			}
		}
	}
}

// Worker takes a Work item and sends the result (serialized json) on the out channel
func Worker(in chan *marc22.Record, out chan []byte, wg *sync.WaitGroup, options JSONConversionOptions) {
	defer wg.Done()
	for record := range in {
		recordMap := RecordMap(record, options.FilterMap, options.IncludeLeader)
		if options.PlainMode {
			b, err := json.Marshal(recordMap)
			if err != nil {
				if !options.IgnoreErrors {
					log.Fatal(err)
				}
				log.Println(err)
				continue
			}
			out <- b
		} else {
			m := map[string]interface{}{
				options.RecordKey: recordMap,
				"meta":            options.MetaMap,
			}
			b, err := json.Marshal(m)
			if err != nil {
				if !options.IgnoreErrors {
					log.Fatal(err)
				}
				log.Println(err)
				continue
			}
			out <- b
		}
	}
}

// FanInWriter writes the channel content to the writer
func FanInWriter(writer io.Writer, in chan []byte, done chan bool) {
	for b := range in {
		writer.Write(b)
		writer.Write([]byte("\n"))
	}
	done <- true
}

// KeyValueStringToMap turns a string like "key1=value1, key2=value2" into a map.
func KeyValueStringToMap(s string) (map[string]string, error) {
	result := make(map[string]string)
	if len(s) > 0 {
		for _, pair := range strings.Split(s, ",") {
			kv := strings.Split(pair, "=")
			if len(kv) != 2 {
				return nil, fmt.Errorf("could not parse key-value parameter: %s", s)
			}
			k := strings.TrimSpace(kv[0])
			v := strings.TrimSpace(kv[1])
			if len(k) == 0 || len(v) == 0 {
				return nil, fmt.Errorf("empty key or values not allowed: %s", s)
			}
			result[k] = v
		}
	}
	return result, nil
}

// StringToMapSet takes a string of the form "val1,val2, val3" and turns it
// into a poor mans set, a map[string]bool that is.
func StringToMapSet(s string) map[string]bool {
	result := make(map[string]bool)
	if len(s) > 0 {
		tags := strings.Split(s, ",")
		for _, value := range tags {
			result[strings.TrimSpace(value)] = true
		}
	}
	return result
}

// RecordLength returns the length of the marc record as stored in the leader
func RecordLength(reader io.Reader) (length int64, err error) {
	data := make([]byte, 24)
	n, err := reader.Read(data)
	if err != nil {
		return 0, err
	}
	if n != 24 {
		return 0, fmt.Errorf("marc: invalid leader: expected 24 bytes, read %d", n)
	}
	l, err := strconv.Atoi(string(data[0:5]))
	if err != nil {
		return 0, fmt.Errorf("marc: invalid record length: %s", err)
	}
	return int64(l), nil
}

// RecordCount count the number of records in marc file
func RecordCount(filename string) int64 {
	handle, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		if err := handle.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	var i, offset int64

	for {
		length, err := RecordLength(handle)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		i++
		offset += length
		handle.Seek(offset, 0)
	}
	return i
}

// IdentifierList returns a slice of strings, containing all ids of the given
// marc file. Set safe to true to use the slower, more safe method of parsing
// each record. Fast method breaks when there are multiple 001 fields (invalid,
// but real-world).
func IdentifierList(filename string, safe bool) []string {
	var fallback bool
	var yaz, awk string
	var err error

	if !safe {
		yaz, err = exec.LookPath("yaz-marcdump")
		if err != nil {
			fallback = true
		}
		awk, err = exec.LookPath("awk")
		if err != nil {
			fallback = true
		}
	}

	var ids []string

	if fallback || safe {
		// use slower iteration over records
		fi, err := os.Open(filename)
		if err != nil {
			log.Fatal(err)
		}
		defer func() {
			if err := fi.Close(); err != nil {
				log.Fatal(err)
			}
		}()

		for {
			record, err := marc22.ReadRecord(fi)
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatal(err)
			}

			fields := record.GetControlFields("001")
			// Cf. https://github.com/ubleipzig/marctools/issues/5
			// In case of multiple identifiers, choose the first.
			if len(fields) == 0 {
				log.Fatalf("missing 001 field")
			}
			ids = append(ids, strings.TrimSpace(fields[0].Data))
		}
	} else {
		// fast version using yaz and awk
		command := fmt.Sprintf("%s '%s' | %s ' /^001 / {print $2}'", yaz, filename, awk)
		out, err := exec.Command("bash", "-c", command).Output()
		if err != nil {
			log.Fatal(err)
		}

		for _, line := range strings.Split(string(out), "\n") {
			line = strings.TrimSpace(line)
			if len(line) == 0 {
				continue
			}
			ids = append(ids, strings.TrimSpace(line))
		}
	}

	return ids
}

// MapEntry contains location information of a single record in a MARC file
type MapEntry struct {
	ID     string
	Offset int64
	Length int64
}

// MarcMapEntries returns a chan of MapEntry structs.
func MarcMapEntries(infile string, safe bool) chan MapEntry {
	c := make(chan MapEntry)
	go func() {
		handle, err := os.Open(infile)
		if err != nil {
			log.Fatal(err)
		}
		defer func() {
			if err := handle.Close(); err != nil {
				log.Fatal(err)
			}
		}()

		ids := IdentifierList(infile, safe)
		var i, offset int64

		for {
			length, err := RecordLength(handle)
			if err == io.EOF {
				break
			}
			c <- MapEntry{ID: ids[i], Offset: offset, Length: length}
			offset += length
			i++
			handle.Seek(offset, 0)
		}
		close(c)
	}()
	return c
}

// MarcMap writes (id, offset, length) TSV of a given MARC file to a io.Writer
func MarcMap(infile string, writer io.Writer, safe bool) {
	for e := range MarcMapEntries(infile, safe) {
		writer.Write([]byte(fmt.Sprintf("%s\t%d\t%d\n", e.ID, e.Offset, e.Length)))
	}
}

// MarcMapSqlite writes (id, offset, length) sqlite3 database of a given MARC file to given output file
func MarcMapSqlite(infile, outfile string, safe bool) {
	db, err := sql.Open("sqlite3", outfile)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	init := `CREATE TABLE IF NOT EXISTS seekmap (id text, offset int, length int)`
	_, err = db.Exec(init)
	if err != nil {
		log.Fatalf("%q: %s\n", err, init)
	}

	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	stmt, err := tx.Prepare("INSERT INTO seekmap VALUES (?, ?, ?)")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	for e := range MarcMapEntries(infile, safe) {
		_, err = stmt.Exec(e.ID, e.Offset, e.Length)
		if err != nil {
			log.Fatal(err)
		}
	}

	_, err = tx.Exec("CREATE INDEX idx_seekmap_id ON seekmap (id)")
	if err != nil {
		log.Fatal(err)
	}
	tx.Commit()
}

// writeSplit writes bytes beginning at offset from file into output
// number of bytes copied is given by the buffer length
func writeSplit(file, output *os.File, offset int64, buffer []byte) {
	file.Seek(offset, 0)
	file.Read(buffer)
	output.Write(buffer)
	err := output.Close()
	if err != nil {
		log.Fatal(err)
	}
}

// createSplitFile returns a writeable file object
func createSplitFile(directory, prefix string, fileno int64) *os.File {
	filename := filepath.Join(directory, fmt.Sprintf("%s%08d", prefix, fileno))
	output, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
	}
	return output
}

// MarcSplitDirectoryPrefix splits a file into parts, each containing at most size records
// and writes the to specified directory, using a specific prefix
func MarcSplitDirectoryPrefix(infile string, size int64, directory, prefix string) {
	file, err := os.Open(infile)
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		if err := file.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	var i, length, cumulative, offset, batch, fileno int64

	for {
		length, err = RecordLength(file)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		if i%size == 0 && i > 0 {
			output := createSplitFile(directory, prefix, fileno)
			buffer := make([]byte, batch)
			writeSplit(file, output, offset, buffer)

			batch = 0
			fileno++
			offset = cumulative
		}
		cumulative += length
		batch += length
		file.Seek(int64(cumulative), 0)
		i++
	}

	output := createSplitFile(directory, prefix, fileno)
	buffer := make([]byte, batch)
	writeSplit(file, output, offset, buffer)
}

// MarcSplitDirectory splits a file into parts, each containing at most size records
// and writes the to specified directory
func MarcSplitDirectory(infile string, size int64, directory string) {
	MarcSplitDirectoryPrefix(infile, size, directory, "split-")
}

// MarcSplit splits a file into parts, each containing at most size records
func MarcSplit(infile string, size int64) {
	MarcSplitDirectoryPrefix(infile, size, ".", "split-")
}

// recordMap converts a record into a map, optionally only the tags given in filter
func recordMap(record *marc22.Record, filter map[string]bool) map[string]interface{} {
	m := make(map[string]interface{})
	hasFilter := len(filter) > 0

	for _, field := range record.ControlFields {
		tag := field.GetTag()
		if hasFilter {
			_, present := filter[tag]
			if !present {
				continue
			}
		}
		m[tag] = field.Data
	}

	for _, field := range record.DataFields {
		tag := field.GetTag()
		if hasFilter {
			_, present := filter[tag]
			if !present {
				continue
			}
		}

		smap := map[string]interface{}{
			"ind1": field.Ind1,
			"ind2": field.Ind2,
		}
		for _, subfield := range field.SubFields {
			code := fmt.Sprintf("%s", subfield.Code)
			_, present := smap[code]
			if present {
				smap[code] = append(smap[code].([]string), subfield.Value)
			} else {
				smap[code] = []string{subfield.Value}
			}
		}
		_, present := m[tag]
		if !present {
			subfields := make([]interface{}, 0)
			m[tag] = subfields
		}
		m[tag] = append(m[tag].([]interface{}), smap)
	}
	return m
}

// RecordMap converts a record to a map, optionally keeping only the tags
// given in filter. If includeLeader is true, the leader is converted as well.
func RecordMap(record *marc22.Record, filter map[string]bool, includeLeader bool) map[string]interface{} {
	rmap := recordMap(record, filter)
	if includeLeader {
		leader := record.LeaderParsed
		l := map[string]string{
			"status":  string(leader.Status),
			"cs":      string(leader.CharacterEncoding),
			"length":  fmt.Sprintf("%d", leader.Length),
			"type":    string(leader.Type),
			"impldef": string(leader.ImplementationDefined[:5]),
			"ic":      fmt.Sprintf("%d", leader.IndicatorCount),
			"lol":     fmt.Sprintf("%d", leader.LengthOfLength),
			"losp":    fmt.Sprintf("%d", leader.LengthOfStartPos),
			"sfcl":    fmt.Sprintf("%d", leader.SubfieldCodeLength),
			"ba":      fmt.Sprintf("%d", leader.BaseAddress),
			"raw":     string(leader.Bytes()),
		}
		rmap["leader"] = l
	}
	return rmap
}

var regexSubfield = regexp.MustCompile(`^([\d]{3})\.([a-z0-9])$`)
var regexControlfield = regexp.MustCompile(`^[\d]{3}$`)

// RecordToSlice returns a string slice with the values of the given tags
func RecordToSlice(record *marc22.Record,
	tags []string,
	fillna, separator string,
	skipIncompleteLines bool) []string {

	var cols []string

	for _, tag := range tags {
		if regexControlfield.MatchString(tag) {
			fields := record.GetControlFields(tag)
			if len(fields) > 0 {
				cols = append(cols, fields[0].Data)
			} else {
				if skipIncompleteLines {
					return []string{}
				}
				cols = append(cols, fillna)
			}
		} else if regexSubfield.MatchString(tag) {
			parts := strings.Split(tag, ".")
			code := parts[1]
			subfields := record.GetSubFields(parts[0], code)
			if len(subfields) > 0 {
				if separator == "" {
					cols = append(cols, subfields[0].Value)
				} else {
					var values []string
					for _, subfield := range subfields {
						values = append(values, subfield.Value)
					}
					cols = append(cols, strings.Join(values, separator))
				}
			} else {
				if skipIncompleteLines {
					return []string{}
				}
				cols = append(cols, fillna)
			}
		} else if strings.HasPrefix(tag, "@") {
			leader := record.LeaderParsed
			switch tag {
			case "@Length":
				cols = append(cols, fmt.Sprintf("%d", leader.Length))
			case "@Status":
				cols = append(cols, string(leader.Status))
			case "@Type":
				cols = append(cols, string(leader.Type))
			case "@ImplementationDefined":
				cols = append(cols, string(leader.ImplementationDefined[:5]))
			case "@CharacterEncoding":
				cols = append(cols, string(leader.CharacterEncoding))
			case "@BaseAddress":
				cols = append(cols, fmt.Sprintf("%d", leader.BaseAddress))
			case "@IndicatorCount":
				cols = append(cols, fmt.Sprintf("%d", leader.IndicatorCount))
			case "@SubfieldCodeLength":
				cols = append(cols, fmt.Sprintf("%d", leader.SubfieldCodeLength))
			case "@LengthOfLength":
				cols = append(cols, fmt.Sprintf("%d", leader.LengthOfLength))
			case "@LengthOfStartPos":
				cols = append(cols, fmt.Sprintf("%d", leader.LengthOfStartPos))
			default:
				log.Fatalf("unknown tag: %s\n", tag)
			}
		} else if !strings.HasPrefix(tag, "-") {
			cols = append(cols, strings.TrimSpace(tag))
		}
	}
	return cols
}

// RecordToTSV turns a single record into a single TSV line
func RecordToTSV(record *marc22.Record,
	tags []string,
	fillna, separator string,
	skipIncompleteLines bool) string {

	cols := RecordToSlice(record, tags, fillna, separator, skipIncompleteLines)
	var result string
	if len(cols) > 0 {
		result = fmt.Sprintf("%s\n", strings.Join(cols, "\t"))
	}
	return result
}
