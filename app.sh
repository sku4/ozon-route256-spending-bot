#!/bin/sh

DB_STRING="host=$POSTGRES_HOST port=$POSTGRES_PORT user=$POSTGRES_USER password=$POSTGRES_PASSWORD dbname=$POSTGRES_DB_NAME sslmode=$POSTGRES_SSL"

/app/goose -dir migrations postgres "$DB_STRING" up
/app/bot