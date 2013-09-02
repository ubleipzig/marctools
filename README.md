lok2tsv
=======

Convert MARC21 [*lok* data](https://wiki.bsz-bw.de/doku.php?id=v-team:daten:datendienste:marc21) into a tabular format, using *001*, *004*,
*005*, *852.a* fields. Why? In a use case, we had a large MARC file which we wanted to convert to a tabular form. Using `yaz-marcdump` and `grep` or
`awk` would work, too, but it's a bit dependent on the printed output of `yaz-marcdump`. Using an XSLT stylesheet on turbomarc gets `xsltproc` killed (out of memory).
So why not try Go? It should be faster than python and [easier](https://gitorious.org/marc21-go/marc21) to implement then [C](http://www.indexdata.com/yaz/doc/marc.html).
Note: There are yet other ways, like splitting the large MARC file into pieces and then apply the XSL transformation.


Build - you'll [need a Go installation](http://golang.org/doc/install):

    $ git clone git@github.com:miku/lok2tsv.git
    $ cd lok2tsv
    $ make

Usage:

    $ ./lok2tsv /tmp/data-lok.mrc
    ...
    014929481   481126031   DE-15-292   20040219000000
    014929481   531827348   DE-15       20090924120312
    014929481   481126880   DE-Ch1      19971112000000
    014929481   481126996   DE-105      20061219132653
    014929481   481127062   DE-Zwi2     19980210000000
    ...


Benchmarks on a single 1.3G file with 5457095 records:

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


P.S. A hacky way to get a similar output would be:

    $ time yaz-marcdump test.mrc | egrep "(^001 |^004 |^005 |^852.*DE-*)" | \
        sed -e 's/852    $a //' | \
        sed -e 's/^001 //' | \
        sed -e 's/^004 //' | \
        sed -e 's/^005 //' | \
        paste - - - -

But that would not handle missing (or multiple values). On the upside, this
takes just 27s (about 202114 records/s).