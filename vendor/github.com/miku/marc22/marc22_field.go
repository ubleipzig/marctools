package marc22

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"strings"
)

// ControlField represents a control field, which contains only a tag and data.
type ControlField struct {
	XMLName xml.Name `xml:"controlfield"`
	Tag     string   `xml:"tag,attr"`
	Data    string   `xml:",chardata"`
}

// ControlField.String returns the ControlField as a string.
func (cf *ControlField) String() string {
	return fmt.Sprintf("%s %s", cf.Tag, cf.Data)
}

// ControlField.GetTag returns the tag for a ControlField.
func (cf *ControlField) GetTag() string {
	return cf.Tag
}

func read_control(reader io.Reader, dent *dirent) (field *ControlField, err error) {
	data := make([]byte, dent.length)
	n, err := reader.Read(data)
	if err != nil {
		return
	}
	if n != dent.length {
		errs := fmt.Sprintf("MARC21: invalid control entry, expected %d bytes, read %d", dent.length, n)
		err = errors.New(errs)
		return
	}
	if data[dent.length-1] != RS {
		errs := fmt.Sprintf("MARC21: invalid control entry, does not end with a field terminator")
		err = errors.New(errs)
		return
	}
	field = &ControlField{Tag: dent.tag, Data: string(data[:dent.length-1])}
	return
}

// Subfield represents a subfield, containing a single-byte code and
// associated data.
type SubField struct {
	XMLName xml.Name `xml:"subfield"`
	Code    string   `xml:"code,attr"`
	Value   string   `xml:",chardata"`
}

// SubField.String returns the subfield as a string.
func (sf SubField) String() string {
	return fmt.Sprintf("(%s) %s", sf.Code, sf.Value)
}

// DataField represents a variable data field, containing a tag, two
// single-byte indicators, and one or more subfields.
type DataField struct {
	XMLName   xml.Name    `xml:"datafield"`
	Tag       string      `xml:"tag,attr"`
	Ind1      string      `xml:"ind1,attr"`
	Ind2      string      `xml:"ind2,attr"`
	SubFields []*SubField `xml:"subfield"`
}

// DataField.GetTag returns the tag for a DataField.
func (df *DataField) GetTag() string {
	return df.Tag
}

// DataField.String returns the DataField as a string.
func (df DataField) String() string {
	subfields := make([]string, 0, len(df.SubFields))
	for _, sf := range df.SubFields {
		subfields = append(subfields, "["+sf.String()+"]")
	}
	return fmt.Sprintf("%s [%s%s] %s", df.Tag, df.Ind1, df.Ind2,
		strings.Join(subfields, ", "))
}

func read_data(reader io.Reader, dent *dirent) (field *DataField, err error) {
	data := make([]byte, dent.length)
	n, err := reader.Read(data)
	if err != nil {
		return
	}
	if n != dent.length {
		errs := fmt.Sprintf("MARC21: invalid data entry, expected %d bytes, read %d", dent.length, n)
		err = errors.New(errs)
		return
	}
	if data[dent.length-1] != RS {
		errs := fmt.Sprintf("MARC21: invalid data entry, does not end with a field terminator")
		err = errors.New(errs)
		return
	}

	df := &DataField{Tag: dent.tag}
	df.Ind1, df.Ind2 = string(data[0]), string(data[1])

	df.SubFields = make([]*SubField, 0, 1)
	for _, sfbytes := range bytes.Split(data[2:dent.length-1], []byte{DELIM}) {
		if len(sfbytes) == 0 {
			continue
		}
		sf := &SubField{Code: string(sfbytes[0]), Value: string(sfbytes[1:])}
		df.SubFields = append(df.SubFields, sf)
	}

	field = df
	return
}
