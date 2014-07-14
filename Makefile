# http://docs.travis-ci.com/user/languages/go/#Default-Test-Script
test:
	go get -d && go test

fmt:
	go fmt ./...

all: fmt test
	go build

install:
	go install

clean:
	go clean
