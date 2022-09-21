
GOHOME ?= ${HOME}
GOPATH ?= ${GOHOME}
GOBIN ?= ${GOPATH}/bin

ZZZGO = $(wildcard *.zzzgo */*.zzzgo */*/*.zzzgo)
ZZZGENERATED = $(patsubst %.zzzgo, %.go, $(ZZZGO))
TOOLS = ${GOBIN}/gofumpt ${GOBIN}/goimports ${GOBIN}/enumer
TEST_ONLY =?

all:	$(ZZZGENERATED) 
	go generate ./...
	go build ./...

test:	$(ZZZGENERATED)
	go generate ./...
	go test -v ./xopjson/... -run TestASingleLine
	go test -v ./xopjson/... -tags xoptesting -run TestParameters -failfast $(TEST_ONLY)
	go test -tags xoptesting ./... -failfast $(TEST_ONLY)
	go test -tags xoptesting -race ./... -failfast $(TEST_ONLY)
	XOPLEVEL_xoptestutil=warn go test -tags xoptesting ./xoptest/xoptestutil -run TestAdjustedLevelLogger 
	XOPLEVEL_xoptestutil=debug go test -tags xoptesting ./xoptest/xoptestutil -run TestAdjustedLevelLogger 

${GOBIN}/gofumpt:;
	go install mvdan.cc/gofumpt@latest

${GOBIN}/goimports:;
	go install golang.org/x/tools/cmd/goimports@latest

${GOBIN}/enumer:;
	go install github.com/dmarkham/enumer@latest

%.go : %.zzzgo tools/xopzzz/xopzzz.go $(TOOLS) Makefile
	go run tools/xopzzz/xopzzz.go < $< > $@.tmp
	-chmod +w $@
	gofumpt -w $@.tmp 
	goimports -w $@.tmp
	-mv $@.tmp $@
	-chmod -w $@

calculate_coverage:
	echo "mode: atomic" > coverage.txt
	for d in $$(go list ./...); do \
	  go test -race -covermode=atomic -coverprofile=profile.out -coverpkg=github.com/muir/xop-go/... $$d; \
	  if [ -f profile.out ]; then \
	    grep -v ^mode profile.out >> coverage.txt; \
	    rm profile.out; \
	  fi; \
	done
	XOPLEVEL_xoptestutil=debug go test -covermode=atomic -tags xoptesting -coverprofile=profile.out -coverpkg=github.com/muir/xop-go/... ./xoptest/xoptestutil -run TestAdjustedLevelLogger 
	if [ -f profile.out ]; then \
	  grep -v ^mode profile.out >> coverage.txt; \
	  rm profile.out; \
	fi

coverage: calculate_coverage
	go tool cover -html=coverage.txt

golanglint:
	# binary will be $(go env GOPATH)/bin/golangci-lint
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin v1.48.0
	golangci-lint --version

lint:;
	golangci-lint run
	
misspell:;
	go install github.com/client9/misspell/cmd/misspell@latest
	misspell -w `find . -name \*.md`

OTEL_TAG="v1.12.0"

../opentelemetry-specification:
	cd ..; git clone https://github.com/open-telemetry/opentelemetry-specification.git

../opentelemetry-go:
	cd ..; git clone https://github.com/open-telemetry/opentelemetry-go.git

otel-generate: ../opentelemetry-specification ../opentelemetry-go
	cd ../opentelemetry-specification && git pull && git checkout tags/$(OTEL_TAG)

