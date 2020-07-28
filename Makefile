all: build build-img push-img

GO111MODULE=on
export GO111MODULE
GOOS?=$(shell uname -s | tr A-Z a-z)
DOCKER_REPO?=songleo/rbac-query-proxy
DOCKER_TAG?=latest

build:
	CGO_ENABLED=0 GOOS=${GOOS} go build -a -o rbac-query-proxy main.go

build-img:
	docker build -t ${DOCKER_REPO}:${DOCKER_TAG} .

push-img:
	docker push ${DOCKER_REPO}:${DOCKER_TAG}