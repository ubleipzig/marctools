// marcxml to es flavoured json
// Example XML

// <?xml version="1.0" encoding="utf-8"?>
// <marc:collection xmlns:marc="http://www.loc.gov/MARC21/slim">
// <marc:record xmlns:marc="http://www.loc.gov/MARC21/slim">
// <marc:leader>     njm a22     2u 4500</marc:leader>
// <marc:controlfield tag="001">NML00000001</marc:controlfield>
// <marc:controlfield tag="003">DE-Khm1</marc:controlfield>
// <marc:controlfield tag="005">20130916115438</marc:controlfield>
// <marc:controlfield tag="006">m||||||||h||||||||</marc:controlfield>
// <marc:controlfield tag="007">cr nnannnu uuu</marc:controlfield>
// <marc:controlfield tag="008">130916s2013</marc:controlfield>
// <marc:datafield tag="028" ind1="1" ind2="1">
// <marc:subfield code="a">8.220369</marc:subfield>
// <marc:subfield code="b">Naxos Digital Services Ltd</marc:subfield>
// </marc:datafield>
// <marc:datafield tag="035" ind1=" " ind2=" ">
// <marc:subfield code="a">(DE-Khm1)NML00000001</marc:subfield>
// </marc:datafield>
// <marc:datafield tag="040" ind1=" " ind2=" ">
// <marc:subfield code="a">DE-Khm1</marc:subfield>
// <marc:subfield code="b">ger</marc:subfield>
// <marc:subfield code="c">DE-Khm1</marc:subfield>
// </marc:datafield>
// </marc:record>
// ...
// </marc:collection>

package main

import (
    "encoding/json"
    "encoding/xml"
    "flag"
    "fmt"
    "io/ioutil"
    "os"
    "strings"
)

const app_version = "1.0"

type ControlField struct {
    Value string `xml:",chardata"`
    Tag   string `xml:"tag,attr"`
}

type Subfield struct {
    Value string `xml:",chardata"`
    Code  string `xml:"code,attr"`
}

type DataField struct {
    Tag       string     `xml:"tag,attr"`
    Ind1      string     `xml:"ind1,attr"`
    Ind2      string     `xml:"ind2,attr"`
    Subfields []Subfield `xml:"http://www.loc.gov/MARC21/slim subfield"`
}

type Record struct {
    Leader        string         `xml:"http://www.loc.gov/MARC21/slim leader"`
    ControlFields []ControlField `xml:"http://www.loc.gov/MARC21/slim controlfield"`
    DataFields    []DataField    `xml:"http://www.loc.gov/MARC21/slim datafield"`
}

type Collection struct {
    XMLName xml.Name `xml:"http://www.loc.gov/MARC21/slim collection"`
    Records []Record `xml:"http://www.loc.gov/MARC21/slim record"`
}

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

// Convert a single unmarshalled record struct into a map.
func (record *Record) ToMap() map[string]interface{} {
    dict := make(map[string]interface{})
    dict["leader"] = record.Leader
    for _, field := range record.ControlFields {
        dict[field.Tag] = field.Value
    }
    for _, field := range record.DataFields {
        submap := make(map[string]interface{})
        submap["ind1"] = field.Ind1
        submap["ind2"] = field.Ind2
        for _, subfield := range field.Subfields {
            value, present := submap[subfield.Code]
            if present {
                switch t := value.(type) {
                default:
                    panic(fmt.Sprintf("unexpected type: %T", t))
                case string:
                    values := make([]string, 0)
                    values = append(values, value.(string))
                    values = append(values, subfield.Value)
                    submap[subfield.Code] = values
                case []string:
                    submap[subfield.Code] = append(submap[subfield.Code].([]string), subfield.Value)
                }
            } else {
                submap[subfield.Code] = subfield.Value
            }
        }
        _, present := dict[field.Tag]
        if !present {
            subfields := make([]interface{}, 0)
            dict[field.Tag] = subfields
        }
        dict[field.Tag] = append(dict[field.Tag].([]interface{}), submap)
    }
    return dict
}

func main() {

    plainVar := flag.Bool("p", false, "plain mode: dump without content and meta")
    metaVar := flag.String("m", "", "a key=value pair to pass to meta")

    var PrintUsage = func() {
        fmt.Fprintf(os.Stderr, "Usage: %s MARCFILE\n", os.Args[0])
        flag.PrintDefaults()
    }

    flag.Parse()

    if flag.NArg() < 1 {
        PrintUsage()
        os.Exit(1)
    }

    handle, err := os.Open(flag.Args()[0])
    if err != nil {
        fmt.Println("Error opening file:", err)
        os.Exit(1)
    }

    defer func() {
        if err := handle.Close(); err != nil {
            panic(err)
        }
    }()

    b, err := ioutil.ReadAll(handle)
    if err != nil {
        fmt.Println("Error reading file:", err)
        os.Exit(1)
    }

    metamap := stringToMap(*metaVar)

    collection := Collection{}
    err = xml.Unmarshal(b, &collection)

    for _, record := range collection.Records {
        recordMap := record.ToMap()
        result := recordMap
        if !*plainVar {
            result = make(map[string]interface{})
            result["content"] = recordMap
            result["meta"] = metamap
        }
        b, err = json.Marshal(result)
        if err != nil {
            panic(fmt.Sprintf("Conversion error: %s", err))
        }
        os.Stdout.Write(b)
        fmt.Println()
    }
}
