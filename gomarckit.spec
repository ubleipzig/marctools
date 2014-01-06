Summary:    Various MARC command line utils in Go
Name:       gomarckit
Version:    1.3.1
Release:    0
License:    GPLv3
BuildArch:  x86_64
BuildRoot:  %{_tmppath}/%{name}-build
Group:      System/Base
Vendor:     UB Leipzig
URL:        https://github.com/miku/gomarckit

%description
Highlights:

* marctotsv, converts MARC to TAB-separated file
* marctojson, converst MARC to JSON

Other:

* marcxmltojson (non-streaming)
* marcsplit
* marccount
* marcdump
* marcuniq


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
install -m 755 marctotsv $RPM_BUILD_ROOT/usr/local/sbin
install -m 755 marctojson $RPM_BUILD_ROOT/usr/local/sbin
install -m 755 marcxmltojson $RPM_BUILD_ROOT/usr/local/sbin
install -m 755 marcsplit $RPM_BUILD_ROOT/usr/local/sbin
install -m 755 marccount $RPM_BUILD_ROOT/usr/local/sbin
install -m 755 marcdump $RPM_BUILD_ROOT/usr/local/sbin
install -m 755 marcuniq $RPM_BUILD_ROOT/usr/local/sbin


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
/usr/local/sbin/marctotsv
/usr/local/sbin/marctojson
/usr/local/sbin/marcxmltojson
/usr/local/sbin/marcsplit
/usr/local/sbin/marccount
/usr/local/sbin/marcdump
/usr/local/sbin/marcuniq



%changelog
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

