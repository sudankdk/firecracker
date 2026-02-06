SHELL := /bin/bash
.SHELLFLAGS := -eu -o pipefail -c

RELEASE_URL := https://github.com/firecracker-microvm/firecracker/releases
ARCH := $(shell uname -m)
LATEST := $(shell curl -fsSLI -o /dev/null -w '%{url_effective}' $(RELEASE_URL)/latest | xargs basename)

TARBALL := firecracker-$(LATEST)-$(ARCH).tgz
RELEASE_DIR := release-$(LATEST)-$(ARCH)
BINARY := firecracker
KERNEL_PATH := $(PWD)/hello-vmlinux.bin
ROOTFS_PATH := $(PWD)/hello-rootfs.ext4

.PHONY: all kvm-perms download extract install clean

all: install

kvm-perms:
	sudo setfacl -m u:$(USER):rw /dev/kvm

download:
	curl -L $(RELEASE_URL)/download/$(LATEST)/$(TARBALL) -o $(TARBALL)

extract: download
	tar -xzf $(TARBALL)

install: extract
	mv $(RELEASE_DIR)/firecracker-$(LATEST)-$(ARCH) $(BINARY)
	chmod +x $(BINARY)

clean:
	rm -rf $(TARBALL) $(RELEASE_DIR) $(BINARY)

run:
	./firecracker \
	  --kernel-path $(KERNEL_PATH) \
	  --rootfs-path $(ROOTFS_PATH)
