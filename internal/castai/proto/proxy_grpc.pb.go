// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v5.27.3
// source: internal/castai/proto/proxy.proto

package proto

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	GCPProxyServer_Proxy_FullMethodName = "/gcpproxy.GCPProxyServer/Proxy"
)

// GCPProxyServerClient is the client API for GCPProxyServer service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type GCPProxyServerClient interface {
	Proxy(ctx context.Context, opts ...grpc.CallOption) (grpc.BidiStreamingClient[HttpResponse, HttpRequest], error)
}

type gCPProxyServerClient struct {
	cc grpc.ClientConnInterface
}

func NewGCPProxyServerClient(cc grpc.ClientConnInterface) GCPProxyServerClient {
	return &gCPProxyServerClient{cc}
}

func (c *gCPProxyServerClient) Proxy(ctx context.Context, opts ...grpc.CallOption) (grpc.BidiStreamingClient[HttpResponse, HttpRequest], error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	stream, err := c.cc.NewStream(ctx, &GCPProxyServer_ServiceDesc.Streams[0], GCPProxyServer_Proxy_FullMethodName, cOpts...)
	if err != nil {
		return nil, err
	}
	x := &grpc.GenericClientStream[HttpResponse, HttpRequest]{ClientStream: stream}
	return x, nil
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type GCPProxyServer_ProxyClient = grpc.BidiStreamingClient[HttpResponse, HttpRequest]

// GCPProxyServerServer is the server API for GCPProxyServer service.
// All implementations must embed UnimplementedGCPProxyServerServer
// for forward compatibility.
type GCPProxyServerServer interface {
	Proxy(grpc.BidiStreamingServer[HttpResponse, HttpRequest]) error
	mustEmbedUnimplementedGCPProxyServerServer()
}

// UnimplementedGCPProxyServerServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedGCPProxyServerServer struct{}

func (UnimplementedGCPProxyServerServer) Proxy(grpc.BidiStreamingServer[HttpResponse, HttpRequest]) error {
	return status.Errorf(codes.Unimplemented, "method Proxy not implemented")
}
func (UnimplementedGCPProxyServerServer) mustEmbedUnimplementedGCPProxyServerServer() {}
func (UnimplementedGCPProxyServerServer) testEmbeddedByValue()                        {}

// UnsafeGCPProxyServerServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to GCPProxyServerServer will
// result in compilation errors.
type UnsafeGCPProxyServerServer interface {
	mustEmbedUnimplementedGCPProxyServerServer()
}

func RegisterGCPProxyServerServer(s grpc.ServiceRegistrar, srv GCPProxyServerServer) {
	// If the following call pancis, it indicates UnimplementedGCPProxyServerServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&GCPProxyServer_ServiceDesc, srv)
}

func _GCPProxyServer_Proxy_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(GCPProxyServerServer).Proxy(&grpc.GenericServerStream[HttpResponse, HttpRequest]{ServerStream: stream})
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type GCPProxyServer_ProxyServer = grpc.BidiStreamingServer[HttpResponse, HttpRequest]

// GCPProxyServer_ServiceDesc is the grpc.ServiceDesc for GCPProxyServer service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var GCPProxyServer_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "gcpproxy.GCPProxyServer",
	HandlerType: (*GCPProxyServerServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Proxy",
			Handler:       _GCPProxyServer_Proxy_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "internal/castai/proto/proxy.proto",
}
