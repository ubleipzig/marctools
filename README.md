gomarckit
=========

Included: [`marcdump`](https://github.com/miku/gomarckit#marcdump), 
[`marctotsv`](https://github.com/miku/gomarckit#marctotsv),
[`marctojson`](https://github.com/miku/gomarckit#marctojson),
[`marcsplit`](https://github.com/miku/gomarckit#marcsplit),
[`marccount`](https://github.com/miku/gomarckit#marccount),
[`marciter`](https://github.com/miku/gomarckit#marciter). Based on [marc21](https://gitorious.org/marc21-go/marc21) by [Dan Scott](https://gitorious.org/~dbs) and William Waites.

Build
-----

You'll [need a Go installation](http://golang.org/doc/install):

    $ git clone git@github.com:miku/gomarckit.git
    $ cd gomarckit
    $ make


marcdump
--------

Just like `yaz-marcdump`. The Go version is about 4-5x slower than the [C version](http://git.indexdata.com/?p=yaz.git;a=blob;f=util/marcdump.c;h=f92204e386431f044f06dddd8baa1c9db08d69c9;hb=HEAD).


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



Convert MARC21 to tsv. If a given tag or subfield has multiple values, just
the first one is considered. Only useful for control fields or non-repeatable fields
with a single subfield.

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

Use a custom `fillna` tag:

    $ ./marctotsv test.mrc -f UNDEF 001 004 005 852.a
    ...
    385410271   087701553   20101125121554  DE-15
    38541028X   087701561   20010420000000  DE-540
    385410298   087701561   20120910090128  UNDEF
    385411057   087701723   20120910090145  DE-540
    ...

Or skip non-complete row entirely:

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

  The resulting JSON file is about 2.1G in size.

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


marcsplit
---------

Just like `yaz-marcdump -s [PREFIX] -C [CHUNKSIZE] FILE`, but a bit faster.


marccount
---------

Count the number of records in a file (fast). Can be about 4 times faster than

    yaz-marcdump -np file.mrc | tail -1 | awk '{print $3}'

for files with a lot of small records. Up to 20 times faster,
when the Linux file system cache is warmed up.


marciter
--------

For development and benchmarks only.


Retired
-------

`loktotsv`, which was just a `marctotsv` with a fixed set of fields.


License
-------

GPLv3.
