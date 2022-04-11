#!/usr/bin/env bash
docker build -t gcr.io/coresystem-171219/c24-media:initial . --no-cache
docker push gcr.io/coresystem-171219/c24-media:initial