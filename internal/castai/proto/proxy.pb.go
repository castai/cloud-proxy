// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.2
// 	protoc        v5.27.3
// source: internal/castai/proto/proxy.proto

package proto

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type StreamCloudProxyRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	MessageId    string        `protobuf:"bytes,1,opt,name=message_id,json=messageId,proto3" json:"message_id,omitempty"`
	HttpResponse *HTTPResponse `protobuf:"bytes,2,opt,name=http_response,json=httpResponse,proto3,oneof" json:"http_response,omitempty"`
}

func (x *StreamCloudProxyRequest) Reset() {
	*x = StreamCloudProxyRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_internal_castai_proto_proxy_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *StreamCloudProxyRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StreamCloudProxyRequest) ProtoMessage() {}

func (x *StreamCloudProxyRequest) ProtoReflect() protoreflect.Message {
	mi := &file_internal_castai_proto_proxy_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StreamCloudProxyRequest.ProtoReflect.Descriptor instead.
func (*StreamCloudProxyRequest) Descriptor() ([]byte, []int) {
	return file_internal_castai_proto_proxy_proto_rawDescGZIP(), []int{0}
}

func (x *StreamCloudProxyRequest) GetMessageId() string {
	if x != nil {
		return x.MessageId
	}
	return ""
}

func (x *StreamCloudProxyRequest) GetHttpResponse() *HTTPResponse {
	if x != nil {
		return x.HttpResponse
	}
	return nil
}

type StreamCloudProxyResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	MessageId   string       `protobuf:"bytes,1,opt,name=message_id,json=messageId,proto3" json:"message_id,omitempty"`
	HttpRequest *HTTPRequest `protobuf:"bytes,2,opt,name=http_request,json=httpRequest,proto3,oneof" json:"http_request,omitempty"`
}

func (x *StreamCloudProxyResponse) Reset() {
	*x = StreamCloudProxyResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_internal_castai_proto_proxy_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *StreamCloudProxyResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StreamCloudProxyResponse) ProtoMessage() {}

func (x *StreamCloudProxyResponse) ProtoReflect() protoreflect.Message {
	mi := &file_internal_castai_proto_proxy_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StreamCloudProxyResponse.ProtoReflect.Descriptor instead.
func (*StreamCloudProxyResponse) Descriptor() ([]byte, []int) {
	return file_internal_castai_proto_proxy_proto_rawDescGZIP(), []int{1}
}

func (x *StreamCloudProxyResponse) GetMessageId() string {
	if x != nil {
		return x.MessageId
	}
	return ""
}

func (x *StreamCloudProxyResponse) GetHttpRequest() *HTTPRequest {
	if x != nil {
		return x.HttpRequest
	}
	return nil
}

type HTTPRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Method  string                  `protobuf:"bytes,1,opt,name=method,proto3" json:"method,omitempty"`
	Path    string                  `protobuf:"bytes,2,opt,name=path,proto3" json:"path,omitempty"`
	Body    []byte                  `protobuf:"bytes,3,opt,name=body,proto3,oneof" json:"body,omitempty"`
	Headers map[string]*HeaderValue `protobuf:"bytes,4,rep,name=headers,proto3" json:"headers,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

func (x *HTTPRequest) Reset() {
	*x = HTTPRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_internal_castai_proto_proxy_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *HTTPRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*HTTPRequest) ProtoMessage() {}

func (x *HTTPRequest) ProtoReflect() protoreflect.Message {
	mi := &file_internal_castai_proto_proxy_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use HTTPRequest.ProtoReflect.Descriptor instead.
func (*HTTPRequest) Descriptor() ([]byte, []int) {
	return file_internal_castai_proto_proxy_proto_rawDescGZIP(), []int{2}
}

func (x *HTTPRequest) GetMethod() string {
	if x != nil {
		return x.Method
	}
	return ""
}

func (x *HTTPRequest) GetPath() string {
	if x != nil {
		return x.Path
	}
	return ""
}

func (x *HTTPRequest) GetBody() []byte {
	if x != nil {
		return x.Body
	}
	return nil
}

func (x *HTTPRequest) GetHeaders() map[string]*HeaderValue {
	if x != nil {
		return x.Headers
	}
	return nil
}

type HTTPResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Body       []byte                  `protobuf:"bytes,1,opt,name=body,proto3,oneof" json:"body,omitempty"`
	Error      *string                 `protobuf:"bytes,2,opt,name=error,proto3,oneof" json:"error,omitempty"`
	Status     string                  `protobuf:"bytes,3,opt,name=status,proto3" json:"status,omitempty"`
	Headers    map[string]*HeaderValue `protobuf:"bytes,4,rep,name=headers,proto3" json:"headers,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	StatusCode int32                   `protobuf:"varint,5,opt,name=statusCode,proto3" json:"statusCode,omitempty"`
}

