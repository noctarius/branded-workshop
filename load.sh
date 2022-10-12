#!/bin/bash
sudo wget https://github.com/noctarius/branded-workshop/raw/main/export-workshop.csv.tar.gz
tar xvfz export-workshop.csv.tar.gz
mv tmp/export-workshop.csv .
rm -rf tmp
sudo -u postgres psql -c "alter user postgres password 'test1234'"
sudo -u postgres psql -c "create table metrics(created timestamp with time zone default now() not null, type_id integer not null, value double precision not null);"
echo "Importing data into database..."
timescaledb-parallel-copy --connection "postgres://postgres:test1234@localhost:5432/postgres" --db-name postgres --table metrics --file "export-workshop.csv" --workers 4 --reporting-period 1s --skip-header
