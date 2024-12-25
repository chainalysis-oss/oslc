// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             (unknown)
// source: oslc/v1/oslc.proto

package oslcv1

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
	OslcService_GetPackageInfo_FullMethodName = "/oslc.v1.OslcService/GetPackageInfo"
)

// OslcServiceClient is the client API for OslcService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type OslcServiceClient interface {
	GetPackageInfo(ctx context.Context, in *GetPackageInfoRequest, opts ...grpc.CallOption) (*GetPackageInfoResponse, error)
}

type oslcServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewOslcServiceClient(cc grpc.ClientConnInterface) OslcServiceClient {
	return &oslcServiceClient{cc}
}

func (c *oslcServiceClient) GetPackageInfo(ctx context.Context, in *GetPackageInfoRequest, opts ...grpc.CallOption) (*GetPackageInfoResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GetPackageInfoResponse)
	err := c.cc.Invoke(ctx, OslcService_GetPackageInfo_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// OslcServiceServer is the server API for OslcService service.
// All implementations must embed UnimplementedOslcServiceServer
// for forward compatibility.
type OslcServiceServer interface {
	GetPackageInfo(context.Context, *GetPackageInfoRequest) (*GetPackageInfoResponse, error)
	mustEmbedUnimplementedOslcServiceServer()
}

// UnimplementedOslcServiceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedOslcServiceServer struct{}

func (UnimplementedOslcServiceServer) GetPackageInfo(context.Context, *GetPackageInfoRequest) (*GetPackageInfoResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetPackageInfo not implemented")
}
func (UnimplementedOslcServiceServer) mustEmbedUnimplementedOslcServiceServer() {}
func (UnimplementedOslcServiceServer) testEmbeddedByValue()                     {}

// UnsafeOslcServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to OslcServiceServer will
// result in compilation errors.
type UnsafeOslcServiceServer interface {
	mustEmbedUnimplementedOslcServiceServer()
}

func RegisterOslcServiceServer(s grpc.ServiceRegistrar, srv OslcServiceServer) {
	// If the following call pancis, it indicates UnimplementedOslcServiceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&OslcService_ServiceDesc, srv)
}

func _OslcService_GetPackageInfo_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetPackageInfoRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(OslcServiceServer).GetPackageInfo(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: OslcService_GetPackageInfo_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(OslcServiceServer).GetPackageInfo(ctx, req.(*GetPackageInfoRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// OslcService_ServiceDesc is the grpc.ServiceDesc for OslcService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var OslcService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "oslc.v1.OslcService",
	HandlerType: (*OslcServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetPackageInfo",
			Handler:    _OslcService_GetPackageInfo_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "oslc/v1/oslc.proto",
}