// nolint: gocritic
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
	type fields struct {
		// tuneMockCredentials func(m *mock_gcp.MockCredentials)
		// httpClient     *http.Client.
	}
	type args struct {
		msgID string
		resp  *http.Response
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *cloudproxyv1alpha.HTTPResponse
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
			c := New(nil, nil, nil, "podName", "clusterID", "version", "apiKey", time.Second, time.Minute)
			got := c.toResponse(tt.args.resp)
			// diff := cmp.Diff(got, tt.want, protocmp.Transform())
			// require.Empty(t, diff).
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
			c := New(nil, nil, nil, "podName", "clusterID", "version", "apiKey", time.Second, time.Minute)
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
		tuneMockCloudClient func(m *mock_proxy.MockCloudClient)
	}
	type args struct {
		in             *cloudproxyv1alpha.StreamCloudProxyResponse
		tuneMockStream func(m *mock_proxy.MockCloudProxyAPI_StreamCloudProxyClient)
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
			wantKeepAlive:        int64(config.KeepAliveDefault),
			wantKeepAliveTimeout: int64(config.KeepAliveTimeoutDefault),
		},
		{
			name: "keep alive",
			args: args{
				in: &cloudproxyv1alpha.StreamCloudProxyResponse{
					MessageId: KeepAliveMessageID,
				},
			},
			wantLastSeenUpdated:  true,
			wantKeepAlive:        int64(config.KeepAliveDefault),
			wantKeepAliveTimeout: int64(config.KeepAliveTimeoutDefault),
		},
		{
			name: "keep alive timeout and keepalive",
			args: args{
				in: &cloudproxyv1alpha.StreamCloudProxyResponse{
					MessageId: KeepAliveMessageID,
					ConfigurationRequest: &cloudproxyv1alpha.ConfigurationRequest{
						KeepAlive:        1,
						KeepAliveTimeout: 2,
					},
				},
			},
			wantLastSeenUpdated:  true,
			wantKeepAlive:        1,
			wantKeepAliveTimeout: 2,
		},
		//{
		//	name: "http error, send error",
		//	args: args{
		//		in: &cloudproxyv1alpha.StreamCloudProxyResponse{
		//			MessageId:   "msgID",
		//			HttpRequest: &cloudproxyv1alpha.HTTPRequest{},
		//		},
		//		tuneMockStream: func(m *mock_proxy.MockCloudProxyAPI_StreamCloudProxyClient) {
		//			m.EXPECT().Send(&cloudproxyv1alpha.StreamCloudProxyRequest{
		//				Request: &cloudproxyv1alpha.StreamCloudProxyRequest_Response{
		//					Response: &cloudproxyv1alpha.ClusterResponse{
		//						ClientMetadata: &cloudproxyv1alpha.ClientMetadata{
		//							PodName:   "podName",
		//							ClusterId: "clusterID",
		//						},
		//						MessageId: "msgID",
		//						HttpResponse: &cloudproxyv1alpha.HTTPResponse{
		//							Error: lo.ToPtr("c.cloudClient.DoHTTPRequest: error"),
		//						},
		//					},
		//				},
		//			}).Return(fmt.Errorf("error"))
		//		},
		//	},
		//	fields: fields{
		//		tuneMockCloudClient: func(m *mock_proxy.MockCloudClient) {
		//			m.EXPECT().DoHTTPRequest(gomock.Any()).Return(nil, fmt.Errorf("error"))
		//		},
		//	},
		//	wantLastSeenUpdated:  false,
		//	wantKeepAlive:        int64(config.KeepAliveDefault),
		//	wantKeepAliveTimeout: int64(config.KeepAliveTimeoutDefault),
		//	wantErrCount:         1,
		//}.
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
			c := New(nil, cloudClient, logrus.New(), "podName", "clusterID", "version", "apiKey", config.KeepAliveDefault, config.KeepAliveTimeoutDefault)
			stream := mock_proxy.NewMockCloudProxyAPI_StreamCloudProxyClient(ctrl)
			if tt.args.tuneMockStream != nil {
				tt.args.tuneMockStream(stream)
			}

			msgStream := make(chan *cloudproxyv1alpha.StreamCloudProxyRequest)
			go func() {
				<-msgStream
			}()

			c.handleMessage(context.Background(), tt.args.in, msgStream)
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
			c := New(nil, cloudClient, logrus.New(), "podName", "clusterID", "version", "apiKey", time.Second, time.Minute)
			if got := c.processHTTPRequest(tt.args.req); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("processHttpRequest() = %v, want %v", got, tt.want)
			}
			require.Equal(t, tt.wantProcessCount, c.processedCount.Load(), "processedCount: %v", c.processedCount.Load())
		})
	}
}

