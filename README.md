lok2tsv
=======

Convert MARC21 *lok* (local) data into a tabular format, using *001*, *004*, 
*005*, *852.a* fields.

Build ([Go needs to be installed](http://golang.org/doc/install)):

    $ git clone git@github.com:miku/lok2tsv.git
    $ cd lok2tsv
    $ make

Usage:

    $ lok2tsv /tmp/data-lok.mrc
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
