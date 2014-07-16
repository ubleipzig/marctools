----

Note: This project is currently being overhauled: getting more standards-compliant and possibly faster as well.

----


gomarckit
=========


Various MARC command line utilities, with an eye on performance. 

[![gh release](http://img.shields.io/github/release/miku/gomarckit.svg?style=flat)](https://github.com/miku/gomarckit/releases)



Included:

* [`marccount`](https://github.com/miku/gomarckit#marccount)
* [`marcdump`](https://github.com/miku/gomarckit#marcdump)
* [`marcmap`](https://github.com/miku/gomarckit#marcmap)
* [`marcsplit`](https://github.com/miku/gomarckit#marcsplit)
* [`marctojson`](https://github.com/miku/gomarckit#marctojson)
* [`marctotsv`](https://github.com/miku/gomarckit#marctotsv)
* [`marcuniq`](https://github.com/miku/gomarckit#marcuniq)
* [`marcxmltojson`](https://github.com/miku/gomarckit#marcxmltojson)

Based on [marc21](https://gitorious.org/marc21-go/marc21) by
[Dan Scott](https://gitorious.org/~dbs) and William Waites.

Build
-----

You'll [need a Go installation](http://golang.org/doc/install).

Go library dependencies:

* [github.com/mattn/go-sqlite3](https://github.com/mattn/go-sqlite3)

Build:

    $ git clone git@github.com:miku/gomarckit.git
    $ cd gomarckit
    $ go get github.com/mattn/go-sqlite3
    $ make


Symlink the executables:

    $ make install-home


Clean everything:

    $ make clean


marcdump
--------

Just like `yaz-marcdump`. The Go version is about 4-5x slower than the
[C version](http://git.indexdata.com/?p=yaz.git;a=blob;f=util/marcdump.c;h=f92204e386431f044f06dddd8baa1c9db08d69c9;hb=HEAD).


    $ ./marcdump test.mrc
    ...
    001 003915646
    003 DE-576
    004 000577812
    005 19951013000000
    008 940607||||||||||||||||ger|||||||
    689 [  ] [(A) g], [(0) 235869880], [(a) Französisch], [(x) 00]
    689 [  ] [(A) s], [(0) 235840734], [(a) Syntax], [(x) 01]
    852 [  ] [(a) DE-Ch1]
    852 [ 1] [(c) ID 5150 boe], [(9) 00]
    936 [ln] [(0) 221790136], [(a) ID 5150]
    ...


marctotsv
---------

    $ marctotsv -h
    Usage: ./marctotsv [OPTIONS] MARCFILE TAG [TAG, ...]
      -f="<NULL>": fill missing values with this
      -i=false: ignore marc errors (not recommended)
      -k=false: skip lines with missing values
      -s="": separator to use for multiple values
      -v=false: prints current program version



Convert MARC21 to TSV (tab-separated values). By default, if a given tag or subfield has multiple values, just the first one is considered. Most useful for control fields or non-repeatable fields with a single subfield.

For repeated fields, use the `-s` flag to specify a value delimiter.

Examples:

    $ ./marctotsv test.mrc 001
    ...
    111859182
    111862493
    111874173
    111879078
    ...

Empty tag values get a default fill value `<NULL>`:

    $ ./marctotsv test.mrc 001 004 005 852.a
    ...
    121187764   01635253X   20040324000000  DE-105
    38541028X   087701561   20010420000000  DE-540
    385410298   087701561   20120910090128  <NULL>
    385411057   087701723   20120910090145  DE-540
    ...

Use a custom *fill NA* tag with `-f`:

    $ ./marctotsv test.mrc -f UNDEF 001 004 005 852.a
    ...
    385410271   087701553   20101125121554  DE-15
    38541028X   087701561   20010420000000  DE-540
    385410298   087701561   20120910090128  UNDEF
    385411057   087701723   20120910090145  DE-540
    ...

Or skip non-complete row entirely with `-k` (hard to see):

    $ ./marctotsv test.mrc -k 001 004 005 852.a
    ...
    121187764   01635253X   20040324000000  DE-105
    121187772   01635253X   20040324000000  DE-105
    38541028X   087701561   20010420000000  DE-540
    385411057   087701723   20120910090145  DE-540
    ...

Access leader with special tags:

    $ ./marctotsv test.mrc -k 001 @Status
    9780415681889   c
    9780415839792   n
    9780415773874   c

    $ ./marctotsv test.mrc 001 @Length @Status 300.a
    9781420069235   5000    c   712 p. :
    9780415458931   6769    c   424 p. :
    9781841846804   3983    c   444 p. :
    ...


Available special tags to access the leader:

* @Length
* @Status
* @Type
* @ImplementationDefined
* @CharacterEncoding
* @BaseAddress
* @IndicatorCount
* @SubfieldCodeLength
* @LengthOfLength
* @LengthOfStartPos


Literals can be passed along (as long as they do not look like tags):


    $ ./marctotsv test.mrc 001 @Status Hello "x y z"
    9781420086454   n   Hello   x y z
    9781420086478   n   Hello   x y z
    9781420094237   c   Hello   x y z
    9780849397776   c   Hello   x y z
    ...


To convert repeated subfield values, you can specifiy a separator string. If
a separator string is present, multiple subfield values will be joined using
that separator:


    $ ./marctotsv -f NA -s "|" test.mrc 001 773.w 800.w 810.w 811.w 830.w
    ...
    007717717 (DE-576)007717709 NA  (DE-576)005674956 NA  NA
    007717725 (DE-576)007717709 NA  (DE-576)005674956 NA  NA
    007717997 NA  NA  NA  NA  (DE-576)008007659|(DE-576)014490781
    ...



marctojson
----------

Similar output to [marctojson](https://github.com/miku/marctojson) (Java version).
The Go version is a bit more lightweight and faster.

----

Note: The JSON format is tailored for a project, which uses [Elasticsearch](http://www.elasticsearch.org/). The idea was to make queries as convenient as possible.
Here a some example queries:

    $ curl localhost:9200/_search?q=content.001:007717717
    $ curl localhost:9200/_search?q=content.020.a:0-684-84328-5

----

Performance data points:

* Extracting 5 fields from 4007756 records from a 4.3G file takes about
  8 minutes, so about 8349 records per second. That's about four times
  faster than the Java version.

* A conversion to a 7 fields plus leader of a 4.3G file in *plain* mode
  takes about 7m27.258s.

  Command:

      $ time marctojson -p -l -r 001,020,100,245,260,700,776 test.mrc > test.json

      real  7m27.258s
      user  5m24.932s
      sys   1m38.876s

  The resulting JSON file is about 2.1G in size. A similar conversion to TSV

      $ time marctotsv -f NA -s "|" test-tit.mrc 001 020.a 020.z 020.9 100.a \
        245.a 245.b 245.c 260.a 260.b 260.c 700.a 776.z > test.tsv

      real  5m39.058s
      user  4m10.628s
      sys   1m29.868s

  takes about 5 minutes with a output file size of about 700M.

* As comparison, the baseline iteration, which only creates the MARC data structures takes about 4 minutes
  in Go, which amounts to about 17425 records per seconds.

* A simple `yaz-marcdump -np` seems to iterate over the same 4007756 records
  in about 30 seconds (133591 records per second) and a `dev/nulled`
  iteration about 65 seconds (61657 records per second). So C is still three three to four times faster.

The upside of the Go version is, that it is quite short, it is about 126 LOC (+ 826 for entiry library),
the Java version has around 600 (with a bit more functionality, but + 11061 for the library)
and the [C version](https://github.com/nla/yaz/blob/dd8b34fa4d19c1bac5e2a8d4e3446c09f0a8cf69/src/marc_read_line.c)
contains about 260 LOC (+ 41938 for the yaz source). (Note: I know this is a skewed
comparison, it's only here for framing).

Example usage:

    $ marctojson
    Usage: marctojson [OPTIONS] MARCFILE
      -i=false: ignore marc errors (not recommended)
      -l=false: dump the leader as well
      -m="": a key=value pair(s) to pass to meta
      -p=false: plain mode: dump without content and meta
      -r="": only dump the given tags (e.g. 001,003)
      -v=false: prints current program version and exit


    $ ./marctojson -r 001,260 test-tit.mrc|head -1|json_pp
    {
       "content_type" : "application/marc",
       "content" : {
          "leader" : {
             "sfcl" : "2",
             "ba" : "265",
             "status" : "c",
             "lol" : "4",
             "impldef" : "m   a",
             "length" : "1078",
             "cs" : "a",
             "type" : "a",
             "losp" : "5",
             "ic" : "2",
             "raw" : "01078cam a2200265  a4500"
          },
          "260" : [
             {
                "c" : "18XX",
                "a" : "Karlsruhe :",
                "b" : "Groos,"
             }
          ],
          "001" : "00002144X"
       },
       "meta" : {}
    }


    $ ./marctojson -r 001,005,260 -m date=`date +"%Y-%-m-%d"` test-tit.mrc | head -1 | json_pp
    {
       "content_type" : "application/marc",
       "content" : {
          "leader" : {
             "sfcl" : "2",
             "ba" : "265",
             "status" : "c",
             "lol" : "4",
             "impldef" : "m   a",
             "length" : "1078",
             "cs" : "a",
             "type" : "a",
             "losp" : "5",
             "ic" : "2",
             "raw" : "01078cam a2200265  a4500"
          },
          "260" : [
             {
                "c" : "18XX",
                "a" : "Karlsruhe :",
                "b" : "Groos,"
             }
          ],
          "005" : "20120312115715.0",
          "001" : "00002144X"
       },
       "meta" : {
          "date" : "2013-10-30"
       }
    }


marcxmltojson
-------------


Convert MARC XML to JSON.

    $ marcxmltojson
    Usage: marcxmltojson MARCFILE
      -m="": a key=value pair to pass to meta
      -p=false: plain mode: dump without content and meta
      -v=false: prints current program version and exit

Example input (snippet):

    <?xml version="1.0" encoding="utf-8"?>
    <marc:collection xmlns:marc="http://www.loc.gov/MARC21/slim">
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
    ...
    <marc:datafield tag="830" ind1=" " ind2="0">
    <marc:subfield code="a"/>
    <marc:subfield code="p">Caucasian Sketches, Suite 1, Op. 10 -- Caucasian
    Sketches, Suite 2, Op. 42, "Iveria"
    </marc:subfield>
    </marc:datafield>
    <marc:datafield tag="856" ind1="4" ind2="0">
    <marc:subfield code="u">
    http://univportal.naxosmusiclibrary.com/catalogue/item.asp?cid=8.220369
    </marc:subfield>
    <marc:subfield code="z">Verbindung zu Naxos Music Library</marc:subfield>
    </marc:datafield>
    <marc:datafield tag="902" ind1=" " ind2=" ">
    <marc:subfield code="a">130916</marc:subfield>
    </marc:datafield>
    </marc:record>
    ...
    </marc:collection>


Go [XML unmarshaller](http://golang.org/pkg/encoding/xml/#Unmarshal) is not streaming,
which limits the size of the files that can be handled. Converting a 500M
XML file takes about 1m and is about twice as fast as a corresponding Python
version.


marcsplit
---------

Just like `yaz-marcdump -s [PREFIX] -C [CHUNKSIZE] FILE`, but a bit faster.


marccount
---------

Count the number of records in a file (fast). Can be about 4 times faster than

    yaz-marcdump -np file.mrc | tail -1 | awk '{print $3}'

for files with a lot of small records. Up to 20 times faster,
when the Linux file system cache is warmed up.


marcuniq
--------

Like `uniq` but consider MARC files and their 001 fields.

    Usage: marcuniq [OPTIONS] MARCFILE
      -i=false: ignore marc errors (not recommended)
      -o="": output file (or stdout if none given)
      -v=false: prints current program version
      -x="": comma separated list of ids to exclude (or filename with one id per line)

The `-x` option is a bit unrelated. It lets in addition specify ids, that
should be ignored completely. It can be a comma separated string or a filename
with one id per line. Example with filename:

    $ marcuniq -x excludes.txt -o filtered.mrc file.mrc

Example with inline excludes:

    $ marcuniq -x '15270, 15298, 15318, 15335' file.mrc > filtered.mrc


marcmap
-------

Generate tab separated values of the id (001), offset and length of the
MARC records in a file:

    Usage: marcmap [OPTIONS] MARCFILE
      -o="": output to sqlite3 file
      -v=false: prints current program version

Example:

    $ marcmap file.mrc|head -5
    1000574288  0 880
    1000572307  880 777
    1000570320  1657  861
    1000570045  2518  908
    1000555798  3426  823

This tool uses `yaz-marcdump` and `awk` internally to be more performant.
A 4.3G file with 4007803 records takes less than two minutes to map.
(The original version without the external commands needed to parse each record
and took about four minutes on the above file.)

----

Note: Since this tools won't parse the record, it cannot at the moment handle
files with records, that have no ID value (no 001 field).

----

Since 1.3.6: You can use the `-o FILE` flag to specify a filename where
a sqlite3 representation of the data is written into a table called
`seekmap(id, offset, length)`. Example:

    $ marcmap -o test.db test-dups.mrc
    $ sqlite3 test.db ".schema seekmap"
    CREATE TABLE seekmap (id text, offset int, length int);
    $ sqlite3 test.db "select count(*) from seekmap"
    102812

marciter
--------

For development and benchmarks only.


Retired
-------

`loktotsv`, which was just a `marctotsv` with a fixed set of fields.


License
-------

GPLv3.
