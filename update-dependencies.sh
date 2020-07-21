#!/bin/bash
export GOPATH=$(pwd)

go env

go get -u golang.org/x/sys
go get -u github.com/jteeuwen/go-bindata/...
go get github.com/syndtr/goleveldb/leveldb
go get github.com/ltearno/go-wasm3
go get gopkg.in/ltearno/go-duktape.v3
go get github.com/gorilla/websocket
go get github.com/golang-collections/go-datastructures/queue

exit 0