func (x *HTTPResponse) Reset() {
	*x = HTTPResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_internal_castai_proto_proxy_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *HTTPResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*HTTPResponse) ProtoMessage() {}

func (x *HTTPResponse) ProtoReflect() protoreflect.Message {
	mi := &file_internal_castai_proto_proxy_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use HTTPResponse.ProtoReflect.Descriptor instead.
func (*HTTPResponse) Descriptor() ([]byte, []int) {
	return file_internal_castai_proto_proxy_proto_rawDescGZIP(), []int{3}
}

func (x *HTTPResponse) GetBody() []byte {
	if x != nil {
		return x.Body
	}
	return nil
}

func (x *HTTPResponse) GetError() string {
	if x != nil && x.Error != nil {
		return *x.Error
	}
	return ""
}

func (x *HTTPResponse) GetStatus() string {
	if x != nil {
		return x.Status
	}
	return ""
}

func (x *HTTPResponse) GetHeaders() map[string]*HeaderValue {
	if x != nil {
		return x.Headers
	}
	return nil
}

func (x *HTTPResponse) GetStatusCode() int32 {
	if x != nil {
		return x.StatusCode
	}
	return 0
}

type HeaderValue struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Value []string `protobuf:"bytes,1,rep,name=value,proto3" json:"value,omitempty"`
}

func (x *HeaderValue) Reset() {
	*x = HeaderValue{}
	if protoimpl.UnsafeEnabled {
		mi := &file_internal_castai_proto_proxy_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *HeaderValue) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*HeaderValue) ProtoMessage() {}

