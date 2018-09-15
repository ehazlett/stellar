FROM golang:1.10 AS build

RUN apt-get update && apt-get install -y --no-install-recommends \
    git make curl build-essential bash git autoconf automake libtool unzip file ca-certificates
RUN git clone https://github.com/google/protobuf /tmp/protobuf && \
    cd /tmp/protobuf && \
    git checkout 3.5.x && \
    ./autogen.sh && \
    ./configure && make install
RUN ldconfig
RUN go get -v github.com/LK4D4/vndr
RUN go get -v github.com/golang/protobuf/protoc-gen-go
RUN go get -v github.com/gogo/protobuf/protoc-gen-gofast
RUN go get -v github.com/gogo/protobuf/proto
RUN go get -v github.com/gogo/protobuf/gogoproto
RUN go get -v github.com/gogo/protobuf/protoc-gen-gogo
RUN go get -v github.com/gogo/protobuf/protoc-gen-gogofast
RUN go get -v github.com/stevvooe/protobuild
RUN go get -v github.com/golang/lint/golint
WORKDIR /go/src/github.com/ehazlett/blackbird
COPY . /go/src/github.com/ehazlett/blackbird
RUN make

FROM scratch
COPY --from=build /go/src/github.com/ehazlett/blackbird/bin/* /bin/
