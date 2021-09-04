#! /usr/local/bin/bash

DOT_FILE=.env-test
source .env-test

if createdb $DB_NAME 2>/dev/null ; then
  echo "created db $DB_NAME"
else
  echo "DB $DB_NAME already exists"
fi

if createuser $DB_USER 2>/dev/null; then
  echo "Created user $DB_USER"
else
  echo "User $DB_USER exists, which is fine."
fi
psql <<SQL
select now();
alter user $DB_USER password '$DB_PASSWD';
grant all privileges on database $DB_NAME to $DB_USER;


SQL

