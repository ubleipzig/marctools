include $(GOROOT)/src/Make.inc

TARG=marc2xml
GOFILES=marc2xml.go

include $(GOROOT)/src/Make.cmd

format:
	gofmt -w *.go

docs:
	gomake clean
	godoc ${TARG} > README.txt
