sources = $(wildcard *.go)
targets = $(basename $(sources))
installed = $(sources:%.go=$(HOME)/bin/%)

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

# helper to build RPM on a RHEL6 VM, to link against glibc 2.12
# Assumes a RHEL6 go installation (http://nareshv.blogspot.de/2013/08/installing-go-lang-11-on-centos-64-64.html)
# And: sudo yum install git rpm-build
# Don't forget to vagrant up :)
rpm-compatible:
	ssh -o StrictHostKeyChecking=no -i /opt/vagrant/embedded/gems/gems/vagrant-1.3.5/keys/vagrant vagrant@127.0.0.1 -p 2222 "GOPATH=/home/vagrant go get github.com/mattn/go-sqlite3"
	ssh -o StrictHostKeyChecking=no -i /opt/vagrant/embedded/gems/gems/vagrant-1.3.5/keys/vagrant vagrant@127.0.0.1 -p 2222 "cd /home/vagrant/github/miku/gomarckit && git pull origin master && GOPATH=/home/vagrant make rpm"
	scp -o port=2222 -o StrictHostKeyChecking=no -i /opt/vagrant/embedded/gems/gems/vagrant-1.3.5/keys/vagrant vagrant@127.0.0.1:/home/vagrant/github/miku/gomarckit/*rpm .
