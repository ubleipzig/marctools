package marc21

import (
	"encoding/xml"
	"errors"
	"io"
	"strings"
)

// ReadRecord returns a single MARC record from a reader.
func ReadRecord(reader io.Reader) (record *Record, err error) {
	record = &Record{}
	record.Fields = make([]Field, 0, 8)

	record.Leader, err = read_leader(reader)
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
		var field Field
		if strings.HasPrefix(dent.tag, "00") {
			if field, err = read_control(reader, dent); err != nil {
				return
			}
		} else {
			if field, err = read_data(reader, dent); err != nil {
				return
			}
		}
		record.Fields = append(record.Fields, field)
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
	XMLName xml.Name `xml:"record"`
	Leader  string   `xml:"leader"`
	Fields  []Field
}

// Record.XML writes a MARCXML representation of the record.
func (record *Record) XML(writer io.Writer) (err error) {
	xmlrec := &RecordXML{Leader: record.Leader.String(), Fields: record.Fields}
	output, err := xml.Marshal(xmlrec)
	writer.Write(output)
	return
}

// Record represents a MARC21 record, consisting of a leader and a number of
// fields.
type Record struct {
	XMLName xml.Name `xml:"record"`
	Leader  *Leader  `xml:"leader"`
	Fields  []Field
}

// Record.String returns the Record as a string.
func (record Record) String() string {
	estrings := make([]string, len(record.Fields))
	for i, entry := range record.Fields {
		estrings[i] = entry.String()
	}
	return strings.Join(estrings, "\n")
}

// Record.GetFields returns a slice of fields that match the given tag.
func (record Record) GetFields(tag string) (fields []Field) {
	fields = make([]Field, 0, 4)
	for _, field := range record.Fields {
		if field.GetTag() == tag {
			fields = append(fields, field)
		}
	}
	return
}

// Record.GetSubFields returns a slice of subfields that match the given tag
// and code.
func (record Record) GetSubFields(tag string, code byte) (subfields []*SubField) {
	subfields = make([]*SubField, 0, 4)
	fields := record.GetFields(tag)
	for _, field := range fields {
		switch data := field.(type) {
		case *DataField:
			for _, subfield := range data.SubFields {
				if subfield.Code == code {
					subfields = append(subfields, subfield)
				}
			}
		}
	}
	return
}
