// Simple tool to test baseline iteration speed of marc records in Go.
// Using [marc21](https://gitorious.org/marc21-go/marc21).
//
// Test file: 4007756 records, 4.3G
//
// A baseline iteration, which only creates the MARC data structures takes about
// 4 minutes in Go, which amounts to about 17425 records per seconds.
//
// A simple yaz-marcdump -np seems to iterate over the same 4007756 records in
// about 30 seconds (133591 records per second) and a dev/nulled iteration about 65
// seconds (61657 records per second). So C is still three three to four times
// faster.
package main

import (
    "./marc21"
    "flag"
    "fmt"
    "io"
    "os"
)

const app_version = "1.3.3"

func main() {

    ignore := flag.Bool("i", false, "ignore marc errors (not recommended)")
    version := flag.Bool("v", false, "prints current program version")

    var PrintUsage = func() {
        fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS] MARCFILE\n", os.Args[0])
        flag.PrintDefaults()
    }

    flag.Parse()

    if *version {
        fmt.Println(app_version)
        os.Exit(0)
    }

    if flag.NArg() < 1 {
        PrintUsage()
        os.Exit(1)
    }

    fi, err := os.Open(flag.Args()[0])
    if err != nil {
        panic(err)
    }

    defer func() {
        if err := fi.Close(); err != nil {
            panic(err)
        }
    }()

    counter := 0
    for {
        _, err := marc21.ReadRecord(fi)
        if err == io.EOF {
            break
        }
        if err != nil {
            if *ignore {
                fmt.Fprintf(os.Stderr, "Skipping, since -i was set. Error: %s\n", err)
                continue
            } else {
                panic(err)
            }
        }
        counter += 1
    }
    fmt.Printf("Looped over %d records.\n", counter)
}
