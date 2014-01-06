sources = $(wildcard *.go)
targets = $(basename $(sources))
installed = $(sources:%.go=$(HOME)/bin/%)

all: $(targets)

%: %.go
	gofmt -w -tabs=false -tabwidth=4 $<
	go build -o $@ $<

$(HOME)/bin/%: %
	ln -s $(shell pwd)/$< $@

install-home: $(targets) $(installed)

clean-installed:
	rm -f $(installed)

clean: clean-installed
	rm -f $(targets)
	rm -f gomarckit-*.rpm

# buildrpm: https://gist.github.com/miku/7874111
rpm: $(targets)
	mkdir -p $(HOME)/rpmbuild/{BUILD,SOURCES,SPECS,RPMS}
	cp gomarckit.spec $(HOME)/rpmbuild/SPECS
	cp marctotsv marctojson marcxmltojson marcdump marcsplit marccount marcuniq $(HOME)/rpmbuild/BUILD
	./buildrpm.sh gomarckit
	cp $(HOME)/rpmbuild/RPMS/x86_64/*rpm .
