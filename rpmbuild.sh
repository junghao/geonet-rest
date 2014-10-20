#!/bin/sh
#
# Expects two argument, the pipeline build number and the git revision.

rel=${1}
rev=${2}

shortrev=${rev:0:7}

mkdir -p rpmbuild/{BUILD,RPMS,SRPMS,SOURCES,SPECS}
topdir="$(pwd)/rpmbuild"
sourcedir=$(pwd)
buildroot="${topdir}/BUILD"

install -D -m 0755 geonet-rest ${buildroot}/usr/bin/geonet-rest
install -D -m 0644 geonet-rest.json ${buildroot}/etc/sysconfig/geonet-rest.json

cp -a README.md api-docs ${buildroot}

# Convert git log to RPM's ChangeLog format (shown with rpm -qp --changelog <rpm file>)
cp geonet-rest.spec $topdir/SPECS/geonet-rest.spec
git log -n 20 --date-order --no-merges --format='* %cd %an <%ae> (%h)%n- %s%n%w(80,2,2)%b' --date=local | sed -r '/^[*]/ s/[0-9]+:[0-9]+:[0-9]+ //' >> $topdir/SPECS/geonet-rest.spec

rpmbuild -bb -v \
	--define="rel $rel" \
	--define="rev $rev" \
	--define="buildroot $buildroot" \
	--define="_topdir $topdir" \
	--define="_sourcedir $sourcedir" \
	--define="_rpmdir $topdir/RPMS" \
	$topdir/SPECS/geonet-rest.spec
