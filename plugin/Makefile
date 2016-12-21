.PHONY: all setup plugin test

export BASE?=$(PWD)/..
export DST?=$(GOPATH)/src/github.com/komand/plugin-sdk-go


setup:
	test -d $(DST) || ln -s $(BASE) $(DST)

all: setup plugin

plugin:
	cd $(DST)/plugin && go get -v ./...
	cd $(DST)/plugin && go build .

test:
	cd $(DST)/plugin && go list ./... | grep -v /vendor/ | go test -v
	cd $(DST)/test/cache/ && go run