// nolint
//func TestClient_sendKeepAlive(t *testing.T) {
//	t.Parallel()
//
//	type args struct {
//		tuneMockStream   func(m *mock_proxy.MockCloudProxyAPI_StreamCloudProxyClient)
//		keepAlive        int64
//		keepAliveTimeout int64
//	}
//	tests := []struct {
//		name           string
//		args           args
//		isLastSeenZero bool
//	}{
//		{
//			name: "end of ticker",
//			args: args{
//				keepAlive: 0,
//				tuneMockStream: func(m *mock_proxy.MockCloudProxyAPI_StreamCloudProxyClient) {
//					m.EXPECT().Send(gomock.Any()).Return(nil).AnyTimes()
//					m.EXPECT().Context().Return(context.Background()).AnyTimes()
//				},
//			},
//		},
//		{
//			name: "send returned error, should exit",
//			args: args{
//				tuneMockStream: func(m *mock_proxy.MockCloudProxyAPI_StreamCloudProxyClient) {
//					m.EXPECT().Send(gomock.Any()).Return(fmt.Errorf("error"))
//					m.EXPECT().Context().Return(context.Background()).AnyTimes()
//				},
//				keepAlive:        int64(time.Second),
//				keepAliveTimeout: int64(10 * time.Minute),
//			},
//			isLastSeenZero: true,
//		},
//	}
//	for _, tt := range tests {
//		tt := tt
//		t.Run(tt.name, func(t *testing.T) {
//			t.Parallel()
//			ctrl := gomock.NewController(t)
//			defer ctrl.Finish()
//
//			c := New(nil, nil, logrus.New(), "podName", "clusterID",
//				"version", "apiKey", config.KeepAliveDefault, config.KeepAliveTimeoutDefault)
//			c.keepAlive.Store(tt.args.keepAlive)
//			c.keepAliveTimeout.Store(tt.args.keepAliveTimeout)
//
//			stream := mock_proxy.NewMockCloudProxyAPI_StreamCloudProxyClient(ctrl)
//			if tt.args.tuneMockStream != nil {
//				tt.args.tuneMockStream(stream)
//			}
//			c.lastSeen.Store(time.Now().UnixNano())
//
//			kaCh := make(chan *cloudproxyv1alpha.StreamCloudProxyRequest)
//			go func() {
//				for {
//					<-kaCh
//				}
//			}()
//
//			c.sendKeepAlive(stream, kaCh)
//			require.Equal(t, tt.isLastSeenZero, c.lastSeen.Load() == 0, "lastSeen: %v", c.lastSeen.Load())
//		})
//	}
//}.

func TestClient_run(t *testing.T) {
	t.Parallel()

	type args struct {
		ctx            func() context.Context
		tuneMockStream func(m *mock_proxy.MockCloudProxyAPI_StreamCloudProxyClient)
	}
	tests := []struct {
		name                string
		args                args
		wantErr             bool
		wantLastSeenUpdated bool
	}{
		{
			name: "send initial error",
			args: args{
				tuneMockStream: func(m *mock_proxy.MockCloudProxyAPI_StreamCloudProxyClient) {
					m.EXPECT().Send(gomock.Any()).Return(fmt.Errorf("test error"))
				},
			},
			wantErr: true,
		},
		{
			name: "context done",
			args: args{
				tuneMockStream: func(m *mock_proxy.MockCloudProxyAPI_StreamCloudProxyClient) {
					ctx, cancel := context.WithCancel(context.Background())
					cancel()

					m.EXPECT().Send(gomock.Any()).Return(nil).AnyTimes() // expected 0 or 1 times.
					m.EXPECT().Context().Return(ctx).AnyTimes()          // expected 0 or 1 times.
				},
			},
			wantLastSeenUpdated: true,
			wantErr:             true,
		},
		{
			name: "stream not alive",
			args: args{
				tuneMockStream: func(m *mock_proxy.MockCloudProxyAPI_StreamCloudProxyClient) {
					m.EXPECT().Send(gomock.Any()).Return(nil).AnyTimes()         // expected 0 or 1 times.
					m.EXPECT().Context().Return(context.Background()).AnyTimes() // expected 0 or 1 times.
					m.EXPECT().Recv().Return(nil, fmt.Errorf("test error"))
				},
			},
			wantLastSeenUpdated: false,
			wantErr:             true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			c := New(nil, nil, logrus.New(), "podName", "clusterID", "version", "apiKey", time.Second, time.Second)
			stream := mock_proxy.NewMockCloudProxyAPI_StreamCloudProxyClient(ctrl)
			if tt.args.tuneMockStream != nil {
				tt.args.tuneMockStream(stream)
			}
			if err := c.run(stream, func() {}); (err != nil) != tt.wantErr {
				t.Errorf("run() error = %v, wantErr %v", err, tt.wantErr)
			}
			require.Equal(t, tt.wantLastSeenUpdated, c.lastSeen.Load() > 0, "lastSeen: %v", c.lastSeen.Load())
		})
	}
}
