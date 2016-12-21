VERSION=$(shell git describe --abbrev=0 --tags)
MAJOR_VERSION=$(shell git describe --abbrev=0 --tags | cut -d"." -f1-2)

setup:
	make -C plugin setup

all: setup plugin

plugin:
	make -C plugin plugin

test:
	make -C plugin test
	
image:
	docker build -t komand/go-plugin .

tag: image
	@echo version is $(VERSION)
	docker tag komand/go-plugin komand/go-plugin:$(VERSION)
	docker tag komand/go-plugin:$(VERSION) komand/go-plugin:$(MAJOR_VERSION)


.PHONY: setup all test image plugin tag
