// Convert marc to tsv.
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
	"time"

	"github.com/miku/marc22"
	"github.com/miku/marctools"
)

type Work struct {
	Record              *marc22.Record // MARC record
	Tags                *[]string      // tags to dump
	FillNA              *string        // placeholder if value is not available
	Separator           *string
	SkipIncompleteLines *bool // skip lines, that do
}

// Worker takes a Work item and sends the result (serialized json) on the out channel
func Worker(in chan *Work, out chan *string, wg *sync.WaitGroup) {
	defer wg.Done()
	for work := range in {
		line := marctools.RecordToTSV(work.Record, work.Tags, work.FillNA, work.Separator, work.SkipIncompleteLines)
		if len(*line) > 0 {
			out <- line
		}
	}
}

// FanInWriter writes the channel content to the writer
func FanInWriter(writer io.Writer, in chan *string, done chan bool) {
	for s := range in {
		writer.Write([]byte(*s))
	}
	done <- true
}

func main() {

	cpuprofile := flag.String("cpuprofile", "", "write cpu profile to file")
	ignoreErrors := flag.Bool("i", false, "ignore marc errors (not recommended)")
	numWorkers := flag.Int("w", runtime.NumCPU(), "number of workers")
	version := flag.Bool("v", false, "prints current program version and exit")

	fillna := flag.String("f", "<NULL>", "fill missing values with this")
	separator := flag.String("s", "", "separator to use for multiple values")
	skipIncompleteLines := flag.Bool("k", false, "skip incomplete lines (missing values)")

	var PrintUsage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS] MARCFILE TAG [TAG, TAG, ...]\n", os.Args[0])
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

	tags := flag.Args()[1:]

	if len(tags) == 0 {
		log.Fatalln("at least one tag is required")
	}

	queue := make(chan *Work)
	results := make(chan *string)
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
		record, err := marc22.ReadRecord(file)
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
			Tags:                &tags,
			FillNA:              fillna,
			Separator:           separator,
			SkipIncompleteLines: skipIncompleteLines}
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
