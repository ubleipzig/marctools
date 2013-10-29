all:
	go build -o loktotsv loktotsv.go
	go build -o marctotsv marctotsv.go
	go build -o marcdump marcdump.go
	go build -o marctojson marctojson.go

clean:
	rm -f loktotsv marctotsv marcdump marctojson

fmt:
	gofmt -w -tabs=false -tabwidth=4 loktotsv.go
	gofmt -w -tabs=false -tabwidth=4 marctotsv.go
	gofmt -w -tabs=false -tabwidth=4 marcdump.go
	gofmt -w -tabs=false -tabwidth=4 marctojson.go
