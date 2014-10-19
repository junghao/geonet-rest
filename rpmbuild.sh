#!/bin/sh
#
# Expects two argument, the pipeline build number and the git revision.

rel=${1}
rev=${2}

shortrev=${rev:0:7}

mkdir -p rpmbuild/{BUILD,RPMS,SRPMS,SOURCES,SPECS}
topdir="$(pwd)/rpmbuild"
sourcedir="${topdir}/SOURCES"

# Build the tarball
git archive --format=tar HEAD | gzip -c > $sourcedir/GeoNet-geonet-rest-${shortrev}.tar.gz

# Convert git log to RPM's ChangeLog format (shown with rpm -qp --changelog <rpm file>)
cp geonet-rest.spec $topdir/SPECS/geonet-rest.spec
git log -n 20 --date-order --no-merges --format='* %cd %an <%ae> (%h)%n- %s%n%w(80,2,2)%b' --date=local | sed -r '/^[*]/ s/[0-9]+:[0-9]+:[0-9]+ //' >> $topdir/SPECS/geonet-rest.spec

rpmbuild -bb -v \
	--define="rel $rel" \
	--define="rev $rev" \
	--define="_topdir $topdir" \
	--define="_sourcedir $sourcedir" \
	--define="_rpmdir $topdir/RPMS" \
	$topdir/SPECS/geonet-rest.spec
