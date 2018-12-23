FROM golang:1.11 AS build

ARG GOOS
ARG GOARCH
RUN apt-get update && apt-get install -y git make curl build-essential git autoconf automake libtool unzip file bzr
RUN git clone https://github.com/google/protobuf /tmp/protobuf && \
    cd /tmp/protobuf && \
    ./autogen.sh && \
    ./configure && make install
RUN go get -v github.com/LK4D4/vndr
RUN go get -v github.com/golang/protobuf/protoc-gen-go
RUN go get -v github.com/gogo/protobuf/protoc-gen-gofast
RUN go get -v github.com/gogo/protobuf/proto
RUN go get -v github.com/gogo/protobuf/gogoproto
RUN go get -v github.com/gogo/protobuf/protoc-gen-gogo
RUN go get -v github.com/gogo/protobuf/protoc-gen-gogofast
RUN go get -v github.com/stevvooe/protobuild
RUN go get -v golang.org/x/lint/golint
RUN go get -v github.com/tebeka/go2xunit

FROM build as stellar
ENV APP stellar
ENV REPO ehazlett/$APP
ARG BUILD
COPY . /go/src/github.com/$REPO
WORKDIR /go/src/github.com/$REPO
RUN make
RUN date > /release.txt
RUN git rev-parse HEAD >> /release.txt

FROM build as runc
RUN apt-get update && apt-get install -y libseccomp-dev
RUN git clone https://github.com/opencontainers/runc /go/src/github.com/opencontainers/runc
WORKDIR /go/src/github.com/opencontainers/runc
RUN git reset --hard 20aff4f0488c6d4b8df4d85b4f63f1f704c11abd
RUN make

FROM build as containerd
RUN apt-get update && apt-get install -y libseccomp-dev btrfs-tools
RUN git clone https://github.com/containerd/containerd /go/src/github.com/containerd/containerd
WORKDIR /go/src/github.com/containerd/containerd
RUN make

FROM build as buildkit
RUN git clone https://github.com/moby/buildkit /go/src/github.com/moby/buildkit
WORKDIR /go/src/github.com/moby/buildkit
RUN mkdir .tmp; \
    PKG=github.com/moby/buildkit VERSION=$(git describe --match 'v[0-9]*' --dirty='.m' --always) REVISION=$(git rev-parse HEAD)$(if ! git diff --no-ext-diff --quiet --exit-code; then echo .m; fi); \
    echo "-X ${PKG}/version.Version=${VERSION} -X ${PKG}/version.Revision=${REVISION} -X ${PKG}/version.Package=${PKG}" | tee .tmp/ldflags

FROM buildkit as buildctl
ENV CGO_ENABLED=0
RUN go build -ldflags "$(cat .tmp/ldflags) -d" -o /usr/bin/buildctl ./cmd/buildctl

FROM buildkit as buildkitd
ENV CGO_ENABLED=1
RUN go build -installsuffix netgo -ldflags "$(cat .tmp/ldflags) -w -extldflags -static" -tags 'seccomp netgo cgo static_build' -o /usr/bin/buildkitd ./cmd/buildkitd


FROM build as cni
RUN apt-get update && apt-get install -y build-essential
RUN git clone https://github.com/containernetworking/plugins /go/src/github.com/containernetworking/plugins
WORKDIR /go/src/github.com/containernetworking/plugins
RUN ./build_linux.sh

FROM scratch as rootfs
COPY --from=stellar /release.txt /
COPY --from=stellar /go/src/github.com/ehazlett/stellar/bin/sctl /usr/local/bin/
COPY --from=stellar /go/src/github.com/ehazlett/stellar/bin/stellar /usr/local/bin/
COPY --from=stellar /go/src/github.com/ehazlett/stellar/bin/stellar-cni-ipam /opt/containerd/bin/
COPY --from=stellar /go/src/github.com/ehazlett/stellar/contrib/containerd.service /etc/systemd/system/
COPY --from=stellar /go/src/github.com/ehazlett/stellar/contrib/buildkit.service /etc/systemd/system/
COPY --from=stellar /go/src/github.com/ehazlett/stellar/contrib/stellar.service /etc/systemd/system/
COPY --from=stellar /go/src/github.com/ehazlett/stellar/contrib/stellar.conf /etc/stellar.conf
COPY --from=runc /go/src/github.com/opencontainers/runc/runc /usr/local/bin/
COPY --from=containerd /go/src/github.com/containerd/containerd/bin/ctr /usr/local/bin/
COPY --from=containerd /go/src/github.com/containerd/containerd/bin/containerd /usr/local/bin/
COPY --from=containerd /go/src/github.com/containerd/containerd/bin/containerd-shim /usr/local/bin/
COPY --from=containerd /go/src/github.com/containerd/containerd/bin/containerd-shim-runc-v1 /usr/local/bin/
COPY --from=buildctl /usr/bin/buildctl /usr/local/bin/
COPY --from=buildkitd /usr/bin/buildkitd /usr/local/bin/
COPY --from=cni /go/src/github.com/containernetworking/plugins/bin/bridge /opt/containerd/bin/
COPY --from=cni /go/src/github.com/containernetworking/plugins/bin/loopback /opt/containerd/bin/

FROM build as release
COPY --from=rootfs / /package
WORKDIR /package
RUN tar czvf /stellar.tar.gz .
RUN chmod 777 /stellar.tar.gz

FROM scratch
COPY --from=release /stellar.tar.gz /
