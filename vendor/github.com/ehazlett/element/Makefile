PACKAGES=$(shell go list ./... | grep -v /vendor/)

protos:
	protobuild --quiet ${PACKAGES}
