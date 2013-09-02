package main

import "./marc21" // https://gitorious.org/marc21-go/marc21
import "fmt"
import "io"
import "os"

func main() {
    if len(os.Args) != 2 {
        fmt.Printf("Usage: lok2tsv MARCFILE\n")
        os.Exit(1)
    }

    fi, err := os.Open(os.Args[1])
    if err != nil { panic(err) }
    defer func() {
        if err := fi.Close(); err != nil {
            panic(err)
        }
    }()

    for {
        record, err := marc21.ReadRecord(fi)
        if err == io.EOF { break }
        if err != nil { panic(err) }

        var fields []marc21.Field
        var epn, ppn, date, sigel string

        fields = record.GetFields("001")
        if len(fields) == 1 {
            epn = fields[0].(*marc21.ControlField).Data
        } else {
            panic("No 001 found. Are you sure this is local data?")
        }

        fields = record.GetFields("004")
        if len(fields) == 1 {
            ppn = fields[0].(*marc21.ControlField).Data
        } else {
            panic("No 004 found. Are you sure this is local data?")
        }

        fields = record.GetFields("005")
        if len(fields) == 1 {
            date = fields[0].(*marc21.ControlField).Data
        } else {
            panic("No 005 found. Are you sure this is local data?")
        }

        var subfields = record.GetSubFields("852", 'a')
        if len(subfields) == 1 {
            sigel = subfields[0].Value
        } else if len(subfields) == 0 {
            fmt.Fprintf(os.Stderr, "[EE] No sigel found.\n")
            fmt.Fprintf(os.Stderr, "%s\n", record.String())
            // panic("No sigel found.")
            continue
        } else if len(subfields) > 1 {
            fmt.Fprintf(os.Stderr, "[EE] More than one sigel found.\n")
            fmt.Fprintf(os.Stderr, "%s\n", record.String())
            // panic("More than one sigel found.")
            continue
        }

        fmt.Printf("%s\t%s\t%s\t%s\n", ppn, epn, date, sigel)

    }
    return
}
