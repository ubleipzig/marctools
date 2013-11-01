all: $(basename *.go)

%: %.go
	gofmt -w -tabs=false -tabwidth=4 $<
	go build -o $@ $<

