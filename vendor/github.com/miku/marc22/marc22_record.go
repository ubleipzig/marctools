package marc22

import (
	"encoding/xml"
	"errors"
	"io"
	"strings"
)

// ReadRecord returns a single MARC record from a reader.
func ReadRecord(reader io.Reader) (record *Record, err error) {
	record = &Record{}
	record.ControlFields = make([]ControlField, 0, 8)
	record.DataFields = make([]DataField, 0, 8)

	record.LeaderParsed, err = read_leader(reader)
	if err != nil {
		return
	}
	dents := make([]*dirent, 0, 8)
	for {
		var dent *dirent
		dent, err = read_dirent(reader)
		if err == ERS {
			err = nil
			break
		}
		if err != nil {
			return
		}
		dents = append(dents, dent)
	}

	for _, dent := range dents {
		if strings.HasPrefix(dent.tag, "00") {
			var field *ControlField
			if field, err = read_control(reader, dent); err != nil {
				return
			}
			record.ControlFields = append(record.ControlFields, *field)
		} else {
			var field *DataField
			if field, err = read_data(reader, dent); err != nil {
				return
			}
			record.DataFields = append(record.DataFields, *field)
		}
	}
	rtbuf := make([]byte, 1)
	_, err = reader.Read(rtbuf)
	if err != nil {
		return
	}
	if rtbuf[0] != RT {
		err = errors.New("MARC21: could not read record terminator")
	}
	return
}

// RecordXML represents a MARCXML record, with a root element named 'record'.
type RecordXML struct {
	XMLName       xml.Name       `xml:"record"`
	Leader        string         `xml:"leader"`
	ControlFields []ControlField `xml:"controlfield"`
	DataFields    []DataField    `xml:"datafield"`
}

// Record.XML writes a MARCXML representation of the record.
func (record *Record) XML(writer io.Writer) (err error) {
	xmlrec := &RecordXML{Leader: record.LeaderParsed.String(), ControlFields: record.ControlFields, DataFields: record.DataFields}
	output, err := xml.Marshal(xmlrec)
	writer.Write(output)
	return
}

type Record struct {
	XMLName       xml.Name `xml:"record"`
	Leader        string   `xml:"leader"`
	LeaderParsed  *Leader
	ControlFields []ControlField `xml:"controlfield"`
	DataFields    []DataField    `xml:"datafield"`
}

// Record.String returns the Record as a string.
func (record Record) String() string {
	var estrings []string
	for _, entry := range record.ControlFields {
		estrings = append(estrings, entry.String())
	}
	for _, entry := range record.DataFields {
		estrings = append(estrings, entry.String())
	}
	return strings.Join(estrings, "\n")
}

// Record.GetFields returns a slice of fields that match the given tag.
func (record Record) GetControlFields(tag string) (fields []ControlField) {
	fields = make([]ControlField, 0, 4)
	for _, field := range record.ControlFields {
		if field.GetTag() == tag {
			fields = append(fields, field)
		}
	}
	return
}

// Record.GetFields returns a slice of fields that match the given tag.
func (record Record) GetDataFields(tag string) (fields []DataField) {
	fields = make([]DataField, 0, 4)
	for _, field := range record.DataFields {
		if field.GetTag() == tag {
			fields = append(fields, field)
		}
	}
	return
}

// Record.GetSubFields returns a slice of subfields that match the given tag
// and code.
func (record Record) GetSubFields(tag string, code string) (subfields []*SubField) {
	subfields = make([]*SubField, 0, 4)
	fields := record.GetDataFields(tag)
	for _, field := range fields {
		for _, subfield := range field.SubFields {
			if subfield.Code == code {
				subfields = append(subfields, subfield)
			}
		}
	}
	return
}
