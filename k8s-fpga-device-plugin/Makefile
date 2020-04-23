GOLANG_VERSION  := 1.14.2

.PHONY: all push

REGISTRY := $(shell cat REGISTRY_URL)
IMAGE := $(shell cat DOCKERHUB_REPO)
TAG := $(shell cat DOCKERHUB_TAG)


all: ubuntu16.04

push:
	docker push "$(REGISTRY)/$(IMAGE):$(TAG)"

ubuntu%:
	@echo Building docker image $(IMAGE):$(TAG)
	docker build    --pull \
			--build-arg UBUNTU_VERSION="$*" \
			--build-arg GOLANG_VERSION="$(GOLANG_VERSION)" \
			--tag "$(REGISTRY)/$(IMAGE):$(TAG)" \
			--file docker/build_k8s-fpga-device-plugin/Dockerfile .
