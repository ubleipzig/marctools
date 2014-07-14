package marctools

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
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
