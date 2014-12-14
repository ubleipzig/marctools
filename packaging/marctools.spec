Summary:    Various MARC command line utils in Go
Name:       marctools
Version:    1.6.0
Release:    0
License:    GPLv3
BuildArch:  x86_64
BuildRoot:  %{_tmppath}/%{name}-build
Group:      System/Base
Vendor:     UB Leipzig
URL:        https://github.com/ubleipzig/marctools

%description
Highlights:

* marctojson -- convert MARC to JSON
* marctotsv  -- convert MARC to TAB-separated file

Other:

* marccount
* marcdb
* marcdump
* marcmap
* marcsplit
* marcuniq
* marcxmltojson


%prep
# the set up macro unpacks the source bundle and changes in to the represented by
# %{name} which in this case would be my_maintenance_scripts. So your source bundle
# needs to have a top level directory inside called my_maintenance _scripts
# %setup -n %{name}

%build
# this section is empty for this example as we're not actually building anything

%install
# create directories where the files will be located
mkdir -p $RPM_BUILD_ROOT/usr/local/sbin

# put the files in to the relevant directories.
# the argument on -m is the permissions expressed as octal. (See chmod man page for details.)
install -m 755 marccount $RPM_BUILD_ROOT/usr/local/sbin
install -m 755 marcdb $RPM_BUILD_ROOT/usr/local/sbin
install -m 755 marcdump $RPM_BUILD_ROOT/usr/local/sbin
install -m 755 marcmap $RPM_BUILD_ROOT/usr/local/sbin
install -m 755 marcsplit $RPM_BUILD_ROOT/usr/local/sbin
install -m 755 marctojson $RPM_BUILD_ROOT/usr/local/sbin
install -m 755 marctotsv $RPM_BUILD_ROOT/usr/local/sbin
install -m 755 marcuniq $RPM_BUILD_ROOT/usr/local/sbin
install -m 755 marcxmltojson $RPM_BUILD_ROOT/usr/local/sbin


%post
# the post section is where you can run commands after the rpm is installed.
# insserv /etc/init.d/my_maintenance

%clean
rm -rf $RPM_BUILD_ROOT
rm -rf %{_tmppath}/%{name}
rm -rf %{_topdir}/BUILD/%{name}

# list files owned by the package here
%files
%defattr(-,root,root)
/usr/local/sbin/marccount
/usr/local/sbin/marcdb
/usr/local/sbin/marcdump
/usr/local/sbin/marcmap
/usr/local/sbin/marcsplit
/usr/local/sbin/marctojson
/usr/local/sbin/marctotsv
/usr/local/sbin/marcuniq
/usr/local/sbin/marcxmltojson


%changelog
* Sun Dec 14 2014 Martin Czygan
- 1.6.0 release
- marctojson got a -recordkey flag
- move marctools import from miku to ubleipzig

* Wed Sep 17 2014 Martin Czygan
- 1.5.5 release
- added marcdb -encode flag

* Wed Sep 17 2014 Martin Czygan
- 1.5.3 release
- new marcdb command

* Wed Aug 13 2014 Martin Czygan
- 1.5.1 release
- switched to github.com/miku/marc22 library, which allows XML-deserialization
- re-added marcxmltojson, this time with a streaming parser

* Sat Jul 19 2014 Martin Czygan
- 1.4.0 release
- rebranded to marctools
- no more marcxmltojson, convert marcxml to marc first and use marctojson
- major overhaul of all tools, marctojson and marctotsv are now multicore-aware
- standardize packaging, make executables go-get-able
- added test with about 70% coverage (make cover)

* Sun May 11 2014 Martin Czygan
- 1.3.8 release
- index seekmap sqlite3 table

* Sun Jan 19 2014 Martin Czygan
- 1.3.7 release
- RHEL6 compatible release (glibc 2.12)
- added `make rpm-compatible` as build helper

* Thu Jan 09 2014 Martin Czygan
- 1.3.6 release
- added marcmap sqlite3 export option via -o

* Thu Jan 07 2014 Martin Czygan
- 1.3.5 release
- make marcmap fast (yaz-marcdump and awk required)

* Tue Jan 07 2014 Martin Czygan
- 1.3.4 release
- added marcmap utility

* Mon Jan 06 2014 Martin Czygan
- 1.3.3 release
- allow files marcuniq -x

* Mon Jan 06 2014 Martin Czygan
- 1.3.2 release
- added -x flag to marcuniq

* Mon Jan 06 2014 Martin Czygan
- 1.3.1 release
- added marcuniq

* Mon Dec 09 2013 Martin Czygan
- 1.3 release
- added Go 1.2 benchmark
- -verbose, dump number of records converted
- added verbose and version flags
- first version of marcxmltojson
- added -l to omit leader and -p for plain mode
- added separator string flag for repeated subfields
- added install-home symlink option to Makefile
- added marccount; version bump to 1.1.0
- added parallel wrapper to marctojson
- added marcsplit, slightly faster than yaz-marcdump's split
- fixed shadowed repeated subfield code bug
- simplified Makefile
- panic on json marshalling errors
- added a few performance data points
- added -r option for tag filtering
- first functional version of marctojson.go
- change naming from x2y to xtoy
- allow literals to be passed alongside tags and leader specs
- added special tags to access leader information
- use licence terms of marc21 library
- use tabwidth=4
- make fmt; and use spaces instead of tabs
- added marc2tsv, marcdump
- implement -v for version
- added a note on use case performance
- added link to marc21/go
- reorder columns
- added marc21 library inline
