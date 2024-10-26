// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v5.27.3
// source: api/proto/v1/service.proto

package v1

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
	GophkeeperService_Login_FullMethodName    = "/api.v1.GophkeeperService/Login"
	GophkeeperService_Register_FullMethodName = "/api.v1.GophkeeperService/Register"
)

// GophkeeperServiceClient is the client API for GophkeeperService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type GophkeeperServiceClient interface {
	// unauthenticated APIs
	Login(ctx context.Context, in *LoginRequest, opts ...grpc.CallOption) (*AuthResponse, error)
	Register(ctx context.Context, in *RegisterRequest, opts ...grpc.CallOption) (*AuthResponse, error)
}

type gophkeeperServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewGophkeeperServiceClient(cc grpc.ClientConnInterface) GophkeeperServiceClient {
	return &gophkeeperServiceClient{cc}
}

func (c *gophkeeperServiceClient) Login(ctx context.Context, in *LoginRequest, opts ...grpc.CallOption) (*AuthResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(AuthResponse)
	err := c.cc.Invoke(ctx, GophkeeperService_Login_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *gophkeeperServiceClient) Register(ctx context.Context, in *RegisterRequest, opts ...grpc.CallOption) (*AuthResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(AuthResponse)
	err := c.cc.Invoke(ctx, GophkeeperService_Register_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// GophkeeperServiceServer is the server API for GophkeeperService service.
// All implementations must embed UnimplementedGophkeeperServiceServer
// for forward compatibility.
type GophkeeperServiceServer interface {
	// unauthenticated APIs
	Login(context.Context, *LoginRequest) (*AuthResponse, error)
	Register(context.Context, *RegisterRequest) (*AuthResponse, error)
	mustEmbedUnimplementedGophkeeperServiceServer()
}

// UnimplementedGophkeeperServiceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedGophkeeperServiceServer struct{}

func (UnimplementedGophkeeperServiceServer) Login(context.Context, *LoginRequest) (*AuthResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Login not implemented")
}
func (UnimplementedGophkeeperServiceServer) Register(context.Context, *RegisterRequest) (*AuthResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Register not implemented")
}
func (UnimplementedGophkeeperServiceServer) mustEmbedUnimplementedGophkeeperServiceServer() {}
func (UnimplementedGophkeeperServiceServer) testEmbeddedByValue()                           {}

// UnsafeGophkeeperServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to GophkeeperServiceServer will
// result in compilation errors.
type UnsafeGophkeeperServiceServer interface {
	mustEmbedUnimplementedGophkeeperServiceServer()
}

func RegisterGophkeeperServiceServer(s grpc.ServiceRegistrar, srv GophkeeperServiceServer) {
	// If the following call pancis, it indicates UnimplementedGophkeeperServiceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&GophkeeperService_ServiceDesc, srv)
}

func _GophkeeperService_Login_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(LoginRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GophkeeperServiceServer).Login(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: GophkeeperService_Login_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GophkeeperServiceServer).Login(ctx, req.(*LoginRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _GophkeeperService_Register_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RegisterRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GophkeeperServiceServer).Register(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: GophkeeperService_Register_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GophkeeperServiceServer).Register(ctx, req.(*RegisterRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// GophkeeperService_ServiceDesc is the grpc.ServiceDesc for GophkeeperService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var GophkeeperService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "api.v1.GophkeeperService",
	HandlerType: (*GophkeeperServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Login",
			Handler:    _GophkeeperService_Login_Handler,
		},
		{
			MethodName: "Register",
			Handler:    _GophkeeperService_Register_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "api/proto/v1/service.proto",
}
