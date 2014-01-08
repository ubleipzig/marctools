marctojson 1 "JANUARY 2014" marctojson 1.3.5 "User Commands"
============================================================

NAME
----

marctojson - convert binary MARC to JSON

SYNOPSIS
--------

`marctojson` [`-ilpv`] [`-m` *key=value,key=value,...* `-r` *tag,tag,...* ] *file* ...

DESCRIPTION
-----------

`marctojson` converts binary MARC to JSON. The JSON format is geared towards
easy queries, in particular with Elasticsearch.

OPTIONS
-------

`-i`
  Ignore marc errors (not recommended).

`-p`
  *Plain mode*: dump without content and meta.

`-l`
  Dump the leader as well.

`-v`
  Prints current program version and exit.

`-m` *key=value,key=value,...*
  `key=value` pairs to pass to the `meta` document section.

`-r` *tag,tag,...*
  Only dump the given tags (e.g. 001,003).

PERFORMANCE
-----------

Tests performed on a i5 M520 Arrandale, Intel 320 SSD (SSDSA2CW120G3),
gomarckit compiled with Go 1.2.

*test-tit.mrc* (#records: 4007803, avg record size: 1126b, total: 4.2GB)

  * `marctojson test-tit.mrc > /dev/null` (24.4min, 2737r/s)

  * `marctojson -r 001 -p test-tit.mrc > /dev/null` (13.8min, 4840r/s)


EXAMPLE OUTPUTS
---------------


`marctojson test-tit.mrc | head -1 | jsonpp`

    {
      "content": {
        "001": "00002144X",
        "003": "DE-576",
        "005": "20120312115715.0",
        "007": "tu",
        "008": "850101n18uuuuuuxx             00 0 ger c",
        "035": [
          {
            "a": "(DE-599)BSZ00002144X",
            "ind1": " ",
            "ind2": " "
          }
        ],
        "040": [
          {
            "a": "DE-576",
            "b": "ger",
            "c": "DE-576",
            "e": "rakwb",
            "ind1": " ",
            "ind2": " "
          }
        ],
        "041": [
          {
            "a": "ger",
            "ind1": "0",
            "ind2": " "
          },
          {
            "a": "dt.",
            "ind1": "0",
            "ind2": "7"
          }
        ],
        "700": [
          {
            "0": [
              "(DE-588)11851802X",
              "(DE-576)160498678"
            ],
            "a": "Buß, Franz Joseph von",
            "d": "1803 - 1878",
            "ind1": "1",
            "ind2": " "
          },
          {
            "0": [
              "(DE-588)100822541",
              "(DE-576)28867958X"
            ],
            "a": "Hepp, Georges Philipp",
            "d": "1792 - 1872",
            "ind1": "1",
            "ind2": " "
          }
        ],
        "936": [
          {
            "a": "PI 2710",
            "b": "Quellenschriften",
            "ind1": "r",
            "ind2": "v",
            "k": [
              "Rechtswissenschaft",
              "Allgemeine Rechtslehre und Rechtstheorie, Rechts- und Staatsphilosophie, Rechtssoziologie",
              "Geschichte der Staats- und Rechtsphilosophie",
              "Staats- und Rechtsphilosophie im 19. Jahrhundert",
              "Quellenschriften"
            ]
          }
        ]
      },
      "meta": {}
    }


`marctojson -p test-tit.mrc | head -1 | jsonpp`

    {
      "001": "00002144X",
      "003": "DE-576",
      "005": "20120312115715.0",
      "007": "tu",
      "008": "850101n18uuuuuuxx             00 0 ger c",
      "035": [
        {
          "a": "(DE-599)BSZ00002144X",
          "ind1": " ",
          "ind2": " "
        }
      ],
      "040": [
        {
          "a": "DE-576",
          "b": "ger",
          "c": "DE-576",
          "e": "rakwb",
          "ind1": " ",
          "ind2": " "
        }
      ],
      "041": [
        {
          "a": "ger",
          "ind1": "0",
          "ind2": " "
        },
        {
          "a": "dt.",
          "ind1": "0",
          "ind2": "7"
        }
      ],
      "700": [
        {
          "0": [
            "(DE-588)11851802X",
            "(DE-576)160498678"
          ],
          "a": "Buß, Franz Joseph von",
          "d": "1803 - 1878",
          "ind1": "1",
          "ind2": " "
        },
        {
          "0": [
            "(DE-588)100822541",
            "(DE-576)28867958X"
          ],
          "a": "Hepp, Georges Philipp",
          "d": "1792 - 1872",
          "ind1": "1",
          "ind2": " "
        }
      ],
      "936": [
        {
          "a": "PI 2710",
          "b": "Quellenschriften",
          "ind1": "r",
          "ind2": "v",
          "k": [
            "Rechtswissenschaft",
            "Allgemeine Rechtslehre und Rechtstheorie, Rechts- und Staatsphilosophie, Rechtssoziologie",
            "Geschichte der Staats- und Rechtsphilosophie",
            "Staats- und Rechtsphilosophie im 19. Jahrhundert",
            "Quellenschriften"
          ]
        }
      ]
    }


`marctojson -p -r 936 test-tit.mrc | head -1 | jsonpp`

    {
      "936": [
        {
          "a": "PI 2710",
          "b": "Quellenschriften",
          "ind1": "r",
          "ind2": "v",
          "k": [
            "Rechtswissenschaft",
            "Allgemeine Rechtslehre und Rechtstheorie, Rechts- und Staatsphilosophie, Rechtssoziologie",
            "Geschichte der Staats- und Rechtsphilosophie",
            "Staats- und Rechtsphilosophie im 19. Jahrhundert",
            "Quellenschriften"
          ]
        }
      ]
    }


`marctojson -l -r 936 test-tit.mrc | head -1 | jsonpp`

    {
      "content": {
        "936": [
          {
            "a": "PI 2710",
            "b": "Quellenschriften",
            "ind1": "r",
            "ind2": "v",
            "k": [
              "Rechtswissenschaft",
              "Allgemeine Rechtslehre und Rechtstheorie, Rechts- und Staatsphilosophie, Rechtssoziologie",
              "Geschichte der Staats- und Rechtsphilosophie",
              "Staats- und Rechtsphilosophie im 19. Jahrhundert",
              "Quellenschriften"
            ]
          }
        ],
        "leader": {
          "ba": "265",
          "cs": "a",
          "ic": "2",
          "impldef": "m   a",
          "length": "1078",
          "lol": "4",
          "losp": "5",
          "raw": "01078cam a2200265  a4500",
          "sfcl": "2",
          "status": "c",
          "type": "a"
        }
      },
      "meta": {}
    }


`marctojson -m about="less is more" -r 001 test-tit.mrc | head -1 | jsonpp`

    {
      "content": {
        "001": "00002144X"
      },
      "meta": {
        "about": "less is more"
      }
    }


AUTHOR
------

Martin Czygan <martin.czygan@gmail.com>

SEE ALSO
--------

[gomarckit](https://github.com/miku/gomarckit)