#!/bin/bash

# This script run the application, migrations and external services locally. Good to test the application locally
echo "## Starting dev containers ##"
cd ../
cd build
exec docker compose -p bp-dev -f docker-compose.yaml up --build --force-recreate --remove-orphans "$@"
