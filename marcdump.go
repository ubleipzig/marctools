package main

// https://gitorious.org/marc21-go/marc21
import "./marc21"
import "flag"
import "fmt"
import "io"
import "os"

const app_version = "1.3.0"

func main() {

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

    for {
        record, err := marc21.ReadRecord(fi)
        if err == io.EOF {
            break
        }
        if err != nil {
            panic(err)
        }

        fmt.Printf("%s\n", record.String())
    }
    return
}
