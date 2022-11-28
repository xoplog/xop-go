package xopup

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/xoplog/xop-go/xopbytes"
	"github.com/xoplog/xop-go/xopproto"
	"github.com/xoplog/xop-go/xoptrace"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	sizeOfSpan    = int(unsafe.Sizeof(xopproto.Span{})) + 8 + 8
	sizeOfLine    = int(unsafe.Sizeof(xopproto.Line{})) + 8
	sizeOfTrace   = int(unsafe.Sizeof(xopproto.Trace{})) + 16
	sizeOfRequest = int(unsafe.Sizeof(xopproto.Request{})) + 8
)

// Config controls the behavior of the xop uploader.  The struct is specifically compatabile with
// https://github.com/muir/nfigure but using nfigure to fill it out is optional.
//
// Namespace defines where the key naming conventions are coming from.  Ideally, this should be set
// at the organization level, but if not, it can be more specific.  Every uploader with the same
// Namespace and Version should use the same semantic conventions.
//
// The Source is used to identify the log uploader.  It can be a program name, a container name, or
// anything else that is useful to identify where logs came from.  The Source does not have to be
// unique.
type Config struct {
	Namespace string `json:"namespace" config:"namespace" env:"XOPNAMESPACE" flag:"namespace"   help:"semantic conventions set"`
	Version   string `json:"version"   config:"version"   env:"XOPVERSION"   flag:"xopversion"  help:"data source version (semver)"`
	Source    string `json:"source"    config:"source"    env:"XOPSOURCE"    flag:"xopsource"   help:"data source name"`
	Address   string `json:"address"   config:"address"   env:"XOPADDRESS"   flag:"xopaddress"  help:"host:port of xop ingest server"`
	Proto     string `json:"proto"     config:"proto"     env:"XOPPROTO"     flag:"xopproto"    help:"protocol to use (includes version)"`
	BufSizeK  int    `json:"bufsizeK"  config:"bufsizeK"  env:"XOPBUFSIZEK"  flag:"xopbufsizek" help:"how much to buffer (in K) per request"`
	OnError   func(error)
}

// TODO: keep?
//	BufDuration time.Duration `json:"bufDuration" config:"buffDuration" env:"XOPBUFSIZEK"  flag:"xopbufdur"   help:"how long to buffer"`
// 	InFlight  int    `json:"inFlight"  config:"inFlight"  env:"XOPINFLIGHT"  flag:"xopinflight" help:"how many outstanding blocks in flight"`

type Uploader struct {
	ctx                 context.Context
	config              Config
	client              xopproto.IngestClient
	conn                *grpc.ClientConn
	lock                sync.Mutex
	fragment            *xopproto.IngestFragment
	requestsInFragment  []*Request
	bytesBuffered       int
	requests            []*Request
	source              xopproto.SourceIdentity
	traceIDIndex        map[[16]byte]int
	requestIDIndex      map[[8]byte]int
	attributesDefined   sync.Map
	enumsDefined        sync.Map
	definitionsComplete sync.Pool
	completion          *sync.Cond
	shutdown            int32 // 1 == shutting down
	fragmentsInFlight   int32
}

type Request struct {
	uploader             *Uploader
	bundle               xoptrace.Bundle
	lineCount            int32
	request              xopbytes.Request
	startTime            int64
	fragmentsOutstanding int32
}

var _ xopbytes.BytesWriter = &Uploader{}
var _ xopbytes.BytesRequest = &Request{}

type attributeKey struct {
	key       string
	namespace string
}

type enumKey struct {
	attributeKey
	value int64
}

// newUploader is lazy: no connection is opened until there is data to send.
func newUploader(ctx context.Context, c Config) *Uploader {
	u := uuid.New()
	return &Uploader{
		ctx:    ctx,
		config: c,
		source: xopproto.SourceIdentity{
			SourceNamespace:        c.Namespace,
			SourceNamespaceVersion: c.Version,
			SourceID:               c.Source,
			SourceStartTime:        time.Now().UnixNano(),
			SourceRandom:           u[:],
		},
		definitionsComplete: sync.Pool{
			New: func() any {
				return &definitionComplete{}
			},
		},
		completion: sync.NewCond(&sync.Mutex{}),
	}
}

// Ping validates that the uploader has a connection
func (u *Uploader) Ping() error {
	var client xopproto.IngestClient
	err := func() error {
		u.lock.Lock()
		defer u.lock.Unlock()
		var err error
		client, err = u.connect()
		return err
	}()
	if err != nil {
		return err
	}
	_, err = client.Ping(u.ctx, &xopproto.Empty{})
	return err
}

// must hold a lock before calling connect
func (u *Uploader) connect() (xopproto.IngestClient, error) {
	if u.client != nil {
		return u.client, nil
	}
	conn, err := grpc.Dial(u.config.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, errors.Wrapf(err, "could not open grpc connection to %s", u.config.Address)
	}
	u.client = xopproto.NewIngestClient(conn)
	u.conn = conn
	return u.client, nil
}

func (u *Uploader) Request(bytesRequest xopbytes.Request) xopbytes.BytesRequest {
	r := &Request{
		uploader:  u,
		request:   bytesRequest,
		bundle:    bytesRequest.GetBundle(),
		startTime: bytesRequest.GetStartTime().UnixNano(),
	}
	u.lock.Lock()
	defer u.lock.Unlock()
	u.requests = append(u.requests, r)
	return r
}

