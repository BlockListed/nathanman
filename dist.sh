#!/bin/bash
PATH=/usr/bin

mkdir -p dist

docker build -t nathanman:latest .
docker save nathanman:latest > dist/nathanman.docker.tar
cp -l dist/nathanman.docker.tar ~/Documents/dev/Ansible/Server/files/discord