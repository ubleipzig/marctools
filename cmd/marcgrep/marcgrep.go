// marcgrep - a basic MARC filter
//
// Examples
//
// Keep only records, where field 007 starts with a small letter 't'
//
//     $ marcgrep -field 007 -regex "^t" file.mrc > filtered.mrc
//
// Keep only records, where any 245.a field contains Berlin
//
//     $ marcgrep -field 245.a -regex ".*Berlin.*" file.mrc > filtered.mrc
//
package main

import (
	"bufio"
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"runtime"
	"runtime/pprof"
	"strings"
	"sync"

	"github.com/miku/marc22"
	"github.com/ubleipzig/marctools"
)

type work struct {
	filename string
	record   *marc22.Record
	offset   int64
	length   int64
}

type workerOptions struct {
	field string
	regex regexp.Regexp
}

type writerOptions struct {
	filename string
	writer   io.Writer
}

func worker(in chan *work, out chan *work, wg *sync.WaitGroup, options *workerOptions) {
	defer wg.Done()
	for w := range in {
		if len(options.field) == 3 {
			fields := w.record.GetControlFields(options.field)
			for _, f := range fields {
				if options.regex.MatchString(f.Data) {
					out <- w
				}
			}
		} else if len(options.field) == 5 {
			parts := strings.Split(options.field, ".")
			code := parts[1]
			subfields := w.record.GetSubFields(parts[0], code)
			for _, subfield := range subfields {
				if options.regex.MatchString(subfield.Value) {
					out <- w
				}
			}
		} else {
			log.Fatalf("unknown field spec: %s", options.field)
		}
	}
}

// stringFanWriter writes the channel content to the writer
func stringFanWriter(in chan *work, done chan bool, options *writerOptions) {
	fi, err := os.Open(options.filename)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := fi.Close(); err != nil {
			log.Fatal(err)
		}
	}()
	for w := range in {
		fi.Seek(w.offset, 0)
		_, err = io.CopyN(options.writer, fi, w.length)
		if err != nil && err != io.EOF {
			log.Fatal(err)
		}
		binary.Write(options.writer, binary.LittleEndian, marc22.RS)
	}
	done <- true
}

func main() {
	field := flag.String("field", "", "field to apply regex on")
	regex := flag.String("regex", ".*", "regex to apply on given field value (re2 syntax)")
	output := flag.String("o", "", "output file")

	version := flag.Bool("v", false, "prints current program version")
	cpuprofile := flag.String("cpuprofile", "", "write cpu profile to file")
	numWorkers := flag.Int("w", runtime.NumCPU(), "number of workers")

	var PrintUsage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS] MARCFILE\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	if *version {
		fmt.Println(marctools.AppVersion)
		os.Exit(0)
	}

	if flag.NArg() != 1 {
		PrintUsage()
		os.Exit(1)
	}

	regexField, err := regexp.Compile(*regex)
	if err != nil {
		log.Fatal(err)
	}

	filename := flag.Args()[0]
	fi, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := fi.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	queue := make(chan *work)
	results := make(chan *work)
	done := make(chan bool)

	var writer *bufio.Writer
	if *output == "" {
		writer = bufio.NewWriter(os.Stdout)
	} else {
		w, err := os.Create(*output)
		writer = bufio.NewWriter(w)
		if err != nil {
			log.Fatal(err)
		}
		defer func() {
			if err := w.Close(); err != nil {
				log.Fatal(err)
			}
		}()
	}

	defer writer.Flush()
	go stringFanWriter(results, done, &writerOptions{filename: filename, writer: writer})

	var wg sync.WaitGroup
	for i := 0; i < *numWorkers; i++ {
		wg.Add(1)
		go worker(queue, results, &wg, &workerOptions{field: *field, regex: *regexField})
	}

	fmt.Printf("Filtering %s with %s\n", *field, *regex)

	var offset, length int64
	for {
		length, err = marctools.RecordLength(fi)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalln(err)
		}

		fi.Seek(-24, 1)
		offset += length

		record, err := marc22.ReadRecord(fi)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("%s\n", err)
		}
		queue <- &work{record: record, offset: offset, length: length}
	}

	close(queue)
	wg.Wait()
	close(results)
	<-done
}
