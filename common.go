package marctools

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"github.com/miku/marc21"
	"io"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

const AppVersion = "1.4.0"

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
