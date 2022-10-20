.PHONY: build clean

BINARY="aqmp-pipeline"

.PHONY: build
build:
	GOOS=linux GOARCH="amd64" go build -o ${BINARY} ./main.go

.PHONY: install
install:
	@govendor sync -v

.PHONY: clean
	@if [ -f ${BINARY} ] ; then rm ${BINARY} ; fi
