
GOHOME ?= ${HOME}
GOPATH ?= ${GOHOME}
GOBIN ?= ${GOPATH}/bin

ZZZGO = $(wildcard ../xop*-go/*.zzzgo ../xop*-go/*/*.zzzgo ../xop*-go/*/*/*.zzzgo)
ZZZGENERATED = $(patsubst %.zzzgo, %.go, $(ZZZGO))
PB = xopproto/ingest.pb.go xopproto/ingest_grpc.pb.go
TOOLS = ${GOBIN}/gofumpt ${GOBIN}/goimports ${GOBIN}/enumer
TEST_ONLY ?= 

RELATED = xopresty-go xopotel-go

all:	$(ZZZGENERATED) $(PB) .gitattributes
	go generate ./...
	go build ./...
	for i in $(RELATED); do (echo $$i ...; cd ../$$i && go generate ./... && go build ./...); done


ci_checkout_peers:;
	branch="$${GITHUB_REF##*/}"; for i in $(RELATED); do (cd ..; git clone https://github.com/xoplog/$$i --depth 1 -b $$branch || git clone https://github.com/xoplog/$$i --depth 1); done

.gitattributes: $(ZZZGENERATED)
	echo '*.zzzgo linguist-language=Go' > $@
	echo 'doc.go linguist-documentation' >> $@
	echo '*.md linguist-documentation' >> $@
	echo '*.pb.go linguist-generated' >> $@
	for i in $(ZZZGENERATED); do echo "$$i linguist-generated" >> $@; done

test:	$(ZZZGENERATED) testadjuster
	go generate ./...
	go test -v ./xopjson/... -run TestASingleLine
	go test -v ./xopjson/... -tags xoptesting -run TestParameters -failfast $(TEST_ONLY)
	go test -tags xoptesting ./... -failfast $(TEST_ONLY)
	for i in $(RELATED); do (echo $$i...; cd ../$$i && go test -tags xoptesting ./... $(TEST_ONLY) ); done
	go test -tags xoptesting -race ./... -failfast $(TEST_ONLY)
	for i in $(RELATED); do (echo $$i...; cd ../$$i && go test -tags xoptesting -race ./... $(TEST_ONLY) ); done


testadjuster: $(ZZZGenerated)
	XOPLEVEL_xoptestutil=warn XOPLEVEL_foo=debug go test -tags xoptesting ./xoptest/xoptestutil -run TestAdjuster -count 1
	XOPLEVEL_xoptestutil=debug XOPLEVEL_foo=warn go test -tags xoptesting ./xoptest/xoptestutil -run TestAdjuster -count 1

citest:
	go test ./... -failfast 
	for i in $(RELATED); do (echo $$i...; cd ../$$i && go test -tags xoptesting ./... $(TEST_ONLY) ); done
	go test -race ./... -failfast 
	for i in $(RELATED); do (echo $$i...; cd ../$$i && go test -tags xoptesting -race ./... $(TEST_ONLY) ); done
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
	grep -v '\.pb.go:' coverage.txt | \
		egrep -v 'xoptestutil/|xopoteltest/' > coverage.txt.tmp
	mv coverage.txt.tmp coverage.txt

coverage: calculate_coverage
	go tool cover -html=coverage.txt

golanglint:
	# binary will be $(go env GOPATH)/bin/golangci-lint
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin v1.52.2
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

xopproto/%.proto: ../xopproto/%.proto
	ln -sf ../$< xopproto/

$(PB): xopproto/xop.proto
	cd xopproto && protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative *.proto

depgraph:
	set -x; godepgraph -s -p `go list -deps ./... | sort | perl -n -e 'next if m!^github\.com/(?:xoplog|muir)!; next unless m!^[^/]+\.!; print' | perl -n -e 'print "$$1\n" if m!^([^/]+)! && $$1 ne "github.com"; print "$$1\n" if m!^(github.com/[^/]+)!' | sort -u | perl -000 -p -e 'chomp; s/\n/,/g'` `go list ./...` | grep -v github.com/muir | dot -Tpng -o godepgraph.png && open godepgraph.png
