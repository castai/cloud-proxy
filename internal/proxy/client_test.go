package proxy

import (
	"bytes"
	mock_cloud "cloud-proxy/internal/proxy/mock"
	cloudproxyv1alpha "cloud-proxy/proto/v1alpha"
	proto "cloud-proxy/proto/v1alpha"
	"context"
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"testing"
	"time"
)

type mockReadCloserErr struct{}

func (m mockReadCloserErr) Read(_ []byte) (int, error) {
	return 0, io.ErrUnexpectedEOF
}
func (m mockReadCloserErr) Close() error { return nil }

func TestClient_toResponse(t *testing.T) {
	t.Parallel()
	type fields struct {
		//tuneMockCredentials func(m *mock_gcp.MockCredentials)
		//httpClient     *http.Client
	}
	type args struct {
		msgID string
		resp  *http.Response
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *proto.HTTPResponse
	}{
		{
			name: "nil response",
			want: &proto.HTTPResponse{
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
			want: &proto.HTTPResponse{
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
			want: &proto.HTTPResponse{
				Body:    []byte("body"),
				Status:  200,
				Headers: map[string]*proto.HeaderValue{"header": {Value: []string{"value"}}},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := New(nil, nil, "clusterID", "version")
			got := c.toResponse(tt.args.resp)
			//diff := cmp.Diff(got, tt.want, protocmp.Transform())
			//require.Empty(t, diff)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestClient_toHTTPRequest(t *testing.T) {
	t.Parallel()

	type args struct {
		req *proto.HTTPRequest
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
				req: &proto.HTTPRequest{
					Path: "\n\t\f",
				},
			},
			wantErr: true,
		},
		{
			name: "success",
			args: args{
				req: &proto.HTTPRequest{
					Method: "GET",
					Headers: map[string]*proto.HeaderValue{
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
			c := New(nil, nil, "clusterID", "version")
			got, err := c.toHTTPRequest(tt.args.req)
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
		tuneMockCloudClient func(m *mock_cloud.MockCloudClient)
	}
	type args struct {
		in             *cloudproxyv1alpha.StreamCloudProxyResponse
		tuneMockStream func(m *mock_cloud.MockStreamCloudProxyClient)
	}
	tests := []struct {
		name                 string
		fields               fields
		args                 args
		wantLastSeenUpdated  bool
		wantKeepAlive        int64
		wantKeepAliveTimeout int64
		wantErrCount         int64
	}{
		{
			name:                 "nil response",
			wantLastSeenUpdated:  false,
			wantKeepAlive:        int64(KeepAliveDefault),
			wantKeepAliveTimeout: int64(KeepAliveTimeoutDefault),
		},
		{
			name: "keep alive",
			args: args{
				in: &cloudproxyv1alpha.StreamCloudProxyResponse{
					MessageId: KeepAliveMessageID,
				},
			},
			wantLastSeenUpdated:  true,
			wantKeepAlive:        int64(KeepAliveDefault),
			wantKeepAliveTimeout: int64(KeepAliveTimeoutDefault),
		},
		{
			name: "keep alive timeout and keepalive",
			args: args{
				in: &cloudproxyv1alpha.StreamCloudProxyResponse{
					MessageId: KeepAliveMessageID,
					ConfigurationRequest: &proto.ConfigurationRequest{
						KeepAlive:        1,
						KeepAliveTimeout: 2,
					},
				},
			},
			wantLastSeenUpdated:  true,
			wantKeepAlive:        1,
			wantKeepAliveTimeout: 2,
		},
		{
			name: "http error, send error",
			args: args{
				in: &cloudproxyv1alpha.StreamCloudProxyResponse{
					MessageId:   "msgID",
					HttpRequest: &proto.HTTPRequest{},
				},
				tuneMockStream: func(m *mock_cloud.MockStreamCloudProxyClient) {
					m.EXPECT().Send(&cloudproxyv1alpha.StreamCloudProxyRequest{
						Request: &cloudproxyv1alpha.StreamCloudProxyRequest_Response{
							Response: &cloudproxyv1alpha.ClusterResponse{
								ClientMetadata: &cloudproxyv1alpha.ClientMetadata{
									ClusterId: "clusterID",
								},
								MessageId: "msgID",
								HttpResponse: &cloudproxyv1alpha.HTTPResponse{
									Error: lo.ToPtr("c.cloudClient.DoHTTPRequest: error"),
								},
							},
						},
					}).Return(fmt.Errorf("error"))
				},
			},
			fields: fields{
				tuneMockCloudClient: func(m *mock_cloud.MockCloudClient) {
					m.EXPECT().DoHTTPRequest(gomock.Any()).Return(nil, fmt.Errorf("error"))
				},
			},
			wantLastSeenUpdated:  false,
			wantKeepAlive:        int64(KeepAliveDefault),
			wantKeepAliveTimeout: int64(KeepAliveTimeoutDefault),
			wantErrCount:         1,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			cloudClient := mock_cloud.NewMockCloudClient(ctrl)
			if tt.fields.tuneMockCloudClient != nil {
				tt.fields.tuneMockCloudClient(cloudClient)
			}
			c := New(cloudClient, logrus.New(), "clusterID", "version")
			stream := mock_cloud.NewMockStreamCloudProxyClient(ctrl)
			if tt.args.tuneMockStream != nil {
				tt.args.tuneMockStream(stream)
			}

			c.handleMessage(tt.args.in, stream)
			require.Equal(t, tt.wantLastSeenUpdated, c.lastSeen.Load() > 0, "lastSeen: %v", c.lastSeen.Load())
			require.Equal(t, tt.wantKeepAlive, c.keepAlive.Load(), "keepAlive: %v", c.keepAlive.Load())
			require.Equal(t, tt.wantKeepAliveTimeout, c.keepAliveTimeout.Load(), "keepAliveTimeout: %v", c.keepAliveTimeout.Load())
			require.Equal(t, tt.wantErrCount, c.errCount.Load(), "errCount: %v", c.errCount.Load())
		})
	}
}

func TestClient_processHttpRequest(t *testing.T) {
	t.Parallel()
	type fields struct {
		tuneMockCloudClient func(m *mock_cloud.MockCloudClient)
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
				req: &proto.HTTPRequest{
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
				req: &proto.HTTPRequest{},
			},
			fields: fields{
				tuneMockCloudClient: func(m *mock_cloud.MockCloudClient) {
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
				req: &proto.HTTPRequest{},
			},
			fields: fields{
				tuneMockCloudClient: func(m *mock_cloud.MockCloudClient) {
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
			cloudClient := mock_cloud.NewMockCloudClient(ctrl)
			if tt.fields.tuneMockCloudClient != nil {
				tt.fields.tuneMockCloudClient(cloudClient)
			}
			c := New(cloudClient, logrus.New(), "clusterID", "version")
			if got := c.processHttpRequest(tt.args.req); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("processHttpRequest() = %v, want %v", got, tt.want)
			}
			require.Equal(t, tt.wantProcessCount, c.processedCount.Load(), "processedCount: %v", c.processedCount.Load())
		})
	}
}

func TestClient_sendKeepAlive(t *testing.T) {
	t.Parallel()

	type args struct {
		ctx              func() context.Context
		tuneMockStream   func(m *mock_cloud.MockStreamCloudProxyClient)
		keepAlive        int64
		keepAliveTimeout int64
	}
	tests := []struct {
		name           string
		args           args
		isLastSeenZero bool
	}{
		{
			name: "end of ticker",
			args: args{
				ctx: func() context.Context {
					return context.Background()
				},
				keepAlive: 0,
			},
		},
		{
			name: "context done",
			args: args{
				ctx: func() context.Context {
					ctx, cancel := context.WithCancel(context.Background())
					cancel()
					return ctx
				},
			},
		},
		{
			name: "send returned error, should exit",
			args: args{
				ctx: func() context.Context {
					return context.Background()
				},
				tuneMockStream: func(m *mock_cloud.MockStreamCloudProxyClient) {
					m.EXPECT().Send(gomock.Any()).Return(fmt.Errorf("error"))
				},
				keepAlive:        int64(time.Second),
				keepAliveTimeout: int64(10 * time.Minute),
			},
			isLastSeenZero: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			c := New(nil, logrus.New(), "clusterID", "version")
			c.keepAlive.Store(tt.args.keepAlive)
			c.keepAliveTimeout.Store(tt.args.keepAliveTimeout)

			stream := mock_cloud.NewMockStreamCloudProxyClient(ctrl)
			if tt.args.tuneMockStream != nil {
				tt.args.tuneMockStream(stream)
			}
			c.lastSeen.Store(time.Now().UnixNano())

			c.sendKeepAlive(tt.args.ctx(), stream)
			require.Equal(t, tt.isLastSeenZero, c.lastSeen.Load() == 0, "lastSeen: %v", c.lastSeen.Load())
		})
	}
}
