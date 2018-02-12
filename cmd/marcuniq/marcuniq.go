package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/miku/marc22"
	"github.com/ubleipzig/marctools"
)

func main() {

	ignore := flag.Bool("i", false, "ignore marc errors (not recommended)")
	version := flag.Bool("v", false, "prints current program version")
	outfile := flag.String("o", "", "output file (or stdout if none given)")
	exclude := flag.String("x", "", "comma separated list of ids to exclude (or filename with one id per line)")

	var PrintUsage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS] MARCFILE\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()

	// display version and exit
	if *version {
		fmt.Println(marctools.AppVersion)
		os.Exit(0)
	}

	if flag.NArg() != 1 {
		PrintUsage()
		os.Exit(1)
	}

	// input file
	fi, err := os.Open(flag.Args()[0])
	if err != nil {
		log.Fatalln(err)
	}

	defer func() {
		if err := fi.Close(); err != nil {
			log.Fatalln(err)
		}
	}()

	// output file or stdout
	var output *os.File
	if *outfile == "" {
		output = os.Stdout
	} else {
		output, err = os.Create(*outfile)
		if err != nil {
			log.Fatalln(err)
		}
		defer func() {
			if err := output.Close(); err != nil {
				log.Fatalln(err)
			}
		}()
	}

	// exclude list
	excludedIds := marctools.NewStringSet()

	if *exclude != "" {
		if _, err := os.Stat(*exclude); err != nil {
			if os.IsNotExist(err) {
				fmt.Fprintf(os.Stderr, "excluded ids interpreted as string\n")
				for _, value := range strings.Split(*exclude, ",") {
					excludedIds.Add(strings.TrimSpace(value))
				}
			} else if err != nil {
				log.Fatalln(err)
			}
		} else {
			fmt.Fprintf(os.Stderr, "excluded ids interpreted as file\n")

			// read one id per line from file
			file, err := os.Open(*exclude)
			if err != nil {
				log.Fatalln(err)
			}

			defer func() {
				if err := file.Close(); err != nil {
					log.Fatalln(err)
				}
			}()

			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				excludedIds.Add(strings.TrimSpace(scanner.Text()))
			}
		}
		fmt.Fprintf(os.Stderr, "%d ids to exclude loaded\n", excludedIds.Size())
	}

	// collect the excluded ids here
	excluded := make([]string, 0, 0)

	// keep track of all ids
	ids := marctools.NewStringSet()
	// collect the duplicate ids; array, since same id may occur many times
	// skipped could be an integer for now, because we do not display the skipped
	// records (TODO: add flag to display skipped records)
	skipped := make([]string, 0, 0)
	// just count the total records and those without id
	var counter, withoutID int

	for {
		head, _ := fi.Seek(0, os.SEEK_CUR)
		record, err := marc22.ReadRecord(fi)
		if err == io.EOF {
			break
		}
		if err != nil {
			if *ignore {
				fmt.Fprintf(os.Stderr, "skipping error: %s\n", err)
				continue
			} else {
				log.Fatalln(err)
			}
		}
		tail, _ := fi.Seek(0, os.SEEK_CUR)
		length := tail - head

		fields := record.GetControlFields("001")
		if len(fields) > 0 {
			id := fields[0].Data
			if ids.Contains(id) {
				skipped = append(skipped, id)
			} else if excludedIds.Contains(id) {
				excluded = append(excluded, id)
			} else {
				ids.Add(id)
				fi.Seek(head, 0)
				buf := make([]byte, length)
				n, err := fi.Read(buf)
				if err != nil {
					log.Fatalln(err)
				}
				if _, err := output.Write(buf[:n]); err != nil {
					log.Fatalln(err)
				}
			}
		} else if len(fields) == 0 {
			withoutID++
		}
		counter++
	}

	fmt.Fprintf(os.Stderr, "%d records read\n", counter)
	fmt.Fprintf(os.Stderr, "%d records written, %d skipped, %d excluded, %d without ID (001)\n",
		ids.Size(), len(skipped), len(excluded), withoutID)
}
