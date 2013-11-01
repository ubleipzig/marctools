package main

import (
    "./marc21" // https://gitorious.org/marc21-go/marc21
    "encoding/json"
    "flag"
    "fmt"
    "io"
    "os"
    "strings"
)

const app_version = "1.0.0"

func stringToMap(s string) map[string]string {
    result := make(map[string]string)
    if len(s) == 0 {
        return result
    }
    for _, pair := range strings.Split(s, ",") {
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

    metaVar := flag.String("m", "", "a key=value pair to pass to meta")
    filterVar := flag.String("r", "", "only dump the given tags (comma separated list)")

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

    metamap := stringToMap(*metaVar)
    filterMap := make(map[string]bool)
    if len(*filterVar) > 0 {
        tags := strings.Split(*filterVar, ",")
        for _, value := range tags {
            filterMap[value] = true
        }
    }
    hasFilter := len(filterMap) > 0

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

        mainMap := make(map[string]interface{})

        marcMap := make(map[string]interface{})
        leaderMap := make(map[string]string)

        leader := record.Leader
        leaderMap["status"] = string(leader.Status)
        leaderMap["cs"] = string(leader.CharacterEncoding)
        leaderMap["length"] = fmt.Sprintf("%d", leader.Length)
        leaderMap["type"] = string(leader.Type)
        leaderMap["impldef"] = string(leader.ImplementationDefined[:5])
        leaderMap["ic"] = fmt.Sprintf("%d", leader.IndicatorCount)
        leaderMap["lol"] = fmt.Sprintf("%d", leader.LengthOfLength)
        leaderMap["losp"] = fmt.Sprintf("%d", leader.LengthOfStartPos)
        leaderMap["sfcl"] = fmt.Sprintf("%d", leader.SubfieldCodeLength)
        leaderMap["ba"] = fmt.Sprintf("%d", leader.BaseAddress)
        leaderMap["raw"] = string(leader.Bytes())

        marcMap["leader"] = leaderMap

        for _, field := range record.Fields {
            tag := field.GetTag()
            if hasFilter {
                _, present := filterMap[tag]
                if !present {
                    continue
                }
            }
            if strings.HasPrefix(tag, "00") {
                marcMap[tag] = field.(*marc21.ControlField).Data
            } else {
                datafield := field.(*marc21.DataField)
                subfieldMap := make(map[string]string)
                for _, subfield := range datafield.SubFields {
                    subfieldMap[fmt.Sprintf("%c", subfield.Code)] = subfield.Value
                }
                _, present := marcMap[tag]
                if !present {
                    subfields := make([]interface{}, 0)
                    marcMap[tag] = subfields
                }
                marcMap[tag] = append(marcMap[tag].([]interface{}), subfieldMap)
            }
        }

        mainMap["content"] = marcMap
        mainMap["meta"] = metamap
        mainMap["content_type"] = "application/marc"

        b, err := json.Marshal(mainMap)
        if err != nil {
            panic(fmt.Sprintf("error: %s", err))
        }
        os.Stdout.Write(b)
        fmt.Println()

    }
}
