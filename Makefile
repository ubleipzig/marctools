sources = $(wildcard *.go)
targets = $(basename $(sources))

all: $(targets)

%: %.go
	gofmt -w -tabs=false -tabwidth=4 $<
	go build -o $@ $<

clean:
	rm $(targets)