func (x *HeaderValue) ProtoReflect() protoreflect.Message {
	mi := &file_internal_castai_proto_proxy_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use HeaderValue.ProtoReflect.Descriptor instead.
func (*HeaderValue) Descriptor() ([]byte, []int) {
	return file_internal_castai_proto_proxy_proto_rawDescGZIP(), []int{4}
}

func (x *HeaderValue) GetValue() []string {
	if x != nil {
		return x.Value
	}
	return nil
}

var File_internal_castai_proto_proxy_proto protoreflect.FileDescriptor

var file_internal_castai_proto_proxy_proto_rawDesc = []byte{
	0x0a, 0x21, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2f, 0x63, 0x61, 0x73, 0x74, 0x61,
	0x69, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x70, 0x72, 0x6f, 0x78, 0x79, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x12, 0x15, 0x63, 0x61, 0x73, 0x74, 0x61, 0x69, 0x2e, 0x63, 0x6c, 0x6f, 0x75,
	0x64, 0x2e, 0x70, 0x72, 0x6f, 0x78, 0x79, 0x2e, 0x76, 0x31, 0x22, 0x99, 0x01, 0x0a, 0x17, 0x53,
	0x74, 0x72, 0x65, 0x61, 0x6d, 0x43, 0x6c, 0x6f, 0x75, 0x64, 0x50, 0x72, 0x6f, 0x78, 0x79, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x1d, 0x0a, 0x0a, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67,
	0x65, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x6d, 0x65, 0x73, 0x73,
	0x61, 0x67, 0x65, 0x49, 0x64, 0x12, 0x4d, 0x0a, 0x0d, 0x68, 0x74, 0x74, 0x70, 0x5f, 0x72, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x23, 0x2e, 0x63,
	0x61, 0x73, 0x74, 0x61, 0x69, 0x2e, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x70, 0x72, 0x6f, 0x78,
	0x79, 0x2e, 0x76, 0x31, 0x2e, 0x48, 0x54, 0x54, 0x50, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x48, 0x00, 0x52, 0x0c, 0x68, 0x74, 0x74, 0x70, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x88, 0x01, 0x01, 0x42, 0x10, 0x0a, 0x0e, 0x5f, 0x68, 0x74, 0x74, 0x70, 0x5f, 0x72, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x96, 0x01, 0x0a, 0x18, 0x53, 0x74, 0x72, 0x65, 0x61,
	0x6d, 0x43, 0x6c, 0x6f, 0x75, 0x64, 0x50, 0x72, 0x6f, 0x78, 0x79, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x12, 0x1d, 0x0a, 0x0a, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x5f, 0x69,
	0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65,
	0x49, 0x64, 0x12, 0x4a, 0x0a, 0x0c, 0x68, 0x74, 0x74, 0x70, 0x5f, 0x72, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x22, 0x2e, 0x63, 0x61, 0x73, 0x74, 0x61,
	0x69, 0x2e, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x70, 0x72, 0x6f, 0x78, 0x79, 0x2e, 0x76, 0x31,
	0x2e, 0x48, 0x54, 0x54, 0x50, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x48, 0x00, 0x52, 0x0b,
	0x68, 0x74, 0x74, 0x70, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x88, 0x01, 0x01, 0x42, 0x0f,
	0x0a, 0x0d, 0x5f, 0x68, 0x74, 0x74, 0x70, 0x5f, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x22,
	0x86, 0x02, 0x0a, 0x0b, 0x48, 0x54, 0x54, 0x50, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12,
	0x16, 0x0a, 0x06, 0x6d, 0x65, 0x74, 0x68, 0x6f, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x06, 0x6d, 0x65, 0x74, 0x68, 0x6f, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x70, 0x61, 0x74, 0x68, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x70, 0x61, 0x74, 0x68, 0x12, 0x17, 0x0a, 0x04, 0x62,
	0x6f, 0x64, 0x79, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0c, 0x48, 0x00, 0x52, 0x04, 0x62, 0x6f, 0x64,
	0x79, 0x88, 0x01, 0x01, 0x12, 0x49, 0x0a, 0x07, 0x68, 0x65, 0x61, 0x64, 0x65, 0x72, 0x73, 0x18,
	0x04, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x2f, 0x2e, 0x63, 0x61, 0x73, 0x74, 0x61, 0x69, 0x2e, 0x63,
	0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x70, 0x72, 0x6f, 0x78, 0x79, 0x2e, 0x76, 0x31, 0x2e, 0x48, 0x54,
	0x54, 0x50, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x2e, 0x48, 0x65, 0x61, 0x64, 0x65, 0x72,
	0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x07, 0x68, 0x65, 0x61, 0x64, 0x65, 0x72, 0x73, 0x1a,
	0x5e, 0x0a, 0x0c, 0x48, 0x65, 0x61, 0x64, 0x65, 0x72, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12,
	0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65,
	0x79, 0x12, 0x38, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x22, 0x2e, 0x63, 0x61, 0x73, 0x74, 0x61, 0x69, 0x2e, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e,
	0x70, 0x72, 0x6f, 0x78, 0x79, 0x2e, 0x76, 0x31, 0x2e, 0x48, 0x65, 0x61, 0x64, 0x65, 0x72, 0x56,
	0x61, 0x6c, 0x75, 0x65, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x42,
	0x07, 0x0a, 0x05, 0x5f, 0x62, 0x6f, 0x64, 0x79, 0x22, 0xb9, 0x02, 0x0a, 0x0c, 0x48, 0x54, 0x54,
	0x50, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x17, 0x0a, 0x04, 0x62, 0x6f, 0x64,
	0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x48, 0x00, 0x52, 0x04, 0x62, 0x6f, 0x64, 0x79, 0x88,
	0x01, 0x01, 0x12, 0x19, 0x0a, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x09, 0x48, 0x01, 0x52, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x88, 0x01, 0x01, 0x12, 0x16, 0x0a,
	0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x73,
	0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x4a, 0x0a, 0x07, 0x68, 0x65, 0x61, 0x64, 0x65, 0x72, 0x73,
	0x18, 0x04, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x30, 0x2e, 0x63, 0x61, 0x73, 0x74, 0x61, 0x69, 0x2e,
	0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x70, 0x72, 0x6f, 0x78, 0x79, 0x2e, 0x76, 0x31, 0x2e, 0x48,
	0x54, 0x54, 0x50, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x2e, 0x48, 0x65, 0x61, 0x64,
	0x65, 0x72, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x07, 0x68, 0x65, 0x61, 0x64, 0x65, 0x72,
	0x73, 0x12, 0x1e, 0x0a, 0x0a, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x43, 0x6f, 0x64, 0x65, 0x18,
	0x05, 0x20, 0x01, 0x28, 0x05, 0x52, 0x0a, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x43, 0x6f, 0x64,
	0x65, 0x1a, 0x5e, 0x0a, 0x0c, 0x48, 0x65, 0x61, 0x64, 0x65, 0x72, 0x73, 0x45, 0x6e, 0x74, 0x72,
	0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03,
	0x6b, 0x65, 0x79, 0x12, 0x38, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x22, 0x2e, 0x63, 0x61, 0x73, 0x74, 0x61, 0x69, 0x2e, 0x63, 0x6c, 0x6f, 0x75,
	0x64, 0x2e, 0x70, 0x72, 0x6f, 0x78, 0x79, 0x2e, 0x76, 0x31, 0x2e, 0x48, 0x65, 0x61, 0x64, 0x65,
	0x72, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38,
	0x01, 0x42, 0x07, 0x0a, 0x05, 0x5f, 0x62, 0x6f, 0x64, 0x79, 0x42, 0x08, 0x0a, 0x06, 0x5f, 0x65,
	0x72, 0x72, 0x6f, 0x72, 0x22, 0x23, 0x0a, 0x0b, 0x48, 0x65, 0x61, 0x64, 0x65, 0x72, 0x56, 0x61,
	0x6c, 0x75, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x01, 0x20, 0x03,
	0x28, 0x09, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x32, 0x8a, 0x01, 0x0a, 0x0d, 0x43, 0x6c,
	0x6f, 0x75, 0x64, 0x50, 0x72, 0x6f, 0x78, 0x79, 0x41, 0x50, 0x49, 0x12, 0x79, 0x0a, 0x10, 0x53,
	0x74, 0x72, 0x65, 0x61, 0x6d, 0x43, 0x6c, 0x6f, 0x75, 0x64, 0x50, 0x72, 0x6f, 0x78, 0x79, 0x12,
	0x2e, 0x2e, 0x63, 0x61, 0x73, 0x74, 0x61, 0x69, 0x2e, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x70,
	0x72, 0x6f, 0x78, 0x79, 0x2e, 0x76, 0x31, 0x2e, 0x53, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x43, 0x6c,
	0x6f, 0x75, 0x64, 0x50, 0x72, 0x6f, 0x78, 0x79, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a,
	0x2f, 0x2e, 0x63, 0x61, 0x73, 0x74, 0x61, 0x69, 0x2e, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x70,
	0x72, 0x6f, 0x78, 0x79, 0x2e, 0x76, 0x31, 0x2e, 0x53, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x43, 0x6c,
	0x6f, 0x75, 0x64, 0x50, 0x72, 0x6f, 0x78, 0x79, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x22, 0x00, 0x28, 0x01, 0x30, 0x01, 0x42, 0x09, 0x5a, 0x07, 0x2e, 0x3b, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_internal_castai_proto_proxy_proto_rawDescOnce sync.Once
	file_internal_castai_proto_proxy_proto_rawDescData = file_internal_castai_proto_proxy_proto_rawDesc
)

func file_internal_castai_proto_proxy_proto_rawDescGZIP() []byte {
	file_internal_castai_proto_proxy_proto_rawDescOnce.Do(func() {
		file_internal_castai_proto_proxy_proto_rawDescData = protoimpl.X.CompressGZIP(file_internal_castai_proto_proxy_proto_rawDescData)
	})
	return file_internal_castai_proto_proxy_proto_rawDescData
}

var file_internal_castai_proto_proxy_proto_msgTypes = make([]protoimpl.MessageInfo, 7)
var file_internal_castai_proto_proxy_proto_goTypes = []any{
	(*StreamCloudProxyRequest)(nil),  // 0: castai.cloud.proxy.v1.StreamCloudProxyRequest
	(*StreamCloudProxyResponse)(nil), // 1: castai.cloud.proxy.v1.StreamCloudProxyResponse
	(*HTTPRequest)(nil),              // 2: castai.cloud.proxy.v1.HTTPRequest
	(*HTTPResponse)(nil),             // 3: castai.cloud.proxy.v1.HTTPResponse
	(*HeaderValue)(nil),              // 4: castai.cloud.proxy.v1.HeaderValue
	nil,                              // 5: castai.cloud.proxy.v1.HTTPRequest.HeadersEntry
	nil,                              // 6: castai.cloud.proxy.v1.HTTPResponse.HeadersEntry
}
var file_internal_castai_proto_proxy_proto_depIdxs = []int32{
	3, // 0: castai.cloud.proxy.v1.StreamCloudProxyRequest.http_response:type_name -> castai.cloud.proxy.v1.HTTPResponse
	2, // 1: castai.cloud.proxy.v1.StreamCloudProxyResponse.http_request:type_name -> castai.cloud.proxy.v1.HTTPRequest
	5, // 2: castai.cloud.proxy.v1.HTTPRequest.headers:type_name -> castai.cloud.proxy.v1.HTTPRequest.HeadersEntry
	6, // 3: castai.cloud.proxy.v1.HTTPResponse.headers:type_name -> castai.cloud.proxy.v1.HTTPResponse.HeadersEntry
	4, // 4: castai.cloud.proxy.v1.HTTPRequest.HeadersEntry.value:type_name -> castai.cloud.proxy.v1.HeaderValue
	4, // 5: castai.cloud.proxy.v1.HTTPResponse.HeadersEntry.value:type_name -> castai.cloud.proxy.v1.HeaderValue
	0, // 6: castai.cloud.proxy.v1.CloudProxyAPI.StreamCloudProxy:input_type -> castai.cloud.proxy.v1.StreamCloudProxyRequest
	1, // 7: castai.cloud.proxy.v1.CloudProxyAPI.StreamCloudProxy:output_type -> castai.cloud.proxy.v1.StreamCloudProxyResponse
	7, // [7:8] is the sub-list for method output_type
	6, // [6:7] is the sub-list for method input_type
	6, // [6:6] is the sub-list for extension type_name
	6, // [6:6] is the sub-list for extension extendee
	0, // [0:6] is the sub-list for field type_name
}

func init() { file_internal_castai_proto_proxy_proto_init() }
func file_internal_castai_proto_proxy_proto_init() {
	if File_internal_castai_proto_proxy_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_internal_castai_proto_proxy_proto_msgTypes[0].Exporter = func(v any, i int) any {
			switch v := v.(*StreamCloudProxyRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_internal_castai_proto_proxy_proto_msgTypes[1].Exporter = func(v any, i int) any {
			switch v := v.(*StreamCloudProxyResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_internal_castai_proto_proxy_proto_msgTypes[2].Exporter = func(v any, i int) any {
			switch v := v.(*HTTPRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_internal_castai_proto_proxy_proto_msgTypes[3].Exporter = func(v any, i int) any {
			switch v := v.(*HTTPResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_internal_castai_proto_proxy_proto_msgTypes[4].Exporter = func(v any, i int) any {
			switch v := v.(*HeaderValue); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	file_internal_castai_proto_proxy_proto_msgTypes[0].OneofWrappers = []any{}
	file_internal_castai_proto_proxy_proto_msgTypes[1].OneofWrappers = []any{}
	file_internal_castai_proto_proxy_proto_msgTypes[2].OneofWrappers = []any{}
	file_internal_castai_proto_proxy_proto_msgTypes[3].OneofWrappers = []any{}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_internal_castai_proto_proxy_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   7,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_internal_castai_proto_proxy_proto_goTypes,
		DependencyIndexes: file_internal_castai_proto_proxy_proto_depIdxs,
		MessageInfos:      file_internal_castai_proto_proxy_proto_msgTypes,
	}.Build()
	File_internal_castai_proto_proxy_proto = out.File
	file_internal_castai_proto_proxy_proto_rawDesc = nil
	file_internal_castai_proto_proxy_proto_goTypes = nil
	file_internal_castai_proto_proxy_proto_depIdxs = nil
}
