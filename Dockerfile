FROM golang:1.16-alpine

RUN apk add bash make git gcc g++ cmake linux-headers

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

RUN apk add mesa-dev glu

RUN make build

RUN mkdir /data
ADD tls.key.pem /data/
ADD tls.cert.pem /data/

ARG UID=1000
ARG GID=1000

RUN addgroup --gid ${GID} my-own-cluster-group || echo "group ${GID} already exists"
RUN adduser --uid ${UID} -G my-own-cluster-group my-own-cluster-user || echo "user ${UID} already exists"

RUN chown -R ${UID}:${GID} /my-own-cluster
RUN chown -R ${UID}:${GID} /data

USER ${UID}

ENTRYPOINT [ "/my-own-cluster/my-own-cluster", "serve", "/data" ]
