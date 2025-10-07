#!/bin/bash

# Set environment variables
export POSTGRES_USER=envini
export POSTGRES_PASSWORD=envini
export POSTGRES_DB=envini
export POSTGRES_PORT=5433

# Run PostgreSQL with custom configuration
docker run --rm \
  --name envini-secret-postgres \
  -e POSTGRES_USER=$POSTGRES_USER \
  -e POSTGRES_PASSWORD=$POSTGRES_PASSWORD \
  -e POSTGRES_DB=$POSTGRES_DB \
  -e POSTGRES_PORT=$POSTGRES_PORT \
  -p 5433:5433 \
  -v $(pwd)/postgresql.conf:/etc/postgresql/postgresql.conf \
  postgres:16-alpine \
  postgres -c config_file=/etc/postgresql/postgresql.conf 