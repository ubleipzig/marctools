gomarckit
=========

Included: `marcdump`, `marc2tsv`, `lok2tsv`. Based on [marc21](https://gitorious.org/marc21-go/marc21) by [Dan Scott](https://gitorious.org/~dbs).

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


marc2tsv
--------

    $ marc2tsv -h
    Usage of ./marc2tsv:
      -f="<NULL>": fill missing values with this
      -i=false: ignore marc errors (not recommended)
      -k=false: skip lines with missing values
      -v=false: prints current program version


Convert MARC21 to tsv. If a given tag or subfield has multiple values, just
the first one is considered. Only useful for control fields or non-repeatable fields
with a single subfield.

Examples:

    $ ./marc2tsv test.mrc 001
    ...
    111859182
    111862493
    111874173
    111879078
    ...

Empty tag values get a default fill value `<NULL>`:

    $ ./marc2tsv test.mrc 001 004 005 852.a
    ...
    121187764   01635253X   20040324000000  DE-105
    38541028X   087701561   20010420000000  DE-540
    385410298   087701561   20120910090128  <NULL>
    385411057   087701723   20120910090145  DE-540
    ...

Use a custom `fillna` tag:

    $ ./marc2tsv test.mrc -f UNDEF 001 004 005 852.a
    ...
    385410271   087701553   20101125121554  DE-15
    38541028X   087701561   20010420000000  DE-540
    385410298   087701561   20120910090128  UNDEF
    385411057   087701723   20120910090145  DE-540
    ...

Or skip non-complete row entirely:

    $ ./marc2tsv test.mrc -k 001 004 005 852.a
    ...
    121187764   01635253X   20040324000000  DE-105
    121187772   01635253X   20040324000000  DE-105
    38541028X   087701561   20010420000000  DE-540
    385411057   087701723   20120910090145  DE-540
    ...

Access leader with special tags:

    $ ./marc2tsv test.mrc -k 001 @Status
    9780415681889   c
    9780415839792   n
    9780415773874   c

    $ ./marc2tsv test.mrc 001 @Length @Status 300.a
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


    $ ./marc2tsv test.mrc 001 @Status Hello "x y z"
    9781420086454   n   Hello   x y z
    9781420086478   n   Hello   x y z
    9781420094237   c   Hello   x y z
    9780849397776   c   Hello   x y z
    ...


lok2tsv
-------

Convert MARC21 [*lok* data](https://wiki.bsz-bw.de/doku.php?id=v-team:daten:datendienste:marc21) into a tabular format, using *001*, *004*,
*005*, *852.a* fields. Why? In a use case, we had a large MARC file which we wanted to convert to a tabular form. Using `yaz-marcdump` and `grep` or
`awk` would work, too, but it's a bit dependent on the printed output of `yaz-marcdump`. Using an XSLT stylesheet on turbomarc gets `xsltproc` killed (out of memory).
So why not try Go? It should be faster than python and [easier](https://gitorious.org/marc21-go/marc21) to implement then [C](http://www.indexdata.com/yaz/doc/marc.html).
Note: There are yet other ways, like splitting the large MARC file into pieces and then apply the XSL transformation.

For our use case, the conversion with Go is about 4x faster than our current pure Python version.


    $ ./lok2tsv /tmp/data-lok.mrc
    ...
    014929481   481126031   DE-15-292   20040219000000
    014929481   531827348   DE-15       20090924120312
    014929481   481126880   DE-Ch1      19971112000000
    014929481   481126996   DE-105      20061219132653
    014929481   481127062   DE-Zwi2     19980210000000
    ...


Benchmarks
..........

On a single 1.3G file with 5457095 records:

    $ wc -c < test.mrc
    1384415730

    $ time yaz-marcdump -np test.mrc | tail -1
    <!-- Record 5457095 offset 1384415509 (0x52848115) -->

    real    0m12.723s
    user    0m12.512s
    sys     0m0.552s

    $ time ./lok2tsv test.mrc > test.tsv

    real    2m24.412s
    user    1m31.512s
    sys     0m48.984s

So about 37896 records/s.



**P.S.** A hacky way to get a similar output would be:

    $ time yaz-marcdump test.mrc | egrep "(^001 |^004 |^005 |^852.*DE-*)" | \
        sed -e 's/852    $a //' | \
        sed -e 's/^001 //' | \
        sed -e 's/^004 //' | \
        sed -e 's/^005 //' | \
        paste - - - -

But that would not handle missing (or multiple values). On the upside, this
takes just 27s (about 202114 records/s).

License
-------

GPLv3.
