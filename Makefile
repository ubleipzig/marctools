all:
	go fmt ./...
	go get -d && go build

clean:
	go clean
