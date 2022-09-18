SHELL := /bin/bash

update:
	go get -u
	go mod tidy

test:
	go test -v -cover $(shell go list ./... | grep -v /vendor/ | grep -v /build/) -coverprofile=coverage-pkg.txt -covermode=atomic
