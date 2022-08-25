
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
	go test -v ./xopjson/... -run TestASingleLine
	go test -v ./xopjson/... -run TestParameters -failfast
	go test ./... -failfast
	go test -race ./... -failfast

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

calculate_coverage:
	echo "mode: atomic" > coverage.txt
	for d in $$(go list ./...); do \
	  go test -race -covermode=atomic -coverprofile=profile.out -coverpkg=github.com/muir/xop-go/... $$d; \
	  if [ -f profile.out ]; then \
	    grep -v ^mode profile.out >> coverage.txt; \
	    rm profile.out; \
	  fi; \
	done


coverage: calculate_coverage
	go tool cover -html=coverage.txt

golanglint:
	# binary will be $(go env GOPATH)/bin/golangci-lint
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin v1.45.2
	golangci-lint --version

lint:;
	golangci-lint run
	
