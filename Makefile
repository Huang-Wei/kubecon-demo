IMAGE_NAME = kubecon-demo
TAG := $(shell cat VERSION)

COMMONENVVAR = GOOS=$(shell uname -s | tr A-Z a-z) GOARCH=$(subst x86_64,amd64,$(patsubst i%86,386,$(shell uname -m)))
BUILDENVVAR = CGO_ENABLED=0

all: build

deps:
	dep ensure

# test:
# 	$(COMMONENVVAR) $(BUILDENVVAR) go test ./... -v

build:
	$(COMMONENVVAR) $(BUILDENVVAR) go build -o main *.go
	$(COMMONENVVAR) $(BUILDENVVAR) go build -o poker/main poker/*.go

image: build
	docker build -t hweicdl/$(IMAGE_NAME):$(TAG) .
	docker build -t hweicdl/poker:$(TAG) ./poker/

push: image
	docker login -u $(DOCKER_USERNAME) -p $(DOCKER_PASSWORD)
	docker push hweicdl/$(IMAGE_NAME):$(TAG)

.PHONY: all deps test build image push
