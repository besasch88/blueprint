#!/bin/sh

# This script run all the migrations. It is used in development environment and it contains hardcoded parameters to connect to DB.
migrate -path "./scripts/migrations" -database "postgres://blueprint:blueprint@127.0.0.1:54323/blueprint?sslmode=disable" up
