GOOS?=linux
GOARCH?=amd64
COMMIT=`git rev-parse --short HEAD`
NAMESPACE?=ehazlett
IMAGE_NAMESPACE?=$(NAMESPACE)
APP=stellar
CLI=sctl
REPO?=$(NAMESPACE)/$(APP)
TAG?=dev
BUILD?=-dev
BUILD_ARGS?=
PACKAGES=$(shell go list ./... | grep -v -e /vendor/)
EXTENSIONS=$(wildcard extensions/*)
CYCLO_PACKAGES=$(shell go list ./... | grep -v /vendor/ | sed "s/github.com\/$(NAMESPACE)\/$(APP)\///g" | tail -n +2)
CWD=$(PWD)

all: binaries

deps:
	@vndr -whitelist github.com/gogo/protobuf

generate:
	@echo ${PACKAGES} | xargs protobuild

docker-generate:
	@echo "** This uses a separate Dockerfile (Dockerfile.build) **"
	@docker build -t $(APP)-dev -f Dockerfile.build .
	@docker run -ti --rm -w /go/src/github.com/$(NAMESPACE)/$(APP) -v $(PWD):/go/src/github.com/$(NAMESPACE)/$(APP) $(APP)-dev sh -c "make generate"

docker-build:
	@echo "** This uses a separate Dockerfile (Dockerfile.build) **"
	@docker build -t $(APP)-dev -f Dockerfile.build .
	@docker run -ti --rm -w /go/src/github.com/$(NAMESPACE)/$(APP) -v $(PWD):/go/src/github.com/$(NAMESPACE)/$(APP) $(APP)-dev sh -c "make binaries"

binaries: daemon cli

cli:
	@echo " -> Building cli $(TAG) version ${COMMIT} (${GOOS}/${GOARCH})"
	@cd cmd/$(CLI) && CGO_ENABLED=0 go build -a -installsuffix cgo -ldflags "-w -X github.com/$(REPO)/version.GitCommit=$(COMMIT) -X github.com/$(REPO)/version.Build=$(BUILD)" .

daemon:
	@echo " -> Building daemon $(TAG) version ${COMMIT} (${GOOS}/${GOARCH})"
	@cd cmd/$(APP) && CGO_ENABLED=0 go build -a -installsuffix cgo -ldflags "-w -X github.com/$(REPO)/version.GitCommit=$(COMMIT) -X github.com/$(REPO)/version.Build=$(BUILD)" .

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

install:
	@install -D -m 755 cmd/$(APP)/$(APP) /usr/local/bin/

vendor:
	@vndr

clean:
	@rm cmd/$(APP)/$(APP)

.PHONY: generate clean docs docker-build docker-generate check test install vendor daemon cli binaries
