package e2etest

import (
	"log"
	"sync"

	"github.com/castai/cloud-proxy/internal/castai/proto"
)

type Dispatcher struct {
	pendingRequests map[string]chan *proto.StreamCloudProxyRequest
	locker          sync.Mutex

	proxyRequestChan  chan<- *proto.StreamCloudProxyResponse
	proxyResponseChan <-chan *proto.StreamCloudProxyRequest

	logger *log.Logger
}

func NewDispatcher(requestChan chan<- *proto.StreamCloudProxyResponse, responseChan <-chan *proto.StreamCloudProxyRequest, logger *log.Logger) *Dispatcher {
	return &Dispatcher{
		pendingRequests:   make(map[string]chan *proto.StreamCloudProxyRequest),
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
				waiter := d.findWaiterForResponse(resp.MessageId)
				waiter <- resp
				d.logger.Println("Sent a response back to caller")
			}
		}
	}()
}

func (d *Dispatcher) SendRequest(req *proto.StreamCloudProxyResponse) (<-chan *proto.StreamCloudProxyRequest, error) {
	waiter := d.addRequestToWaitingList(req.MessageId)
	d.proxyRequestChan <- req
	return waiter, nil
}

func (d *Dispatcher) addRequestToWaitingList(requestID string) <-chan *proto.StreamCloudProxyRequest {
	waiter := make(chan *proto.StreamCloudProxyRequest, 1)
	d.locker.Lock()
	d.pendingRequests[requestID] = waiter
	d.locker.Unlock()
	return waiter
}

func (d *Dispatcher) findWaiterForResponse(requestID string) chan *proto.StreamCloudProxyRequest {
	d.locker.Lock()
	val, ok := d.pendingRequests[requestID]
	if !ok {
		d.logger.Panicln("Trying to send a response for non-existent request", requestID)
	}
	delete(d.pendingRequests, requestID)
	d.locker.Unlock()

	return val
}
