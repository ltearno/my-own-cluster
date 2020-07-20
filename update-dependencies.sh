#!/bin/bash
export GOPATH=$(pwd)

go env

go get -u github.com/jteeuwen/go-bindata/...
go get github.com/syndtr/goleveldb/leveldb
go get github.com/ltearno/go-wasm3
go get gopkg.in/ltearno/go-duktape.v3
go get golang.org/x/sys/unix
go get github.com/gorilla/websocket

exit 0