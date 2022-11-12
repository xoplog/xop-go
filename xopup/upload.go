package xopup

import (
	"context"
	"fmt"
	"sync"
	"time"
	"unsafe"

	"github.com/xoplog/xop-go/trace"
	"github.com/xoplog/xop-go/xopat"
	"github.com/xoplog/xop-go/xopbytes"
	"github.com/xoplog/xop-go/xopproto"

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
	InFlight  int    `json:"inFlight"  config:"inFlight"  env:"XOPINFLIGHT"  flag:"xopinflight" help:"how many outstanding blocks in flight"`
	OnError   func(error)
}

// XXX keep?
//	BufDuration time.Duration `json:"bufDuration" config:"buffDuration" env:"XOPBUFSIZEK"  flag:"xopbufdur"   help:"how long to buffer"`

type Uploader struct {
	sequenceNumber    int64
	ctx               context.Context
	config            Config
	client            xopproto.IngestClient
	conn              *grpc.ClientConn
	lock              sync.Mutex
	fragment          *xopproto.IngestFragmentBody
	bytesBuffered     int
	requests          []*Request
	source            xopproto.SourceIdentity
	traceIDIndex      map[[16]byte]int
	requestIDIndex    map[[8]byte]int
	attributesDefined map[attributeKey]struct{}
	enumsDefined      map[enumKey]struct{}
}

