SHELL := /bin/bash
TARGETS = marccount marcdump marcmap marcsplit marctojson marctotsv marcuniq marcxmltojson

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
	rm -f coverage.out
	rm -f marccount marcdump marcmap marcsplit marctojson marctotsv marcuniq marcxmltojson
	rm -f marctools-*.x86_64.rpm
	rm -f debian/marctools*.deb
	rm -rf debian/marctools/usr

cover:
	go get -d && go test -v	-coverprofile=coverage.out
	go tool cover -html=coverage.out

marccount:
	go build cmd/marccount/marccount.go

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
	cp $(HOME)/rpmbuild/RPMS/x86_64/*rpm .

# Helper to build RPM on a RHEL6 VM, to link against glibc 2.12
vagrant.key:
	curl -sL "https://raw.githubusercontent.com/mitchellh/vagrant/master/keys/vagrant" > vagrant.key
	chmod 0600 vagrant.key

# Don't forget to vagrant up :) - and add your public key to the guests authorized_keys
setup: vagrant.key
	$(SSHCMD) "sudo yum install -y sudo yum install http://ftp.riken.jp/Linux/fedora/epel/6/i386/epel-release-6-8.noarch.rpm"
	$(SSHCMD) "sudo yum install -y golang git rpm-build"
	$(SSHCMD) "mkdir -p /home/vagrant/src/github.com/miku"
	$(SSHCMD) "cd /home/vagrant/src/github.com/miku && git clone https://github.com/miku/marctools.git"

rpm-compatible: vagrant.key
	$(SSHCMD) "cd /home/vagrant/src/github.com/miku/marctools && GOPATH=/home/vagrant go get"
	$(SSHCMD) "cd /home/vagrant/src/github.com/miku/marctools && git pull origin master && pwd && GOPATH=/home/vagrant make rpm"
	$(SCPCMD) vagrant@127.0.0.1:/home/vagrant/src/github.com/miku/marctools/*rpm .

# local rpm publishing
REPOPATH = /usr/share/nginx/html/repo/CentOS/6/x86_64

publish: rpm-compatible
	cp marctools-*.rpm $(REPOPATH)
	createrepo $(REPOPATH)