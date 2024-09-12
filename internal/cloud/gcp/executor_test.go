package gcp

import (
	"bytes"
	"cloud-proxy/internal/cloud/gcp/gcpauth"
	proto "cloud-proxy/proto/v1alpha"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"testing"
)

type mockReadCloserErr struct{}

func (m mockReadCloserErr) Read(p []byte) (int, error) {
	return 0, io.ErrUnexpectedEOF
}
func (m mockReadCloserErr) Close() error { return nil }

func TestClient_toResponse(t *testing.T) {
	t.Parallel()
	type fields struct {
		credentialsSrc *gcpauth.CredentialsSource
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
		want   *proto.StreamCloudProxyRequest
	}{
		{
			name: "nil response",
			want: &proto.StreamCloudProxyRequest{
				Request: &proto.StreamCloudProxyRequest_Response{
					Response: &proto.ClusterResponse{
						MessageId: "",
						HttpResponse: &proto.HTTPResponse{
							Error: lo.ToPtr("nil response"),
						},
					},
				},
			},
		},
		{
			name: "error reading response body",
			args: args{
				resp: &http.Response{
					Body: &mockReadCloserErr{},
				},
			},
			want: &proto.StreamCloudProxyRequest{
				Request: &proto.StreamCloudProxyRequest_Response{
					Response: &proto.ClusterResponse{
						MessageId: "",
						HttpResponse: &proto.HTTPResponse{
							Error: lo.ToPtr("io.ReadAll: body for msgID= error: unexpected EOF"),
						},
					},
				},
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
			want: &proto.StreamCloudProxyRequest{
				Request: &proto.StreamCloudProxyRequest_Response{
					Response: &proto.ClusterResponse{
						MessageId: "msgID",
						HttpResponse: &proto.HTTPResponse{
							Body:    []byte("body"),
							Status:  200,
							Headers: map[string]*proto.HeaderValue{"header": {Value: []string{"value"}}},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := &Client{
				credentialsSrc: tt.fields.credentialsSrc,
				//httpClient:     tt.fields.httpClient,
			}
			got := c.toResponse(tt.args.msgID, tt.args.resp)
			//diff := cmp.Diff(got, tt.want, protocmp.Transform())
			//require.Empty(t, diff)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestClient_toGCPRequest(t *testing.T) {
	t.Parallel()
	type fields struct {
		credentialsSrc *gcpauth.CredentialsSource
		//httpClient     *http.Client
	}
	type args struct {
		req *proto.StreamCloudProxyResponse
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *http.Request
		wantErr bool
	}{
		{
			name:    "req is nil",
			wantErr: true,
		},
		{
			name:    "httpRequest is nil",
			args:    args{req: &proto.StreamCloudProxyResponse{}},
			wantErr: true,
		},
		{
			name:    "creds nil",
			args:    args{req: &proto.StreamCloudProxyResponse{HttpRequest: &proto.HTTPRequest{}}},
			wantErr: true,
		},
		{
			name: "error getting creds",
			args: args{req: &proto.StreamCloudProxyResponse{HttpRequest: &proto.HTTPRequest{}}},
			fields: fields{
				credentialsSrc: &gcpauth.CredentialsSource{},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := &Client{
				credentialsSrc: tt.fields.credentialsSrc,
				//httpClient:     tt.fields.httpClient,
			}
			got, err := c.toGCPRequest(tt.args.req)
			require.Equal(t, tt.wantErr, err != nil, err)
			require.Equal(t, tt.want, got)
		})
	}
}
