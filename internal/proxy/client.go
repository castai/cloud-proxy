//go:generate mockgen -package=mock_proxy -source $GOFILE -destination mock/$GOFILE .
//go:generate mockgen -package=mock_proxy -destination mock/stream.go cloud-proxy/proto/gen/proto/v1alpha CloudProxyAPI_StreamCloudProxyClient

package proxy

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"cloud-proxy/internal/config"
	cloudproxyv1alpha "cloud-proxy/proto/gen/proto/v1alpha"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc/backoff"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	KeepAliveMessageID = "keep-alive"
)

type CloudClient interface {
	DoHTTPRequest(request *http.Request) (*http.Response, error)
}

type Client struct {
	cfg *config.Config

	cloudClient CloudClient
	log         *logrus.Logger
	podName     string
	clusterID   string

	errCount       atomic.Int64
	processedCount atomic.Int64

	lastSeenReceive atomic.Int64
	lastSeenSend    atomic.Int64

	keepAlive        atomic.Int64
	keepAliveTimeout atomic.Int64

	version string
}

func New(cloudClient CloudClient, logger *logrus.Logger, version string, cfg *config.Config) *Client {
	c := &Client{
		cfg:         cfg,
		cloudClient: cloudClient,
		log:         logger,
		podName:     cfg.PodMetadata.PodName,
		clusterID:   cfg.ClusterID,
		version:     version,
	}
	c.keepAlive.Store(int64(cfg.KeepAlive))
	c.keepAliveTimeout.Store(int64(cfg.KeepAliveTimeout))

	return c
}

func (c *Client) Run(ctx context.Context) error {
	authCtx := metadata.NewOutgoingContext(ctx, metadata.Pairs(
		"authorization", fmt.Sprintf("Token %s", c.cfg.CastAI.APIKey),
	))

	t := time.NewTimer(time.Millisecond)

	for {
		select {
		case <-authCtx.Done():
			return authCtx.Err()
		case <-t.C:
			c.log.Info("Starting proxy client")
			err := c.prepareAndRun(authCtx)
			if err != nil {
				c.log.Errorf("Restarting proxy client in %vs: due to error: %v", time.Duration(c.keepAlive.Load()).Seconds(), err)
				t.Reset(time.Duration(c.keepAlive.Load()))
			}
		}
	}
}

func (c *Client) getStream(ctx context.Context) (cloudproxyv1alpha.CloudProxyAPI_StreamCloudProxyClient, func(), error) {
	c.log.Info("Connecting to castai")
	dialOpts := make([]grpc.DialOption, 0)
	if c.cfg.CastAI.DisableGRPCTLS {
		// ONLY For testing purposes.
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	} else {
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(credentials.NewTLS(nil)))
	}

	connectParams := grpc.ConnectParams{
		Backoff: backoff.Config{
			BaseDelay:  2 * time.Second,
			Jitter:     0.1,
			MaxDelay:   5 * time.Second,
			Multiplier: 1.2,
		},
		MinConnectTimeout: 2 * time.Minute,
	}
	dialOpts = append(dialOpts, grpc.WithConnectParams(connectParams))

	c.log.Infof(
		"Creating grpc channel against (%s) with connection config (%+v) and TLS enabled=%v",
		c.cfg.CastAI.GrpcURL,
		connectParams,
		!c.cfg.CastAI.DisableGRPCTLS,
	)

	conn, err := grpc.NewClient(c.cfg.CastAI.GrpcURL, dialOpts...)
	if err != nil {
		return nil, nil, fmt.Errorf("grpc.NewClient: %w", err)
	}

	cancelFunc := func() {
		c.log.Info("Closing grpc connection")
		err := conn.Close()
		if err != nil {
			c.log.Errorf("error closing grpc connection %v", err)
		}
	}

	apiClient := cloudproxyv1alpha.NewCloudProxyAPIClient(conn)
	stream, err := apiClient.StreamCloudProxy(ctx)
	if err != nil {
		return nil, cancelFunc, fmt.Errorf("proxyCastAIClient.StreamCloudProxy: %w", err)
	}

	c.log.Info("Connected to castai, sending initial metadata")
	return stream, cancelFunc, nil
}

