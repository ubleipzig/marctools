include $(GOROOT)/src/Make.inc

TARG=bitbucket.org/ww/marc21
GOFILES=marc21.go marcxml.go

include $(GOROOT)/src/Make.pkg

format:
	gofmt -w *.go

docs:
	gomake clean
	godoc ${TARG} > README.txt
