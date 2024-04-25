ONIGMO_VERSION ?= 6.1.3
ONIG_REPOSITORY := https://github.com/k-takata/Onigmo
BASE_PATH := $(shell pwd)
BUILD_PATH := $(BASE_PATH)/.build

.PHONY: install-onigmo clean

$(BUILD_PATH):
	mkdir -p $(BUILD_PATH)

install-onigmo: $(BUILD_PATH)
	cd ${BUILD_PATH} && \
	wget ${ONIG_REPOSITORY}/releases/download/Onigmo-${ONIGMO_VERSION}/onigmo-${ONIGMO_VERSION}.tar.gz && \
	tar -xvzf onigmo-${ONIGMO_VERSION}.tar.gz && \
	cd onigmo-${ONIGMO_VERSION} && \
	./configure --prefix=/usr && make && sudo make install

clean:
	rm -rf $(BUILD_PATH)