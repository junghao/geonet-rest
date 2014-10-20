# Disable the stupid stuff rpm distros include in the build process by default:
#   Disable any prep shell actions. replace them with simply 'true'
%define __spec_prep_post true
%define __spec_prep_pre true
#   Disable any build shell actions. replace them with simply 'true'
%define __spec_build_post true
%define __spec_build_pre true
#   Disable any install shell actions. replace them with simply 'true'
%define __spec_install_post true
%define __spec_install_pre true
#   Disable any clean shell actions. replace them with simply 'true'
%define __spec_clean_post true
%define __spec_clean_pre true
# Disable checking for unpackaged files ?
#%undefine __check_files

%define debug_package   %{nil}

%if 0%{!?rev:1}
%define rev             %(git rev-parse HEAD)
%endif
%define shortrev        %(r=%{rev}; echo ${r:0:7})

%define gh_user         GeoNet
%define gh_name         geonet-rest
%define gh_tar          %{gh_user}-%{gh_name}-%{shortrev}
%define import_path     github.com/%{gh_user}/%{gh_name}

Name:		geonet-rest
Version:	0.1
Release:	%{?rel}git%{shortrev}%{?dist}
Summary:	Rest API for GeoNet web site data.

Group:		Applications/Webapps
License:	GNS
URL:		https://%{import_path}
Source0:	https://%{import_path}/tarball/master/%{gh_tar}.tar.gz

BuildRequires:	golang

%description
GeoNet REST API

The data provided here is used for the GeoNet web site and other similar services.
If you are looking for data for research or other purposes then please check the full [range of data available](http://info.geonet.org.nz/x/DYAO) from GeoNet.  

%prep
# noop

%build
# noop

%install
# noop

%clean
# noop

%files
%defattr(-,root,root,-)
%doc README.md api-docs
%config(noreplace) %{_sysconfdir}/sysconfig/geonet-rest.json
%attr(755,root,root) %{_bindir}/geonet-rest


%changelog
