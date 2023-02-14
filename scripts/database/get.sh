#!/bin/bash
export MIGRATE_VERSION=v4.15.2
export PLATFORM=linux

echo https://github.com/golang-migrate/migrate/releases/download/$MIGRATE_VERSION/migrate.$PLATFORM-amd64.tar.gz
wget -P ./bin https://github.com/golang-migrate/migrate/releases/download/$MIGRATE_VERSION/migrate.$PLATFORM-amd64.tar.gz
tar -xvf ./bin/migrate.linux-amd64.tar.gz -C ./bin

