#!/bin/bash

ddl_dir=$(dirname $0)/../ddl

user=postgres
db_user=${1:-$user}
export PGPASSWORD=$2

# A script to initialise the database.
#
# usage: initdb-93.sh 'db_super_user_name' 'db_super_user_password'
#     for GeoNet Vagrant boxes run with no args e.g.,: initdb-93.sh
# 
# Install postgres and postgis.
# There are comprehensive instructions here http://wiki.openstreetmap.org/wiki/Mapnik/PostGIS
#
# Set the default timezone to UTC and set the timezone abbreviations.  
# Assuming a yum install this will be `/var/lib/pgsql/data/postgresql.conf`
# ...
# timezone = UTC
# timezone_abbreviations = 'Default'
#
# For testing do not set a password for postgres and in /var/lib/pgsql/data/pg_hba.conf set
# connections for local ans host connections to trust:
#
# local   all             all                                     trust
# host    all             all             127.0.0.1/32            trust
#
# Restart postgres.
#
dropdb --host=127.0.0.1 --username=$db_user hazard
psql --host=127.0.0.1 -d postgres --username=$db_user --file=${ddl_dir}/drop-create-users.ddl
psql --host=127.0.0.1 -d postgres --username=$db_user --file=${ddl_dir}/create-db.ddl

# Function security means adding postgis has to be done as a superuser - here that is the postgres user.
# On AWS RDS the created functions have to be transfered to the rds_superuser.
# http://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/Appendix.PostgreSQL.CommonDBATasks.html#Appendix.PostgreSQL.CommonDBATasks.PostGIS
psql --host=127.0.0.1 -d hazard --username=$db_user -c 'create extension postgis;'

psql --host=127.0.0.1 --quiet --username=$db_user --dbname=hazard --file=${ddl_dir}/qrt-drop-create.ddl
psql --host=127.0.0.1 --quiet --username=$db_user hazard -f ${ddl_dir}/qrt-clean.ddl
psql --host=127.0.0.1 --quiet --username=$db_user hazard -f ${ddl_dir}/qrt-locality.ddl
psql --host=127.0.0.1 --quiet --username=$db_user hazard -f ${ddl_dir}/qrt-locality-values-copy.ddl
psql --host=127.0.0.1 --quiet --username=$db_user hazard -f ${ddl_dir}/qrt-region.ddl
psql --host=127.0.0.1 --quiet --username=$db_user hazard -f ${ddl_dir}/qrt-region-values.ddl
psql --host=127.0.0.1 --quiet --username=$db_user hazard -f ${ddl_dir}/qrt-views.ddl
psql --host=127.0.0.1 --quiet --username=$db_user hazard -f ${ddl_dir}/qrt-soh.ddl
psql --host=127.0.0.1 --quiet --username=$db_user hazard -f ${ddl_dir}/impact-create.ddl
psql --host=127.0.0.1 --quiet --username=$db_user hazard -f ${ddl_dir}/impact-functions.ddl
psql --host=127.0.0.1 --quiet --username=$db_user hazard -f ${ddl_dir}/user-permissions.ddl
#
# Event test data.
# Change the year of the quake times in the test data so that the quakes will
# still be in the qrt.quakeinternal view which is limited to the last year
# for performance reasons.
#
psql --host=127.0.0.1 --quiet --username=$db_user hazard -f ${ddl_dir}/event-test-data.ddl
psql --host=127.0.0.1 --quiet --username=$db_user hazard -f ${ddl_dir}/event-date-change.ddl
psql --host=127.0.0.1 --quiet --username=$db_user hazard -f ${ddl_dir}/soh-test-data.ddl
psql --host=127.0.0.1 --quiet --username=$db_user hazard -f ${ddl_dir}/impact-test-data.ddl
