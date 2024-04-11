// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v4.25.3
// source: spec.proto

package orchestrator

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// SandboxesServiceClient is the client API for SandboxesService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type SandboxesServiceClient interface {
	// SandboxList is a gRPC service that returns a list of all the sandboxes.
	SandboxCreate(ctx context.Context, in *SandboxCreateRequest, opts ...grpc.CallOption) (*NewSandbox, error)
	SandboxList(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*SandboxListResponse, error)
	SandboxDelete(ctx context.Context, in *SandboxRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
}

type sandboxesServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewSandboxesServiceClient(cc grpc.ClientConnInterface) SandboxesServiceClient {
	return &sandboxesServiceClient{cc}
}

func (c *sandboxesServiceClient) SandboxCreate(ctx context.Context, in *SandboxCreateRequest, opts ...grpc.CallOption) (*NewSandbox, error) {
	out := new(NewSandbox)
	err := c.cc.Invoke(ctx, "/SandboxesService/SandboxCreate", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *sandboxesServiceClient) SandboxList(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*SandboxListResponse, error) {
	out := new(SandboxListResponse)
	err := c.cc.Invoke(ctx, "/SandboxesService/SandboxList", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *sandboxesServiceClient) SandboxDelete(ctx context.Context, in *SandboxRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/SandboxesService/SandboxDelete", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// SandboxesServiceServer is the server API for SandboxesService service.
// All implementations must embed UnimplementedSandboxesServiceServer
// for forward compatibility
type SandboxesServiceServer interface {
	// SandboxList is a gRPC service that returns a list of all the sandboxes.
	SandboxCreate(context.Context, *SandboxCreateRequest) (*NewSandbox, error)
	SandboxList(context.Context, *emptypb.Empty) (*SandboxListResponse, error)
	SandboxDelete(context.Context, *SandboxRequest) (*emptypb.Empty, error)
	mustEmbedUnimplementedSandboxesServiceServer()
}

// UnimplementedSandboxesServiceServer must be embedded to have forward compatible implementations.
type UnimplementedSandboxesServiceServer struct {
}

func (UnimplementedSandboxesServiceServer) SandboxCreate(context.Context, *SandboxCreateRequest) (*NewSandbox, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SandboxCreate not implemented")
}
func (UnimplementedSandboxesServiceServer) SandboxList(context.Context, *emptypb.Empty) (*SandboxListResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SandboxList not implemented")
}
func (UnimplementedSandboxesServiceServer) SandboxDelete(context.Context, *SandboxRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SandboxDelete not implemented")
}
func (UnimplementedSandboxesServiceServer) mustEmbedUnimplementedSandboxesServiceServer() {}

// UnsafeSandboxesServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to SandboxesServiceServer will
// result in compilation errors.
type UnsafeSandboxesServiceServer interface {
	mustEmbedUnimplementedSandboxesServiceServer()
}

func RegisterSandboxesServiceServer(s grpc.ServiceRegistrar, srv SandboxesServiceServer) {
	s.RegisterService(&SandboxesService_ServiceDesc, srv)
}

func _SandboxesService_SandboxCreate_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SandboxCreateRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SandboxesServiceServer).SandboxCreate(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/SandboxesService/SandboxCreate",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SandboxesServiceServer).SandboxCreate(ctx, req.(*SandboxCreateRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _SandboxesService_SandboxList_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SandboxesServiceServer).SandboxList(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/SandboxesService/SandboxList",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SandboxesServiceServer).SandboxList(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _SandboxesService_SandboxDelete_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SandboxRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SandboxesServiceServer).SandboxDelete(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/SandboxesService/SandboxDelete",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SandboxesServiceServer).SandboxDelete(ctx, req.(*SandboxRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// SandboxesService_ServiceDesc is the grpc.ServiceDesc for SandboxesService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var SandboxesService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "SandboxesService",
	HandlerType: (*SandboxesServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "SandboxCreate",
			Handler:    _SandboxesService_SandboxCreate_Handler,
		},
		{
			MethodName: "SandboxList",
			Handler:    _SandboxesService_SandboxList_Handler,
		},
		{
			MethodName: "SandboxDelete",
			Handler:    _SandboxesService_SandboxDelete_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "spec.proto",
}
