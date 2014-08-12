package marctools

import (
	"encoding/xml"
)

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