func (c *Client) sendInitialRequest(stream cloudproxyv1alpha.CloudProxyAPI_StreamCloudProxyClient) error {
	c.log.Info("Sending initial request to castai")

	err := stream.Send(&cloudproxyv1alpha.StreamCloudProxyRequest{
		Request: &cloudproxyv1alpha.StreamCloudProxyRequest_InitialRequest{
			InitialRequest: &cloudproxyv1alpha.InitialCloudProxyRequest{
				ClientMetadata: &cloudproxyv1alpha.ClientMetadata{
					PodName:   c.podName,
					ClusterId: c.clusterID,
				},
				Version: c.version,
			},
		},
	})
	if err != nil {
		return fmt.Errorf("stream.Send: initial request %w", err)
	}
	c.lastSeenReceive.Store(time.Now().UnixNano())

	c.log.Info("Stream to castai started successfully")

	return nil
}

func (c *Client) prepareAndRun(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	stream, closeConnection, err := c.getStream(ctx)
	if err != nil {
		return fmt.Errorf("c.getStream: %w", err)
	}
	defer closeConnection()

	c.lastSeenReceive.Store(time.Now().UnixNano())
	c.lastSeenSend.Store(time.Now().UnixNano())

	return c.sendAndReceive(ctx, stream)
}

func (c *Client) sendAndReceive(ctx context.Context, stream cloudproxyv1alpha.CloudProxyAPI_StreamCloudProxyClient) error {
	err := c.sendInitialRequest(stream)
	if err != nil {
		return fmt.Errorf("c.Connect: %w", err)
	}

	eg, egctx := errgroup.WithContext(ctx)

	sendCh := make(chan *cloudproxyv1alpha.StreamCloudProxyRequest, 10)
	defer close(sendCh)

	eg.Go(func() error {
		err := c.sendKeepAlive(egctx, stream, sendCh)
		if err != nil {
			c.log.Errorf("stopping keep-alive loop: %v", err)
		}
		return err
	})

	eg.Go(func() error {
		err := c.receive(egctx, stream, sendCh)
		if err != nil {
			c.log.Errorf("stopping receive loop: %v", err)
		}
		return err
	})

	eg.Go(func() error {
		err := c.send(egctx, stream, sendCh)
		if err != nil {
			c.log.Errorf("stopping send loop: %v", err)
		}
		return err
	})

	err = eg.Wait()
	if err != nil {
		c.log.Errorf("sendAndReceive: closing with error: %v", err)
	}

	return err
}

func (c *Client) send(ctx context.Context, stream cloudproxyv1alpha.CloudProxyAPI_StreamCloudProxyClient, sendCh chan *cloudproxyv1alpha.StreamCloudProxyRequest) error {
	defer func() {
		c.log.Info("Closing send channel")
		_ = stream.CloseSend()
	}()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-stream.Context().Done():
			return fmt.Errorf("stream closed %w", stream.Context().Err())
		case req := <-sendCh:
			c.log.Printf("Sending message to stream %v", req.GetResponse().GetMessageId())
			if err := stream.Send(req); err != nil {
				c.log.WithError(err).Warn("failed to send message to stream")
				return fmt.Errorf("failed to send message to stream: %w", err)
			}
			c.lastSeenSend.Store(time.Now().UnixNano())

		case <-time.After(time.Duration(c.keepAlive.Load())):
			if err := c.isAlive(); err != nil {
				return err
			}
		}
	}
}

