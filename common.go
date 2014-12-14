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
const AppVersion = "1.5.5"

// Work represents the input for a conversion of a single record to JSON
type Work struct {
	Record        *marc22.Record     // MARC record
	FilterMap     *map[string]bool   // which tags to include
	MetaMap       *map[string]string // meta information
	IncludeLeader bool
	PlainMode     bool // only dump the content
	IgnoreErrors  bool
	RecordKey     string
}

// Worker takes a Work item and sends the result (serialized json) on the out channel
func Worker(in chan *Work, out chan *[]byte, wg *sync.WaitGroup) {
	defer wg.Done()
	for work := range in {
		recordMap := RecordToMap(work.Record, work.FilterMap, work.IncludeLeader)
		if work.PlainMode {
			b, err := json.Marshal(*recordMap)
			if err != nil {
				if !work.IgnoreErrors {
					log.Fatalln(err)
				}
				log.Printf("error: %s\n", err)
				continue
			}
			out <- &b
		} else {
			m := map[string]interface{}{
				work.RecordKey: *recordMap,
				"meta":         *work.MetaMap,
			}
			b, err := json.Marshal(m)
			if err != nil {
				if !work.IgnoreErrors {
					log.Fatalln(err)
				}
				log.Printf("[EE] %s\n", err)
				continue
			}
			out <- &b
		}
	}
}

