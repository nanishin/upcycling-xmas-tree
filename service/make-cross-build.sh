#!/bin/sh

rm upcycling-xmas-tree-service
rm -rf pkg src

CURRENT_PATH=`pwd`

# If proxy is used
#export http_proxy="http://x.x.x.x:x"
#export https_proxy="http://x.x.x.x:x"

export GOPATH=$CURRENT_PATH

# Target OS
export GOOS=linux

# Target Architecture
export GOARCH=arm

go get -v github.com/kellydunn/go-opc

go build upcycling-xmas-tree-service.go
