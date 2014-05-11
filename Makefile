sources = $(wildcard *.go)
targets = $(basename $(sources))
installed = $(sources:%.go=$(HOME)/bin/%)

SSH = ssh -o StrictHostKeyChecking=no -i vagrant.key vagrant@127.0.0.1 -p 2222

all: $(targets)

%: %.go
	gofmt -w -tabs=false -tabwidth=4 $<
	go build -o $@ $<

$(HOME)/bin/%: %
	ln -s $(shell pwd)/$< $@

install-home: $(targets) $(installed)

clean-installed:
	rm -f $(installed)

clean: clean-installed
	rm -f $(targets)
	rm -f gomarckit-*.rpm

# buildrpm: https://gist.github.com/miku/7874111
rpm: $(targets)
	mkdir -p $(HOME)/rpmbuild/{BUILD,SOURCES,SPECS,RPMS}
	cp gomarckit.spec $(HOME)/rpmbuild/SPECS
	cp marctotsv marctojson marcxmltojson marcdump marcsplit marccount \
	   marcuniq marcmap $(HOME)/rpmbuild/BUILD
	./buildrpm.sh gomarckit
	cp $(HOME)/rpmbuild/RPMS/x86_64/*rpm .

vagrant.key:
	curl -sL "https://raw2.github.com/mitchellh/vagrant/master/keys/vagrant" > vagrant.key
	chmod 0600 vagrant.key

# helper to build RPM on a RHEL6 VM, to link against glibc 2.12
# Assumes a RHEL6 go installation (http://nareshv.blogspot.de/2013/08/installing-go-lang-11-on-centos-64-64.html)
# And: sudo yum install git rpm-build
# Don't forget to vagrant up :)
vm-setup: vagrant.key
	$(SSH) "sudo yum install -y https://dl.fedoraproject.org/pub/epel/6/i386/epel-release-6-8.noarch.rpm"
	$(SSH) "sudo yum install -y golang git rpm-build"
	$(SSH) "mkdir -p /home/vagrant/github/miku"
	$(SSH) "cd /home/vagrant/github/miku && git clone https://github.com/miku/gomarckit.git"

rpm-compatible: vagrant.key
	$(SSH) "GOPATH=/home/vagrant go get github.com/mattn/go-sqlite3"
	$(SSH) "cd /home/vagrant/github/miku/gomarckit && git pull origin master && GOPATH=/home/vagrant make rpm"
	scp -o port=2222 -o StrictHostKeyChecking=no -i vagrant.key vagrant@127.0.0.1:/home/vagrant/github/miku/gomarckit/*rpm .
