#!/bin/bash

binDir="./bin"
mainDir="./cmd"

if [[ ! -d "$binDir" ]]; then
  mkdir $binDir
fi

echo "Building - Windows amd64"
# Windows amd64 build
env GOOS=windows GOARCH=amd64 go build -o $binDir/server-backup-win-amd64.exe $mainDir

echo "Building - Linux amd64"
# Linux amd64 build
env GOOS=linux GOARCH=amd64 go build -o $binDir/server-backup-linux-amd64 $mainDir

echo "Building - Linux arm64"
# Linux arm64 build
env GOOS=linux GOARCH=arm64 go build -o $binDir/server-backup-linux-arm64 $mainDir

echo "Building - Mac amd64"
# Mac amd64 build
env GOOS=darwin GOARCH=amd64 go build -o $binDir/server-backup-mac-amd64 $mainDir

echo "Building - Mac arm64"
# Mac arm64 build
env GOOS=darwin GOARCH=arm64 go build -o $binDir/server-backup-mac-arm64 $mainDir

echo "Completed"
