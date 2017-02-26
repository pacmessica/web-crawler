// Code generated by protoc-gen-go.
// source: PageGetter.proto
// DO NOT EDIT!

/*
Package PageGetter is a generated protocol buffer package.

It is generated from these files:
	PageGetter.proto

It has these top-level messages:
	Result
	Request
	Search
*/
package PageGetter

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

import (
	client "github.com/micro/go-micro/client"
	server "github.com/micro/go-micro/server"
	context "golang.org/x/net/context"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type Result struct {
	Pageids []string `protobuf:"bytes,1,rep,name=pageids" json:"pageids,omitempty"`
}

func (m *Result) Reset()                    { *m = Result{} }
func (m *Result) String() string            { return proto.CompactTextString(m) }
func (*Result) ProtoMessage()               {}
func (*Result) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *Result) GetPageids() []string {
	if m != nil {
		return m.Pageids
	}
	return nil
}

type Request struct {
	Id     string  `protobuf:"bytes,1,opt,name=id" json:"id,omitempty"`
	Search *Search `protobuf:"bytes,2,opt,name=search" json:"search,omitempty"`
}

func (m *Request) Reset()                    { *m = Request{} }
func (m *Request) String() string            { return proto.CompactTextString(m) }
func (*Request) ProtoMessage()               {}
func (*Request) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *Request) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func (m *Request) GetSearch() *Search {
	if m != nil {
		return m.Search
	}
	return nil
}

type Search struct {
	Or     *Search_Or  `protobuf:"bytes,1,opt,name=or" json:"or,omitempty"`
	And    *Search_And `protobuf:"bytes,2,opt,name=and" json:"and,omitempty"`
	Term   string      `protobuf:"bytes,3,opt,name=term" json:"term,omitempty"`
	Phrase string      `protobuf:"bytes,4,opt,name=phrase" json:"phrase,omitempty"`
}

func (m *Search) Reset()                    { *m = Search{} }
func (m *Search) String() string            { return proto.CompactTextString(m) }
func (*Search) ProtoMessage()               {}
func (*Search) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *Search) GetOr() *Search_Or {
	if m != nil {
		return m.Or
	}
	return nil
}

func (m *Search) GetAnd() *Search_And {
	if m != nil {
		return m.And
	}
	return nil
}

func (m *Search) GetTerm() string {
	if m != nil {
		return m.Term
	}
	return ""
}

func (m *Search) GetPhrase() string {
	if m != nil {
		return m.Phrase
	}
	return ""
}

type Search_Or struct {
	Search []*Search `protobuf:"bytes,1,rep,name=search" json:"search,omitempty"`
}

func (m *Search_Or) Reset()                    { *m = Search_Or{} }
func (m *Search_Or) String() string            { return proto.CompactTextString(m) }
func (*Search_Or) ProtoMessage()               {}
func (*Search_Or) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2, 0} }

func (m *Search_Or) GetSearch() []*Search {
	if m != nil {
		return m.Search
	}
	return nil
}

type Search_And struct {
	Search []*Search `protobuf:"bytes,1,rep,name=search" json:"search,omitempty"`
}

func (m *Search_And) Reset()                    { *m = Search_And{} }
func (m *Search_And) String() string            { return proto.CompactTextString(m) }
func (*Search_And) ProtoMessage()               {}
func (*Search_And) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2, 1} }

func (m *Search_And) GetSearch() []*Search {
	if m != nil {
		return m.Search
	}
	return nil
}

