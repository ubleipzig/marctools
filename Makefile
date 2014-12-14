SHELL := /bin/bash
TARGETS = marccount marcdb marcdump marcmap marcsplit marctojson marctotsv marcuniq marcxmltojson

# http://docs.travis-ci.com/user/languages/go/#Default-Test-Script
test:
	go get -d && go test -v

fmt:
	go fmt ./...

imports:
	goimports -w .

all: fmt test
	go build

install:
	go install

clean:
	go clean
	rm -fv coverage.out
	rm -fv marccount marcdb marcdump marcmap marcsplit marctojson marctotsv marcuniq marcxmltojson
	rm -fv *.x86_64.rpm
	rm -fv debian/marctools*.deb
	rm -rfv debian/marctools/usr

cover:
	go get -d && go test -v	-coverprofile=coverage.out
	go tool cover -html=coverage.out

marccount:
	go build cmd/marccount/marccount.go

marcdb:
	go build cmd/marcdb/marcdb.go

marcdump:
	go build cmd/marcdump/marcdump.go

marcmap:
	go build cmd/marcmap/marcmap.go

marcsplit:
	go build cmd/marcsplit/marcsplit.go

marctojson:
	go build cmd/marctojson/marctojson.go

marctotsv:
	go build cmd/marctotsv/marctotsv.go

marcuniq:
	go build cmd/marcuniq/marcuniq.go

marcxmltojson:
	go build cmd/marcxmltojson/marcxmltojson.go

# experimental deb building
deb: $(TARGETS)
	mkdir -p debian/marctools/usr/sbin
	cp $(TARGETS) debian/marctools/usr/sbin
	cd debian && fakeroot dpkg-deb --build marctools .

# rpm building via vagrant
SSHCMD = ssh -o StrictHostKeyChecking=no -i vagrant.key vagrant@127.0.0.1 -p 2222
SCPCMD = scp -o port=2222 -o StrictHostKeyChecking=no -i vagrant.key

rpm: $(TARGETS)
	mkdir -p $(HOME)/rpmbuild/{BUILD,SOURCES,SPECS,RPMS}
	cp ./packaging/marctools.spec $(HOME)/rpmbuild/SPECS
	cp $(TARGETS) $(HOME)/rpmbuild/BUILD
	./packaging/buildrpm.sh marctools
	cp $(HOME)/rpmbuild/RPMS/x86_64/marctools*rpm .

# Helper to build RPM on a RHEL6 VM, to link against glibc 2.12
vagrant.key:
	curl -sL "https://raw.githubusercontent.com/mitchellh/vagrant/master/keys/vagrant" > vagrant.key
	chmod 0600 vagrant.key

# Don't forget to vagrant up :) - and add your public key to the guests authorized_keys
setup: vagrant.key
	$(SSHCMD) "sudo yum install -y sudo yum install http://ftp.riken.jp/Linux/fedora/epel/6/i386/epel-release-6-8.noarch.rpm"
	$(SSHCMD) "sudo yum install -y golang git rpm-build"
	$(SSHCMD) "mkdir -p /home/vagrant/src/github.com/ubleipzig"
	$(SSHCMD) "cd /home/vagrant/src/github.com/ubleipzig && git clone https://github.com/ubleipzig/marctools.git"

rpm-compatible: vagrant.key
	$(SSHCMD) "cd /home/vagrant/src/github.com/ubleipzig/marctools && GOPATH=/home/vagrant go get ./..."
	$(SSHCMD) "cd /home/vagrant/src/github.com/ubleipzig/marctools && git pull origin master && pwd && GOPATH=/home/vagrant make clean rpm"
	$(SCPCMD) vagrant@127.0.0.1:/home/vagrant/src/github.com/ubleipzig/marctools/*rpm .

# local rpm publishing
REPOPATH = /usr/share/nginx/html/repo/CentOS/6/x86_64

publish: rpm-compatible
	cp marctools-*.rpm $(REPOPATH)
	createrepo $(REPOPATH)
