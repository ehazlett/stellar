GOOS?=linux
GOARCH?=amd64
COMMIT=`git rev-parse --short HEAD`
NAMESPACE?=ehazlett
IMAGE_NAMESPACE?=$(NAMESPACE)
APP=element
REPO?=$(NAMESPACE)/$(APP)
TAG?=dev
BUILD?=-dev
BUILD_ARGS?=
PACKAGES=$(shell go list ./... | grep -v -e /vendor/ -e /test/rttf-image/)
EXTENSIONS=$(wildcard extensions/*)
CYCLO_PACKAGES=$(shell go list ./... | grep -v /vendor/ | sed "s/github.com\/$(NAMESPACE)\/$(APP)\///g" | tail -n +2)
CWD=$(PWD)

all: binary

deps:
	@vndr -whitelist github.com/gogo/protobuf

generate:
	@echo ${PACKAGES} | xargs protobuild

docker-generate:
	@echo "** This uses a separate Dockerfile (Dockerfile.build) **"
	@docker build -t $(APP)-dev -f Dockerfile.build.$(GOOS).$(GOARCH) .
	@docker run -ti --rm -v $(PWD):/go/src/github.com/$(NAMESPACE)/$(APP) $(APP)-dev ash -c "echo ${PACKAGES} | xargs /go/bin/protobuild"

binary:
	@echo " -> Building $(TAG)"
	@cd cmd/$(APP) && go build -v -a -tags "netgo static_build" -installsuffix netgo -ldflags "-w -X github.com/$(REPO)/version.GitCommit=$(COMMIT) -X github.com/$(REPO)/version.Build=$(BUILD)" .
	@echo "Built ${APP} version ${COMMIT} (${GOOS}/${GOARCH})"

image:
	@docker build $(BUILD_ARGS) --build-arg GOOS=$(GOOS) --build-arg GOARCH=$(GOARCH) --build-arg TAG=$(TAG) --build-arg BUILD=$(BUILD) -t $(IMAGE_NAMESPACE)/$(APP):$(TAG) -f Dockerfile.$(GOOS).$(GOARCH) .
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

clean:
	@rm cmd/$(APP)/$(APP)

.PHONY: binary generate clean check test
