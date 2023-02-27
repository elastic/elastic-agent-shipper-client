// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License;
// you may not use this file except in compliance with the Elastic License.

// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.19.4
// source: shipper.proto

package proto

import (
	context "context"
	messages "github.com/elastic/elastic-agent-shipper-client/pkg/proto/messages"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// ProducerClient is the client API for Producer service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ProducerClient interface {
	// Publishes a list of events via the Elastic agent shipper.
	// Blocks until all processing steps complete and data is written to the queue.
	//
	// If the queue could not accept some events from the request, this returns a successful response
	// containing a count of events that were accepted by the queue.
	// The client is expected to retry sending the rest of the events in a separate request.
	//
	// The client is also expected to have some kind of backoff strategy
	//
	//	in case of a reply with an accepted count < the amount of sent events.
	PublishEvents(ctx context.Context, in *messages.PublishRequest, opts ...grpc.CallOption) (*messages.PublishReply, error)
	// Returns the shipper's uuid and its current position in the event stream (persisted index).
	PersistedIndex(ctx context.Context, in *messages.PersistedIndexRequest, opts ...grpc.CallOption) (Producer_PersistedIndexClient, error)
}

type producerClient struct {
	cc grpc.ClientConnInterface
}

func NewProducerClient(cc grpc.ClientConnInterface) ProducerClient {
	return &producerClient{cc}
}

func (c *producerClient) PublishEvents(ctx context.Context, in *messages.PublishRequest, opts ...grpc.CallOption) (*messages.PublishReply, error) {
	out := new(messages.PublishReply)
	err := c.cc.Invoke(ctx, "/elastic.agent.shipper.v1.Producer/PublishEvents", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *producerClient) PersistedIndex(ctx context.Context, in *messages.PersistedIndexRequest, opts ...grpc.CallOption) (Producer_PersistedIndexClient, error) {
	stream, err := c.cc.NewStream(ctx, &Producer_ServiceDesc.Streams[0], "/elastic.agent.shipper.v1.Producer/PersistedIndex", opts...)
	if err != nil {
		return nil, err
	}
	x := &producerPersistedIndexClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type Producer_PersistedIndexClient interface {
	Recv() (*messages.PersistedIndexReply, error)
	grpc.ClientStream
}

type producerPersistedIndexClient struct {
	grpc.ClientStream
}

func (x *producerPersistedIndexClient) Recv() (*messages.PersistedIndexReply, error) {
	m := new(messages.PersistedIndexReply)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// ProducerServer is the server API for Producer service.
// All implementations must embed UnimplementedProducerServer
// for forward compatibility
type ProducerServer interface {
	// Publishes a list of events via the Elastic agent shipper.
	// Blocks until all processing steps complete and data is written to the queue.
	//
	// If the queue could not accept some events from the request, this returns a successful response
	// containing a count of events that were accepted by the queue.
	// The client is expected to retry sending the rest of the events in a separate request.
	//
	// The client is also expected to have some kind of backoff strategy
	//
	//	in case of a reply with an accepted count < the amount of sent events.
	PublishEvents(context.Context, *messages.PublishRequest) (*messages.PublishReply, error)
	// Returns the shipper's uuid and its current position in the event stream (persisted index).
	PersistedIndex(*messages.PersistedIndexRequest, Producer_PersistedIndexServer) error
	mustEmbedUnimplementedProducerServer()
}

// UnimplementedProducerServer must be embedded to have forward compatible implementations.
type UnimplementedProducerServer struct {
}

func (UnimplementedProducerServer) PublishEvents(context.Context, *messages.PublishRequest) (*messages.PublishReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PublishEvents not implemented")
}
func (UnimplementedProducerServer) PersistedIndex(*messages.PersistedIndexRequest, Producer_PersistedIndexServer) error {
	return status.Errorf(codes.Unimplemented, "method PersistedIndex not implemented")
}
func (UnimplementedProducerServer) mustEmbedUnimplementedProducerServer() {}

// UnsafeProducerServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ProducerServer will
// result in compilation errors.
type UnsafeProducerServer interface {
	mustEmbedUnimplementedProducerServer()
}

func RegisterProducerServer(s grpc.ServiceRegistrar, srv ProducerServer) {
	s.RegisterService(&Producer_ServiceDesc, srv)
}

func _Producer_PublishEvents_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(messages.PublishRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ProducerServer).PublishEvents(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/elastic.agent.shipper.v1.Producer/PublishEvents",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ProducerServer).PublishEvents(ctx, req.(*messages.PublishRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Producer_PersistedIndex_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(messages.PersistedIndexRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(ProducerServer).PersistedIndex(m, &producerPersistedIndexServer{stream})
}

type Producer_PersistedIndexServer interface {
	Send(*messages.PersistedIndexReply) error
	grpc.ServerStream
}

type producerPersistedIndexServer struct {
	grpc.ServerStream
}

func (x *producerPersistedIndexServer) Send(m *messages.PersistedIndexReply) error {
	return x.ServerStream.SendMsg(m)
}

// Producer_ServiceDesc is the grpc.ServiceDesc for Producer service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Producer_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "elastic.agent.shipper.v1.Producer",
	HandlerType: (*ProducerServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "PublishEvents",
			Handler:    _Producer_PublishEvents_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "PersistedIndex",
			Handler:       _Producer_PersistedIndex_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "shipper.proto",
}
