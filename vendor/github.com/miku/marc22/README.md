MARC 22
=======

An experimental fork of [marc21](https://github.com/miku/marc21).

The main difference for now is that marc22 will be able to read MARC from
MARCXML, whereas marc21 can only write XML.

marc21 uses a `[]Field` interface slice to hold both control and data fields, which
will prevent [XML parsing](http://golang.org/src/pkg/encoding/xml/read.go#L345).

See http://golang.org/src/pkg/encoding/xml/read.go#L345:

```golang
...
345		case reflect.Interface:
346			// TODO: For now, simply ignore the field. In the near
347			//       future we may choose to unmarshal the start
348			//       element on it, if not nil.
349			return p.Skip()
...
```

marc22 is a workaround for Go 1.3. The situation might improve with coming
versions.
