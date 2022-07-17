
GOHOME ?= ${HOME}
GOPATH ?= ${GOHOME}
GOBIN ?= ${GOPATH}/bin

ZZZGO = $(wildcard *.zzzgo */*.zzzgo */*/*.zzzgo)
GENERATED = $(patsubst %.zzzgo, %.go, $(ZZZGO))
TOOLS = ${GOBIN}/gofumpt ${GOBIN}/goimports

all:	$(GENERATED)
	go build ./...

test:	$(GENERATED)
	go test ./...

${GOBIN}/gofumpt:;
	go install mvdan.cc/gofumpt@latest

${GOBIN}/goimports:;
	go install golang.org/x/tools/cmd/goimports@latest

%.go : %.zzzgo tools/xopzzz/xopzzz.go $(TOOLS) Makefile
	go run tools/xopzzz/xopzzz.go < $< > $@.tmp
	-chmod +w $@
	gofumpt < $@.tmp | goimports > $@
	-chmod -w $@
	-rm $@.tmp
