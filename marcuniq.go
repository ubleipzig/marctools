package main

import (
    "./marc21"
    "bufio"
    "flag"
    "fmt"
    "io"
    "os"
    "strings"
)

const app_version = "1.3.6"

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

    ignore := flag.Bool("i", false, "ignore marc errors (not recommended)")
    version := flag.Bool("v", false, "prints current program version")
    outfile := flag.String("o", "", "output file (or stdout if none given)")
    exclude := flag.String("x", "", "comma separated list of ids to exclude (or filename with one id per line)")

    var PrintUsage = func() {
        fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS] MARCFILE\n", os.Args[0])
        flag.PrintDefaults()
    }

    flag.Parse()

    // display version and exit
    if *version {
        fmt.Println(app_version)
        os.Exit(0)
    }

    if flag.NArg() != 1 {
        PrintUsage()
        os.Exit(1)
    }

    // input file
    fi, err := os.Open(flag.Args()[0])
    if err != nil {
        panic(err)
    }

    defer func() {
        if err := fi.Close(); err != nil {
            panic(err)
        }
    }()

    // output file or stdout
    var output *os.File

    if *outfile == "" {
        output = os.Stdout
    } else {
        output, err = os.Create(*outfile)
        if err != nil {
            panic(err)
        }
        defer func() {
            if err := output.Close(); err != nil {
                panic(err)
            }
        }()
    }

    // exclude list
    excludedIds := NewStringSet()

    if *exclude != "" {
        if _, err := os.Stat(*exclude); err != nil {
            if os.IsNotExist(err) {
                fmt.Fprintf(os.Stderr, "excluded ids interpreted as string\n")
                for _, value := range strings.Split(*exclude, ",") {
                    excludedIds.Add(strings.TrimSpace(value))
                }
            } else if err != nil {
                panic(err)
            }
        } else {
            fmt.Fprintf(os.Stderr, "excluded ids interpreted as file\n")

            // read one id per line from file
            xfi, err := os.Open(*exclude)
            if err != nil {
                panic(err)
            }

            defer func() {
                if err := xfi.Close(); err != nil {
                    panic(err)
                }
            }()

            scanner := bufio.NewScanner(xfi)
            for scanner.Scan() {
                excludedIds.Add(strings.TrimSpace(scanner.Text()))
            }
        }
        fmt.Fprintf(os.Stderr, "%d ids to exclude loaded\n", excludedIds.Size())
    }

    // collect the excluded ids here
    excluded := make([]string, 0, 0)

    // keep track of all ids
    ids := NewStringSet()
    // collect the duplicate ids; array, since same id may occur many times
    // skipped could be an integer for now, because we do not display the skipped
    // records (TODO: add flag to display skipped records)
    skipped := make([]string, 0, 0)
    // just count the total records and those without id
    var counter, without_id int

    for {
        head, _ := fi.Seek(0, os.SEEK_CUR)
        record, err := marc21.ReadRecord(fi)
        if err == io.EOF {
            break
        }
        if err != nil {
            if *ignore {
                fmt.Fprintf(os.Stderr, "Skipping, since -i: %s\n", err)
                continue
            } else {
                panic(err)
            }
        }
        tail, _ := fi.Seek(0, os.SEEK_CUR)
        length := tail - head

        fields := record.GetFields("001")
        if len(fields) > 0 {
            id := fields[0].(*marc21.ControlField).Data
            if ids.Contains(id) {
                skipped = append(skipped, id)
            } else if excludedIds.Contains(id) {
                excluded = append(excluded, id)
            } else {
                ids.Add(id)
                fi.Seek(head, 0)
                buf := make([]byte, length)
                n, err := fi.Read(buf)
                if err != nil {
                    panic(err)
                }
                if _, err := output.Write(buf[:n]); err != nil {
                    panic(err)
                }
            }
        } else if len(fields) == 0 {
            without_id += 1
        }
        counter += 1
    }

    fmt.Fprintf(os.Stderr, "%d records read\n", counter)
    fmt.Fprintf(os.Stderr, "%d records written, %d skipped, %d excluded, %d without ID (001)\n",
        ids.Size(), len(skipped), len(excluded), without_id)
}
