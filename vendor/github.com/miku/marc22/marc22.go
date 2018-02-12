// Copyright (C) 2011 William Waites
// Copyright (C) 2012 Dan Scott <dan@coffeecode.net>

// This program is free software: you can redistribute it and/or
// modify it under the terms of the GNU Lesser General Public License
// as published by the Free Software Foundation, either version 3 of
// the License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public
// License and the GNU General Public License along with this program
// (the files COPYING and GPL3 respectively).  If not, see
// <http://www.gnu.org/licenses/>.

/*
Package marc22 reads and writes MARC21 bibliographic catalogue records.

This package is an experimental fork of https://github.com/miku/marc21.
The main difference for now is that marc22 will be able to read MARC from
MARCXML, whereas marc21 can currently only write XML.

Usage is straightforward. Example on the playground: http://play.golang.org/p/mnn6UggTHa.

	marcfile, err := os.Open("somedata.mrc")
	record, err := marc22.ReadRecord(marcfile)
	err = record.XML(os.Stdout)
*/
package marc22

import (
	"errors"
	"fmt"
	"io"
	"strconv"
)

// Leader represents the record leader, containing structural data about the
// MARC record.
type Leader struct {
	Length                             int
	Status, Type                       byte
	ImplementationDefined              [5]byte
	CharacterEncoding                  byte
	BaseAddress                        int
	IndicatorCount, SubfieldCodeLength int
	LengthOfLength, LengthOfStartPos   int
}

// Leader.Bytes() returns the leader as a slice of 24 bytes.
func (leader Leader) Bytes() (buf []byte) {
	buf = make([]byte, 24)
	copy(buf[0:5], []byte(fmt.Sprintf("%05d", leader.Length)))
	buf[5] = leader.Status
	buf[6] = leader.Type
	copy(buf[7:9], leader.ImplementationDefined[0:2])
	buf[9] = leader.CharacterEncoding
	copy(buf[10:11], fmt.Sprintf("%d", leader.IndicatorCount))
	copy(buf[11:12], fmt.Sprintf("%d", leader.SubfieldCodeLength))
	copy(buf[12:17], fmt.Sprintf("%05d", leader.BaseAddress))
	copy(buf[17:20], leader.ImplementationDefined[2:5])
	copy(buf[20:21], fmt.Sprintf("%d", leader.LengthOfLength))
	copy(buf[21:22], fmt.Sprintf("%d", leader.LengthOfStartPos))
	buf[22] = '0'
	buf[23] = '0'
	return
}

// Leader.String() returns the leader as a string.
func (leader Leader) String() string {
	return string(leader.Bytes())
}

func read_leader(reader io.Reader) (leader *Leader, err error) {
	data := make([]byte, 24)
	n, err := reader.Read(data)
	if err != nil {
		return
	}
	if n != 24 {
		errs := fmt.Sprintf("MARC21: invalid leader: expected 24 bytes, read %d", n)
		err = errors.New(errs)
		return
	}
	leader = &Leader{}
	leader.Length, err = strconv.Atoi(string(data[0:5]))
	if err != nil {
		errs := fmt.Sprintf("MARC21: invalid record length: %s", err)
		err = errors.New(errs)
		return
	}
	leader.Status = data[5]
	leader.Type = data[6]
	copy(leader.ImplementationDefined[0:2], data[7:9])
	leader.CharacterEncoding = data[9]

	leader.IndicatorCount, err = strconv.Atoi(string(data[10:11]))
	if err != nil || leader.IndicatorCount != 2 {
		errs := fmt.Sprintf("MARC21: erroneous indicator count, expected '2', got %u", data[10])
		err = errors.New(errs)
		return
	}
	leader.SubfieldCodeLength, err = strconv.Atoi(string(data[11:12]))
	if err != nil || leader.SubfieldCodeLength != 2 {
		errs := fmt.Sprintf("MARC21: erroneous subfield code length, expected '2', got %u", data[11])
		err = errors.New(errs)
		return
	}

	leader.BaseAddress, err = strconv.Atoi(string(data[12:17]))

	if err != nil {
		errs := fmt.Sprintf("MARC21: invalid base address: %s", err)
		err = errors.New(errs)
		return
	}

	copy(leader.ImplementationDefined[2:5], data[17:20])

	leader.LengthOfLength, err = strconv.Atoi(string(data[20:21]))
	if err != nil || leader.LengthOfLength != 4 {
		errs := fmt.Sprintf("MARC21: invalid length of length, expected '4', got %u", data[20])
		err = errors.New(errs)
		return
	}
	leader.LengthOfStartPos, err = strconv.Atoi(string(data[21:22]))
	if err != nil || leader.LengthOfStartPos != 5 {
		errs := fmt.Sprintf("MARC21: invalid length of starting character position, expected '5', got %u", data[21])
		err = errors.New(errs)
		return
	}
	return
}

type dirent struct {
	tag          string
	length       int
	startCharPos int
}

// Record terminator.
const RT = 0x1D

// Record separator.
const RS = 0x1E

// Subfield delimiter.
const DELIM = 0x1F

var ERS = errors.New("Record Separator (field terminator)")

func read_dirent(reader io.Reader) (dent *dirent, err error) {
	data := make([]byte, 12)
	_, err = reader.Read(data[0:1])
	if err != nil {
		return
	}
	if data[0] == RS {
		err = ERS
		return
	}
	n, err := reader.Read(data[1:])
	if err != nil {
		return
	}
	if n != 11 {
		errs := fmt.Sprintf("MARC21: invalid directory entry, expected 12 bytes, got %d", n)
		err = errors.New(errs)
		return
	}
	dent = &dirent{}
	dent.tag = string(data[0:3])
	dent.length, err = strconv.Atoi(string(data[3:7]))
	if err != nil {
		return
	}
	dent.startCharPos, err = strconv.Atoi(string(data[7:12]))
	if err != nil {
		return
	}

	return
}
