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
	fragments := server.getFragments()
}
