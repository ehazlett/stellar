GOOS?=linux
GOARCH?=amd64
COMMIT=`git rev-parse --short HEAD`
NAMESPACE?=ehazlett
IMAGE_NAMESPACE?=$(NAMESPACE)
APP=blackbird
CLI=bctl
REPO?=$(NAMESPACE)/$(APP)
TAG?=dev
BUILD?=-dev
BUILD_ARGS?=
PACKAGES=$(shell go list ./... | grep -v -e /vendor/)
EXTENSIONS=$(wildcard extensions/*)
CYCLO_PACKAGES=$(shell go list ./... | grep -v /vendor/ | sed "s/github.com\/$(NAMESPACE)\/$(APP)\///g" | tail -n +2)
VNDR_ARGS=-whitelist github.com/gogo/protobuf -whitelist github.com/xenolf/lego -whitelist gopkg.in/square -whitelist github.com/mholt/caddy
CWD=$(PWD)

all: binaries

deps:
	@vndr $(VNDR_ARGS)

deps-init:
	@vndr $(VNDR_ARGS) init
	@rm -rf vendor/github.com/gogo/protobuf/.git
	@rm -rf vendor/github.com/xenolf/lego/.git
	@rm -rf vendor/github.com/mholt/caddy/.git
	@rm -rf vendor/gopkg.in/square/.git

generate:
	@echo ${PACKAGES} | xargs protobuild -quiet

docker-generate:
	@echo "** This uses a separate Dockerfile (Dockerfile.build) **"
	@docker build -t $(APP)-dev -f Dockerfile.build .
	@docker run --rm -w /go/src/github.com/$(NAMESPACE)/$(APP) $(APP)-dev sh -c "make generate; find api -name \"*.pb.go\" | tar -T - -cf -" | tar -xvf -

docker-build: bindir
	@echo "** This uses a separate Dockerfile (Dockerfile.build) **"
	@docker build -t $(APP)-dev -f Dockerfile.build .
	@docker run --rm -e GOOS=${GOOS} -e GOARCH=${GOARCH} -w /go/src/github.com/$(NAMESPACE)/$(APP) $(APP)-dev sh -c "make cli daemon cni-ipam; tar -C ./bin -cf - ." | tar -C ./bin -xf -
	@echo " -> Built $(TAG) version ${COMMIT} (${GOOS}/${GOARCH})"

binaries: daemon cli
	@echo " -> Built $(TAG) version ${COMMIT} (${GOOS}/${GOARCH})"

bindir:
	@mkdir -p bin

daemon: bindir
	@cd cmd/$(APP) && CGO_ENABLED=0 go build -installsuffix cgo -ldflags "-w -X github.com/$(REPO)/version.GitCommit=$(COMMIT) -X github.com/$(REPO)/version.Build=$(BUILD)" -o ../../bin/$(APP) .

cli: bindir
	@cd cmd/$(CLI) && CGO_ENABLED=0 go build -installsuffix cgo -ldflags "-w -X github.com/$(REPO)/version.GitCommit=$(COMMIT) -X github.com/$(REPO)/version.Build=$(BUILD)" -o ../../bin/$(CLI) .

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

check: vet lint

test:
	@go test -short -v -cover $(TEST_ARGS) ${PACKAGES}

install:
	@install -D -m 755 cmd/$(APP)/$(APP) /usr/local/bin/

vendor:
	@vndr

clean:
	@rm -rf bin/

.PHONY: generate clean docs docker-build docker-generate check test install vendor daemon cli binaries
