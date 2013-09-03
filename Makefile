all:
	go build -o lok2tsv lok2tsv.go
	go build -o marcdump marcdump.go
	go build -o marc2tsv marc2tsv.go

clean:
	rm -f lok2tsv marcdump marc2tsv