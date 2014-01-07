// Create a seekmap of the form (sorted by OFFSET)
// ID OFFSET LENGTH
package main

import (
    "./marc21"
    "flag"
    "fmt"
    "io"
    "os"
)

const app_version = "1.3.4"

func main() {

    version := flag.Bool("v", false, "prints current program version")
    ignore := flag.Bool("i", false, "ignore marc errors (not recommended)")

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
        fmt.Printf("%s\n", err)
        os.Exit(1)
    }

    defer func() {
        if err := fi.Close(); err != nil {
            panic(err)
        }
    }()

    for {
        head, _ := fi.Seek(0, os.SEEK_CUR)
        record, err := marc21.ReadRecord(fi)
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
        tail, _ := fi.Seek(0, os.SEEK_CUR)
        length := tail - head

        fields := record.GetFields("001")
        id := fields[0].(*marc21.ControlField).Data
        fmt.Printf("%s\t%d\t%d\n", id, head, length)
    }
}
