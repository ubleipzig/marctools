// Create a seekmap of the form (sorted by OFFSET)
// ID OFFSET LENGTH
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime/pprof"

	"github.com/miku/marctools"
)

func main() {

	version := flag.Bool("v", false, "prints current program version")
	output := flag.String("o", "", "output to sqlite3 file")
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

	filename := flag.Args()[0]

	if *output == "" {
		marctools.MarcMap(filename, os.Stdout)
	} else {
		marctools.MarcMapSqlite(filename, *output)
	}
}
