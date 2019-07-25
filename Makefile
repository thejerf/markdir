.PHONY: bin install

BIN_NAME="markdir"

bin:
	@go build -o ${BIN_NAME}

install: bin
	@cp ${BIN_NAME} ~/bin/
