package main

import (
    "./marc21"
    "flag"
    "fmt"
    "io"
    "log"
    "os"
)

// poor mans string set
type StringSet struct {
    set map[string]bool
}

func NewStringSet() *StringSet {
    return &StringSet{set: make(map[string]bool)}
}

func (set *StringSet) Add(s string) bool {
    _, found := set.set[s]
    set.set[s] = true
    return !found //False if it existed already
}

func (set *StringSet) Contains(s string) bool {
    _, found := set.set[s]
    return found
}

func (set *StringSet) Size() int {
    return len(set.set)
}

func main() {

    const app_version = "1.3.0"

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

    if flag.NArg() != 1 {
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

    ids := NewStringSet()
    skipped := make([]string, 0, 0)

    counter := 0
    for {
        record, err := marc21.ReadRecord(fi)
        if err == io.EOF {
            break
        }
        if err != nil {
            if *ignore {
                fmt.Fprintf(os.Stderr, "Skipping, since -i was set. Error: %s\n",
                    err)
                continue
            } else {
                panic(err)
            }
        }

        fields := record.GetFields("001")
        id := fields[0].(*marc21.ControlField).Data
        if ids.Contains(id) {
            // log.Printf("Skipping duplicate: %s", id)
            skipped = append(skipped, id)
        } else {
            ids.Add(id)
        }
        counter += 1
    }

    log.Printf("Looped over %d records.\n", counter)
    log.Printf("Uniq: %d\n", ids.Size())
    // log.Printf("Skipped: %d\n", skipped.Size())
    log.Printf("Skipped: %d\n", len(skipped))
}
