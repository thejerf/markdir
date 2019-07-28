.PHONY: build install

BIN_NAME="markdir"

build:
	@go build -o ${BIN_NAME}

install:
	packr2 install
