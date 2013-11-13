sources = $(wildcard *.go)
targets = $(basename $(sources))
installed = $(sources:%.go=$(HOME)/bin/%)

all: $(targets)

%: %.go
	gofmt -w -tabs=false -tabwidth=4 $<
	go build -o $@ $<

$(HOME)/bin/%: %
	ln -s $(shell pwd)/$< $@

install-home: $(installed)
	@echo "Installed:"
	@echo $(installed)

clean-bin:
	rm -f $(installed)

clean: clean-bin
	rm $(targets)
