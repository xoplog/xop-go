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
	"github.com/xoplog/xop-go/xoptrace"
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
	fragment := combineFragments(t, server.getFragments())
	var requestCount int
	var lineCount int
	for i, trace := range fragment.Traces {
		traceID := xoptrace.NewHexBytes16FromSlice(trace.TraceID)
		require.Falsef(t, traceID.IsZero(), "traceID not zero, trace #%d", i)
		requestCount += len(trace.Requests)
		for _, request := range trace.Requests {
			requestID := xoptrace.NewHexBytes8FromSlice(request.RequestID)
			lineCount += len(request.Lines)
			for i, line := range request.Lines {
				require.NotNilf(t, line, "line %d in trace %s in request %s is nil", i, traceID, requestID)
			}
		}
	}
	assert.Equal(t, len(tlog.Requests), requestCount, "count of requests")
	assert.Equal(t, len(tlog.Lines), lineCount, "count of lines")
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
func combineFragments(t *testing.T, fragments []*xopproto.IngestFragment) *xopproto.IngestFragment {
	combined := &xopproto.IngestFragment{
		Source: fragments[0].Source,
	}
	traceMap := make(map[[16]byte]*OrderedTrace)
	var allTraces []*OrderedTrace
	t.Logf("combining %d fragments", len(fragments))
	for fi, fragment := range fragments {
		t.Logf(" fragment %d has %d traces", fi, len(fragment.Traces))
		combined.AttributeDefinitions = append(combined.AttributeDefinitions, fragment.AttributeDefinitions...)
		for ti, trace := range fragment.Traces {
			require.Equal(t, 16, len(trace.TraceID), "traceID length")
			traceID := xoptrace.NewHexBytes16FromSlice(trace.TraceID)
			t.Logf("  trace %d (%s) has %d requests", ti, traceID, len(trace.Requests))
			ot, existingTrace := traceMap[traceID.Array()]
			if !existingTrace {
				ot = &OrderedTrace{
					Trace:      *trace,
					RequestMap: make(map[[8]byte]*OrderedRequest),
				}
				traceMap[traceID.Array()] = ot
				allTraces = append(allTraces, ot)
			}
			for ri, request := range trace.Requests {
				require.Equal(t, 8, len(request.RequestID), "requestID length")
				requestID := xoptrace.NewHexBytes8FromSlice(request.RequestID)
				t.Logf("   request %d (%s) has %d lines", ri, requestID, len(request.Lines))
				combinedRequests, ok := ot.RequestMap[requestID.Array()]
				if !ok {
					if request.PriorLinesInRequest != 0 {
						t.Logf("   prior lines in %s: %d (new)", requestID, request.PriorLinesInRequest)
						newLines := make([]*xopproto.Line, int32(len(request.Lines))+request.PriorLinesInRequest)
						copy(newLines[request.PriorLinesInRequest:], request.Lines)
						request.Lines = newLines
					}
					or := &OrderedRequest{
						Request: *request,
						SpanMap: make(map[[8]byte]int),
					}
					for i, span := range request.Spans {
						var spanID [8]byte
						copy(spanID[:], span.SpanID)
						or.SpanMap[spanID] = i
					}
					ot.RequestMap[requestID.Array()] = or
					if existingTrace {
						t.Logf("   appending request")
						ot.Trace.Requests = append(ot.Trace.Requests, request)
					}
					continue
				}
				if int(request.PriorLinesInRequest)+len(request.Lines) < len(combinedRequests.Lines) {
					t.Logf("   prior lines in %s: %d (shifting)", requestID, request.PriorLinesInRequest)
					newLines := make([]*xopproto.Line, len(request.Lines)+int(request.PriorLinesInRequest))
					copy(newLines, combinedRequests.Lines)
					copy(newLines[request.PriorLinesInRequest:], request.Lines)
					combinedRequests.Lines = newLines
				}
				for _, span := range request.Spans {
					var spanID [8]byte
					copy(spanID[:], span.SpanID)
					if existingIndex, ok := combinedRequests.SpanMap[spanID]; ok {
						existing := combinedRequests.Spans[existingIndex]
						if span.Version > existing.Version {
							combinedRequests.Spans[existingIndex] = span
						}
					} else {
						combinedRequests.SpanMap[spanID] = len(combinedRequests.Spans)
						combinedRequests.Spans = append(combinedRequests.Spans, span)
					}
				}
			}
		}
	}
	for _, trace := range allTraces {
		combined.Traces = append(combined.Traces, &trace.Trace)
	}
	return combined
}
