package gocsi

import (
	goctx "context"
	"fmt"
	"strconv"
	"sync/atomic"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const requestIDKey = "csi.requestid"

type requestIDInjector struct {
	id uint64
}

// NewServerRequestIDInjector returns a new UnaryServerInterceptor
// that reads a unique request ID from the incoming context's gRPC
// metadata. If the incoming context does not contain gRPC metadata or
// a request ID, then a new request ID is generated.
func NewServerRequestIDInjector() grpc.UnaryServerInterceptor {
	return newRequestIDInjector().handleServer
}

// NewClientRequestIDInjector provides a UnaryClientInterceptor
// that injects the outgoing context with gRPC metadata that contains
// a unique ID.
func NewClientRequestIDInjector() grpc.UnaryClientInterceptor {
	return newRequestIDInjector().handleClient
}

func newRequestIDInjector() *requestIDInjector {
	return &requestIDInjector{}
}

func (s *requestIDInjector) handleServer(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (interface{}, error) {

	// storeID is a flag that indicates whether or not the request ID
	// should be atomically stored in the interceptor's id field at
	// the end of this function. If the ID was found in the incoming
	// request and could be parsed successfully then the ID is stored.
	// If the ID was generated server-side then the ID is not stored.
	storeID := true

	// Retrieve the gRPC metadata from the incoming context.
	md, mdOK := metadata.FromIncomingContext(ctx)

	// If no gRPC metadata was found then create some and ensure the
	// context is a gRPC incoming context.
	if !mdOK {
		md = metadata.Pairs()
		ctx = metadata.NewIncomingContext(ctx, md)
	}

	// Check the metadata from the request ID.
	szID, szIDOK := md[requestIDKey]

	// If the metadata does not contain a request ID then create a new
	// request ID and inject it into the metadata.
	if !szIDOK || len(szID) != 1 {
		szID = []string{fmt.Sprintf("%d", atomic.AddUint64(&s.id, 1))}
		md[requestIDKey] = szID
		storeID = false
	}

	// Parse the request ID from the
	id, err := strconv.ParseUint(szID[0], 10, 64)
	if err != nil {
		id = atomic.AddUint64(&s.id, 1)
		storeID = false
	}

	if storeID {
		atomic.StoreUint64(&s.id, id)
	}

	return handler(ctx, req)
}

func (s *requestIDInjector) handleClient(
	ctx context.Context,
	method string,
	req, rep interface{},
	cc *grpc.ClientConn,
	invoker grpc.UnaryInvoker,
	opts ...grpc.CallOption) error {

	// Ensure there is an outgoing gRPC context with metadata.
	md, mdOK := metadata.FromOutgoingContext(ctx)
	if !mdOK {
		md = metadata.Pairs()
		ctx = metadata.NewOutgoingContext(ctx, md)
	}

	// Ensure the request ID is set in the metadata.
	if szID, szIDOK := md[requestIDKey]; !szIDOK || len(szID) != 1 {
		szID = []string{fmt.Sprintf("%d", atomic.AddUint64(&s.id, 1))}
		md[requestIDKey] = szID
	}

	return invoker(ctx, method, req, rep, cc, opts...)
}

// GetRequestID inspects the context for gRPC metadata and returns
// its request ID if available.
func GetRequestID(ctx goctx.Context) (uint64, bool) {
	var (
		szID   []string
		szIDOK bool
	)

	// Prefer the incoming context, but look in both types.
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		szID, szIDOK = md[requestIDKey]
	} else if md, ok := metadata.FromOutgoingContext(ctx); ok {
		szID, szIDOK = md[requestIDKey]
	}

	if szIDOK && len(szID) == 1 {
		if id, err := strconv.ParseUint(szID[0], 10, 64); err == nil {
			return id, true
		}
	}

	return 0, false
}
