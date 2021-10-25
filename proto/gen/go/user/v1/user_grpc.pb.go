// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package userv1

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// UserServiceClient is the client API for UserService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type UserServiceClient interface {
	// SMS OTP 요청
	RequestSmsOtp(ctx context.Context, in *RequestSmsOtpRequest, opts ...grpc.CallOption) (*RequestSmsOtpResponse, error)
	// SMS OTP 검증
	VerifySmsOtp(ctx context.Context, in *VerifySmsOtpRequest, opts ...grpc.CallOption) (*VerifySmsOtpResponse, error)
	// 회원가입
	Register(ctx context.Context, in *RegisterRequest, opts ...grpc.CallOption) (*RegisterResponse, error)
	// 로그인
	SignIn(ctx context.Context, in *SignInRequest, opts ...grpc.CallOption) (*SignInResponse, error)
	// 로그아웃
	SignOut(ctx context.Context, in *SignOutRequest, opts ...grpc.CallOption) (*SignOutResponse, error)
	// 토큰 갱신
	RefreshToken(ctx context.Context, in *RefreshTokenRequest, opts ...grpc.CallOption) (*RefreshTokenResponse, error)
}

type userServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewUserServiceClient(cc grpc.ClientConnInterface) UserServiceClient {
	return &userServiceClient{cc}
}

func (c *userServiceClient) RequestSmsOtp(ctx context.Context, in *RequestSmsOtpRequest, opts ...grpc.CallOption) (*RequestSmsOtpResponse, error) {
	out := new(RequestSmsOtpResponse)
	err := c.cc.Invoke(ctx, "/user.v1.UserService/RequestSmsOtp", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *userServiceClient) VerifySmsOtp(ctx context.Context, in *VerifySmsOtpRequest, opts ...grpc.CallOption) (*VerifySmsOtpResponse, error) {
	out := new(VerifySmsOtpResponse)
	err := c.cc.Invoke(ctx, "/user.v1.UserService/VerifySmsOtp", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *userServiceClient) Register(ctx context.Context, in *RegisterRequest, opts ...grpc.CallOption) (*RegisterResponse, error) {
	out := new(RegisterResponse)
	err := c.cc.Invoke(ctx, "/user.v1.UserService/Register", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *userServiceClient) SignIn(ctx context.Context, in *SignInRequest, opts ...grpc.CallOption) (*SignInResponse, error) {
	out := new(SignInResponse)
	err := c.cc.Invoke(ctx, "/user.v1.UserService/SignIn", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *userServiceClient) SignOut(ctx context.Context, in *SignOutRequest, opts ...grpc.CallOption) (*SignOutResponse, error) {
	out := new(SignOutResponse)
	err := c.cc.Invoke(ctx, "/user.v1.UserService/SignOut", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *userServiceClient) RefreshToken(ctx context.Context, in *RefreshTokenRequest, opts ...grpc.CallOption) (*RefreshTokenResponse, error) {
	out := new(RefreshTokenResponse)
	err := c.cc.Invoke(ctx, "/user.v1.UserService/RefreshToken", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// UserServiceServer is the server API for UserService service.
// All implementations should embed UnimplementedUserServiceServer
// for forward compatibility
type UserServiceServer interface {
	// SMS OTP 요청
	RequestSmsOtp(context.Context, *RequestSmsOtpRequest) (*RequestSmsOtpResponse, error)
	// SMS OTP 검증
	VerifySmsOtp(context.Context, *VerifySmsOtpRequest) (*VerifySmsOtpResponse, error)
	// 회원가입
	Register(context.Context, *RegisterRequest) (*RegisterResponse, error)
	// 로그인
	SignIn(context.Context, *SignInRequest) (*SignInResponse, error)
	// 로그아웃
	SignOut(context.Context, *SignOutRequest) (*SignOutResponse, error)
	// 토큰 갱신
	RefreshToken(context.Context, *RefreshTokenRequest) (*RefreshTokenResponse, error)
}

// UnimplementedUserServiceServer should be embedded to have forward compatible implementations.
type UnimplementedUserServiceServer struct {
}

func (UnimplementedUserServiceServer) RequestSmsOtp(context.Context, *RequestSmsOtpRequest) (*RequestSmsOtpResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RequestSmsOtp not implemented")
}
func (UnimplementedUserServiceServer) VerifySmsOtp(context.Context, *VerifySmsOtpRequest) (*VerifySmsOtpResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method VerifySmsOtp not implemented")
}
func (UnimplementedUserServiceServer) Register(context.Context, *RegisterRequest) (*RegisterResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Register not implemented")
}
func (UnimplementedUserServiceServer) SignIn(context.Context, *SignInRequest) (*SignInResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SignIn not implemented")
}
func (UnimplementedUserServiceServer) SignOut(context.Context, *SignOutRequest) (*SignOutResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SignOut not implemented")
}
func (UnimplementedUserServiceServer) RefreshToken(context.Context, *RefreshTokenRequest) (*RefreshTokenResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RefreshToken not implemented")
}

// UnsafeUserServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to UserServiceServer will
// result in compilation errors.
type UnsafeUserServiceServer interface {
	mustEmbedUnimplementedUserServiceServer()
}

func RegisterUserServiceServer(s grpc.ServiceRegistrar, srv UserServiceServer) {
	s.RegisterService(&UserService_ServiceDesc, srv)
}

func _UserService_RequestSmsOtp_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RequestSmsOtpRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UserServiceServer).RequestSmsOtp(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/user.v1.UserService/RequestSmsOtp",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UserServiceServer).RequestSmsOtp(ctx, req.(*RequestSmsOtpRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _UserService_VerifySmsOtp_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(VerifySmsOtpRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UserServiceServer).VerifySmsOtp(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/user.v1.UserService/VerifySmsOtp",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UserServiceServer).VerifySmsOtp(ctx, req.(*VerifySmsOtpRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _UserService_Register_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RegisterRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UserServiceServer).Register(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/user.v1.UserService/Register",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UserServiceServer).Register(ctx, req.(*RegisterRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _UserService_SignIn_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SignInRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UserServiceServer).SignIn(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/user.v1.UserService/SignIn",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UserServiceServer).SignIn(ctx, req.(*SignInRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _UserService_SignOut_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SignOutRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UserServiceServer).SignOut(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/user.v1.UserService/SignOut",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UserServiceServer).SignOut(ctx, req.(*SignOutRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _UserService_RefreshToken_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RefreshTokenRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UserServiceServer).RefreshToken(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/user.v1.UserService/RefreshToken",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UserServiceServer).RefreshToken(ctx, req.(*RefreshTokenRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// UserService_ServiceDesc is the grpc.ServiceDesc for UserService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var UserService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "user.v1.UserService",
	HandlerType: (*UserServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "RequestSmsOtp",
			Handler:    _UserService_RequestSmsOtp_Handler,
		},
		{
			MethodName: "VerifySmsOtp",
			Handler:    _UserService_VerifySmsOtp_Handler,
		},
		{
			MethodName: "Register",
			Handler:    _UserService_Register_Handler,
		},
		{
			MethodName: "SignIn",
			Handler:    _UserService_SignIn_Handler,
		},
		{
			MethodName: "SignOut",
			Handler:    _UserService_SignOut_Handler,
		},
		{
			MethodName: "RefreshToken",
			Handler:    _UserService_RefreshToken_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "user/v1/user.proto",
}