func (c *Client) receive(ctx context.Context, stream cloudproxyv1alpha.CloudProxyAPI_StreamCloudProxyClient, respCh chan<- *cloudproxyv1alpha.StreamCloudProxyRequest) error {
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("context ended with %w", ctx.Err())
		case <-stream.Context().Done():
			return fmt.Errorf("stream ended with %w", stream.Context().Err())
		default:
			if err := c.isAlive(); err != nil {
				return err
			}
		}

		c.log.Debugf("Polling stream for messages")

		in, err := stream.Recv()
		if err != nil {
			c.log.Errorf("stream.Recv: got error: %v", err)
			c.lastSeenReceive.Store(0)
			return fmt.Errorf("stream.Recv: %w", err)
		}

		c.log.Debugf("Handling message from castai")
		go c.handleMessage(stream.Context(), in, respCh)
	}
}

func (c *Client) handleMessage(ctx context.Context, in *cloudproxyv1alpha.StreamCloudProxyResponse, respCh chan<- *cloudproxyv1alpha.StreamCloudProxyRequest) {
	if in == nil {
		c.log.Error("nil message")
		return
	}

	c.lastSeenReceive.Store(time.Now().UnixNano())
	c.processConfigurationRequest(in)

	// skip processing http request if keep alive message.
	if in.GetMessageId() == KeepAliveMessageID {
		c.log.Debugf("Received keep-alive message from castai for %s", in.GetClientMetadata().GetPodName())
		return
	}

	c.log.Debugf("Received request for proxying msg_id=%v path=%v from castai", in.GetMessageId(), in.GetHttpRequest().GetPath())
	resp := c.processHTTPRequest(in.GetHttpRequest())
	if resp.GetError() != "" {
		c.log.Errorf("Failed to proxy request msg_id=%v with %v", in.GetMessageId(), resp.GetError())
	} else {
		c.log.Debugf("Proxied request msg_id=%v, sending response to castai", in.GetMessageId())
	}

	select {
	case <-ctx.Done():
		return

	case respCh <- &cloudproxyv1alpha.StreamCloudProxyRequest{
		Request: &cloudproxyv1alpha.StreamCloudProxyRequest_Response{
			Response: &cloudproxyv1alpha.ClusterResponse{
				ClientMetadata: &cloudproxyv1alpha.ClientMetadata{
					PodName:   c.podName,
					ClusterId: c.clusterID,
				},
				MessageId:    in.GetMessageId(),
				HttpResponse: resp,
			},
		},
	}:
		return
	}
}

func (c *Client) processConfigurationRequest(in *cloudproxyv1alpha.StreamCloudProxyResponse) {
	if in.ConfigurationRequest == nil {
		return
	}

	if in.ConfigurationRequest.GetKeepAlive() != 0 {
		c.keepAlive.Store(in.ConfigurationRequest.GetKeepAlive())
	}
	if in.ConfigurationRequest.GetKeepAliveTimeout() != 0 {
		c.keepAliveTimeout.Store(in.ConfigurationRequest.GetKeepAliveTimeout())
	}

	c.log.Debugf("Updated keep-alive configuration to %v and keep-alive timeout to %v", time.Duration(c.keepAlive.Load()).Seconds(), time.Duration(c.keepAliveTimeout.Load()).Seconds())
}

func (c *Client) processHTTPRequest(req *cloudproxyv1alpha.HTTPRequest) *cloudproxyv1alpha.HTTPResponse {
	if req == nil {
		return &cloudproxyv1alpha.HTTPResponse{
			Error: lo.ToPtr("nil http request"),
		}
	}
	httpReq, err := c.toHTTPRequest(req)
	if err != nil {
		return &cloudproxyv1alpha.HTTPResponse{
			Error: lo.ToPtr(fmt.Sprintf("toHTTPRequest: %v", err)),
		}
	}
	resp, err := c.cloudClient.DoHTTPRequest(httpReq)
	if err != nil {
		c.errCount.Add(1)
		return &cloudproxyv1alpha.HTTPResponse{
			Error: lo.ToPtr(fmt.Sprintf("c.cloudClient.DoHTTPRequest: %v", err)),
		}
	}
	c.processedCount.Add(1)

	return c.toResponse(resp)
}

var errAlive = fmt.Errorf("client connection is not alive")

