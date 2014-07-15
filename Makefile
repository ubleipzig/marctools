# http://docs.travis-ci.com/user/languages/go/#Default-Test-Script
test:
	go get -d && go test -v

fmt:
	go fmt ./...

all: fmt test
	go build

install:
	go install

clean:
	go clean

cover:
	go get -d && go test -v	-coverprofile=coverage.out
	go tool cover -html=coverage.out
