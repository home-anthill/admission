// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v5.29.3
// source: api/grpc/register/register.proto

package register

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
	Registration_Register_FullMethodName = "/register.Registration/Register"
)

// RegistrationClient is the client API for Registration service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type RegistrationClient interface {
	Register(ctx context.Context, in *RegisterRequest, opts ...grpc.CallOption) (*RegisterReply, error)
}

type registrationClient struct {
	cc grpc.ClientConnInterface
}

func NewRegistrationClient(cc grpc.ClientConnInterface) RegistrationClient {
	return &registrationClient{cc}
}

func (c *registrationClient) Register(ctx context.Context, in *RegisterRequest, opts ...grpc.CallOption) (*RegisterReply, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(RegisterReply)
	err := c.cc.Invoke(ctx, Registration_Register_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// RegistrationServer is the server API for Registration service.
// All implementations must embed UnimplementedRegistrationServer
// for forward compatibility.
type RegistrationServer interface {
	Register(context.Context, *RegisterRequest) (*RegisterReply, error)
	mustEmbedUnimplementedRegistrationServer()
}

// UnimplementedRegistrationServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedRegistrationServer struct{}

func (UnimplementedRegistrationServer) Register(context.Context, *RegisterRequest) (*RegisterReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Register not implemented")
}
func (UnimplementedRegistrationServer) mustEmbedUnimplementedRegistrationServer() {}
func (UnimplementedRegistrationServer) testEmbeddedByValue()                      {}

// UnsafeRegistrationServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to RegistrationServer will
// result in compilation errors.
type UnsafeRegistrationServer interface {
	mustEmbedUnimplementedRegistrationServer()
}

func RegisterRegistrationServer(s grpc.ServiceRegistrar, srv RegistrationServer) {
	// If the following call pancis, it indicates UnimplementedRegistrationServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&Registration_ServiceDesc, srv)
}

func _Registration_Register_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RegisterRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RegistrationServer).Register(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Registration_Register_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RegistrationServer).Register(ctx, req.(*RegisterRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Registration_ServiceDesc is the grpc.ServiceDesc for Registration service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Registration_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "register.Registration",
	HandlerType: (*RegistrationServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Register",
			Handler:    _Registration_Register_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "api/grpc/register/register.proto",
}