func init() {
	proto.RegisterType((*Result)(nil), "Result")
	proto.RegisterType((*Request)(nil), "Request")
	proto.RegisterType((*Search)(nil), "Search")
	proto.RegisterType((*Search_Or)(nil), "Search.Or")
	proto.RegisterType((*Search_And)(nil), "Search.And")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ client.Option
var _ server.Option

// Publisher API

type Publisher interface {
	Publish(ctx context.Context, msg interface{}, opts ...client.PublishOption) error
}

type publisher struct {
	c     client.Client
	topic string
}

func (p *publisher) Publish(ctx context.Context, msg interface{}, opts ...client.PublishOption) error {
	return p.c.Publish(ctx, p.c.NewPublication(p.topic, msg), opts...)
}

func NewPublisher(topic string, c client.Client) Publisher {
	if c == nil {
		c = client.NewClient()
	}
	return &publisher{c, topic}
}

// Subscriber API

func RegisterSubscriber(topic string, s server.Server, h interface{}, opts ...server.SubscriberOption) error {
	return s.Subscribe(s.NewSubscriber(topic, h, opts...))
}

// Client API for PageGetter service

type PageGetterClient interface {
	GetPagesFromQuery(ctx context.Context, in *Request, opts ...client.CallOption) (*Result, error)
}

type pageGetterClient struct {
	c           client.Client
	serviceName string
}

func NewPageGetterClient(serviceName string, c client.Client) PageGetterClient {
	if c == nil {
		c = client.NewClient()
	}
	if len(serviceName) == 0 {
		serviceName = "pagegetter"
	}
	return &pageGetterClient{
		c:           c,
		serviceName: serviceName,
	}
}

func (c *pageGetterClient) GetPagesFromQuery(ctx context.Context, in *Request, opts ...client.CallOption) (*Result, error) {
	req := c.c.NewRequest(c.serviceName, "PageGetter.GetPagesFromQuery", in)
	out := new(Result)
	err := c.c.Call(ctx, req, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for PageGetter service

type PageGetterHandler interface {
	GetPagesFromQuery(context.Context, *Request, *Result) error
}

func RegisterPageGetterHandler(s server.Server, hdlr PageGetterHandler, opts ...server.HandlerOption) {
	s.Handle(s.NewHandler(&PageGetter{hdlr}, opts...))
}

type PageGetter struct {
	PageGetterHandler
}

func (h *PageGetter) GetPagesFromQuery(ctx context.Context, in *Request, out *Result) error {
	return h.PageGetterHandler.GetPagesFromQuery(ctx, in, out)
}

func init() { proto.RegisterFile("PageGetter.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 253 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x84, 0x90, 0xc1, 0x4a, 0xc4, 0x30,
	0x10, 0x86, 0x6d, 0xba, 0xb4, 0xee, 0x14, 0x44, 0xe7, 0x20, 0xa1, 0x20, 0x96, 0x82, 0xd2, 0x53,
	0x0f, 0x15, 0x3c, 0x78, 0xdb, 0x8b, 0x7b, 0x5c, 0x8d, 0x4f, 0x10, 0xcd, 0xb0, 0x5b, 0x70, 0x9b,
	0x3a, 0x49, 0x0f, 0xbe, 0x9c, 0xcf, 0x26, 0xcd, 0x66, 0x51, 0xf0, 0xe0, 0x2d, 0xff, 0x97, 0x49,
	0xf2, 0xe7, 0x83, 0xf3, 0x27, 0xbd, 0xa5, 0x35, 0x79, 0x4f, 0xdc, 0x8e, 0x6c, 0xbd, 0xad, 0x6b,
	0xc8, 0x14, 0xb9, 0xe9, 0xdd, 0xa3, 0x84, 0x7c, 0xd4, 0x5b, 0xea, 0x8d, 0x93, 0x49, 0x95, 0x36,
	0x4b, 0x75, 0x8c, 0xf5, 0x03, 0xe4, 0x8a, 0x3e, 0x26, 0x72, 0x1e, 0xcf, 0x40, 0xf4, 0x46, 0x26,
	0x55, 0xd2, 0x2c, 0x95, 0xe8, 0x0d, 0x5e, 0x43, 0xe6, 0x48, 0xf3, 0xdb, 0x4e, 0x8a, 0x2a, 0x69,
	0x8a, 0x2e, 0x6f, 0x5f, 0x42, 0x54, 0x11, 0xd7, 0x5f, 0x09, 0x64, 0x07, 0x84, 0x25, 0x08, 0xcb,
	0xe1, 0x6c, 0xd1, 0x41, 0x9c, 0x6b, 0x37, 0xac, 0x84, 0x65, 0xbc, 0x82, 0x54, 0x0f, 0x26, 0x5e,
	0x52, 0x1c, 0x37, 0x57, 0x83, 0x51, 0x33, 0x47, 0x84, 0x85, 0x27, 0xde, 0xcb, 0x34, 0x3c, 0x1c,
	0xd6, 0x78, 0x09, 0xd9, 0xb8, 0x63, 0xed, 0x48, 0x2e, 0x02, 0x8d, 0xa9, 0xbc, 0x01, 0xb1, 0xe1,
	0x5f, 0xc5, 0xe6, 0xcf, 0xfc, 0x2d, 0x56, 0xde, 0x42, 0xba, 0x1a, 0xcc, 0xbf, 0x73, 0xdd, 0x3d,
	0xc0, 0x8f, 0x34, 0x6c, 0xe0, 0x62, 0x4d, 0x7e, 0x06, 0xee, 0x91, 0xed, 0xfe, 0x79, 0x22, 0xfe,
	0xc4, 0xd3, 0x36, 0xea, 0x29, 0xf3, 0xf6, 0x20, 0xb3, 0x3e, 0x79, 0xcd, 0x82, 0xdf, 0xbb, 0xef,
	0x00, 0x00, 0x00, 0xff, 0xff, 0xae, 0x98, 0x46, 0x09, 0x73, 0x01, 0x00, 0x00,
}
