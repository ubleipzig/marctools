// convert marc to json
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
	// MARC record
	Record *marc21.Record
	// which tags to include
	FilterMap *map[string]bool
	// meta information
	MetaMap       *map[string]string
	IncludeLeader bool
	// only dump the content
	PlainMode    bool
	IgnoreErrors bool
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
				}
			}
			out <- &b
		}
	}
}

// FanInWriter takes a writer and a channel of byte slices and writes the out
func FanInWriter(writer io.Writer, in chan *[]byte, done chan bool) {
	for b := range in {
		writer.Write(*b)
		writer.Write([]byte("\n"))
	}
	done <- true
}

func main() {

	ignore := flag.Bool("i", false, "ignore marc errors (not recommended)")
	version := flag.Bool("v", false, "prints current program version and exit")

	metaVar := flag.String("m", "", "a key=value pair to pass to meta")
	filterVar := flag.String("r", "", "only dump the given tags (e.g. 001,003)")
	leaderVar := flag.Bool("l", false, "dump the leader as well")
	plainVar := flag.Bool("p", false, "plain mode: dump without content and meta")
	cpuprofile := flag.String("cpuprofile", "", "write cpu profile to file")
	numWorkers := flag.Int("w", runtime.NumCPU(), "number of workers")

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

	// could use some other writer here, via flag?
	writer := bufio.NewWriter(os.Stdout)
	defer writer.Flush()

	// start the writer
	go FanInWriter(writer, results, done)

	var wg sync.WaitGroup

	// start the workers
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
			if *ignore {
				fmt.Fprintf(os.Stderr, "skipping error: %s\n", err)
				continue
			} else {
				log.Fatalln(err)
			}
		}

		work := Work{Record: record,
			FilterMap:     &filterMap,
			MetaMap:       &metaMap,
			IncludeLeader: *leaderVar,
			PlainMode:     *plainVar,
			IgnoreErrors:  *ignore}
		queue <- &work
	}

	// cleanup
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
