// Count records in a MARC file
package main

import (
	"flag"
	"fmt"
	"github.com/miku/marctools"
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

	if flag.NArg() < 1 {
		PrintUsage()
		os.Exit(1)
	}

	count := marctools.RecordCount(flag.Args()[0])
	fmt.Println(count)
}
