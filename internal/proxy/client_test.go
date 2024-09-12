package proxy

import (
	"bytes"
	proto "cloud-proxy/proto/v1alpha"
	"context"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/url"
	"testing"
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
