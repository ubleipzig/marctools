// Count records in a MARC file
package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/miku/marctools"
	"io"
	"os"
	"strconv"
)

func record_length(reader io.Reader) (length int64, err error) {
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

func main() {

	version := flag.Bool("v", false, "prints current program version")

	var PrintUsage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS] MARCFILE\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()

	if *version {
		fmt.Println(marctools.AppVersion)
		os.Exit(0)
	}

	if flag.NArg() < 1 {
		PrintUsage()
		os.Exit(1)
	}

	handle, err := os.Open(flag.Args()[0])
	if err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(1)
	}

	defer func() {
		if err := handle.Close(); err != nil {
			panic(err)
		}
	}()

	var i, cumulative int64

	for {
		length, err := record_length(handle)
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}
		i += 1
		cumulative += length
		handle.Seek(cumulative, 0)
	}
	fmt.Println(i)
}
