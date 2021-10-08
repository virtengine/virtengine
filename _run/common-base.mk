include $(abspath $(dir $(lastword $(MAKEFILE_LIST)))/../make/init.mk)

VIRTENGINE_RUN_NAME := $(notdir $(CURDIR))

VIRTENGINE_HOME ?= $(VIRTENGINE_RUN)/$(VIRTENGINE_RUN_NAME)

.PHONY: all
all:
	(cd "$(VIRTENGINE_ROOT)" && make all)

.PHONY: bins
bins:
	(cd "$(VIRTENGINE_ROOT)" && make bins)

.PHONY: virtengine
virtengine:
	(cd "$(VIRTENGINE_ROOT)" && make)

.PHONY: image-minikube
image-minikube:
	(cd "$(VIRTENGINE_ROOT)" && make image-minikube)
