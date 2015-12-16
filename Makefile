GO   := GO15VENDOREXPERIMENT=1 go

ifdef DEBUG
        bindata_flags = -debug
endif


all: build

build:
	./scripts/build.sh
