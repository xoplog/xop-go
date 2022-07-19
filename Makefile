
GOHOME ?= ${HOME}
GOPATH ?= ${GOHOME}
GOBIN ?= ${GOPATH}/bin

ZZZGO = $(wildcard *.zzzgo */*.zzzgo */*/*.zzzgo)
ZZZGENERATED = $(patsubst %.zzzgo, %.go, $(ZZZGO))
TOOLS = ${GOBIN}/gofumpt ${GOBIN}/goimports ${GOBIN}/enumer

all:	$(ZZZGENERATED) 
	go generate ./...
	go build ./...

test:	$(ZZZGENERATED)
	go generate ./...
	go test ./...

${GOBIN}/gofumpt:;
	go install mvdan.cc/gofumpt@latest

${GOBIN}/goimports:;
	go install golang.org/x/tools/cmd/goimports@latest

${GOBIN}/enumer:;
	go install github.com/dmarkham/enumer@latest

%.go : %.zzzgo tools/xopzzz/xopzzz.go $(TOOLS) Makefile
	go run tools/xopzzz/xopzzz.go < $< > $@.tmp
	-chmod +w $@
	gofumpt < $@.tmp | goimports > $@
	-chmod -w $@
	-rm $@.tmp