// FanInWriter writes the channel content to the writer
func FanInWriter(writer io.Writer, in chan *[]byte, done chan bool) {
	for b := range in {
		writer.Write(*b)
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

// recordLength returns the length of the marc record as stored in the leader
func recordLength(reader io.Reader) (length int64, err error) {
	var l int
	data := make([]byte, 24)
	n, err := reader.Read(data)
	if err != nil {
		return 0, err
	}
	if n != 24 {
		return 0, fmt.Errorf("marc: invalid leader: expected 24 bytes, read %d", n)
	}
	l, err = strconv.Atoi(string(data[0:5]))
	if err != nil {
		return 0, fmt.Errorf("marc: invalid record length: %s", err)
	}
	return int64(l), nil
}

// RecordCount count the number of records in marc file
func RecordCount(filename string) int64 {
	handle, err := os.Open(filename)
	if err != nil {
		log.Fatalf("%s\n", err)
	}

	defer func() {
		if err := handle.Close(); err != nil {
			log.Fatalf("%s\n", err)
		}
	}()

	var i, cumulative int64

	for {
		length, err := recordLength(handle)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("%s\n", err)
		}
		i++
		cumulative += length
		handle.Seek(cumulative, 0)
	}
	return i
}

// IDList returns a slice of strings, containing all ids of the given marc file
func IDList(filename string) []string {
	fallback := false
	yaz, err := exec.LookPath("yaz-marcdump")
	if err != nil {
		fallback = true
	}

	awk, err := exec.LookPath("awk")
	if err != nil {
		fallback = true
	}

	var ids []string

	if fallback {
		// use slower iteration over records
		fi, err := os.Open(filename)
		if err != nil {
			log.Fatalf("%s\n", err)
		}
		defer func() {
			if err := fi.Close(); err != nil {
				log.Fatalf("%s\n", err)
			}
		}()

		for {
			record, err := marc22.ReadRecord(fi)
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatalf("%s\n", err)
			}

			fields := record.GetControlFields("001")
			if len(fields) != 1 {
				log.Fatalf("unusual 001 field count: %d\n", len(fields))
			}
			ids = append(ids, strings.TrimSpace(fields[0].Data))
		}
	} else {
		// fast version using yaz and awk
		command := fmt.Sprintf("%s '%s' | %s ' /^001 / {print $2}'", yaz, filename, awk)
		out, err := exec.Command("bash", "-c", command).Output()
		if err != nil {
			log.Fatalln(err)
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

// MarcMap writes (id, offset, length) TSV of a given MARC file to a io.Writer
func MarcMap(infile string, writer io.Writer) {
	fi, err := os.Open(infile)

	if err != nil {
		log.Fatalln(err)
	}

	defer func() {
		if err := fi.Close(); err != nil {
			log.Fatalln(err)
		}
	}()

	ids := IDList(infile)
	var i, offset int64

	for {
		length, err := recordLength(fi)
		if err == io.EOF {
			break
		}
		writer.Write([]byte(fmt.Sprintf("%s\t%d\t%d\n", ids[i], offset, length)))
		offset += length
		i++
		fi.Seek(offset, 0)
	}
}

// MarcMapSqlite writes (id, offset, length) sqlite3 database of a given MARC file to given output file
func MarcMapSqlite(infile, outfile string) {
	fi, err := os.Open(infile)

	if err != nil {
		log.Fatalln(err)
	}

	// dump results into sqlite3
	db, err := sql.Open("sqlite3", outfile)
	if err != nil {
		log.Fatalln(err)
	}
	defer db.Close()

	init := `CREATE TABLE IF NOT EXISTS seekmap (id text, offset int, length int)`
	_, err = db.Exec(init)
	if err != nil {
		log.Fatalf("%q: %s\n", err, init)
	}

	tx, err := db.Begin()
	if err != nil {
		log.Fatalln(err)
	}
	stmt, err := tx.Prepare("INSERT INTO seekmap VALUES (?, ?, ?)")
	if err != nil {
		log.Fatalln(err)
	}
	defer stmt.Close()

	ids := IDList(infile)
	var i, offset int64

	for {
		length, err := recordLength(fi)
		if err == io.EOF {
			break
		}
		_, err = stmt.Exec(ids[i], offset, length)
		if err != nil {
			log.Fatalln(err)
		}

		offset += length
		i++
		fi.Seek(offset, 0)
	}

	// create index
	_, err = tx.Exec("CREATE INDEX idx_seekmap_id ON seekmap (id)")
	if err != nil {
		log.Fatalln(err)
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
		log.Fatalln(err)
	}
}

// createSplitFile returns a writeable file object
func createSplitFile(directory, prefix string, fileno int64) *os.File {
	filename := filepath.Join(directory, fmt.Sprintf("%s%08d", prefix, fileno))
	output, err := os.Create(filename)
	if err != nil {
		log.Fatalln(err)
	}
	return output
}

// MarcSplitDirectoryPrefix splits a file into parts, each containing at most size records
// and writes the to specified directory, using a specific prefix
func MarcSplitDirectoryPrefix(infile string, size int64, directory, prefix string) {
	file, err := os.Open(infile)
	if err != nil {
		log.Fatalln(err)
	}

	defer func() {
		if err := file.Close(); err != nil {
			log.Fatalln(err)
		}
	}()

	var i, length, cumulative, offset, batch, fileno int64

	for {
		length, err = recordLength(file)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalln(err)
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

// recordToLeaderMap extracts the leader from a given record and puts it in a map
func recordToLeaderMap(record *marc22.Record) *map[string]string {
	leaderMap := make(map[string]string)
	leader := record.LeaderParsed
	leaderMap["status"] = string(leader.Status)
	leaderMap["cs"] = string(leader.CharacterEncoding)
	leaderMap["length"] = fmt.Sprintf("%d", leader.Length)
	leaderMap["type"] = string(leader.Type)
	leaderMap["impldef"] = string(leader.ImplementationDefined[:5])
	leaderMap["ic"] = fmt.Sprintf("%d", leader.IndicatorCount)
	leaderMap["lol"] = fmt.Sprintf("%d", leader.LengthOfLength)
	leaderMap["losp"] = fmt.Sprintf("%d", leader.LengthOfStartPos)
	leaderMap["sfcl"] = fmt.Sprintf("%d", leader.SubfieldCodeLength)
	leaderMap["ba"] = fmt.Sprintf("%d", leader.BaseAddress)
	leaderMap["raw"] = string(leader.Bytes())
	return &leaderMap
}

// recordToContentMap converts a record into a map, optionally only the tags given in filterMap
func recordToContentMap(record *marc22.Record, filterMap *map[string]bool) *map[string]interface{} {
	contentMap := make(map[string]interface{})

	filter := *filterMap
	hasFilter := len(filter) > 0

	for _, field := range record.ControlFields {
		tag := field.GetTag()
		if hasFilter {
			_, present := filter[tag]
			if !present {
				continue
			}
		}
		contentMap[tag] = field.Data
	}

	for _, field := range record.DataFields {
		tag := field.GetTag()
		if hasFilter {
			_, present := filter[tag]
			if !present {
				continue
			}
		}

		subfieldMap := make(map[string]interface{})
		subfieldMap["ind1"] = field.Ind1
		subfieldMap["ind2"] = field.Ind2
		for _, subfield := range field.SubFields {
			code := fmt.Sprintf("%s", subfield.Code)
			_, present := subfieldMap[code]
			if present {
				subfieldMap[code] = append(subfieldMap[code].([]string), subfield.Value)
			} else {
				subfieldMap[code] = []string{subfield.Value}
			}
		}
		_, present := contentMap[tag]
		if !present {
			subfields := make([]interface{}, 0)
			contentMap[tag] = subfields
		}
		contentMap[tag] = append(contentMap[tag].([]interface{}), subfieldMap)

	}

	return &contentMap
}

// RecordToMap converts a record to a map, optionally keeping only the tags
// given in filterMap. If includeLeader is true, the leader is converted as well.
func RecordToMap(record *marc22.Record, filterMap *map[string]bool, includeLeader bool) *map[string]interface{} {
	contentMap := *recordToContentMap(record, filterMap)
	if includeLeader {
		contentMap["leader"] = recordToLeaderMap(record)
	}
	return &contentMap
}

var regexSubfield = regexp.MustCompile(`^([\d]{3})\.([a-z0-9])$`)
var regexControlfield = regexp.MustCompile(`^[\d]{3}$`)

// RecordToTSV turns a single record into a single TSV line
func RecordToTSV(record *marc22.Record,
	tags *[]string,
	fillna, separator *string,
	skipIncompleteLines *bool) *string {

	var line []string
	skipThisLine := false

	for _, tag := range *tags {
		if regexControlfield.MatchString(tag) {
			fields := record.GetControlFields(tag)
			if len(fields) > 0 {
				line = append(line, fields[0].Data)
			} else {
				if *skipIncompleteLines {
					skipThisLine = true
					break
				}
				line = append(line, *fillna) // or any fill value
			}
		} else if regexSubfield.MatchString(tag) {
			parts := strings.Split(tag, ".")
			code := parts[1]
			subfields := record.GetSubFields(parts[0], code)
			if len(subfields) > 0 {
				if *separator == "" {
					// only use the first value
					line = append(line, subfields[0].Value)
				} else {
					var values []string
					for _, subfield := range subfields {
						values = append(values, subfield.Value)
					}
					line = append(line, strings.Join(values, *separator))
				}
			} else {
				if *skipIncompleteLines {
					skipThisLine = true
					break
				}
				line = append(line, *fillna) // or any fill value
			}
		} else if strings.HasPrefix(tag, "@") {
			leader := record.LeaderParsed
			switch tag {
			case "@Length":
				line = append(line, fmt.Sprintf("%d", leader.Length))
			case "@Status":
				line = append(line, string(leader.Status))
			case "@Type":
				line = append(line, string(leader.Type))
			case "@ImplementationDefined":
				line = append(line, string(leader.ImplementationDefined[:5]))
			case "@CharacterEncoding":
				line = append(line, string(leader.CharacterEncoding))
			case "@BaseAddress":
				line = append(line, fmt.Sprintf("%d", leader.BaseAddress))
			case "@IndicatorCount":
				line = append(line, fmt.Sprintf("%d", leader.IndicatorCount))
			case "@SubfieldCodeLength":
				line = append(line, fmt.Sprintf("%d", leader.SubfieldCodeLength))
			case "@LengthOfLength":
				line = append(line, fmt.Sprintf("%d", leader.LengthOfLength))
			case "@LengthOfStartPos":
				line = append(line, fmt.Sprintf("%d", leader.LengthOfStartPos))
			default:
				log.Fatalf("unknown tag: %s", tag)
			}
		} else if !strings.HasPrefix(tag, "-") {
			line = append(line, strings.TrimSpace(tag))
		}
	}

	result := ""
	if !skipThisLine {
		result = fmt.Sprintf("%s\n", strings.Join(line, "\t"))
	}
	return &result
}
