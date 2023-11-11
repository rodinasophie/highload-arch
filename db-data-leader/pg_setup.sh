#!/bin/bash
set -e
cp /etc/my_postgresql.conf ${PGDATA}/postgresql.conf
cp /etc/my_pg_hba.conf ${PGDATA}/pg_hba.conf
su - -c '/usr/lib/postgresql/14/bin/pg_ctl -D /data restart' postgres