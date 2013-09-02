PACKAGE

package marc21
    import "gitorious.org/marc21-go/marc21.git"

    Package marc21 reads and writes MARC21 bibliographic catalogue records.

    Usage is straightforward. For example,

	marcfile, err := os.Open("somedata.mrc")
	record, err := marc21.ReadRecord(marcfile)
	err = record.XML(os.Stdout)

CONSTANTS

const DELIM = 0x1F
    Subfield delimiter.
const RS = 0x1E
    Record separator.
const RT = 0x1D
    Record terminator.


VARIABLES

var ERS = errors.New("Record Separator (field terminator)")


TYPES

type ControlField struct {
    XMLName xml.Name `xml:"controlfield"`
    Tag     string   `xml:"tag,attr"`
    Data    string   `xml:",chardata"`
}
    ControlField represents a control field, which contains only a tag and
    data.

func (cf *ControlField) GetTag() string
    ControlField.GetTag returns the tag for a ControlField.

func (cf *ControlField) String() string
    ControlField.String returns the ControlField as a string.

type DataField struct {
    XMLName   xml.Name `xml:"datafield"`
    Tag       string   `xml:"tag,attr"`
    Ind1      byte     `xml:"ind1,attr"`
    Ind2      byte     `xml:"ind2,attr"`
    SubFields []*SubField
}
    DataField represents a variable data field, containing a tag, two
    single-byte indicators, and one or more subfields.

func (df *DataField) GetTag() string
    DataField.GetTag returns the tag for a DataField.

func (df *DataField) String() string
    DataField.String returns the DataField as a string.

type Field interface {
    String() string
    GetTag() string
}
    Field defines an interface that is satisfied by the Control and Data
    field types.

type Leader struct {
    Length                             int
    Status, Type                       byte
    ImplementationDefined              [5]byte
    CharacterEncoding                  byte
    BaseAddress                        int
    IndicatorCount, SubfieldCodeLength int
    LengthOfLength, LengthOfStartPos   int
}
    Leader represents the record leader, containing structural data about
    the MARC record.

func (leader Leader) Bytes() (buf []byte)
    Leader.Bytes() returns the leader as a slice of 24 bytes.

func (leader Leader) String() string
    Leader.String() returns the leader as a string.

type Record struct {
    XMLName xml.Name `xml:"record"`
    Leader  *Leader  `xml:"leader"`
    Fields  []Field
}
    Record represents a MARC21 record, consisting of a leader and a number
    of fields.

func ReadRecord(reader io.Reader) (record *Record, err error)
    ReadRecord returns a single MARC record from a reader.

func (record Record) GetFields(tag string) (fields []Field)
    Record.GetFields returns a slice of fields that match the given tag.

func (record Record) GetSubFields(tag string, code byte) (subfields []*SubField)
    Record.GetSubFields returns a slice of subfields that match the given
    tag and code.

func (record Record) String() string
    Record.String returns the Record as a string.

func (record *Record) XML(writer io.Writer) (err error)
    Record.XML writes a MARCXML representation of the record.

type RecordXML struct {
    XMLName xml.Name `xml:"record"`
    Leader  string   `xml:"leader"`
    Fields  []Field
}
    RecordXML represents a MARCXML record, with a root element named
    'record'.

type SubField struct {
    XMLName xml.Name `xml:"subfield"`
    Code    byte     `xml:"code,attr"`
    Value   string   `xml:",chardata"`
}
    Subfield represents a subfield, containing a single-byte code and
    associated data.

func (sf SubField) String() string
    SubField.String returns the subfield as a string.


SUBDIRECTORIES

	marc2xml