func (u *Uploader) Buffered() bool { return true }

func (u *Uploader) Close() {
	u.flush()
	atomic.StoreInt32(&u.shutdown, 1)
	u.completion.L.Lock()
	for atomic.LoadInt32(&u.fragmentsInFlight) > 0 {
		u.completion.Wait()
	}
	u.completion.L.Unlock()
	u.lock.Lock()
	defer u.lock.Unlock()
	if u.conn != nil {
		u.conn.Close()
		u.conn = nil
		u.client = nil
	}
}

// Flush waits for the data to be sent
func (r *Request) Flush() error {
	func() {
		r.uploader.lock.Lock()
		defer r.uploader.lock.Unlock()
		request, _ := r.uploader.getRequest(r, false)
		if request != nil {
			request.ErrorCount = r.request.GetErrorCount()
			request.AlertCount = r.request.GetAlertCount()
			r.uploader.flush()
		}
	}()
	r.uploader.completion.L.Lock()
	for atomic.LoadInt32(&r.fragmentsOutstanding) > 0 {
		r.uploader.completion.Wait()
	}
	r.uploader.completion.L.Unlock()
	return nil
}

func (r *Request) ReclaimMemory() {}

// TODO: do this before adding things rather than after
// must be locked before calling
func (u *Uploader) noteBytes(count int) error {
	u.bytesBuffered += count
	if u.bytesBuffered < u.config.BufSizeK*1024 {
		return nil
	}
	return u.flush()
}

// must be locked before calling
func (u *Uploader) flush() error {
	if u.bytesBuffered == 0 {
		return nil
	}
	if atomic.LoadInt32(&u.shutdown) == 1 {
		return fmt.Errorf("uploader is shutdown")
	}
	client, err := u.connect()
	if err != nil {
		return err
	}
	u.fragment.Source = &u.source
	fragment := u.fragment
	requests := u.requestsInFragment
	u.fragment = nil
	u.requestsInFragment = make([]*Request, 0, len(u.requestsInFragment)*2)
	atomic.AddInt32(&u.fragmentsInFlight, 1)
	go func() {
		// TODO: retry N times if atomic.LoadInt32(&u.shutdown) == 0
		// TODO: track and limit total memory use
		pbErr, err := client.UploadFragment(u.ctx, fragment)
		if err == nil && pbErr.Text != "" {
			err = fmt.Errorf("upload error: %s", pbErr.Text)
		}
		if err != nil {
			if u.config.OnError != nil {
				u.config.OnError(err)
			}
		}
		for _, request := range requests {
			atomic.AddInt32(&request.fragmentsOutstanding, -1)
		}
		atomic.AddInt32(&u.fragmentsInFlight, -1)
		u.completion.Broadcast()
	}()
	return nil
}

func (u *Uploader) getFragment() *xopproto.IngestFragment {
	if u.fragment == nil {
		u.fragment = &xopproto.IngestFragment{}
		u.traceIDIndex = make(map[[16]byte]int)
		u.requestIDIndex = make(map[[8]byte]int)
	}
	return u.fragment
}

func (u *Uploader) getRequest(r *Request, makeNew bool) (*xopproto.Request, int) {
	var size int
	fragment := u.getFragment()
	traceIndex, ok := u.traceIDIndex[r.bundle.Trace.TraceID().Array()]
	if !ok {
		if !makeNew {
			return nil, 0
		}
		traceIndex = len(fragment.Traces)
		u.traceIDIndex[r.bundle.Trace.TraceID().Array()] = traceIndex
		fragment.Traces = append(fragment.Traces, &xopproto.Trace{
			TraceID: r.bundle.Trace.TraceID().Bytes(),
		})
		if r.bundle.Trace.TraceID().IsZero() {
			panic("zero trace")
		}
		size += sizeOfTrace
	}
	requestIndex, ok := u.requestIDIndex[r.bundle.Trace.SpanID().Array()]
	if !ok {
		if !makeNew {
			return nil, 0
		}
		request := &xopproto.Request{
			RequestID:           r.bundle.Trace.SpanID().Bytes(),
			ParentSpanID:        r.bundle.Parent.SpanID().Bytes(),
			PriorLinesInRequest: r.lineCount,
			StartTime:           r.startTime,
		}
		if r.bundle.ParentTraceIsDifferent() {
			request.ParentTraceID = r.bundle.Parent.TraceID().Bytes()
		}
		requestIndex = len(fragment.Traces[traceIndex].Requests)
		u.requestIDIndex[r.bundle.Trace.SpanID().Array()] = requestIndex
		fragment.Traces[traceIndex].Requests = append(fragment.Traces[traceIndex].Requests, request)
		size += sizeOfRequest
		u.requestsInFragment = append(u.requestsInFragment, r)
		atomic.AddInt32(&r.fragmentsOutstanding, 1)
	}
	return fragment.Traces[traceIndex].Requests[requestIndex], size
}

func pointerToInt64OrNil(i int64) *int64 {
	if i == 0 {
		return nil
	}
	return &i
}