type Request struct {
	uploader  *Uploader
	bundle    trace.Bundle
	lineCount int32
	request   xopbytes.Request
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

// New is lazy: no connection is opened until there is data to send.
func (c Config) New(ctx context.Context) *Uploader {
	u := uuid.New()
	return &Uploader{
		ctx:    ctx,
		config: c,
		source: xopproto.SourceIdentity{
			SourceID:        c.Source,
			SourceStartTime: time.Now().UnixNano(),
			SourceRandom:    u[:],
		},
		traceIDIndex:   make(map[[16]byte]int),
		requestIDIndex: make(map[[8]byte]int),
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

func (u *Uploader) Request(request xopbytes.Request) xopbytes.BytesRequest {
	r := &Request{
		uploader: u,
		request:  request,
	}
	u.lock.Lock()
	defer u.lock.Unlock()
	u.requests = append(u.requests, r)
	return r
}

func (u *Uploader) Buffered() bool { return true }

func (u *Uploader) Close() {
	u.lock.Lock()
	defer u.lock.Unlock()
	if u.conn != nil {
		u.conn.Close()
		u.conn = nil
		u.client = nil
	}
}

func (u *Uploader) DefineAttribute(a *xopat.Attribute) {
	attributeKey := attributeKey{
		key:       a.Key(),
		namespace: a.Namespace(),
	}
	u.lock.Lock()
	defer u.lock.Unlock()
	if _, ok := u.attributesDefined[attributeKey]; ok {
		return
	}

	u.attributesDefined[attributeKey] = struct{}{}
	fragment := u.getFragment()
	definition := xopproto.AttributeDefinition{
		Key:             a.Key(),
		Description:     a.Description(),
		Namespace:       a.Namespace(),
		NamespaceSemver: a.SemverString(),
		Type:            xopproto.AttributeType(a.SubType()),
		ShouldIndex:     a.Indexed(),
		Prominance:      int32(a.Prominence()),
		Locked:          a.Locked(),
		Distinct:        a.Distinct(),
		Multiple:        a.Multiple(),
	}
	fragment.AttributeDefinitions = append(fragment.AttributeDefinitions, &definition)
}

func (u *Uploader) DefineEnum(a *xopat.EnumAttribute, e xopat.Enum) {
	enumKey := enumKey{
		attributeKey: attributeKey{
			key:       a.Key(),
			namespace: a.Namespace(),
		},
		value: e.Int64(),
	}
	u.lock.Lock()
	defer u.lock.Unlock()
	if _, ok := u.enumsDefined[enumKey]; ok {
		return
	}

	u.enumsDefined[enumKey] = struct{}{}
	enum := xopproto.EnumDefinition{
		AttributeKey:    a.Key(),
		Namespace:       a.Namespace(),
		NamespaceSemver: a.SemverString(),
		String_:         e.String(),
		IntValue:        e.Int64(),
	}
	fragment := u.getFragment()
	fragment.EnumDefinitions = append(fragment.EnumDefinitions, &enum)
}

func (r *Request) Flush() error {
	r.uploader.lock.Lock()
	defer r.uploader.lock.Unlock()
	request, _ := r.uploader.getRequest(r, true)
	if request != nil {
		request.ErrorCount = r.request.GetErrorCount()
		request.AlertCount = r.request.GetAlertCount()
	}
	return r.uploader.flush()
}

func (r *Request) ReclaimMemory() {}

func (r *Request) Span(span xopbytes.Span, buffer xopbytes.Buffer) error {
	bundle := span.GetBundle()
	pbSpan := xopproto.Span{
		SpanID:    bundle.Trace.GetSpanID().Bytes(),
		ParentID:  bundle.Parent.GetSpanID().Bytes(),
		JsonData:  buffer.AsBytes(),
		StartTime: span.GetStartTime().UnixNano(),
		EndTime:   pointerToInt64OrNil(span.GetEndTimeNano()),
	}
	if span.IsRequest() {
		pbSpan.IsRequest = true
		pbSpan.Baggage = bundle.Baggage.Bytes()
		pbSpan.TraceState = bundle.State.Bytes()
	}
	r.uploader.lock.Lock()
	defer r.uploader.lock.Unlock()
	request, byteCount := r.uploader.getRequest(r, true)
	request.Spans = append(request.Spans, &pbSpan)
	r.uploader.noteBytes(byteCount + sizeOfSpan + len(pbSpan.JsonData) + len(pbSpan.Baggage) + len(pbSpan.TraceState))
	return nil
}

func (r *Request) Line(line xopbytes.Line) error {
	pbLine := xopproto.Line{
		SpanID:    line.GetSpanID().Bytes(),
		LogLevel:  int32(line.GetLevel()),
		Timestamp: line.GetTime().UnixNano(),
		JsonData:  line.AsBytes(),
	}
	r.uploader.lock.Lock()
	defer r.uploader.lock.Unlock()
	r.lineCount++
	request, byteCount := r.uploader.getRequest(r, true)
	request.Lines = append(request.Lines, &pbLine)
	return r.uploader.noteBytes(byteCount + sizeOfLine + len(pbLine.JsonData))
}

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
	client, err := u.connect()
	if err != nil {
		return err
	}
	frag := &xopproto.IngestFragment{
		Source:   &u.source,
		Fragment: u.fragment,
	}
	pbErr, err := client.UploadFragment(u.ctx, frag)
	if err != nil {
		return err
	}
	if pbErr.Text != "" {
		return fmt.Errorf("upload error: %s", pbErr.Text)
	}
	u.fragment = nil
	return nil
}

func (u *Uploader) getFragment() *xopproto.IngestFragmentBody {
	if u.fragment == nil {
		u.fragment = &xopproto.IngestFragmentBody{}
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
		size += sizeOfTrace
	}
	requestIndex, ok := u.requestIDIndex[r.bundle.Trace.SpanID().Array()]
	if !ok {
		if !makeNew {
			return nil, 0
		}
		requestIndex = len(fragment.Traces[traceIndex].Requests)
		u.requestIDIndex[r.bundle.Trace.SpanID().Array()] = requestIndex
		request := &xopproto.Request{
			RequestID:           r.bundle.Trace.SpanID().Bytes(),
			ParentSpanID:        r.bundle.Parent.SpanID().Bytes(),
			PriorLinesInRequest: r.lineCount,
		}
		if r.bundle.ParentTraceIsDifferent() {
			request.ParentTraceID = r.bundle.Parent.TraceID().Bytes()
		}
		fragment.Traces[traceIndex].Requests = append(fragment.Traces[traceIndex].Requests, request)
		size += sizeOfRequest
	}
	return fragment.Traces[traceIndex].Requests[requestIndex], size
}

// XXX
func (r *Request) AttributeReferenced(*xopat.Attribute) error { return nil }

func pointerToInt64OrNil(i int64) *int64 {
	if i == 0 {
		return nil
	}
	return &i
}
