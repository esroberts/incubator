#!/bin/sh

export SRC_DIR=`pwd`

# source_relative ignores go_package in .proto files.
# parent directory of .proto file is implicit in relative output path
protoc -I. --go_out=paths=source_relative:. internal/proto/message.proto