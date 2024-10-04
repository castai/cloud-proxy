// nolint: govet,gocritic
package proxy

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"

	"cloud-proxy/internal/config"
	mock_proxy "cloud-proxy/internal/proxy/mock"
	cloudproxyv1alpha "cloud-proxy/proto/gen/proto/v1alpha"
)

type mockReadCloserErr struct{}

func (m mockReadCloserErr) Read(_ []byte) (int, error) {
	return 0, io.ErrUnexpectedEOF
}
func (m mockReadCloserErr) Close() error { return nil }

func TestClient_toResponse(t *testing.T) {
	t.Parallel()
	type args struct {
		msgID string
		resp  *http.Response
	}
	tests := []struct {
		name string
		args args
		want *cloudproxyv1alpha.HTTPResponse
	}{
		{
			name: "nil response",
			want: &cloudproxyv1alpha.HTTPResponse{
				Error: lo.ToPtr("nil response"),
			},
		},
		{
			name: "error reading response body",
			args: args{
				resp: &http.Response{
					Body: &mockReadCloserErr{},
				},
			},
			want: &cloudproxyv1alpha.HTTPResponse{
				Error: lo.ToPtr("io.ReadAll: body for error: unexpected EOF"),
			},
		},
		{
			name: "success",
			args: args{
				msgID: "msgID",
				resp: &http.Response{
					StatusCode: 200,
					Header:     http.Header{"header": {"value"}},
					Body:       io.NopCloser(bytes.NewReader([]byte("body"))),
				},
			},
			want: &cloudproxyv1alpha.HTTPResponse{
				Body:    []byte("body"),
				Status:  200,
				Headers: map[string]*cloudproxyv1alpha.HeaderValue{"header": {Value: []string{"value"}}},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := toResponse(tt.args.resp)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestClient_toHTTPRequest(t *testing.T) {
	t.Parallel()

	type args struct {
		req *cloudproxyv1alpha.HTTPRequest
	}
	tests := []struct {
		name    string
		args    args
		want    *http.Request
		wantErr bool
	}{
		{
			name:    "req is nil",
			wantErr: true,
		},
		{
			name: "error creating http request",
			args: args{
				req: &cloudproxyv1alpha.HTTPRequest{
					Path: "\n\t\f",
				},
			},
			wantErr: true,
		},
		{
			name: "success",
			args: args{
				req: &cloudproxyv1alpha.HTTPRequest{
					Method: "GET",
					Headers: map[string]*cloudproxyv1alpha.HeaderValue{
						"header": {Value: []string{"value"}},
					},
					Body: []byte("body"),
				},
			},
			want: &http.Request{
				Proto:         "HTTP/1.1",
				ContentLength: int64(len("body")),
				ProtoMajor:    1,
				ProtoMinor:    1,
				Method:        "GET",
				Header: http.Header{
					"Header": {"value"},
				},
				URL:  &url.URL{},
				Body: io.NopCloser(bytes.NewReader([]byte("body"))),
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := toHTTPRequest(tt.args.req)
			require.Equal(t, tt.wantErr, err != nil, err)
			if err != nil {
				return
			}
			got.GetBody = nil
			tt.want = tt.want.WithContext(context.Background())
			require.Equal(t, tt.want, got)
		})
	}
}

func TestClient_handleMessage(t *testing.T) {
	t.Parallel()

	type fields struct {
		tuneMockCloudClient func(m *mock_proxy.MockCloudClient)
	}
	type args struct {
		ctx    func() context.Context
		in     *cloudproxyv1alpha.StreamCloudProxyResponse
		sendCh chan *cloudproxyv1alpha.StreamCloudProxyRequest
	}
	tests := []struct {
		name                 string
		fields               fields
		args                 args
		wantKeepAlive        int64
		wantKeepAliveTimeout int64
		wantErrCount         int64
		wantMsgID            string
		wantMsgError         string
	}{
		{
			name: "nil response",
			args: args{
				ctx: func() context.Context {
					return context.Background()
				},
			},
			wantKeepAlive:        int64(config.KeepAliveDefault),
			wantKeepAliveTimeout: int64(config.KeepAliveTimeoutDefault),
		},
		{
			name: "keep alive",
			args: args{
				ctx: func() context.Context {
					return context.Background()
				},
				in: &cloudproxyv1alpha.StreamCloudProxyResponse{
					MessageId: KeepAliveMessageID,
				},
			},
			wantKeepAlive:        int64(config.KeepAliveDefault),
			wantKeepAliveTimeout: int64(config.KeepAliveTimeoutDefault),
		},
		{
			name: "keep alive timeout and keepalive",
			args: args{
				ctx: func() context.Context {
					return context.Background()
				},
				in: &cloudproxyv1alpha.StreamCloudProxyResponse{
					MessageId: KeepAliveMessageID,
					Response: &cloudproxyv1alpha.StreamCloudProxyResponse_ConfigurationRequest{
						ConfigurationRequest: &cloudproxyv1alpha.ConfigurationRequest{
							KeepAlive:        1,
							KeepAliveTimeout: 2,
						},
					},
				},
			},
			wantKeepAlive:        1,
			wantKeepAliveTimeout: 2,
		},
		{
			name: "http request is nil",
			args: args{
				ctx: func() context.Context {
					ctx, _ := context.WithTimeout(context.Background(), time.Second*2)
					return ctx
				},
				sendCh: make(chan *cloudproxyv1alpha.StreamCloudProxyRequest, 1),
				in: &cloudproxyv1alpha.StreamCloudProxyResponse{
					MessageId: "msgID",
				},
			},
			wantErrCount:         0,
			wantKeepAlive:        int64(config.KeepAliveDefault),
			wantKeepAliveTimeout: int64(config.KeepAliveTimeoutDefault),
			wantMsgID:            "msgID",
			wantMsgError:         "nil http request",
		},
		{
			name: "send is full exit on context done",
			args: args{
				ctx: func() context.Context {
					ctx, _ := context.WithTimeout(context.Background(), time.Second*2)
					return ctx
				},
				sendCh: make(chan *cloudproxyv1alpha.StreamCloudProxyRequest),
				in: &cloudproxyv1alpha.StreamCloudProxyResponse{
					MessageId: "msgID",
				},
			},
			wantErrCount:         0,
			wantKeepAlive:        int64(config.KeepAliveDefault),
			wantKeepAliveTimeout: int64(config.KeepAliveTimeoutDefault),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			cloudClient := mock_proxy.NewMockCloudClient(ctrl)
			if tt.fields.tuneMockCloudClient != nil {
				tt.fields.tuneMockCloudClient(cloudClient)
			}
			c := &Client{
				sendCh:      tt.args.sendCh,
				log:         logrus.New(),
				cloudClient: cloudClient,
			}
			c.keepAlive.Store(int64(config.KeepAliveDefault))
			c.keepAliveTimeout.Store(int64(config.KeepAliveTimeoutDefault))

			c.handleMessage(tt.args.ctx(), tt.args.in)
			require.Equal(t, tt.wantKeepAlive, c.keepAlive.Load(), "keepAlive: %v", c.keepAlive.Load())
			require.Equal(t, tt.wantKeepAliveTimeout, c.keepAliveTimeout.Load(), "keepAliveTimeout: %v", c.keepAliveTimeout.Load())
			require.Equal(t, tt.wantErrCount, c.errCount.Load(), "errCount: %v", c.errCount.Load())
			var msg *cloudproxyv1alpha.StreamCloudProxyRequest
			select {
			case msg = <-c.sendCh:
			default:
				break
			}
			require.Equal(t, tt.wantMsgID, msg.GetResponse().GetMessageId(), "msgID: %v", c.sendCh)
			require.Equal(t, tt.wantMsgError, msg.GetResponse().GetHttpResponse().GetError(), "msg response error: %v", c.sendCh)
		})
	}
}

func TestClient_processHttpRequest(t *testing.T) {
	t.Parallel()
	type fields struct {
		tuneMockCloudClient func(m *mock_proxy.MockCloudClient)
	}
	type args struct {
		req *cloudproxyv1alpha.HTTPRequest
	}
	tests := []struct {
		name             string
		fields           fields
		args             args
		want             *cloudproxyv1alpha.HTTPResponse
		wantProcessCount int64
	}{
		{
			name: "nil request",
			want: &cloudproxyv1alpha.HTTPResponse{
				Error: lo.ToPtr("nil http request"),
			},
		},
		{
			name: "error creating http request",
			args: args{
				req: &cloudproxyv1alpha.HTTPRequest{
					Path: "\n\t\f",
				},
			},
			want: &cloudproxyv1alpha.HTTPResponse{
				Error: lo.ToPtr("toHTTPRequest: http.NewRequest: error: parse \"\\n\\t\\f\": net/url: invalid control character in URL"),
			},
		},
		{
			name: "cloud client error",
			args: args{
				req: &cloudproxyv1alpha.HTTPRequest{},
			},
			fields: fields{
				tuneMockCloudClient: func(m *mock_proxy.MockCloudClient) {
					m.EXPECT().DoHTTPRequest(gomock.Any()).Return(nil, fmt.Errorf("error"))
				},
			},
			want: &cloudproxyv1alpha.HTTPResponse{
				Error: lo.ToPtr("c.cloudClient.DoHTTPRequest: error"),
			},
		},
		{
			name: "success",
			args: args{
				req: &cloudproxyv1alpha.HTTPRequest{},
			},
			fields: fields{
				tuneMockCloudClient: func(m *mock_proxy.MockCloudClient) {
					m.EXPECT().DoHTTPRequest(gomock.Any()).Return(&http.Response{}, nil)
				},
			},
			want:             &cloudproxyv1alpha.HTTPResponse{},
			wantProcessCount: 1,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			cloudClient := mock_proxy.NewMockCloudClient(ctrl)
			if tt.fields.tuneMockCloudClient != nil {
				tt.fields.tuneMockCloudClient(cloudClient)
			}
			c := New(cloudClient, logrus.New(), "version", &config.Config{
				ClusterID: "clusterID",
				PodMetadata: config.PodMetadata{
					PodName: "podName",
				},
				KeepAlive:        time.Second,
				KeepAliveTimeout: time.Minute,
			})
			if got := c.processHTTPRequest(tt.args.req); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("processHttpRequest() = %v, want %v", got, tt.want)
			}
			require.Equal(t, tt.wantProcessCount, c.processedCount.Load(), "processedCount: %v", c.processedCount.Load())
		})
	}
}

func TestClient_sendAndReceive(t *testing.T) {
	t.Parallel()

	type args struct {
		ctx            func() context.Context
		tuneMockStream func(m *mock_proxy.MockCloudProxyAPI_StreamCloudProxyClient)
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "context done",
			args: args{
				ctx: func() context.Context {
					ctx, cancel := context.WithCancel(context.Background())
					cancel()
					return ctx
				},
				tuneMockStream: func(m *mock_proxy.MockCloudProxyAPI_StreamCloudProxyClient) {
					m.EXPECT().Send(gomock.Any()).Return(nil).AnyTimes() // expected 0 or 1 times.
					m.EXPECT().Context().DoAndReturn(func() context.Context {
						ctx, cancel := context.WithCancel(context.Background())
						cancel()
						return ctx
					}).AnyTimes() // expected 0 or 1 times.
					m.EXPECT().CloseSend().Return(nil)
				},
			},
			wantErr: true,
		},
		{
			name: "receive not alive",
			args: args{
				ctx: func() context.Context {
					return context.Background()
				},
				tuneMockStream: func(m *mock_proxy.MockCloudProxyAPI_StreamCloudProxyClient) {
					m.EXPECT().Send(gomock.Any()).Return(nil).AnyTimes()         // expected 0 or 1 times.
					m.EXPECT().Context().Return(context.Background()).AnyTimes() // expected 0 or 1 times.
					m.EXPECT().Recv().Return(nil, fmt.Errorf("test error")).AnyTimes()
					m.EXPECT().CloseSend().Return(nil).AnyTimes()
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			c := New(nil, logrus.New(), "version", &config.Config{
				ClusterID: "clusterID",
				PodMetadata: config.PodMetadata{
					PodName: "podName",
				},
				KeepAlive:        time.Second,
				KeepAliveTimeout: time.Hour,
			})
			c.lastSeenReceive.Store(time.Now().UnixNano())
			c.lastSeenSend.Store(time.Now().UnixNano())
			stream := mock_proxy.NewMockCloudProxyAPI_StreamCloudProxyClient(ctrl)
			if tt.args.tuneMockStream != nil {
				tt.args.tuneMockStream(stream)
			}

			if err := c.sendAndReceive(tt.args.ctx(), stream); (err != nil) != tt.wantErr {
				t.Errorf("run() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClient_Run(t *testing.T) {
	t.Parallel()
	type args struct {
		ctx func() context.Context
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "context done",
			args: args{
				ctx: func() context.Context {
					ctx, cancel := context.WithCancel(context.Background())
					cancel()
					return ctx
				},
			},
			wantErr: true,
		},
		{
			name: "get stream error",
			args: args{
				ctx: func() context.Context {
					ctx, _ := context.WithTimeout(context.Background(), time.Second*2)
					return ctx
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			c := New(nil, logrus.New(), "tt.fields.version", &config.Config{KeepAlive: time.Second, KeepAliveTimeout: time.Hour})
			if err := c.Run(tt.args.ctx()); (err != nil) != tt.wantErr {
				t.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
