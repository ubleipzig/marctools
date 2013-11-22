// WIP, prototype
// marcxml to es flavoured json
package main

import (
    "encoding/json"
    "encoding/xml"
    "fmt"
    "os"
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

func (record *Record) AsMap() map[string]interface{} {
    dict := make(map[string]interface{})
    dict["leader"] = record.Leader
    for _, field := range record.ControlFields {
        dict[field.Tag] = field.Value
    }
    for _, field := range record.DataFields {
        submap := make(map[string]string)
        for _, subfield := range field.Subfields {
            submap[subfield.Code] = subfield.Value
        }
        dict[field.Tag] = submap
    }
    return dict
}

func main() {

    data := `
<marc:record xmlns:marc="http://www.loc.gov/MARC21/slim">
<marc:leader>     njm a22     2u 4500</marc:leader>
<marc:controlfield tag="001">NML00000001</marc:controlfield>
<marc:controlfield tag="003">DE-Khm1</marc:controlfield>
<marc:controlfield tag="005">20130916115438</marc:controlfield>
<marc:controlfield tag="006">m||||||||h||||||||</marc:controlfield>
<marc:controlfield tag="007">cr nnannnu uuu</marc:controlfield>
<marc:controlfield tag="008">130916s2013</marc:controlfield>
<marc:datafield tag="028" ind1="1" ind2="1">
<marc:subfield code="a">8.220369</marc:subfield>
<marc:subfield code="b">Naxos Digital Services Ltd</marc:subfield>
</marc:datafield>
<marc:datafield tag="035" ind1=" " ind2=" ">
<marc:subfield code="a">(DE-Khm1)NML00000001</marc:subfield>
</marc:datafield>
<marc:datafield tag="040" ind1=" " ind2=" ">
<marc:subfield code="a">DE-Khm1</marc:subfield>
<marc:subfield code="b">ger</marc:subfield>
<marc:subfield code="c">DE-Khm1</marc:subfield>
</marc:datafield>
<marc:datafield tag="041" ind1=" " ind2=" ">
<marc:subfield code="a">ger</marc:subfield>
</marc:datafield>
<marc:datafield tag="100" ind1="1" ind2=" ">
<marc:subfield code="a">Ippolitov-Ivanov, Mikhail Mikhaylovich</marc:subfield>
<marc:subfield code="d">1859-1935</marc:subfield>
</marc:datafield>
<marc:datafield tag="245" ind1="0" ind2="0">
<marc:subfield code="a">IPPOLITOV-IVANOV : Caucasian Sketches</marc:subfield>
<marc:subfield code="h">[elektronische Ressource]</marc:subfield>
</marc:datafield>
<marc:datafield tag="260" ind1=" " ind2=" ">
<marc:subfield code="b">Marco Polo</marc:subfield>
</marc:datafield>
<marc:datafield tag="500" ind1=" " ind2=" ">
<marc:subfield code="a">Streaming audio.</marc:subfield>
</marc:datafield>
<marc:datafield tag="505" ind1="0" ind2="0">
<marc:subfield code="t">Caucasian Sketches, Suite 1, Op. 10: In a Mountain Pass / In the Village / In the Mosque / Procession of the Sardar / Ippolitov-Ivanov, Mikhail Mikhaylovich</marc:subfield>
<marc:subfield code="g">00:25:11 --</marc:subfield>
<marc:subfield code="t">Caucasian Sketches, Suite 2, Op. 42, "Iveria": Introduction / Berceuse / Lesghinka / Caucasian War March / Ippolitov-Ivanov, Mikhail Mikhaylovich</marc:subfield>
<marc:subfield code="g">00:23:33 --</marc:subfield>
</marc:datafield>
<marc:datafield tag="506" ind1=" " ind2=" ">
<marc:subfield code="a">Der Zugang ist auf registrierte Benutzer beschränkt.</marc:subfield>
</marc:datafield>
<marc:datafield tag="511" ind1="0" ind2=" ">
<marc:subfield code="a">Ippolitov-Ivanov, Mikhail Mikhaylovich, Komponist -- Lyndon-Gee, Christopher, Dirigent -- Sydney Symphony Orchestra, Orchester</marc:subfield>
</marc:datafield>
<marc:datafield tag="538" ind1=" " ind2=" ">
<marc:subfield code="a">Systemvoraussetzungen (minimale Voraussetzungen): MS Windows 98SE, 2000, XP with MS IE 6.0 / Mozilla 1.7.1 / FireFox 1PR / Netscape 7.1 / Opera 7.53 and Media Player 9.0 / 10.0.</marc:subfield>
</marc:datafield>
<marc:datafield tag="648" ind1=" " ind2="4">
<marc:subfield code="y">20. Jahrhundert</marc:subfield>
</marc:datafield>
<marc:datafield tag="650" ind1=" " ind2="0">
<marc:subfield code="a">Musik</marc:subfield>
<marc:subfield code="x">Computer- und Netzwerkressourcen.</marc:subfield>
</marc:datafield>
<marc:datafield tag="655" ind1=" " ind2="4">
<marc:subfield code="a">Orchestermusik</marc:subfield>
</marc:datafield>
<marc:datafield tag="700" ind1="1" ind2=" ">
<marc:subfield code="a">Ippolitov-Ivanov, Mikhail Mikhaylovich</marc:subfield>
<marc:subfield code="c">Komponist</marc:subfield>
<marc:subfield code="d">1859-1935</marc:subfield>
</marc:datafield>
<marc:datafield tag="710" ind1="2" ind2=" ">
<marc:subfield code="a">Sydney Symphony Orchestra</marc:subfield>
</marc:datafield>
<marc:datafield tag="773" ind1="0" ind2=" ">
<marc:subfield code="t">Naxos Music Library.</marc:subfield>
</marc:datafield>
<marc:datafield tag="776" ind1="0" ind2=" ">
<marc:subfield code="d">Hong Kong, Naxos Digital Services Ltd.</marc:subfield>
</marc:datafield>
<marc:datafield tag="830" ind1=" " ind2="0">
<marc:subfield code="a"/>
<marc:subfield code="p">Caucasian Sketches, Suite 1, Op. 10 -- Caucasian Sketches, Suite 2, Op. 42, "Iveria"</marc:subfield>
</marc:datafield>
<marc:datafield tag="856" ind1="4" ind2="0">
<marc:subfield code="u">http://univportal.naxosmusiclibrary.com/catalogue/item.asp?cid=8.220369</marc:subfield>
<marc:subfield code="z">Verbindung zu Naxos Music Library</marc:subfield>
</marc:datafield>
<marc:datafield tag="902" ind1=" " ind2=" ">
<marc:subfield code="a">130916</marc:subfield>
</marc:datafield>
</marc:record>    
    `

    v := Record{}
    err := xml.Unmarshal([]byte(data), &v)
    if err != nil {
        fmt.Printf("%v", err)
        os.Exit(1)
    } else {
        // fmt.Println(v.AsMap())
        // fmt.Printf("%v", v)
        b, err := json.Marshal(v.AsMap())
        if err != nil {
            panic(fmt.Sprintf("error: %s", err))
        }
        os.Stdout.Write(b)
        fmt.Println()

    }
}
