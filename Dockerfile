FROM golang:1.13-alpine

RUN apk add bash make git gcc g++ cmake

RUN mkdir /.cache && chmod ugo+rw /.cache

WORKDIR /wasm3
RUN git clone https://github.com/ltearno/wasm3.git
WORKDIR /wasm3/wasm3
RUN mkdir build && cd build && pwd && cmake .. && make

WORKDIR /my-own-cluster

ADD Makefile ./
ADD src/my-own-cluster ./src/my-own-cluster
ADD assets ./assets/
ADD update-dependencies.sh ./

RUN make build-prepare
RUN cp /wasm3/wasm3/build/source/libm3.a /my-own-cluster/src/github.com/ltearno/go-wasm3/lib/linux/

RUN make build-embed-assets

RUN make build

RUN chown -R 1000:1000 /my-own-cluster

RUN mkdir /data
ADD tls.key.pem /data/
ADD tls.cert.pem /data/

ENTRYPOINT [ "/my-own-cluster/my-own-cluster", "serve", "/data" ]
