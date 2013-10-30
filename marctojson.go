package main

import (
    "./marc21" // https://gitorious.org/marc21-go/marc21
    "flag"
    "fmt"
    "io"
    "os"
    "strings"
    "encoding/json"
)

const app_version = "1.0.0"

func stringToMap(s string) map[string]string {
    result := make(map[string]string)
    for _, pair := range strings.Split(s, ",") {
        fmt.Println(pair)
        kv := strings.Split(pair, "=")
        if len(kv) != 2 {
            panic(fmt.Sprintf("Could not parse key-value parameter: %s", s))
        } else {
            result[kv[0]] = kv[1]
        }
    }
    return result
}

func main() {

    ignore := flag.Bool("i", false, "ignore marc errors (not recommended)")
    version := flag.Bool("v", false, "prints current program version")

    var meta *string = flag.String("m", "", "a key=value pair to pass to meta")

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

    metamap := stringToMap(*meta)

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

        mainmap := make(map[string]interface{})

        marcmap := make(map[string]interface{})
        leadermap := make(map[string]string)

        leader := record.Leader
        leadermap["status"] = string(leader.Status)
        leadermap["cs"] = string(leader.CharacterEncoding)
        leadermap["length"] = fmt.Sprintf("%d", leader.Length)
        leadermap["type"] = string(leader.Type)
        leadermap["impldef"] = string(leader.ImplementationDefined[:5])
        leadermap["ic"] = fmt.Sprintf("%d", leader.IndicatorCount)
        leadermap["lol"] = fmt.Sprintf("%d", leader.LengthOfLength)
        leadermap["losp"] = fmt.Sprintf("%d", leader.LengthOfStartPos)
        leadermap["sfcl"] = fmt.Sprintf("%d", leader.SubfieldCodeLength)
        leadermap["ba"] = fmt.Sprintf("%d", leader.BaseAddress)
        leadermap["raw"] = string(leader.Bytes())

        marcmap["leader"] = leadermap

        for _, field := range record.Fields {
            tag := field.GetTag()
            if strings.HasPrefix(tag, "00") {
                marcmap[tag] = field.(*marc21.ControlField).Data
            } else {
                datafield := field.(*marc21.DataField)
                subfieldmap := make(map[string]string)
                for _, subfield := range datafield.SubFields {
                    subfieldmap[fmt.Sprintf("%c", subfield.Code)] = subfield.Value
                }
                _, present := marcmap[tag]
                if !present {
                    subfields := make([]interface{}, 0)
                    marcmap[tag] = subfields
                }
                marcmap[tag] = append(marcmap[tag].([]interface{}), subfieldmap)
            }
        }

        mainmap["content"] = marcmap
        mainmap["meta"] = metamap
        mainmap["content_type"] = "application/marc"

        b, err := json.Marshal(mainmap)
        if err != nil {
            fmt.Println("error:", err)
        }
        os.Stdout.Write(b)
        fmt.Println()

    }
}
