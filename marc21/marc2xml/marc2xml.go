package main

/*
   Go Language MARC21 to XML Converter
   Copyright (C) 2011 William Waites

   This program is free software: you can redistribute it and/or
   modify it under the terms of the GNU General Public License as
   published by the Free Software Foundation, either version 3 of the
   License, or (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
   General Public License for more details.

   You should have received a copy of the GNU General Public License
   and the GNU General Public License along with this program (the
   named GPL3).  If not, see <http://www.gnu.org/licenses/>.
*/

import (
	"git.gitorious.org/marc21-go/marc21.git"
	"compress/gzip"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
)

var gzoutput bool

func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "\n%s\n\n", "Go MARC21 - XML Converter")
		fmt.Fprintf(os.Stderr, "\nCopyright (c) 2011 William Waites\n")
		fmt.Fprintf(os.Stderr, "This program comes with ABSOLUTELY NO WARRANTY\n")
		fmt.Fprintf(os.Stderr, "This is free software, and you are welcome to redistribute it\n")
		fmt.Fprintf(os.Stderr, "under certain conditions. See the GPL and LGPL (version 3 or later)\n")
		fmt.Fprintf(os.Stderr, "for details.\n\n")
		fmt.Fprintf(os.Stderr, "Usage: %s [flags] infile.mrc [outfile.xml]\n\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\n")
	}
	flag.BoolVar(&gzoutput, "z", false, "Gzip Output")
}

func main() {
	flag.Parse()

	nargs := flag.NArg()
	if flag.NArg() == 0 || flag.NArg() > 2 {
		flag.Usage()
		os.Exit(255)
	}

	var infile io.Reader
	var outfile io.Writer

	infile, err := os.Open(flag.Arg(0))
	if err != nil {
		log.Fatal(err)
	}

	if nargs == 2 {
		outfile, err = os.Open(flag.Arg(1))
		if err != nil {
			log.Fatal(err)
		}
	} else {
		outfile = os.Stdout
	}

	if gzoutput {
		outfile, err = gzip.NewWriter(outfile)
		if err != nil {
			log.Fatal(err)
		}
	}

	err = marc2xml(infile, outfile)
	if err != nil {
		log.Fatal(err)
	}

	closer, ok := infile.(io.Closer)
	if ok {
		closer.Close()
	}

	closer, ok = outfile.(io.Closer)
	if ok {
		closer.Close()
	}
}

func marc2xml(reader io.Reader, writer io.Writer) (err error) {
	records := make(chan *marc21.Record)

	go func() {
		defer close(records)
		for {
			record, err := marc21.ReadRecord(reader)
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatal(err)
			}
			records <- record
		}
	}()

	_, err = writer.Write([]byte(`<?xml version="1.0" encoding="utf-8" ?>
<collection xmlns="http://www.loc.gov/MARC21/slim">\n`))
	if err != nil {
		return
	}

	for record := range records {
		err = record.XML(writer)
		if err != nil {
			return
		}
	}

	_, err = writer.Write([]byte("</collection>\n"))
	return
}
