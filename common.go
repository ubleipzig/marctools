package marctools

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"github.com/miku/marc21"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

const AppVersion = "1.4.0"

// KeyValueStringToMap turns a string like "key1=value1, key2=value2" into a map.
func KeyValueStringToMap(s string) (map[string]string, error) {
	result := make(map[string]string)
	var err error
	if len(s) > 0 {
		for _, pair := range strings.Split(s, ",") {
			kv := strings.Split(pair, "=")
			if len(kv) != 2 {
				err = errors.New(fmt.Sprintf("Could not parse key-value parameter: %s", s))
			} else {
				key := strings.TrimSpace(kv[0])
				value := strings.TrimSpace(kv[1])
				if len(key) == 0 || len(value) == 0 {
					err = errors.New(fmt.Sprintf("Empty key or values not allowed: %s", s))
				} else {
					result[key] = value
				}
			}
		}
	}
	return result, err
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
	} else {
		if n != 24 {
			errs := fmt.Sprintf("MARC21: invalid leader: expected 24 bytes, read %d", n)
			err = errors.New(errs)
		} else {
			l, err = strconv.Atoi(string(data[0:5]))
			if err != nil {
				errs := fmt.Sprintf("MARC21: invalid record length: %s", err)
				err = errors.New(errs)
			}
		}
	}
	return int64(l), err
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
		i += 1
		cumulative += length
		handle.Seek(cumulative, 0)
	}
	return i
}

// IdList returns a slice of strings, containing all ids of the given marc file
func IdList(filename string) []string {
	fallback := false
	yaz, err := exec.LookPath("yaz-marcdump")
	if err != nil {
		// log.Fatalln("yaz-marcdump is required")
		fallback = true
	}

	awk, err := exec.LookPath("awk")
	if err != nil {
		// log.Fatalln("awk is required")
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
			record, err := marc21.ReadRecord(fi)
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatalf("%s\n", err)
			}

			fields := record.GetFields("001")
			if len(fields) == 1 {
				ids = append(ids, strings.TrimSpace(fields[0].(*marc21.ControlField).Data))
			} else {
				log.Fatalf("Unusual 001 field count: %s\n", len(fields))
			}
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
			} else {
				ids = append(ids, strings.TrimSpace(line))
			}
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

	ids := IdList(infile)
	var i, offset int64

	for {
		length, err := recordLength(fi)
		if err == io.EOF {
			break
		}
		writer.Write([]byte(fmt.Sprintf("%s\t%d\t%d\n", ids[i], offset, length)))
		offset += length
		i += 1
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

	ids := IdList(infile)
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
		i += 1
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

// MarcSplit splits a file into parts, each containing at most size records
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
			fileno += 1
			offset = cumulative
		}
		cumulative += length
		batch += length
		file.Seek(int64(cumulative), 0)
		i += 1
	}

	output := createSplitFile(directory, prefix, fileno)
	buffer := make([]byte, batch)
	writeSplit(file, output, offset, buffer)
}

// MarcSplit splits a file into parts, each containing at most size records
// and writes the to specified directory
func MarcSplitDirectory(infile string, size int64, directory string) {
	MarcSplitDirectoryPrefix(infile, size, directory, "split-")
}

// MarcSplit splits a file into parts, each containing at most size records
func MarcSplit(infile string, size int64) {
	MarcSplitDirectoryPrefix(infile, size, ".", "split-")
}

// recordToLeaderMap extracts the leader from a given record and puts it in a map
func recordToLeaderMap(record *marc21.Record) map[string]string {
	leaderMap := make(map[string]string)
	leader := record.Leader
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
	return leaderMap
}

// recordToContentMap converts a record into a map, optionally only the tags
// given in filterMap
func recordToContentMap(record *marc21.Record, filterMap map[string]bool) map[string]interface{} {
	contentMap := make(map[string]interface{})
	hasFilter := len(filterMap) > 0

	for _, field := range record.Fields {
		tag := field.GetTag()
		if hasFilter {
			_, present := filterMap[tag]
			if !present {
				continue
			}
		}
		if strings.HasPrefix(tag, "00") {
			contentMap[tag] = field.(*marc21.ControlField).Data
		} else {
			datafield := field.(*marc21.DataField)
			subfieldMap := make(map[string]interface{})
			subfieldMap["ind1"] = string(datafield.Ind1)
			subfieldMap["ind2"] = string(datafield.Ind2)
			for _, subfield := range datafield.SubFields {
				code := fmt.Sprintf("%c", subfield.Code)
				value, present := subfieldMap[code]
				if present {
					switch t := value.(type) {
					default:
						log.Fatalf("unexpected type: %T", t)
					case string:
						values := make([]string, 0)
						values = append(values, value.(string))
						values = append(values, subfield.Value)
						subfieldMap[code] = values
					case []string:
						subfieldMap[code] = append(subfieldMap[code].([]string), subfield.Value)
					}
				} else {
					subfieldMap[code] = subfield.Value
				}
			}
			_, present := contentMap[tag]
			if !present {
				subfields := make([]interface{}, 0)
				contentMap[tag] = subfields
			}
			contentMap[tag] = append(contentMap[tag].([]interface{}), subfieldMap)
		}
	}
	return contentMap
}

// recordToMap converts a record to a map, optionally keeping only the tags
// given in filterMap. If includeLeader is true, the leader is converted as well.
func recordToMap(record *marc21.Record, filterMap map[string]bool, includeLeader bool) map[string]interface{} {
	contentMap := recordToContentMap(record, filterMap)
	if includeLeader {
		contentMap["leader"] = recordToLeaderMap(record)
	}
	return contentMap
}

func MarcToJsonFile(infile, metaString, filterString string, outfile *os.File, includeLeader, plainMode, ignore bool) {
	file, err := os.Open(infile)
	if err != nil {
		log.Fatalln(err)
	}

	defer func() {
		if err := file.Close(); err != nil {
			log.Fatalln(err)
		}
	}()

	// poor mans set of tags, that should be converted
	filterMap := StringToMapSet(filterString)

	writer := bufio.NewWriter(outfile)
	defer writer.Flush()

	for {
		record, err := marc21.ReadRecord(file)
		if err == io.EOF {
			break
		}
		if err != nil {
			if ignore {
				fmt.Fprintf(os.Stderr, "Skipping, since -i was set. Error: %s\n", err)
				continue
			} else {
				log.Fatalln(err)
			}
		}

		if plainMode {
			b, err := json.Marshal(recordToMap(record, filterMap, includeLeader))
			if err != nil {
				log.Fatalf("error: %s", err)
			}
			writer.Write(b)
			writer.Write([]byte("\n"))
		} else {
			// the final map
			mainMap := make(map[string]interface{})

			mainMap["content"] = recordToMap(record, filterMap, includeLeader)

			metamap, err := KeyValueStringToMap(metaString)
			if err != nil {
				log.Fatalln(err)
			}

			mainMap["meta"] = metamap

			b, err := json.Marshal(mainMap)
			if err != nil {
				log.Fatalf("error: %s", err)
			}
			writer.Write(b)
			writer.Write([]byte("\n"))
		}
	}
}

// MarcToJsonFile with leader included, non-plain mode and strict error checking
func MarcToJson(infile, metaString, filterString string) {
	MarcToJsonFile(infile, metaString, filterString, os.Stdout, true, false, false)
}
