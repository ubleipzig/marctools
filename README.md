marctools
=========

Various MARC utilities with an eye on performance.

[![Build Status](http://img.shields.io/travis/miku/marctools.svg?style=flat)](https://travis-ci.org/miku/marctools)

Installation
------------

For native RPM or DEB packages see: [Releases](https://github.com/ubleipzig/marctools/releases)

If you have a local Go installation, you can just

    go get github.com/miku/marctools/cmd/{marccount,marcdump,marcmap,marcsplit,marctojson,...}

Executables available:

    marccount
    marcdump
    marcmap
    marcsplit
    marctojson
    marctotsv
    marcuniq

marccount
---------

Prints the number of records found in a file and then exits.

    $ marccount fixtures/journals.mrc
    10

marcdump
--------

Dumps MARC to stdout, similar to [`yaz-marcdump`](http://www.indexdata.com/yaz/doc/yaz-marcdump.html):

    $ marcdump fixtures/testbug2.mrc
    001 testbug2
    005 20110419140028.0
    008 110214s1992    it a     b    001 0 ita d
    020 [  ] [(a) 8820737493]
    035 [  ] [(a) (OCoLC)ocm30585539]
    040 [  ] [(a) RBN], [(c) RBN], [(d) OCLCG], [(d) PVU]
    041 [1 ] [(a) ita], [(a) lat], [(h) lat]
    043 [  ] [(a) e-it---]
    050 [14] [(a) DG848.15], [(b) .V53 1992]
    049 [  ] [(a) PVUM]
    100 [1 ] [(a) Vico, Giambattista,], [(d) 1668-1744.]
    240 [10] [(a) Principum Neapolitanorum coniurationis anni MDCCI historia.], [(l) Italian & Latin]
    245 [13] [(a) La congiura dei Principi Napoletani 1701 :], [(b) (prima e seconda stesura) /], ...
    250 [  ] [(a) Fictional edition.]
    260 [  ] [(a) Morano :], [(b) Centro di Studi Vichiani,], [(c) 1992.]
    300 [  ] [(a) 296 p. :], [(b) ill. ;], [(c) 24 cm.]
    490 [1 ] [(a) Opere di Giambattista Vico ;], [(v) 2/1]
    500 [  ] [(a) Italian and Latin.]
    504 [  ] [(a) Includes bibliographical references (p. [277]-281) and index.]
    520 [3 ] [(a) Sample abstract.]
    590 [  ] [(a) April11phi]
    651 [ 0] [(a) Naples (Kingdom)], [(x) History], [(y) Spanish rule, 1442-1707], [(v) Sources.]
    700 [1 ] [(a) Pandolfi, Claudia.]
    800 [1 ] [(a) Vico, Giambattista,], [(d) 1668-1744.], [(t) Works.], [(f) 1982 ;], [(v) 2, pt. 1.]
    856 [40] [(u) http://fictional.com/sample/url]
    994 [  ] [(a) C0], [(b) PVU]

marcmap
-------

Dumps a list of *id, offset, length* tuples to stdout (TSV) or to a sqlite3 database:

By default write to stdout:

    $ marcmap fixtures/journals.mrc
    testsample1 0   1571
    testsample2 1571    1195
    testsample3 2766    1057
    testsample4 3823    1361
    testsample5 5184    1707
    testsample6 6891    1532
    testsample7 8423    1426
    testsample8 9849    1251
    testsample9 11100   2173
    testsample10    13273   1195

Dump listing into an sqlite database with `-o FILENAME`:

    $ marcmap -o seekmap.db fixtures/journals.mrc
    $ sqlite3 seekmap.db 'select id, offset, length from seekmap'
    testsample1|0|1571
    testsample2|1571|1195
    testsample3|2766|1057
    testsample4|3823|1361
    testsample5|5184|1707
    testsample6|6891|1532
    testsample7|8423|1426
    testsample8|9849|1251
    testsample9|11100|2173
    testsample10|13273|1195

marcsplit
---------

Splits a MARC file into smaller pieces.

    $ marcsplit
    Usage of marcsplit:
      -C=1: number of records per file
      -cpuprofile="": write cpu profile to file
      -d=".": directory to write to
      -s="split-": split file prefix
      -v=false: prints current program version

    $ marcsplit -d /tmp -C 3 -s "example-prefix-" fixtures/journals.mrc
    $ ls -1 /tmp/example-prefix-0000000*
    /tmp/example-prefix-00000000
    /tmp/example-prefix-00000001
    /tmp/example-prefix-00000002
    /tmp/example-prefix-00000003

marctojson
----------

Converts MARC to JSON.

    $ marctojson
    Usage of marctojson:
      -cpuprofile="": write cpu profile to file
      -i=false: ignore marc errors (not recommended)
      -l=false: dump the leader as well
      -m="": a key=value pair to pass to meta
      -p=false: plain mode: dump without content and meta
      -r="": only dump the given tags (e.g. 001,003)
      -v=false: prints current program version and exit
      -w=4: number of workers

Default conversion (abbreviated, [pretty-printed](https://github.com/jmhodges/jsonpp)):

    $ marctojson fixtures/testbug2.mrc | jsonpp
    {
      "content": {
        "001": "testbug2",
        "005": "20110419140028.0",
        "008": "110214s1992    it a     b    001 0 ita d",
        "020": [
          {
            "a": "8820737493",
            "ind1": " ",
            "ind2": " "
          }
        ],
        ...
        "040": [
          {
            "a": "RBN",
            "c": "RBN",
            "d": [
              "OCLCG",
              "PVU"
            ],
            "ind1": " ",
            "ind2": " "
          }
        ],
        ...
        "856": [
          {
            "ind1": "4",
            "ind2": "0",
            "u": "http://fictional.com/sample/url"
          }
        ]
      },
      ...
      "meta": {}
    }

Dump the leader as well with `-l` and only dump field 040 with `-r 040`:

    $ marctojson -l -r 040 fixtures/testbug2.mrc | jsonpp
    {
      "content": {
        "040": [
          {
            "a": "RBN",
            "c": "RBN",
            "d": [
              "OCLCG",
              "PVU"
            ],
            "ind1": " ",
            "ind2": " "
          }
        ],
        "leader": {
          "ba": "337",
          "cs": "a",
          "ic": "2",
          "impldef": "m Ma ",
          "length": "1234",
          "lol": "4",
          "losp": "5",
          "raw": "01234cam a2200337Ma 4500",
          "sfcl": "2",
          "status": "c",
          "type": "a"
        }
      },
      "meta": {}
    }

Restrict JSON to 001 and 245, and use plain mode with `-p`, which has no `meta` or
`content` key:

    $ marctojson -r "001, 245" -p fixtures/testbug2.mrc | jsonpp
    {
      "001": "testbug2",
      "245": [
        {
          "a": "La congiura dei Principi Napoletani 1701 :",
          "b": "(prima e seconda stesura) /",
          "c": "Giambattista Vico ; a cura di Claudia Pandolfi.",
          "ind1": "1",
          "ind2": "3"
        }
      ]
    }

Add some value (here the current date) to the meta map:

    $ marctojson -r "001, 245" -m date="$(date)" fixtures/testbug2.mrc | jsonpp
    {
      "content": {
        "001": "testbug2",
        "245": [
          {
            "a": "La congiura dei Principi Napoletani 1701 :",
            "b": "(prima e seconda stesura) /",
            "c": "Giambattista Vico ; a cura di Claudia Pandolfi.",
            "ind1": "1",
            "ind2": "3"
          }
        ]
      },
      "meta": {
        "date": "Sun Jul 20 17:36:37 CEST 2014"
      }
    }

marctotsv
---------

Converts selected MARC tags to tab-separated values (TSV).

    $ marctotsv
    Usage: marctotsv [OPTIONS] MARCFILE TAG [TAG, TAG, ...]
      -cpuprofile="": write cpu profile to file
      -f="<NULL>": fill missing values with this
      -i=false: ignore marc errors (not recommended)
      -k=false: skip incomplete lines (missing values)
      -s="": separator to use for multiple values
      -v=false: prints current program version and exit
      -w=4: number of workers

Extract a single column:

    $ marctotsv fixtures/journals.mrc 001
    testsample1
    testsample2
    testsample3
    testsample4
    testsample5
    testsample6
    testsample7
    testsample8
    testsample9
    testsample10

Extract two columns:

    $ marctotsv fixtures/journals.mrc 001 245.a
    testsample1 Journal of rational emotive therapy :
    testsample2 Rational living.
    testsample3 Psychotherapy in private practice.
    testsample4 Journal of quantitative criminology.
    testsample5 The Journal of parapsychology.
    testsample6 Journal of mathematics and mechanics.
    testsample7 The Journal of psychology.
    testsample8 Journal of psychosomatic research.
    testsample9 The journal of sex research
    testsample10    Journal of phenomenological psychology.

Use a custom value for undefined fields with `-f UNDEF`:

    $ marctotsv -f UNDEF fixtures/journals.mrc  001 245.a 245.b
    testsample1 Journal of rational emotive therapy :   the journal of the Institute for ...
    testsample2 Rational living.    UNDEF
    testsample3 Psychotherapy in private practice.  UNDEF
    testsample4 Journal of quantitative criminology.    UNDEF
    testsample5 The Journal of parapsychology.  UNDEF
    testsample6 Journal of mathematics and mechanics.   UNDEF
    testsample7 The Journal of psychology.  UNDEF
    testsample8 Journal of psychosomatic research.  UNDEF
    testsample9 The journal of sex research UNDEF
    testsample10    Journal of phenomenological psychology. UNDEF

Only keep complete rows with `-k`:

    $ marctotsv -k fixtures/journals.mrc  001 245.a 245.b
    testsample1 Journal of rational emotive therapy :   the journal of the Institute for ...

Include all values, separated by a pipe via `- s "|"`:

    $ marctotsv -s "|" fixtures/journals.mrc  001 710.a
    testsample1 Institute for Rational-Emotive Therapy (New York, N.Y.)
    testsample2 Institute for Rational-Emotive Therapy (New York, N.Y.)|Institute for ...
    testsample3 <NULL>
    testsample4 LINK (Online service)
    testsample5 Duke University.|ProQuest Psychology Journals.
    testsample6 Indiana University.|Indiana University.
    testsample7 ProQuest Psychology Journals.
    testsample8 ScienceDirect (Online service).
    testsample9 Society for the Scientific Study of Sex (U.S.)|Society for ...|JSTOR
    testsample10    Ingenta (Firm).

marcuniq
--------

    $ marcuniq
    Usage: marcuniq [OPTIONS] MARCFILE
      -i=false: ignore marc errors (not recommended)
      -o="": output file (or stdout if none given)
      -v=false: prints current program version
      -x="": comma separated list of ids to exclude (or filename with one id per line)

Exclude three IDs and dump do file:

    $ marcuniq -x "testsample1,testsample2,testsample3" -o filtered.mrc fixtures/journals.mrc
    excluded ids interpreted as string
    3 ids to exclude loaded
    10 records read
    7 records written, 0 skipped, 3 excluded, 0 without ID (001)

    $ marctotsv filtered.mrc 001
    testsample4
    testsample5
    testsample6
    testsample7
    testsample8
    testsample9
    testsample10

Development
-----------

To run the tests just type:

    make

To open a coverage report in you browser, run:

    make cover

To package an DEB adjust `debian/marctools/DEBIAN/control`, e.g. [update the
version](https://github.com/ubleipzig/marctools/commit/0279811c32e8ad78ceeef821e3b950ceb74e22aa), then run:

    make deb

To package an RPM, adjust `packaging/marctools.spec`, e.g. update the version, then run:

    make rpm

To package an RPM on a CentOS 6.2 with libc **2.12** setup a VM with
veewee and vagrant. Then run:

    vagrant up
    make vm-setup

Subsequently build RPMs against libc 2.12 with

    make rpm-compatible

Previous versions
-----------------

Versions 1.0 up to 1.3.8 (named gomarckit) used a non-standard project layout and lacked
tests. Their version history is preserved under the [1.3.8-maint](https://github.com/miku/marctools/tree/1.3.8-maint) branch.

Todo
----

* Perform and include some performance benchmarks in README.
* The MARC21 library used might issue more system calls than needed, e.g.
  in the main [Record create loop](https://github.com/miku/marc21/blob/4f0c7faee66f15b198c7a550fb78e2a80a0010ea/marc21_record.go#L33) each data and control field will issue a read system call. It could
  be more efficient to read MARC in larger block and distribute the Record
  parsing itself to the workers.
* Add more tests for more fancy MARC files (encodings, broken dirents, etc.).
