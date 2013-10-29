package main

// https://gitorious.org/marc21-go/marc21
import "./marc21"
import "flag"
import "fmt"
import "io"
import "os"

const app_version = "1.0.0"

func main() {

    version := flag.Bool("v", false, "prints current program version")
    strict := flag.Bool("s", false, "panic on missing sigel")
    ignore := flag.Bool("i", false, "ignore all errors")

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

    for {
        record, err := marc21.ReadRecord(fi)
        if err == io.EOF {
            break
        }
        if err != nil {
            panic(err)
        }

        var fields []marc21.Field
        var epn, ppn, date, sigel string

        fields = record.GetFields("001")
        if len(fields) > 0 {
            epn = fields[0].(*marc21.ControlField).Data
        } else {
            if *ignore {
                continue
            } else {
                panic(`No 001. Are you sure this is local data?`)
            }
        }

        fields = record.GetFields("004")
        if len(fields) > 0 {
            ppn = fields[0].(*marc21.ControlField).Data
        } else {
            if *ignore {
                continue
            } else {
                panic(`No 004. Are you sure this is local data?`)
            }
        }

        fields = record.GetFields("005")
        if len(fields) > 0 {
            date = fields[0].(*marc21.ControlField).Data
        } else {
            if *ignore {
                continue
            } else {
                panic(`No 005. Are you sure this is local data?`)
            }
        }

        var subfields = record.GetSubFields("852", 'a')
        if len(subfields) > 0 {
            // *Note* if more than one sigel is found, take the first only!
            sigel = subfields[0].Value
        } else {
            fmt.Fprintf(os.Stderr, "[EE] No sigel found.\n")
            fmt.Fprintf(os.Stderr, "%s\n", record.String())
            if *strict {
                panic("Sigels required in strict mode")
            } else {
                continue
            }
        }
        fmt.Printf("%s\t%s\t%s\t%s\n", ppn, epn, sigel, date)

    }
    return
}
