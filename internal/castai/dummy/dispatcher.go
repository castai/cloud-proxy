package dummy

import (
	"log"
	"sync"

	"github.com/google/uuid"

	"github.com/castai/cloud-proxy/internal/castai/proto"
)

type Dispatcher struct {
	pendingRequests map[string]chan *proto.HTTPResponse
	locker          sync.Mutex

	proxyRequestChan  chan<- *proto.StreamCloudProxyResponse
	proxyResponseChan <-chan *proto.StreamCloudProxyRequest

	logger *log.Logger
}

func NewDispatcher(requestChan chan<- *proto.StreamCloudProxyResponse, responseChan <-chan *proto.StreamCloudProxyRequest, logger *log.Logger) *Dispatcher {
	return &Dispatcher{
		pendingRequests:   make(map[string]chan *proto.HTTPResponse),
		locker:            sync.Mutex{},
		proxyRequestChan:  requestChan,
		proxyResponseChan: responseChan,
		logger:            logger,
	}
}

func (d *Dispatcher) Run() {
	go func() {
		d.logger.Println("starting response returning loop")
		for {
			for resp := range d.proxyResponseChan {
				waiter := d.findWaiterForResponse(resp.GetMessageId())
				waiter <- resp.GetHttpResponse()
				d.logger.Println("Sent a response back to caller")
			}
		}
	}()
}

func (d *Dispatcher) SendRequest(req *proto.HTTPRequest) (<-chan *proto.HTTPResponse, error) {
	requestID := uuid.New().String()

	waiter := d.addRequestToWaitingList(requestID)
	d.proxyRequestChan <- &proto.StreamCloudProxyResponse{
		MessageId:   requestID,
		HttpRequest: req,
	}
	return waiter, nil
}

func (d *Dispatcher) addRequestToWaitingList(requestID string) <-chan *proto.HTTPResponse {
	waiter := make(chan *proto.HTTPResponse, 1)
	d.locker.Lock()
	d.pendingRequests[requestID] = waiter
	d.locker.Unlock()
	return waiter
}

func (d *Dispatcher) findWaiterForResponse(requestID string) chan *proto.HTTPResponse {
	d.locker.Lock()
	val, ok := d.pendingRequests[requestID]
	if !ok {
		d.logger.Panicln("Trying to send a response for non-existent request", requestID)
	}
	delete(d.pendingRequests, requestID)
	d.locker.Unlock()

	return val
}
