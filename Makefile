all:
	go build -o lok2tsv lok2tsv.go
	go build -o marc2tsv marc2tsv.go
	go build -o marcdump marcdump.go

clean:
	rm -f lok2tsv marc2tsv marcdump

fmt:
	gofmt -w -tabs=false -tabwidth=4 lok2tsv.go
	gofmt -w -tabs=false -tabwidth=4 marc2tsv.go
	gofmt -w -tabs=false -tabwidth=4 marcdump.go
