lok2tsv
=======

Convert MARC21 [*lok* data](https://wiki.bsz-bw.de/doku.php?id=v-team:daten:datendienste:marc21) into a tabular format, using *001*, *004*,
*005*, *852.a* fields. Why? In a use case, we had a large MARC which we wanted to convert to a tabular form. Using `yaz-marcdump` and `grep` and
`awk` would work, too, but it's dependent on the printed output of `yaz-marcdump`. Using an XSLT stylesheet on turbomarc gets `xsltproc` killed.
So why not try Go? It should be faster than python and easier to implement then C++. Note: There are yet other ways, like splitting
the large MARC file into pieces and then apply the XSL transformation.


Build - you'll [need a Go installation](http://golang.org/doc/install):

    $ git clone git@github.com:miku/lok2tsv.git
    $ cd lok2tsv
    $ make

Usage:

    $ ./lok2tsv /tmp/data-lok.mrc
    ...
    014929481   481126031   20040219000000  DE-15-292
    014929481   531827348   20090924120312  DE-15
    014929481   481126880   19971112000000  DE-Ch1
    014929481   481126996   20061219132653  DE-105
    014929481   481127062   19980210000000  DE-Zwi2
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