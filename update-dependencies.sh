#!/bin/bash
export GOPATH=$(pwd)

go env

go get github.com/jteeuwen/go-bindata/...
go get github.com/syndtr/goleveldb/leveldb
go get github.com/ltearno/go-wasm3
go get gopkg.in/ltearno/go-duktape.v3
go get github.com/gorilla/websocket
go get github.com/golang-collections/go-datastructures/queue
#go get -u golang.org/x/sys
git clone https://github.com/golang/sys.git --depth=1 src/golang.org/x/sys
go get github.com/rs/xid
go get github.com/lestrrat-go/jwx
go get github.com/lestrrat-go/iter/arrayiter

exit 0