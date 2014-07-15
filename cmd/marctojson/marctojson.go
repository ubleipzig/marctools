package main

import (
	"flag"
	"fmt"
	"github.com/miku/marctools"
	"os"
)

func main() {

	ignore := flag.Bool("i", false, "ignore marc errors (not recommended)")
	version := flag.Bool("v", false, "prints current program version and exit")

	metaVar := flag.String("m", "", "a key=value pair to pass to meta")
	filterVar := flag.String("r", "", "only dump the given tags (e.g. 001,003)")
	leaderVar := flag.Bool("l", false, "dump the leader as well")
	plainVar := flag.Bool("p", false, "plain mode: dump without content and meta")

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

	filename := flag.Args()[0]
	marctools.MarcToJsonFile(filename, *metaVar, *filterVar, os.Stdout, *leaderVar, *plainVar, *ignore)
}
