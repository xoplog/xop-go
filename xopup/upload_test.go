package xopup_test

import (
	"context"
	"fmt"
	"net"
	"sync"
	"testing"

	"github.com/xoplog/xop-go"
	"github.com/xoplog/xop-go/xopproto"
	"github.com/xoplog/xop-go/xoptest"
	"github.com/xoplog/xop-go/xoptest/xoptestutil"
	"github.com/xoplog/xop-go/xopup"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

type Server struct {
	xopproto.UnimplementedIngestServer
	pingError error
	fragments []*xopproto.IngestFragment
	lock      sync.Mutex
}

func (s *Server) Ping(_ context.Context, e *xopproto.Empty) (*xopproto.Empty, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	return &xopproto.Empty{}, s.pingError
}

func (s *Server) UploadFragment(ctx context.Context, fragment *xopproto.IngestFragment) (*xopproto.Error, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.fragments = append(s.fragments, fragment)
	return &xopproto.Error{}, nil
}

func (s *Server) getFragments() []*xopproto.IngestFragment {
	s.lock.Lock()
	defer s.lock.Unlock()
	return s.fragments
}

func (s *Server) reset() {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.pingError = nil
	s.fragments = nil
}

func TestUpload(t *testing.T) {
	listen, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err, "listen")
	defer listen.Close()
	config := xopup.Config{
		InFlight:  6,
		Address:   listen.Addr().String(),
		Source:    t.Name(),
		Namespace: "xoptest",
	}
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	server := &Server{}
	xopproto.RegisterIngestServer(grpcServer, server)
	go grpcServer.Serve(listen)
	uploader := xopup.New(context.Background(), config)

	assert.NoError(t, uploader.Uploader.Ping(), "ping")
	server.pingError = fmt.Errorf("ooka")
	e := uploader.Uploader.Ping()
	if assert.Error(t, e, "expected ooka") {
		assert.Contains(t, e.Error(), "ooka", "error")
	}
	server.reset()
	assert.NoError(t, uploader.Uploader.Ping(), "ping")

	for _, mc := range xoptestutil.MessageCases {
		mc := mc
		t.Run(mc.Name, func(t *testing.T) {
			defer server.reset()
			tlog := xoptest.New(t)
			seed := xop.NewSeed(
				xop.WithBase(uploader),
				xop.WithBase(tlog),
			)
			if len(mc.SeedMods) != 0 {
				t.Logf("Applying %d extra seed mods", len(mc.SeedMods))
				seed = seed.Copy(mc.SeedMods...)
			}
			log := seed.Request(t.Name())
			mc.Do(t, log, tlog)
			verify(t, tlog, server)
		})
	}
}

func verify(t *testing.T, tlog *xoptest.TestLogger, server *Server) {
	fragments := combineFragments(server.getFragments())

}

type OrderedTrace struct {
	xopproto.Trace
	RequestMap map[[8]byte]*OrderedRequest
}

type OrderedRequest struct {
	xopproto.Request
	SpanMap map[[8]byte]int
}

// combineFragments creates a new fragment that represents the combination of
// multiple fragments.  It is assumbed tht all the fragments come from the same
// source.
func combineFragments(fragments []*xopproto.IngestFragment) *xopproto.IngestFragment {
	traceMap := make(map[[16]byte]*ReorderedTrace)
	var allTraces []*ReorderedTrace
	for _, fragment := range fragments {
		for _, trace := range fragment.Traces {
			var traceID [16]byte
			copy(traceID[:], trace.TraceID)
			ot, ok := traceMap[traceID]
			if !ok {
				ot = &OrderedTrace{
					Trace:      *trace,
					RequestMap: make(map[[8]byte]*OrderedRequest),
				}
				traceMap[traceID] = ot
				allTraces = append(allTraces, trace)
			}
			for _, request := range trace.Requests {
				var requestID [8]byte
				copy(requestID[:], request.RequestID)
				combinedRequests, ok := ot.RequestMap[requestID]
				if !ok {
					if request.PriorLinesInRequest != 0 {
						newLines := make([]*xopproto.Line, len(request.Lines)+request.PriorLinesInRequest)
						copy(newLines[request.PriorLinesInRequest:], request.Lines)
						request.Lines = newLines
					}
					or := &OrderedRequest{
						Request: *request,
						SpanMap: make(map[[8]byte]*xopproto.Span),
					}
					for i, span := range request.Span {
						var spanID [8]byte
						copy(spanID[:], span.SpanID)
						spanMap[spanID] = i
					}
					ot.RequestMap[requestID] = or
					continue
				}
				if request.PriorLinesInRequest+len(request.Lines) < len(combinedRequests.Lines) {
					newLines := make([]*xopproto.Line, len(request.Lines)+request.PriorLinesInRequest)
					copy(newLines, combinedRequests.Lines)
					copy(newLines[request.PriorLinesInRequest:], request.Lines)
					combinedRequests.Lines = newLines
				}
				for _, span := range request.Spans {
					var spanID [8]byte
					copy(spanID[:], span.SpanID)
					if existingIndex, ok := combinedRequests.spanMap[spanID]; ok {
						existing := combinedRequests.Spans[existingIndex]
						if span.Version > existing.Version {
							combinedRequests.Spans[existingIndex] = span
						}
					} else {
						combinedRequests.spamMap[spanID] = len(combinedRequests.Spans)
						combinedRequests.Spans = append(combinedRequests.Spans, span)
					}
				}
			}
		}
	}
	combined := &xopproto.IngestFragment{
		Source: fragments[0].Source,
	}
	for _, trace := range allTraces {
		combined.Traces = append(combined.Traces, &trace.Trace)
	}
}
