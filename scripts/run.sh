#!/bin/sh
echo "*** TEMPORARY SCRIPT BEFORE IMPLEMENTING MIGRATIONS ***"

. ./scripts/values.sh

setup() {
    PGPASSWORD=$PGPASSWORD psql --host=$PGHOST --username=$PGUSER $PGDATABASE -c "CREATE USER $DEV_USERNAME WITH PASSWORD '$DEV_PASSWORD'"
    PGPASSWORD=$PGPASSWORD psql --host=$PGHOST --username=$PGUSER $PGDATABASE -c "CREATE DATABASE $DEV_DATABASE"
    PGPASSWORD=$PGPASSWORD psql --host=$PGHOST --username=$PGUSER $PGDATABASE -c "GRANT ALL PRIVILEGES ON DATABASE $DEV_DATABASE TO $DEV_USERNAME"

    PGPASSWORD=$DEV_PASSWORD psql --host=$PGHOST --port=$PGPORT --username=$DEV_USERNAME $DEV_DATABASE -f "sqlc/schema.sql"
}

teardown() {
    PGPASSWORD=$PGPASSWORD psql --host=$PGHOST --username=$PGUSER $PGDATABASE -c "SELECT pg_terminate_backend(pg_stat_activity.pid) FROM pg_stat_activity WHERE pg_stat_activity.datname = '$DEV_DATABASE' AND pid <> pg_backend_pid()"
    PGPASSWORD=$PGPASSWORD psql --host=$PGHOST --username=$PGUSER $PGDATABASE -c "DROP DATABASE $DEV_DATABASE"
    PGPASSWORD=$PGPASSWORD psql --host=$PGHOST --username=$PGUSER $PGDATABASE -c "DROP USER $DEV_USERNAME"
}

case "$1" in
setup)
    setup
;;
teardown)
    teardown
;;
*)
echo "Usage: $0 (setup|teardown)"
;;
esac

# values.sh should contain these values (fill in passwords accordingly)

# PGHOST=localhost
# PGUSER=postgres
# PGPASSWORD=
# PGDATABASE=postgres
# DEV_USERNAME=kirbyuser
# DEV_PASSWORD=
# DEV_DATABASE=kirbydb

# run the following command to start the database (manually add in the default pg user, password, and database, THESE WILL NOT BE USED FOR THE KIRBY DATABASE)

# docker run -d \
#   --name dev-postgres \
#   -e POSTGRES_USER=<VALUES PGUSER> \
#   -e POSTGRES_PASSWORD=<VALUES PGPASSWORD> \
#   -e POSTGRES_DB=<VALUES PGDATABASE> \
#   -p 5432:5432 postgres