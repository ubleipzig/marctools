all:
	go build -o loktotsv loktotsv.go
	go build -o marctotsv marctotsv.go
	go build -o marcdump marcdump.go

clean:
	rm -f loktotsv marctotsv marcdump

fmt:
	gofmt -w -tabs=false -tabwidth=4 loktotsv.go
	gofmt -w -tabs=false -tabwidth=4 marctotsv.go
	gofmt -w -tabs=false -tabwidth=4 marcdump.go
