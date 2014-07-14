package main

import (
	"flag"
	"fmt"
	"github.com/miku/marc21"
	"github.com/miku/marctools"
	"io"
	"log"
	"os"
)

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

	if flag.NArg() != 1 {
		PrintUsage()
		os.Exit(1)
	}

	fi, err := os.Open(flag.Args()[0])
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

		fmt.Printf("%s\n", record.String())
	}
	return
}
