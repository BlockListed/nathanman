#!/bin/bash
PATH=/usr/bin

mkdir -p dist

docker build -t gitea.box/aaron/nathanman:latest .
docker push gitea.box/aaron/nathanman:latest