func (c *Client) isAlive() error {
	lastSeenReceive := c.lastSeenReceive.Load()
	lastSeenSend := c.lastSeenSend.Load()
	keepAliveTimeout := c.keepAliveTimeout.Load()
	lastSeenReceiveDiff := time.Now().UnixNano() - lastSeenReceive
	lastSeenSendDiff := time.Now().UnixNano() - lastSeenSend

	if lastSeenReceiveDiff > keepAliveTimeout || lastSeenSendDiff > keepAliveTimeout {
		c.log.Warnf("last seen receive %v, last seen send %v",
			time.Duration(lastSeenReceiveDiff).Seconds(), time.Duration(lastSeenSendDiff).Seconds())
		return errAlive
	}

	return nil
}

func (c *Client) sendKeepAlive(ctx context.Context, stream cloudproxyv1alpha.CloudProxyAPI_StreamCloudProxyClient, sendCh chan<- *cloudproxyv1alpha.StreamCloudProxyRequest) error {
	ticker := time.NewTimer(time.Duration(c.keepAlive.Load()))
	defer ticker.Stop()

	c.log.Info("Starting keep-alive loop")

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("context ended with %w", ctx.Err())
		case <-ticker.C:
			if time.Now().UnixNano()-c.lastSeenSend.Load() <= c.keepAlive.Load()/2 {
				ticker.Reset(time.Duration(c.keepAlive.Load()))
			} else {
				select {
				case sendCh <- &cloudproxyv1alpha.StreamCloudProxyRequest{
					Request: &cloudproxyv1alpha.StreamCloudProxyRequest_ClientStats{
						ClientStats: &cloudproxyv1alpha.ClientStats{
							ClientMetadata: &cloudproxyv1alpha.ClientMetadata{
								PodName:   c.podName,
								ClusterId: c.clusterID,
							},
							Stats: &cloudproxyv1alpha.ClientStats_Stats{
								Status:    cloudproxyv1alpha.ClientStats_Stats_OK,
								Timestamp: time.Now().UnixNano(),
							},
						},
					},
				}:
					ticker.Reset(time.Duration(c.keepAlive.Load()))

				default:
					if stream.Context().Err() != nil {
						return fmt.Errorf("stream ended with %w", stream.Context().Err())
					}
					if err := c.isAlive(); err != nil {
						return err
					}
				}
			}
		}
	}
}

var errBadRequest = fmt.Errorf("bad request")

func (c *Client) toHTTPRequest(req *cloudproxyv1alpha.HTTPRequest) (*http.Request, error) {
	if req == nil {
		return nil, fmt.Errorf("nil http request %w", errBadRequest)
	}

	reqHTTP, err := http.NewRequestWithContext(context.Background(), req.GetMethod(), req.GetPath(), bytes.NewReader(req.GetBody()))
	if err != nil {
		return nil, fmt.Errorf("http.NewRequest: error: %w", err)
	}

	for header, values := range req.GetHeaders() {
		for _, value := range values.Value {
			reqHTTP.Header.Add(header, value)
		}
	}

	return reqHTTP, nil
}

func (c *Client) toResponse(resp *http.Response) *cloudproxyv1alpha.HTTPResponse {
	if resp == nil {
		return &cloudproxyv1alpha.HTTPResponse{
			Error: lo.ToPtr("nil response"),
		}
	}
	var headers map[string]*cloudproxyv1alpha.HeaderValue
	if resp.Header != nil {
		headers = make(map[string]*cloudproxyv1alpha.HeaderValue)
		for h, v := range resp.Header {
			headers[h] = &cloudproxyv1alpha.HeaderValue{Value: v}
		}
	}
	var bodyResp []byte
	var errMessage *string
	if resp.Body != nil {
		var err error
		bodyResp, err = io.ReadAll(resp.Body)
		if err != nil {
			errMessage = lo.ToPtr(fmt.Sprintf("io.ReadAll: body for error: %v", err))
			bodyResp = nil
		}
	}

	return &cloudproxyv1alpha.HTTPResponse{
		Body:    bodyResp,
		Error:   errMessage,
		Status:  int64(resp.StatusCode),
		Headers: headers,
	}
}
