// Go version of "yaz-marcdump -s prefix -C 1000 file.mrc"
package main

import (
	"flag"
	"fmt"
	"github.com/miku/marctools"
	"log"
	"os"
)

func main() {

	directory := flag.String("d", ".", "directory to write to")
	prefix := flag.String("s", "split-", "split file prefix")
	size := flag.Int64("C", 1, "number of records per file")
	version := flag.Bool("v", false, "prints current program version")
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

	fi, err := os.Stat(*directory)
	if os.IsNotExist(err) {
		log.Fatalf("no such file or directory: %s\n", *directory)
	}
	if !fi.IsDir() {
		log.Fatalf("arg to -d must be directory: %s\n", *directory)
	}
	filename := flag.Args()[0]
	marctools.MarcSplitDirectoryPrefix(filename, *size, *directory, *prefix)
}
