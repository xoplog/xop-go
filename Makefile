
ME = xop-go
RELATED = xopresty-go xopotel-go
EXTRA_TEST_DEPS = testadjuster

include Makefile.common

ci_checkout_peers:;
	branch="$${GITHUB_REF##*/}"; for i in $(RELATED); do (cd ..; git clone https://github.com/xoplog/$$i --depth $(CLONE_DEPTH) -b $$branch || git clone https://github.com/xoplog/$$i --depth $(CLONE_DEPTH)); done

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
