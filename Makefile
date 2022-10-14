
GOHOME ?= ${HOME}
GOPATH ?= ${GOHOME}
GOBIN ?= ${GOPATH}/bin

ZZZGO = $(wildcard *.zzzgo */*.zzzgo */*/*.zzzgo)
ZZZGENERATED = $(patsubst %.zzzgo, %.go, $(ZZZGO))
PB = xopproto/ingest.pb.go xopproto/ingest_grpc.pb.go
TOOLS = ${GOBIN}/gofumpt ${GOBIN}/goimports ${GOBIN}/enumer
TEST_ONLY =?

all:	$(ZZZGENERATED) $(PB) .gitattributes
	go generate ./...
	go build ./...

.gitattributes: $(ZZZGENERATED)
	echo '*.zzzgo linguist-language=Go' > $@
	echo 'doc.go linguist-documentation' >> $@
	echo '*.md linguist-documentation' >> $@
	for i in $(ZZZGENERATED); do echo "$$i linguist-generated" >> $@; done

test:	$(ZZZGENERATED) testadjuster
	go generate ./...
	go test -v ./xopjson/... -run TestASingleLine
	go test -v ./xopjson/... -tags xoptesting -run TestParameters -failfast $(TEST_ONLY)
	go test -tags xoptesting ./... -failfast $(TEST_ONLY)
	go test -tags xoptesting -race ./... -failfast $(TEST_ONLY)


testadjuster: $(ZZZGenerated)
	XOPLEVEL_xoptestutil=warn XOPLEVEL_foo=debug go test -tags xoptesting ./xoptest/xoptestutil -run TestAdjuster -count 1
	XOPLEVEL_xoptestutil=debug XOPLEVEL_foo=warn go test -tags xoptesting ./xoptest/xoptestutil -run TestAdjuster -count 1

citest:
	go test ./... -failfast 
	go test -race ./... -failfast 
	XOPLEVEL_xoptestutil=warn XOPLEVEL_foo=debug go test -tags xoptesting ./xoptest/xoptestutil -run TestAdjuster -count 1
	XOPLEVEL_xoptestutil=debug XOPLEVEL_foo=warn go test -tags xoptesting ./xoptest/xoptestutil -run TestAdjuster -count 1

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
	  go test -race -covermode=atomic -coverprofile=profile.out -coverpkg=github.com/xoplog/xop-go/... $$d; \
	  if [ -f profile.out ]; then \
	    grep -v ^mode profile.out >> coverage.txt; \
	    rm profile.out; \
	  fi; \
	done
	XOPLEVEL_xoptestutil=debug go test -covermode=atomic -tags xoptesting -coverprofile=profile.out -coverpkg=github.com/xoplog/xop-go/... ./xoptest/xoptestutil -run TestAdjustedLevelLogger -count 1
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

../xopproto/ingest.proto:
	cd ..; git clone https://github.com/xoplog/xopproto.git

xopproto/ingest.proto: ../xopproto/ingest.proto 
	ln -sf ../../xopproto/ingest.proto xopproto/

$(PB): xopproto/ingest.proto
	cd xopproto && protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative ingest.proto
