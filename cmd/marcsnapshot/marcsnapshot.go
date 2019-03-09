// Keep the newest records among multiple versions in a set of files
package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"runtime/pprof"
	"strconv"
	"strings"

	"github.com/ubleipzig/marctools"
)

type seekInfo struct {
	Offset int
	Length int
}

func lineCounter(r io.Reader) (int, error) {
	buf := make([]byte, 32784)
	count := 0
	lineSep := []byte{'\n'}

	for {
		c, err := r.Read(buf)
		if err != nil && err != io.EOF {
			return count, err
		}

		count += bytes.Count(buf[:c], lineSep)

		if err == io.EOF {
			break
		}
	}
	return count, nil
}

// fileMapper returns the filename of the map file for a given filename
func fileMapper(filename string) string {
	file, err := ioutil.TempFile("", "marcsnapshot-")
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Fatal(err)
		}
	}()
	marctools.MarcMap(filename, file, false)
	return file.Name()
}

func main() {
	version := flag.Bool("v", false, "prints current program version")
	outputFilename := flag.String("o", "", "output filename")
	length := flag.Int("l", 9, "prefix length to use for comparison")
	cpuprofile := flag.String("cpuprofile", "", "write cpu profile to file")
	force := flag.Bool("f", false, "overwrite existing files")
	verbose := flag.Bool("verbose", false, "be verbose")

	var PrintUsage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS] MARCFILE [MARCFILE, ...]\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	if *version {
		fmt.Println(marctools.AppVersion)
		os.Exit(0)
	}

	if flag.NArg() < 1 {
		PrintUsage()
		os.Exit(1)
	}

	// ensure yaz
	yaz, err := exec.LookPath("yaz-marcdump")
	if err != nil {
		log.Fatal(err)
	}

	// ensure awk
	awk, err := exec.LookPath("awk")
	if err != nil {
		log.Fatal(err)
	}

	for _, name := range []string{"cut", "sort", "tac", "uniq"} {
		_, err = exec.LookPath(name)
		if err != nil {
			log.Fatal(err)
		}
	}

	filenames := flag.Args()

	// a single output file for all (001, 005, filename) tuples
	file, err := ioutil.TempFile("", "marcsnapshot-")
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		err := os.Remove(file.Name())
		if err != nil {
			log.Println(err)
		}
	}()

	mapfiles := make(map[string]string)

	for _, fn := range filenames {
		if *verbose {
			log.Printf("extracting record map and timestamps from %s\n", fn)
		}
		mapfile := fileMapper(fn)
		defer func() {
			err := os.Remove(mapfile)
			if err != nil {
				log.Println(err)
			}
		}()
		mapfiles[fn] = mapfile

		// timestamps
		s := fmt.Sprintf(`%s '%s' | %s ' /^001 / {printf $2"\t"}; /^005 / {print $2"\t%s"}'`, yaz, fn, awk, fn)
		cmd := exec.Command("bash", "-c", s)
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			log.Fatal(err)
		}
		if err := cmd.Start(); err != nil {
			log.Fatal(err)
		}
		io.Copy(file, stdout)
	}

	err = file.Sync()
	if err != nil {
		log.Fatal(err)
	}

	// total number of records
	file.Seek(0, os.SEEK_SET)
	total, err := lineCounter(file)
	if err != nil {
		log.Fatal(err)
	}

	// write a sorted file, that cotains the (id, filename) of the latest records
	sfile, err := ioutil.TempFile("", "marcsnapshot-")
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		err := os.Remove(sfile.Name())
		if err != nil {
			log.Println(err)
		}
	}()

	if *verbose {
		log.Printf("sort and filter %s\n", sfile.Name())
	}
	s := fmt.Sprintf(`LANG=C sort -k1,2 %s | tac | uniq -w %d | cut -f 1,3`, file.Name(), *length)
	cmd := exec.Command("bash", "-c", s)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}
	io.Copy(sfile, stdout)

	// filtered number of records
	sfile.Seek(0, os.SEEK_SET)
	filtered, err := lineCounter(sfile)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("%d %d (%0.2f%%)\n", total, filtered, (100/float64(total))*float64(filtered))

	if *verbose {
		log.Println("gathering ids...")
	}
	// for each filename, keep a list of IDs to keep
	idmap := make(map[string][]string)
	for k := range mapfiles {
		idmap[k] = make([]string, 0)
	}

	ff, err := os.Open(sfile.Name())
	if err != nil {
		log.Fatal(err)
	}
	f := bufio.NewReader(ff)
	for {
		line, err := f.ReadString('\n')
		if err != nil && err != io.EOF {
			log.Fatal(err)
		}
		if line == "" {
			break
		}
		fields := strings.Fields(line)
		if len(fields) != 2 {
			log.Fatalf("invalid map, expected (id, path), got: %s", line)
		}
		idmap[fields[1]] = append(idmap[fields[1]], fields[0])
	}

	if *verbose {
		log.Println("building lookup table...")
	}
	// keep the seek information for all files in memory
	table := make(map[string]map[string]seekInfo)

	for k, v := range mapfiles {
		ff, err := os.Open(v)
		if err != nil {
			log.Fatal(err)
		}
		f := bufio.NewReader(ff)
		for {
			line, err := f.ReadString('\n')
			if err != nil && err != io.EOF {
				log.Fatal(err)
			}
			if line == "" {
				break
			}
			fields := strings.Fields(line)
			if len(fields) != 3 {
				log.Fatalf("invalid map, expected (id, offset, length), got: %s", line)
			}
			offset, err := strconv.Atoi(fields[1])
			if err != nil {
				log.Fatalf("invalid offset in %s", line)
			}
			length, err := strconv.Atoi(fields[2])
			if err != nil {
				log.Fatalf("invalid length in %s", line)
			}
			_, ok := table[k]
			if !ok {
				table[k] = make(map[string]seekInfo)
			}
			table[k][fields[0]] = seekInfo{Offset: offset, Length: length}
		}
		ff.Close()
	}

	// final filtered output
	var output *bufio.Writer
	defer func() {
		if err := output.Flush(); err != nil {
			log.Fatal(err)
		}
	}()

	if *outputFilename == "" {
		output = bufio.NewWriter(os.Stdout)
	} else {
		if !*force {
			if _, err := os.Stat(*outputFilename); err == nil {
				log.Fatal("file exists, will not overwrite")
			}
		}
		out, err := os.Create(*outputFilename)
		if err != nil {
			log.Fatal(err)
		}
		output = bufio.NewWriter(out)
	}

	for k, ids := range idmap {
		if *verbose {
			log.Printf("extracting %d records from %s\n", len(ids), k)
		}
		ff, err := os.Open(k)
		if err != nil {
			log.Fatal(err)
		}
		defer func() {
			if err := ff.Close(); err != nil {
				log.Fatal(err)
			}
		}()
		for _, id := range ids {
			seekinfo := table[k][id]
			_, err := ff.Seek(int64(seekinfo.Offset), os.SEEK_SET)
			if err != nil {
				log.Fatal(err)
			}
			_, err = io.CopyN(output, ff, int64(seekinfo.Length))
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}
