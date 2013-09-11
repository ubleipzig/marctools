package main

import (
    "./marc21" // https://gitorious.org/marc21-go/marc21
    "flag"
    "fmt"
    "io"
    "os"
    "regexp"
    "strings"
)

const version = "1.0.0"

func main() {

    fillna := flag.String("f", "<NULL>", "fill missing values with this")
    ignore := flag.Bool("i", false, "ignore marc errors (not recommended)")
    skip := flag.Bool("k", false, "skip lines with missing values")
    version := flag.Bool("v", false, "prints current program version")

    var PrintUsage = func() {
        fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS] MARCFILE TAG [TAG, ...]\n", os.Args[0])
        flag.PrintDefaults()
    }

    flag.Parse()

    if *version {
        fmt.Println(version)
        os.Exit(0)
    }

    if flag.NArg() < 2 {
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

    resf, err := regexp.Compile(`^([\d]{3})\.([a-z0-9])$`)
    if err != nil {
        panic("invalid regex")
    }
    recf, err := regexp.Compile(`^[\d]{3}$`)
    if err != nil {
        panic("invalid regex")
    }

    tags := flag.Args()[1:]

    for {
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

        var line []string
        skipline := false

        for _, tag := range tags {
            if recf.MatchString(tag) {
                fields := record.GetFields(tag)
                if len(fields) > 0 {
                    line = append(line, fields[0].(*marc21.ControlField).Data)
                } else {
                    if *skip {
                        skipline = true
                        break
                    }
                    line = append(line, *fillna) // or any fill value
                }
            }

            if resf.MatchString(tag) {
                parts := strings.Split(tag, ".")
                subfields := record.GetSubFields(parts[0], []byte(parts[1])[0])
                if len(subfields) > 0 {
                    line = append(line, subfields[0].Value)
                } else {
                    if *skip {
                        skipline = true
                        break
                    }
                    line = append(line, *fillna) // or any fill value  
                }
            }
        }
        if skipline {
            continue
        } else {
            fmt.Printf("%s\n", strings.Join(line, "\t"))
            line = line[:0]
        }
    }
    return
}
