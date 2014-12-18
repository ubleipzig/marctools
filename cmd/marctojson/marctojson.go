// Convert marc to json.

// Performance data point: Converting 6537611 records (7G) into /dev/null
// take about 9m31s on a Core i5-3470 (about 11k records/s).
// To take a cpu profile use -cpuprofile flag (example output: https://cdn.mediacru.sh/5rLMpxn5qnJk.svg).
// A previous [single threaded Java app](https://github.com/miku/marctojson)
// took about 46m for the same file (2k records/s).
package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
	"sync"

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
	recordKey := flag.String("recordkey", "record", "key name of the record")
	plainMode := flag.Bool("p", false, "plain mode: dump without content and meta")

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
		log.Fatal(err)
	}

	defer func() {
		if err := file.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	filterMap := marctools.StringToMapSet(*filterVar)
	metaMap, err := marctools.KeyValueStringToMap(*metaVar)
	if err != nil {
		log.Fatal(err)
	}

	queue := make(chan *marc22.Record)
	results := make(chan *[]byte)
	done := make(chan bool)

	writer := bufio.NewWriter(os.Stdout)
	defer writer.Flush()
	go marctools.FanInWriter(writer, results, done)

	var wg sync.WaitGroup
	options := marctools.JsonConversionOptions{
		FilterMap:     &filterMap,
		MetaMap:       &metaMap,
		IncludeLeader: *includeLeader,
		PlainMode:     *plainMode,
		IgnoreErrors:  *ignoreErrors,
		RecordKey:     *recordKey,
	}
	for i := 0; i < *numWorkers; i++ {
		wg.Add(1)
		go marctools.Worker(queue, results, &wg, &options)
	}

	for {
		record, err := marc22.ReadRecord(file)
		if err == io.EOF {
			break
		}
		if err != nil {
			if *ignoreErrors {
				log.Println(err)
				continue
			} else {
				log.Fatal(err)
			}
		}

		queue <- record
	}

	close(queue)
	wg.Wait()
	close(results)
	<-done
}
