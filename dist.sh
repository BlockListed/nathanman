#!/bin/bash
PATH=/usr/bin

mkdir -p dist

docker build -t registry.box/nathanman:latest .
docker push registry.box/nathanman:latest