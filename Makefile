all:
	go fmt ./...
	go get -d && go build

install:
	go install

clean:
	go clean
