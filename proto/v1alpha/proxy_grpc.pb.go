// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v5.27.3
// source: proto/v1alpha/proxy.proto

package cloudproxyv1alpha

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
	CloudProxyAPI_StreamCloudProxy_FullMethodName = "/castai.cloud.proxy.v1alpha.CloudProxyAPI/StreamCloudProxy"
	CloudProxyAPI_SendToProxy_FullMethodName      = "/castai.cloud.proxy.v1alpha.CloudProxyAPI/SendToProxy"
)

// CloudProxyAPIClient is the client API for CloudProxyAPI service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
//
// CloudProxyAPI provides the API for proxying cloud requests for CAST AI External Provisioner.
type CloudProxyAPIClient interface {
	// Stream from cluster to mothership.
	StreamCloudProxy(ctx context.Context, opts ...grpc.CallOption) (grpc.BidiStreamingClient[StreamCloudProxyRequest, StreamCloudProxyResponse], error)
	// Unary CloudProxy request from mothership.
	SendToProxy(ctx context.Context, in *SendToProxyRequest, opts ...grpc.CallOption) (*SendToProxyResponse, error)
}

type cloudProxyAPIClient struct {
	cc grpc.ClientConnInterface
}

func NewCloudProxyAPIClient(cc grpc.ClientConnInterface) CloudProxyAPIClient {
	return &cloudProxyAPIClient{cc}
}

func (c *cloudProxyAPIClient) StreamCloudProxy(ctx context.Context, opts ...grpc.CallOption) (grpc.BidiStreamingClient[StreamCloudProxyRequest, StreamCloudProxyResponse], error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	stream, err := c.cc.NewStream(ctx, &CloudProxyAPI_ServiceDesc.Streams[0], CloudProxyAPI_StreamCloudProxy_FullMethodName, cOpts...)
	if err != nil {
		return nil, err
	}
	x := &grpc.GenericClientStream[StreamCloudProxyRequest, StreamCloudProxyResponse]{ClientStream: stream}
	return x, nil
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type CloudProxyAPI_StreamCloudProxyClient = grpc.BidiStreamingClient[StreamCloudProxyRequest, StreamCloudProxyResponse]

func (c *cloudProxyAPIClient) SendToProxy(ctx context.Context, in *SendToProxyRequest, opts ...grpc.CallOption) (*SendToProxyResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(SendToProxyResponse)
	err := c.cc.Invoke(ctx, CloudProxyAPI_SendToProxy_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// CloudProxyAPIServer is the server API for CloudProxyAPI service.
// All implementations must embed UnimplementedCloudProxyAPIServer
// for forward compatibility.
//
// CloudProxyAPI provides the API for proxying cloud requests for CAST AI External Provisioner.
type CloudProxyAPIServer interface {
	// Stream from cluster to mothership.
	StreamCloudProxy(grpc.BidiStreamingServer[StreamCloudProxyRequest, StreamCloudProxyResponse]) error
	// Unary CloudProxy request from mothership.
	SendToProxy(context.Context, *SendToProxyRequest) (*SendToProxyResponse, error)
	mustEmbedUnimplementedCloudProxyAPIServer()
}

// UnimplementedCloudProxyAPIServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedCloudProxyAPIServer struct{}

func (UnimplementedCloudProxyAPIServer) StreamCloudProxy(grpc.BidiStreamingServer[StreamCloudProxyRequest, StreamCloudProxyResponse]) error {
	return status.Errorf(codes.Unimplemented, "method StreamCloudProxy not implemented")
}
func (UnimplementedCloudProxyAPIServer) SendToProxy(context.Context, *SendToProxyRequest) (*SendToProxyResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SendToProxy not implemented")
}
func (UnimplementedCloudProxyAPIServer) mustEmbedUnimplementedCloudProxyAPIServer() {}
func (UnimplementedCloudProxyAPIServer) testEmbeddedByValue()                       {}

// UnsafeCloudProxyAPIServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to CloudProxyAPIServer will
// result in compilation errors.
type UnsafeCloudProxyAPIServer interface {
	mustEmbedUnimplementedCloudProxyAPIServer()
}

func RegisterCloudProxyAPIServer(s grpc.ServiceRegistrar, srv CloudProxyAPIServer) {
	// If the following call pancis, it indicates UnimplementedCloudProxyAPIServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&CloudProxyAPI_ServiceDesc, srv)
}

func _CloudProxyAPI_StreamCloudProxy_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(CloudProxyAPIServer).StreamCloudProxy(&grpc.GenericServerStream[StreamCloudProxyRequest, StreamCloudProxyResponse]{ServerStream: stream})
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type CloudProxyAPI_StreamCloudProxyServer = grpc.BidiStreamingServer[StreamCloudProxyRequest, StreamCloudProxyResponse]

func _CloudProxyAPI_SendToProxy_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SendToProxyRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CloudProxyAPIServer).SendToProxy(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: CloudProxyAPI_SendToProxy_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CloudProxyAPIServer).SendToProxy(ctx, req.(*SendToProxyRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// CloudProxyAPI_ServiceDesc is the grpc.ServiceDesc for CloudProxyAPI service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var CloudProxyAPI_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "castai.cloud.proxy.v1alpha.CloudProxyAPI",
	HandlerType: (*CloudProxyAPIServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "SendToProxy",
			Handler:    _CloudProxyAPI_SendToProxy_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "StreamCloudProxy",
			Handler:       _CloudProxyAPI_StreamCloudProxy_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "proto/v1alpha/proxy.proto",
}
