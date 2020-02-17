#!/bin/bash
export GOPATH=$(pwd)

go env

go get -u github.com/jteeuwen/go-bindata/...
go get github.com/julienschmidt/httprouter
go get github.com/syndtr/goleveldb/leveldb
go get github.com/ltearno/go-wasm3
go get golang.org/x/sys/unix
go get gopkg.in/olebedev/go-duktape.v3

exit 0