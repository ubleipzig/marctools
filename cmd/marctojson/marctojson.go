// Convert marc to json.

// Performance data point: Converting 6537611 records (7G) into /dev/null
// take about 9m31s on a Core i5-3470 (about 11k records/s).
// To take a cpu profile use -cpuprofile flag (example output: https://cdn.mediacru.sh/5rLMpxn5qnJk.svg).
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/miku/marc21"
	"github.com/miku/marctools"
	"io"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
	"sync"
	"time"
)

type Work struct {
	Record        *marc21.Record     // MARC record
	FilterMap     *map[string]bool   // which tags to include
	MetaMap       *map[string]string // meta information
	IncludeLeader bool
	PlainMode     bool // only dump the content
	IgnoreErrors  bool
}

// Worker takes a Work item and sends the result (serialized json) on the out channel
func Worker(in chan *Work, out chan *[]byte, wg *sync.WaitGroup) {
	defer wg.Done()
	for work := range in {
		recordMap := marctools.RecordToMap(work.Record, work.FilterMap, work.IncludeLeader)
		if work.PlainMode {
			b, err := json.Marshal(recordMap)
			if err != nil {
				if !work.IgnoreErrors {
					log.Fatalln(err)
				} else {
					log.Printf("[EE] %s\n", err)
					continue
				}
			}
			out <- &b
		} else {
			m := map[string]interface{}{
				"content": recordMap,
				"meta":    *work.MetaMap,
			}
			b, err := json.Marshal(m)
			if err != nil {
				if !work.IgnoreErrors {
					log.Fatalln(err)
				} else {
					log.Printf("[EE] %s\n", err)
					continue
				}
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

func main() {

	cpuprofile := flag.String("cpuprofile", "", "write cpu profile to file")
	ignoreErrors := flag.Bool("i", false, "ignore marc errors (not recommended)")
	numWorkers := flag.Int("w", runtime.NumCPU(), "number of workers")
	version := flag.Bool("v", false, "prints current program version and exit")

	filterVar := flag.String("r", "", "only dump the given tags (e.g. 001,003)")
	includeLeader := flag.Bool("l", false, "dump the leader as well")
	metaVar := flag.String("m", "", "a key=value pair to pass to meta")
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

	queue := make(chan *Work)
	results := make(chan *[]byte)
	done := make(chan bool)

	writer := bufio.NewWriter(os.Stdout)
	defer writer.Flush()
	go FanInWriter(writer, results, done)

	var wg sync.WaitGroup
	for i := 0; i < *numWorkers; i++ {
		wg.Add(1)
		go Worker(queue, results, &wg)
	}

	for {
		record, err := marc21.ReadRecord(file)
		if err == io.EOF {
			break
		}
		if err != nil {
			if *ignoreErrors {
				log.Printf("[EE] %s\n", err)
				continue
			} else {
				log.Fatalln(err)
			}
		}

		work := Work{Record: record,
			FilterMap:     &filterMap,
			MetaMap:       &metaMap,
			IncludeLeader: *includeLeader,
			PlainMode:     *plainMode,
			IgnoreErrors:  *ignoreErrors}
		queue <- &work
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
