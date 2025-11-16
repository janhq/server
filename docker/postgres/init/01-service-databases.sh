#!/bin/bash
set -e

# Bootstrap additional service databases and users for media-api and response-api

echo "Creating service users and databases..."

# Create users if they don't exist
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    DO \$\$
    BEGIN
        IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'media') THEN
            CREATE ROLE media LOGIN PASSWORD 'media';
        END IF;
    END
    \$\$;

    DO \$\$
    BEGIN
        IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'response_api') THEN
            CREATE ROLE response_api LOGIN PASSWORD 'response_api';
        END IF;
    END
    \$\$;
EOSQL

# Create media_api database if it doesn't exist
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    SELECT 'CREATE DATABASE media_api OWNER media'
    WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'media_api')\gexec
EOSQL

# Create response_api database if it doesn't exist
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    SELECT 'CREATE DATABASE response_api OWNER response_api'
    WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'response_api')\gexec
EOSQL

# Grant privileges
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    GRANT ALL PRIVILEGES ON DATABASE media_api TO media;
    GRANT ALL PRIVILEGES ON DATABASE response_api TO response_api;
EOSQL

echo "Service databases created successfully"
