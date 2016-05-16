GO   := GO15VENDOREXPERIMENT=1 go

ifdef DEBUG
        bindata_flags = -debug
endif


all: clean build

build:
	./scripts/build.sh


.PHONY: clean
clean:
	-rm ./bin/*
