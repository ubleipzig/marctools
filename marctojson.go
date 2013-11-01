// Turn MARC into JSON.
// The JSON format was choosen with Elasticsearch queries in mind. Also
// the format proposed [in this blog post](http://dilettantes.code4lib.org/blog/2010/09/a-proposal-to-serialize-marc-in-json/)
// had some verbose elements, such as an extra subfields key.
//
// Java version (without original and sha1)
//
// $ time target/marctojson -i ../gomarckit/test-tit.mrc -o ../gomarckit/test-tit.json.3
//
// real    25m34.243s
// user    24m13.468s
// sys      0m17.348s
//
// $ time ./marctojson test-tit.mrc > test-tit.json.1
//
// real    12m32.742s
// user     9m52.720s
// sys      1m47.288s
//
// $ time ./marctojson.py MarcToJSONMerged --filename test-tit.mrc --workers 4
// DEBUG: [SplitMarc.run] 51.48869
// DEBUG: [MarcToJSON.run] 209.79476
// DEBUG: [MarcToJSON.run] 1.84064
// DEBUG: [MarcToJSON.run] 229.05073
// DEBUG: [MarcToJSON.run] 260.67292
// DEBUG: [MarcToJSON.run] 261.16039
// DEBUG: [MarcToJSONMerged.run] 170.03256

// real     8m3.518s
// user    10m0.568s
// sys      2m17.676s

//
// Based on https://gitorious.org/marc21-go/marc21, which is included here.
package main

import (
    "./marc21"
    "encoding/json"
    "flag"
    "fmt"
    "io"
    "os"
    "strings"
)

const app_version = "1.1.0"

// Turn a list of key=value,key=value strings into a map.
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
    filterVar := flag.String("r", "", "only dump the given tags (e.g. 001,003)")

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

        // the final map
        mainMap := make(map[string]interface{})
        // content map
        marcMap := make(map[string]interface{})
        // the expanded leader
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
                subfieldMap := make(map[string]interface{})
                subfieldMap["ind1"] = string(datafield.Ind1)
                subfieldMap["ind2"] = string(datafield.Ind2)
                for _, subfield := range datafield.SubFields {
                    code := fmt.Sprintf("%c", subfield.Code)
                    value, present := subfieldMap[code]
                    if present {
                        switch t := value.(type) {
                        default:
                            panic(fmt.Sprintf("unexpected type: %T", t))
                        case string:
                            values := make([]string, 0)
                            values = append(values, value.(string))
                            values = append(values, subfield.Value)
                            subfieldMap[code] = values
                        case []string:
                            subfieldMap[code] = append(subfieldMap[code].([]string), subfield.Value)
                        }
                    } else {
                        subfieldMap[code] = subfield.Value
                    }
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
