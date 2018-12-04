FROM golang:1.10-alpine AS build

ARG GOOS
ARG GOARCH
RUN apk add -U git make curl build-base bash git autoconf automake libtool unzip file
RUN git clone https://github.com/google/protobuf /tmp/protobuf && \
    cd /tmp/protobuf && \
    ./autogen.sh && \
    ./configure && make install
RUN go get -v github.com/LK4D4/vndr \
    go get -u github.com/golang/dep/cmd/dep \
    go get -v github.com/golang/protobuf/protoc-gen-go \
    go get -v github.com/gogo/protobuf/protoc-gen-gofast \
    go get -v github.com/gogo/protobuf/proto \
    go get -v github.com/gogo/protobuf/gogoproto \
    go get -v github.com/gogo/protobuf/protoc-gen-gogo \
    go get -v github.com/gogo/protobuf/protoc-gen-gogofast \
    go get -v github.com/stevvooe/protobuild \
    go get -v github.com/golang/lint/golint
ENV APP radiant
ENV REPO stellarproject/$APP
COPY . /go/src/github.com/$REPO
WORKDIR /go/src/github.com/$REPO
