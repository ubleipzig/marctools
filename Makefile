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
	mkdir -p $(HOME)/rpmbuild/BUILD
	mkdir -p $(HOME)/rpmbuild/SOURCES
	mkdir -p $(HOME)/rpmbuild/SPECS
	mkdir -p $(HOME)/rpmbuild/RPMS
	cp gomarckit.spec $(HOME)/rpmbuild/SPECS
	cp marctotsv $(HOME)/rpmbuild/BUILD
	cp marctojson $(HOME)/rpmbuild/BUILD
	cp marcxmltojson $(HOME)/rpmbuild/BUILD
	cp marcdump $(HOME)/rpmbuild/BUILD
	cp marcsplit $(HOME)/rpmbuild/BUILD
	cp marccount $(HOME)/rpmbuild/BUILD
	./buildrpm.sh gomarckit
	cp $(HOME)/rpmbuild/RPMS/x86_64/*rpm .
