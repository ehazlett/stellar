GOOS?=linux
GOARCH?=amd64
COMMIT=`git rev-parse --short HEAD`
NAMESPACE?=ehazlett
IMAGE_NAMESPACE?=$(NAMESPACE)
APP=stellar
CLI=sctl
CNI_IPAM=stellar-cni-ipam
REPO?=$(NAMESPACE)/$(APP)
TAG?=dev
BUILD?=-dev
BUILD_ARGS?=
PACKAGES=$(shell go list ./... | grep -v -e /vendor/)
EXTENSIONS=$(wildcard extensions/*)
CYCLO_PACKAGES=$(shell go list ./... | grep -v /vendor/ | sed "s/github.com\/$(NAMESPACE)\/$(APP)\///g" | tail -n +2)
CWD=$(PWD)
VNDR_ARGS=-whitelist github.com/gogo/protobuf -whitelist github.com/xenolf/lego -whitelist gopkg.in/square -whitelist github.com/ehazlett/blackbird -whitelist github.com/mholt/caddy

all: binaries

deps:
	@vndr $(VNDR_ARGS)
deps-init:
	@vndr $(VNDR_ARGS) init
	@# perform fixups
	@rm -rf vendor/github.com/gogo/protobuf/.git
	@rm -rf vendor/github.com/containerd/containerd/.git
	@rm -rf vendor/github.com/containerd/containerd/vendor
	@rm -rf vendor/github.com/xenolf/lego/.git
	@rm -rf vendor/github.com/mholt/caddy/.git
	@rm -rf vendor/github.com/ehazlett/blackbird/.git
	@rm -rf vendor/gopkg.in/square/.git

generate:
	@echo ${PACKAGES} | xargs protobuild -quiet

docker-generate:
	@echo "** This uses a separate Dockerfile (Dockerfile.dev) **"
	@docker build -t $(APP)-dev -f Dockerfile.dev .
	@docker run --rm -w /go/src/github.com/$(NAMESPACE)/$(APP) $(APP)-dev sh -c "make generate; find api -name \"*.pb.go\" | tar -T - -cf -" | tar -xvf -

docker-build: bindir
	@echo "** This uses a separate Dockerfile (Dockerfile.dev) **"
	@docker build -t $(APP)-dev -f Dockerfile.dev .
	@docker run --rm -e GOOS=${GOOS} -e GOARCH=${GOARCH} -w /go/src/github.com/$(NAMESPACE)/$(APP) $(APP)-dev sh -c "make cli daemon cni-ipam; tar -C ./bin -cf - ." | tar -C ./bin -xf -
	@echo " -> Built $(TAG) version ${COMMIT} (${GOOS}/${GOARCH})"

binaries: daemon cli cni-ipam
	@echo " -> Built $(TAG) version ${COMMIT} (${GOOS}/${GOARCH})"

bindir:
	@mkdir -p bin

cli: bindir
	@echo " -> building cli ${COMMIT}${BUILD}"
	@cd cmd/$(CLI) && CGO_ENABLED=0 go build -installsuffix cgo -ldflags "-w -X github.com/$(REPO)/version.GitCommit=$(COMMIT) -X github.com/$(REPO)/version.Build=$(BUILD)" -o ../../bin/$(CLI) .

daemon: bindir
	@echo " -> building daemon ${COMMIT}${BUILD}"
	@cd cmd/$(APP) && CGO_ENABLED=0 go build -installsuffix cgo -ldflags "-w -X github.com/$(REPO)/version.GitCommit=$(COMMIT) -X github.com/$(REPO)/version.Build=$(BUILD)" -o ../../bin/$(APP) .

cni-ipam: bindir
	@echo " -> building cni-ipam ${COMMIT}${BUILD}"
	@cd cmd/$(CNI_IPAM) && CGO_ENABLED=0 go build -installsuffix cgo -ldflags "-w -X github.com/$(REPO)/version.GitCommit=$(COMMIT) -X github.com/$(REPO)/version.Build=$(BUILD)" -o ../../bin/$(CNI_IPAM) .

docs:
	@docker build -t $(APP)-docs -f Dockerfile.docs .
	@mkdir -p _site
	@docker run --rm --entrypoint tar $(APP)-docs -C /usr/share/nginx/html -cf - . | tar -C _site -xf -

docs-netlify:
	@mkdocs build -d _site --clean

docs-serve: docs
	@echo "serving docs on http://localhost:9000"
	@docker run -ti -p 9000:80 --rm $(APP)-docs nginx -g "daemon off;" -c /etc/nginx/nginx.conf

image:
	@docker build $(BUILD_ARGS) --build-arg GOOS=$(GOOS) --build-arg GOARCH=$(GOARCH) --build-arg TAG=$(TAG) --build-arg BUILD=$(BUILD) -t $(IMAGE_NAMESPACE)/$(APP):$(TAG) -f Dockerfile .
	@echo "Image created: $(REPO):$(TAG)"

vet:
	@echo " -> $@"
	@test -z "$$(go vet ${PACKAGES} 2>&1 | tee /dev/stderr)"

lint:
	@echo " -> $@"
	@golint -set_exit_status ${PACKAGES}

cyclo:
	@echo " -> $@"
	@gocyclo -over 20 ${CYCLO_PACKAGES}

check: vet lint

test:
	@go test -short -v -cover $(TEST_ARGS) ${PACKAGES}

test-xunit:
	@2>&1 go test -short -v -cover $(TEST_ARGS) ${PACKAGES} | go2xunit -output tests.xml

test-buildkit:
	@buildctl build --frontend=dockerfile.v0 --frontend-opt filename=Dockerfile.test --local context=. --local dockerfile=. --progress plain --exporter=local --exporter-opt output=./
	@# fixup for ci
	@touch tests.xml

build-buildkit:
	@buildctl build --frontend=dockerfile.v0 --frontend-opt filename=Dockerfile.build --frontend-opt build-arg:BUILD=${BUILD} --local context=. --local dockerfile=. --progress plain --exporter=local --exporter-opt output=./build
	@chmod -R 775 ./build

release:
	@buildctl build --frontend=dockerfile.v0 --frontend-opt filename=Dockerfile.build --local context=. --local dockerfile=. --progress plain --exporter=image --exporter-opt name=docker.io/$(REPO):latest

package:
	@buildctl build --frontend=dockerfile.v0 --frontend-opt filename=Dockerfile.package --frontend-opt build-arg:BUILD=${BUILD} --local context=. --local dockerfile=. --progress plain --exporter=local --exporter-opt output=./build
	@chmod -R 775 ./build

install:
	@install -D -m 755 cmd/$(APP)/$(APP) /usr/local/bin/

vendor:
	@vndr

clean:
	@rm -rf bin/

.PHONY: generate clean docs docker-build docker-generate check test install vendor daemon cli binaries release test-buildkit build-buildkit package
