// Reimplementation of
//
//     yaz-marcdump -s prefix -C 1000 file.mrc
//
// in Go.
package main

import (
    "errors"
    "flag"
    "fmt"
    "io"
    "os"
    "path/filepath"
    "strconv"
)

const app_version = "1.3.3"

func record_length(reader io.Reader) (length int64, err error) {
    data := make([]byte, 24)
    n, err := reader.Read(data)
    if err != nil {
        return
    }
    if n != 24 {
        errs := fmt.Sprintf("MARC21: invalid leader: expected 24 bytes, read %d", n)
        err = errors.New(errs)
        return
    }
    _length, err := strconv.Atoi(string(data[0:5]))
    if err != nil {
        errs := fmt.Sprintf("MARC21: invalid record length: %s", err)
        err = errors.New(errs)
        return
    }
    return int64(_length), err
}

func main() {

    directory := flag.String("d", ".", "directory to write to")
    prefix := flag.String("s", "split", "split file prefix")
    size := flag.Int64("C", 1, "number of records per file")
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

    fi, err := os.Stat(*directory)
    if os.IsNotExist(err) {
        fmt.Printf("no such file or directory: %s\n", *directory)
        os.Exit(1)
    }
    if !fi.IsDir() {
        fmt.Printf("arg to -d must be directory: %s\n", *directory)
        os.Exit(1)
    }

    handle, err := os.Open(flag.Args()[0])
    if err != nil {
        fmt.Printf("%s\n", err)
        os.Exit(1)
    }

    defer func() {
        if err := handle.Close(); err != nil {
            panic(err)
        }
    }()

    var i, length, cumulative, offset, batch, fileno int64

    for {
        length, err = record_length(handle)
        i += 1
        if err == io.EOF {
            break
        }
        if err != nil {
            panic(err)
        }
        if i%*size == 0 {
            if i > 0 {

                filename := filepath.Join(*directory, fmt.Sprintf("%s-%08d", *prefix, fileno))
                output, err := os.Create(filename)
                if err != nil {
                    panic(err)
                }

                buffer := make([]byte, batch)

                handle.Seek(offset, 0)
                handle.Read(buffer)
                output.Write(buffer)
                err = output.Close()
                if err != nil {
                    panic(err)
                }

                batch = 0
                fileno += 1
                offset = cumulative
            }
        }
        cumulative += length
        batch += length
        handle.Seek(int64(cumulative), 0)
    }

    filename := fmt.Sprintf("%s-%08d", prefix, fileno)
    output, err := os.Create(filename)
    if err != nil {
        panic(err)
    }

    buffer := make([]byte, batch)

    handle.Seek(offset, 0)
    handle.Read(buffer)
    output.Write(buffer)
    err = output.Close()
    if err != nil {
        panic(err)
    }
}
