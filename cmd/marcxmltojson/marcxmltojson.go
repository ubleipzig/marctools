package main

import (
	"bufio"
	"encoding/xml"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
	"sync"
	"time"

	"github.com/miku/marc22"
	"github.com/ubleipzig/marctools"
)

func main() {

	cpuprofile := flag.String("cpuprofile", "", "write cpu profile to file")
	ignoreErrors := flag.Bool("i", false, "ignore marc errors (not recommended)")
	numWorkers := flag.Int("w", runtime.NumCPU(), "number of workers")
	version := flag.Bool("v", false, "prints current program version and exit")

	filterVar := flag.String("r", "", "only dump the given tags (e.g. 001,003)")
	includeLeader := flag.Bool("l", false, "dump the leader as well")
	metaVar := flag.String("m", "", "a key=value pair to pass to meta")
	plainMode := flag.Bool("p", false, "plain mode: dump without content and meta")
	recordKey := flag.String("recordkey", "record", "key name of the record")

	var PrintUsage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS] MARCFILE\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()

	if *numWorkers > 0 {
		runtime.GOMAXPROCS(*numWorkers)
	}

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

	if flag.NArg() < 1 {
		PrintUsage()
		os.Exit(1)
	}

	file, err := os.Open(flag.Args()[0])
	if err != nil {
		log.Fatalln(err)
	}

	defer func() {
		if err := file.Close(); err != nil {
			log.Fatalln(err)
		}
	}()

	filterMap := marctools.StringToMapSet(*filterVar)
	metaMap, err := marctools.KeyValueStringToMap(*metaVar)
	if err != nil {
		log.Fatalln(err)
	}

	queue := make(chan *marc22.Record)
	results := make(chan []byte)
	done := make(chan bool)

	writer := bufio.NewWriter(os.Stdout)
	defer writer.Flush()
	go marctools.FanInWriter(writer, results, done)

	options := marctools.JSONConversionOptions{
		FilterMap:     filterMap,
		MetaMap:       metaMap,
		IncludeLeader: *includeLeader,
		PlainMode:     *plainMode,
		IgnoreErrors:  *ignoreErrors,
		RecordKey:     *recordKey,
	}

	var wg sync.WaitGroup
	for i := 0; i < *numWorkers; i++ {
		wg.Add(1)
		go marctools.Worker(queue, results, &wg, options)
	}

	decoder := xml.NewDecoder(file)

	for {
		t, _ := decoder.Token()
		if t == nil {
			break
		}
		switch se := t.(type) {
		case xml.StartElement:
			if se.Name.Local == "record" {
				var record marc22.Record
				decoder.DecodeElement(&record, &se)
				queue <- &record
			}
		}
	}

	close(queue)
	wg.Wait()
	close(results)
	select {
	case <-time.After(1e9):
		break
	case <-done:
		break
	}
}
