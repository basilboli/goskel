#! /bin/bash
# This script is used to import sample data into the MongoDB database on the start of the container.

mongoimport --host mongo-ci-test --db db-test --collection jobs --type json --file /mongo-seed/jobs.json --jsonArray
mongoimport --host mongo-ci-test --db db-test --collection data --type json --file /mongo-seed/data.json --jsonArray

sleep infinity
