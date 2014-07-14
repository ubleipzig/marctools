# http://docs.travis-ci.com/user/languages/go/#Default-Test-Script
test:
	go test

fmt:
	go fmt ./...

all: fmt
	go get -d && go build

install:
	go install

clean:
	go clean
