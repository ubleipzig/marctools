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
	Record        *marc21.Record
	FilterMap     *map[string]bool
	MetaMap       *map[string]string
	IncludeLeader bool
	PlainMode     bool
	IgnoreErrors  bool
}

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

func FanInWriter(writer io.Writer, in chan *[]byte, done chan bool) {
	for b := range in {
		writer.Write(*b)
		writer.Write([]byte("\n"))
	}
	done <- true
}

// preformance data points:
// 9798925 records, sequential
// $ time go run cmd/marctojson/marctojson.go fixtures/biglok.mrc > /dev/null
// real	7m18.731s
// user	6m16.256s
// sys	1m13.612s

// 9798925 records, single short-lived goroutine
// $ time go run cmd/marctojson/marctojson.go fixtures/biglok.mrc > /dev/null
// real	12m49.862s
// user	12m39.992s
// sys	3m23.380s

// 9798925 records, NumCPU parallel
// $ time ./marctojson.go fixtures/biglok.mrc > /dev/null
// real    9m4.779s
// user    13m52.302s
// sys 6m41.470s
func main() {

	runtime.GOMAXPROCS(runtime.NumCPU())

	ignore := flag.Bool("i", false, "ignore marc errors (not recommended)")
	version := flag.Bool("v", false, "prints current program version and exit")

	metaVar := flag.String("m", "", "a key=value pair to pass to meta")
	filterVar := flag.String("r", "", "only dump the given tags (e.g. 001,003)")
	leaderVar := flag.Bool("l", false, "dump the leader as well")
	plainVar := flag.Bool("p", false, "plain mode: dump without content and meta")
	cpuprofile := flag.String("cpuprofile", "", "write cpu profile to file")

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
	for i := 0; i < runtime.NumCPU(); i++ {
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

	close(queue)
	wg.Wait()

	// wait for the writer to finish, but not too long
	close(results)
	select {
	case <-time.After(1e9):
		break
	case <-done:
		break
	}
}
